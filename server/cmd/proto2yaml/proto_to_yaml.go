// proto_to_yaml.go
// Usage:
//   go run proto_to_yaml.go -input=./protos -out=services.yaml
// or
//   go run proto_to_yaml.go -input=file1.proto,file2.proto -out=services.yaml

// Requires:
//   go get github.com/emicklei/proto
//   go get gopkg.in/yaml.v3

package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/emicklei/proto"
	"gopkg.in/yaml.v3"
)

type HttpRule struct {
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
	Body   string `yaml:"body,omitempty"`
}

type Method struct {
	Name     string    `yaml:"name"`
	Request  string    `yaml:"request"`
	Response string    `yaml:"response"`
	Http     *HttpRule `yaml:"http,omitempty"`
}

type ServiceYAML struct {
	File    string   `yaml:"file,omitempty"`
	Service string   `yaml:"service"`
	Methods []Method `yaml:"methods"`
}

func parseProtoFile(path string) ([]ServiceYAML, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	parser := proto.NewParser(strings.NewReader(string(b)))
	def, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	var out []ServiceYAML

	// Find package name
	var packageName string
	for _, elem := range def.Elements {
		if pkg, ok := elem.(*proto.Package); ok {
			packageName = pkg.Name
			break
		}
	}

	for _, elem := range def.Elements {
		switch e := elem.(type) {
		case *proto.Service:
			serviceName := e.Name
			if packageName != "" {
				serviceName = packageName + "." + serviceName
			}

			sy := ServiceYAML{
				File:    path,
				Service: serviceName,
			}
			for _, se := range e.Elements {
				switch r := se.(type) {
				case *proto.RPC:
					m := Method{
						Name:     r.Name,
						Request:  r.RequestType,
						Response: r.ReturnsType,
					}
					for _, el := range r.Elements {
						if opt, ok := el.(*proto.Option); ok && opt.Name == "(google.api.http)" {
							rule := &HttpRule{}
							for _, pair := range opt.Constant.OrderedMap {
								val := strings.Trim(pair.Source, `"`)
								switch pair.Name {
								case "get", "post", "put", "delete", "patch":
									rule.Method = strings.ToUpper(pair.Name)
									rule.Path = val
								case "body":
									rule.Body = val
								}
							}
							if rule.Method != "" && rule.Path != "" {
								m.Http = rule
							}
						}
					}
					sy.Methods = append(sy.Methods, m)
				}
			}
			out = append(out, sy)
		}
	}

	return out, nil
}

func collectProtoFiles(input string) ([]string, error) {
	// If input contains comma, treat as list of files.
	if strings.Contains(input, ",") {
		parts := strings.Split(input, ",")
		var files []string
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			files = append(files, p)
		}
		return files, nil
	}

	// If input is a directory, walk and collect *.proto
	st, err := os.Stat(input)
	if err == nil && st.IsDir() {
		var files []string
		err := filepath.WalkDir(input, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, ".proto") {
				files = append(files, path)
			}
			return nil
		})
		return files, err
	}

	// Otherwise treat input as a single file path
	return []string{input}, nil
}

func main() {
	in := flag.String("input", "./", "input directory or comma-separated .proto files")
	out := flag.String("out", "services.yaml", "output YAML file path")
	flag.Parse()

	files, err := collectProtoFiles(*in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to collect proto files: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "no .proto files found")
		os.Exit(1)
	}

	var all []ServiceYAML
	for _, f := range files {
		svcs, err := parseProtoFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to parse %s: %v\n", f, err)
			continue
		}
		all = append(all, svcs...)
	}

	// Marshal to YAML
	y, err := yaml.Marshal(map[string][]ServiceYAML{
		"services": all,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal yaml: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*out, y, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("wrote %d services to %s\n", len(all), *out)
}
