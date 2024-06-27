package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/go-github/v62/github"
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
	personalAccessToken := os.Getenv("PERSONAL_ACCESS_TOKEN")
	files := listTFFiles(workingDirectory)
	for _, f := range files {
		config := getConfig(f)

		for _, m := range config.Modules {
			var latestVersion string
			var moduleVersion string
			if strings.HasPrefix(m.Source, "git::") {
				latestVersion = getLatestGitHubModuleVersion(m.Source, personalAccessToken)
				moduleVersion = getGitModuleVersion(m.Source)
			} else {
				latestVersion = getLatestRegistryModuleVersion(m.Source)
				moduleVersion = m.Version
			}

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

func getLatestRegistryModuleVersion(source string) string {

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

func getLatestGitHubModuleVersion(source, personalAccessToken string) string {

	source = strings.TrimPrefix(source, "git::https://github.com/")
	source = strings.Split(source, ".git")[0]
	owner := strings.Split(source, "/")[0]
	repo := strings.Split(source, "/")[1]

	// Create a GitHub client
	client := github.NewClient(nil).WithAuthToken(personalAccessToken)

	// List tags
	opts := &github.ListOptions{PerPage: 10}
	var allTags []*github.RepositoryTag

	for {
		tags, resp, err := client.Repositories.ListTags(context.Background(), owner, repo, opts)
		if err != nil {
			log.Fatalf("Error fetching tags: %v", err)
		}
		allTags = append(allTags, tags...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	if len(allTags) == 0 {
		log.Println("No tags found.")
		return ""
	}

	// Sort tags by their name (assuming tags are in semantic versioning format)
	sort.Slice(allTags, func(i, j int) bool {
		return allTags[i].GetName() > allTags[j].GetName()
	})

	latestTag := allTags[0]
	return latestTag.GetName()
}

func getGitModuleVersion(source string) string {
	return strings.Split(source, "?ref=")[1]
}
