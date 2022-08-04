// This source file is part of the EdgeDB open source project.
//
// Copyright 2020-present EdgeDB Inc. and the EdgeDB authors.
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
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	edgedb "github.com/edgedb/edgedb-go/internal/client"
	"github.com/edgedb/edgedb-go/internal/descriptor"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	//go:embed templates/*.template
	templates embed.FS

	typeGen = Generator{typeNameLookup: make(map[lookupKey][]byte)}

	queryFiles MultiStringFlag
	outFile    = flag.String("out", "",
		"output file name; default srcdir/<file>_edgeql.go")
	isJSON = flag.Bool("json", false,
		"use Query(Single)JSON instead of Query(Single)")
)

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), ""+
		"Generate go functions from edgeql files.\n"+
		"\n"+
		"USAGE:\n"+
		"    %s [OPTIONS]\n"+
		"\n"+
		"OPTIONS:\n", os.Args[0])
	flag.PrintDefaults()
}

// MultiStringFlag is a string flag that can be specified multiple times.
type MultiStringFlag []string

// String returns a semicolon delimited list of strings
func (s *MultiStringFlag) String() string { return strings.Join(*s, ";") }

// Set appends a value
func (s *MultiStringFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("edgeql-go: ")

	flag.Usage = usage
	flag.Var(&queryFiles, "file",
		".edgeql file to parse; may be specified multiple times; required")
	flag.Parse()

	if len(queryFiles) == 0 {
		log.Fatal("-file must be specified at least once")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	c, err := edgedb.CreateClient(ctx, edgedb.Options{})
	if err != nil {
		log.Fatalf("creating client: %s", err) // nolint:gocritic
	}
	client := edgedb.InstrospectionClient{Client: c}

	if *outFile != "" {
		if !strings.HasSuffix(*outFile, ".go") {
			log.Fatal("the filename specified by -out must end with .go")
		}
	} else {
		*outFile = getOutFile(queryFiles[0])
	}
	*outFile, err = filepath.Abs(*outFile)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	queries := make([]*Query, len(queryFiles))
	for i, queryFile := range queryFiles {
		wg.Add(1)
		go func(i int, queryFile string) {
			defer wg.Done()
			q, e := newQuery(ctx, client, queryFile, *outFile)
			if e != nil {
				log.Fatalf("processing %s: %s", *outFile, e)
			}
			queries[i] = q
		}(i, queryFile)
	}

	packageName, err := getPackageName(*outFile)
	if err != nil {
		log.Fatal(err)
	}

	t, err := template.ParseFS(templates, "templates/*.template")
	if err != nil {
		log.Fatal(err)
	}

	wg.Wait()
	var buf bytes.Buffer
	err = t.Execute(&buf, map[string]any{
		"packageName": packageName,
		"queries":     queries,
	})
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(*outFile, buf.Bytes(), 0644)
	if err != nil {
		log.Fatalf("error writing file %q: %s", queryFiles, err)
	}

	cmd := exec.Command("gofmt", "-s", "-w", *outFile)
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalln("formatting code:", err)
	}
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

	return strings.ToLower(filepath.Base(dirname)), nil
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

func snakeToUpperCamelCase(s string) string {
	// todo allow passing spoken language as an option?
	title := cases.Title(language.English)

	parts := strings.Split(s, "_")
	for i := 0; i < len(parts); i++ {
		parts[i] = title.String(parts[i])
	}

	return strings.Join(parts, "")
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
