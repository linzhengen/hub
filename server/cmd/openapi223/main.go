package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"gopkg.in/yaml.v3"
)

func main() {
	var files []string
	root := "./openapi"
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			if strings.HasSuffix(filepath.Base(path), "service.swagger.json") {
				files = append(files, path)
			}
			if strings.HasSuffix(filepath.Base(path), "model.swagger.json") {
				if err := os.Remove(path); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return
	}
	for _, file := range files {
		convertSwaggerToOpenAPI3(file)
	}
}

func convertSwaggerToOpenAPI3(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal("Error closing file:", err)
		}
	}(file)

	var swaggerDoc openapi2.T
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&swaggerDoc); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	oas3, err := openapi2conv.ToV3(&swaggerDoc)
	if err != nil {
		fmt.Println("Error converting to OpenAPI 3:", err)
		return
	}

	outputYAML, err := yaml.Marshal(oas3)
	if err != nil {
		fmt.Println("Error encoding OpenAPI 3 YAML:", err)
		return
	}

	outputYAMLFilename := strings.ReplaceAll(filename[:len(filename)-5], ".swagger", "") + ".yaml"
	if err := os.WriteFile(outputYAMLFilename, outputYAML, 0644); err != nil {
		fmt.Println("Error writing YAML file:", err)
		return
	}
	if err := os.Remove(filename); err != nil {
		fmt.Println("Error removing file:", err)
		return
	}

	fmt.Println("Conversion successful! OpenAPI 3 schema saved to", outputYAMLFilename)
}
