// This source file is part of the EdgeDB open source project.
//
// Copyright EdgeDB Inc. and the EdgeDB authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	gel "github.com/geldata/gel-go/internal/client"
	"github.com/geldata/gel-go/internal/descriptor"
	toml "github.com/pelletier/go-toml/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	//go:embed templates/*.template
	templates embed.FS
)

func usage() {
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), ""+
		"Generate go functions from edgeql files.\n"+
		"\n"+
		"USAGE:\n"+
		"  %s [OPTIONS]\n\n"+
		"OPTIONS:\n", os.Args[0])
	flag.PrintDefaults()
}

type cmdConfig struct {
	mixedCaps bool
	pubfuncs  bool
	pubtypes  bool
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("edgeql-go: ")

	flag.Usage = usage
	mixedCaps := flag.Bool("mixedcaps", false,
		"Change snake_case names in shapes "+
			"to MixedCaps names in go structs")
	pubfuncs := flag.Bool("pubfuncs", false,
		"Make generated functions public.")
	pubtypes := flag.Bool("pubtypes", false,
		"Make generated types public.")
	flag.Parse()

	cfg := &cmdConfig{
		mixedCaps: *mixedCaps,
		pubfuncs:  *pubfuncs,
		pubtypes:  *pubtypes,
	}

	timer := time.AfterFunc(200*time.Millisecond, func() {
		log.Println("connecting to Gel")
	})
	defer timer.Stop()

	ctx := context.Background()
	c, err := gel.CreateClient(ctx, gel.Options{})
	if err != nil {
		log.Fatalf("creating client: %s", err) // nolint:gocritic
	}

	fileQueue := queueFilesInBackground()

	t, err := template.ParseFS(templates, "templates/*.template")
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	for queryFile := range fileQueue {
		wg.Add(1)
		go func(queryFile string) {
			defer wg.Done()
			outFile := getOutFile(queryFile)
			q, e := newQuery(ctx, c, queryFile, outFile, cfg)
			if e != nil {
				log.Fatalf("processing %s: %s", queryFile, e)
			}

			e = writeGoFile(t, outFile, []*Query{q})
			if e != nil {
				log.Fatalf("processing %s: %s", queryFile, e)
			}
		}(queryFile)
	}
	wg.Wait()
}

func isEdgeDBTOML(file string) (bool, error) {
	info, err := os.Stat(file)
	if err == nil {
		if info.Mode() == fs.ModeDir {
			return false, fmt.Errorf(
				"expected %q to be a file not a directory",
				file,
			)
		}
		return true, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, err
}

type project struct {
	rootDir       string
	migrationsDir string
}

func getProject() (*project, error) {
	dir, err := filepath.Abs(".")
	if err != nil {
		return nil, err
	}

	for {
		parent := filepath.Dir(dir)
		if dir == parent {
			return nil, fmt.Errorf(
				"could not find gel.toml, " +
					"fix this by initializing a project, run: " +
					" gel project init",
			)
		}

		file := filepath.Join(dir, "gel.toml")
		isTOML, err := isEdgeDBTOML(file)
		if err != nil {
			return nil, err
		}

		if !isTOML {
			file = filepath.Join(dir, "edgedb.toml")
			isTOML, err = isEdgeDBTOML(file)
			if err != nil {
				return nil, err
			}
		}

		if isTOML {
			data, err := os.ReadFile(file)
			if err != nil {
				return nil, err
			}

			var x struct {
				Project struct {
					SchemaDir string `toml:"schema-dir"`
				}
			}
			x.Project.SchemaDir = "dbschema"
			err = toml.Unmarshal(data, &x)
			if err != nil {
				return nil, err
			}

			return &project{
				rootDir: dir,
				migrationsDir: filepath.Join(
					dir, x.Project.SchemaDir, "migrations"),
			}, nil
		}
		dir = parent
	}
}

func queueFilesInBackground() chan string {
	queue := make(chan string)
	p, err := getProject()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		er := filepath.WalkDir(
			p.rootDir,
			func(f string, d fs.DirEntry, e error) error {
				if e != nil {
					return e
				}

				if d.IsDir() &&
					f == p.migrationsDir {
					return fs.SkipDir
				}

				if !d.IsDir() && strings.HasSuffix(f, ".edgeql") {
					queue <- f
				}

				return nil
			},
		)

		if er != nil {
			log.Fatalf("detecting .edgeql files: %s", er)
		}
		close(queue)
	}()

	return queue
}

func writeGoFile(
	t *template.Template,
	outFile string,
	queries []*Query,
) error {
	packageName, err := getPackageName(outFile)
	if err != nil {
		log.Fatal(err)
	}

	var imports []string
	for _, q := range queries {
		imports = append(imports, q.imports...)
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, map[string]any{
		"PackageName":  packageName,
		"ExtraImports": imports,
		"Queries":      queries,
	})
	if err != nil {
		return err
	}

	err = os.WriteFile(outFile, buf.Bytes(), 0644)
	if err != nil {
		return err
	}

	cmd := exec.Command("gofmt", "-s", "-w", outFile)
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("formatting %s: %w", outFile, err)
	}

	return nil
}

// getPackageName looks up the package name from the first adjacent .go file it
// finds. If there are no adjacent .go files it uses the lower case version of
// the directory name as the package name.
func getPackageName(outFile string) (string, error) {
	outFile, err := filepath.Abs(outFile)
	if err != nil {
		return "", err
	}

	dirname := filepath.Dir(outFile)
	entries, err := os.ReadDir(dirname)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.HasSuffix(entry.Name(), ".go") {
			src, err := os.ReadFile(filepath.Join(dirname, entry.Name()))
			if err != nil {
				return "", err
			}

			fset := token.NewFileSet()
			f, err := parser.ParseFile(
				fset,
				entry.Name(),
				src,
				parser.PackageClauseOnly,
			)
			if err != nil {
				return "", err
			}

			return f.Name.Name, nil
		}
	}

	return strings.ReplaceAll(
		strings.ToLower(filepath.Base(dirname)),
		"-",
		"",
	), nil
}

func getOutFile(queryFile string) string {
	base := filepath.Base(queryFile)
	base = strings.TrimSuffix(base, ".edgeql")
	base += "_edgeql.go"
	return filepath.Join(filepath.Dir(queryFile), base)
}

func isNumberedArgs(desc descriptor.Descriptor) bool {
	if len(desc.Fields) == 0 {
		return false
	}

	for i, field := range desc.Fields {
		if field.Name != strconv.Itoa(i) {
			return false
		}
	}

	return true
}

func isNumberedArgsV2(desc *descriptor.V2) bool {
	if len(desc.Fields) == 0 {
		return false
	}

	for i, field := range desc.Fields {
		if field.Name != strconv.Itoa(i) {
			return false
		}
	}

	return true
}

func snakeToUpperMixedCase(s string) string {
	title := cases.Title(language.English)

	parts := strings.Split(s, "_")
	for i := 0; i < len(parts); i++ {
		parts[i] = title.String(parts[i])
	}

	return strings.Join(parts, "")
}

func snakeToLowerMixedCase(s string) string {
	title := cases.Title(language.English)
	lower := cases.Lower(language.English)

	parts := strings.Split(s, "_")
	for i := 0; i < len(parts); i++ {
		if i == 0 {
			parts[i] = lower.String(parts[i])
		} else {
			parts[i] = title.String(parts[i])
		}
	}

	return strings.Join(parts, "")
}
