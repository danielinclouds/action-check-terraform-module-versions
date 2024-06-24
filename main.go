package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

type Config struct {
	Modules []Module `hcl:"module,block"`
}

type Module struct {
	Name    string `hcl:"name,label"`
	Source  string `hcl:"source"`
	Version string `hcl:"version"`
}

func main() {

	workingDirectory := os.Getenv("WORKING_DIRECTORY")
	files := listTFFiles(workingDirectory)
	for _, f := range files {
		config := getConfig(f)

		for _, m := range config.Modules {
			latestVersion := getLatestGCPModuleVersion(m.Source)
			moduleVersion := m.Version

			if moduleVersion != latestVersion {
				fmt.Printf("File: %s Module: %s Source: %s has version %s, latest version is %s\n", strings.TrimPrefix(f, workingDirectory), m.Name, m.Source, moduleVersion, latestVersion)
			}
		}
	}

}

func listTFFiles(dir string) []string {

	// Walk the directory tree
	var tfFiles []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check if the file has a .tf extension
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".tf") {
			tfFiles = append(tfFiles, path)
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	return tfFiles
}

func getConfig(path string) Config {

	hclContent, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var config Config
	hclsimple.Decode("dummy.hcl", hclContent, nil, &config)

	return config
}

func getLatestGCPModuleVersion(source string) string {

	url := fmt.Sprintf("https://registry.terraform.io/v1/modules/%s", source)

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	var moduleVersion struct {
		Version string `json:"version"`
	}

	if resp.StatusCode == http.StatusOK {
		err := json.NewDecoder(resp.Body).Decode(&moduleVersion)
		if err != nil {
			panic(err)
		}
	}

	return moduleVersion.Version
}
