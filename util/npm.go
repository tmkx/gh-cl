package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

func IsNpm(name string) bool {
	matched, _ := regexp.MatchString("^(@[a-z\\d-~][a-z\\d-._~]*/)?[a-z\\d-~][a-z\\d-._~]*$", name)
	return matched
}

func GetNpmRepoFromUrl(url string) string {
	r1 := regexp.MustCompile(`github\.com/([\w-]+/[\w-.]+)\b`)
	g1 := r1.FindStringSubmatch(url)
	if len(g1) > 0 {
		return strings.TrimSuffix(g1[1], ".git")
	}
	r2 := regexp.MustCompile(`\b([\w-]+)\.github\.io/([\w-.]+)\b`)
	g2 := r2.FindStringSubmatch(url)
	if len(g2) > 0 {
		return g2[1] + "/" + g2[2]
	}
	return ""
}

type npmRepository struct {
	Type string `json:"type"`
	Url  string `json:"url"`
}

type npmPackage struct {
	Homepage   string        `json:"homepage"`
	Repository npmRepository `json:"repository"`
}

func GetNpmRepo(name string) (string, error) {
	out, err := exec.Command("npm", "view", name, "--json").Output()
	if err != nil {
		return "", err
	}

	var pkg npmPackage
	err = json.Unmarshal(out, &pkg)
	if err != nil {
		log.Fatal(err)
	}
	repo := GetNpmRepoFromUrl(pkg.Repository.Url)
	if repo != "" {
		return repo, nil
	}
	repo = GetNpmRepoFromUrl(pkg.Homepage)
	if repo != "" {
		return repo, nil
	}

	return "", errors.New(fmt.Sprintf("Cannot get repo from npm package, %#v\n", pkg))
}
