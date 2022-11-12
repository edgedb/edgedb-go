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
	"syscall"
	"text/template"
	"time"

	edgedb "github.com/edgedb/edgedb-go/internal/client"
	"github.com/edgedb/edgedb-go/internal/descriptor"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	typeGen = Generator{typeNameLookup: make(map[lookupKey]Type)}

	//go:embed templates/*.template
	templates embed.FS
)

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), ""+
		"Generate go functions from edgeql files.\n"+
		"\n"+
		"USAGE:\n"+
		"    %s\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("edgeql-go: ")

	flag.Usage = usage
	flag.Parse()

	timer := time.AfterFunc(200*time.Millisecond, func() {
		log.Println("connecting to EdgeDB")
	})
	defer timer.Stop()

	ctx := context.Background()
	c, err := edgedb.CreateClient(ctx, edgedb.Options{})
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
			q, e := newQuery(ctx, c, queryFile, outFile)
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

func getProjectRoot() (string, error) {
	dir, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}

	for {
		parent := filepath.Dir(dir)
		if dir == parent {
			return "", fmt.Errorf(
				"could not find edgedb.toml, " +
					"fix this by initializing a project, run: " +
					" edgedb project init",
			)
		}

		file := filepath.Join(dir, "edgedb.toml")
		isTOML, err := isEdgeDBTOML(file)
		if err != nil {
			return "", err
		}

		if isTOML {
			return dir, nil
		}
		dir = parent
	}
}

func queueFilesInBackground() chan string {
	queue := make(chan string)
	root, err := getProjectRoot()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		er := filepath.WalkDir(
			root,
			func(f string, d fs.DirEntry, e error) error {
				if e != nil {
					if errors.Is(e, syscall.EACCES) {
						return nil
					}
					return e
				}

				if d.IsDir() &&
					f == filepath.Join(root, "dbschema/migrations") {
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
		imports = append(imports, q.Imports()...)
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

func snakeToLowerCamelCase(s string) string {
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
