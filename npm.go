package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"regexp"
)

func isNpm(name string) bool {
	matched, _ := regexp.MatchString("^(@[a-z\\d-~][a-z\\d-._~]*/)?[a-z\\d-~][a-z\\d-._~]*$", name)
	return matched
}

func getNpmRepoFromUrl(url string) string {
	r1 := regexp.MustCompile(`github\.com/([\w-]+/[\w-]+)\b`)
	g1 := r1.FindStringSubmatch(url)
	if len(g1) > 0 {
		return g1[1]
	}
	r2 := regexp.MustCompile(`\b([\w-]+)\.github\.io/([\w-]+)\b`)
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

func getNpmRepo(name string) string {
	out, err := exec.Command("npm", "view", name, "--json").Output()
	if err != nil {
		log.Fatal(err)
	}

	var pkg npmPackage
	err = json.Unmarshal(out, &pkg)
	if err != nil {
		log.Fatal(err)
	}
	repo := getNpmRepoFromUrl(pkg.Repository.Url)
	if repo != "" {
		return repo
	}
	repo = getNpmRepoFromUrl(pkg.Homepage)
	if repo != "" {
		return repo
	}

	panic(fmt.Sprintf("Cannot get repo from npm package, %#v\n", pkg))
}
