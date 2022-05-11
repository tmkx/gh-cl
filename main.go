package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		panic("Unsupported params")
	}
	pkgName := flag.Args()[0]

	repo := pkgName

	if isNpm(pkgName) {
		fmt.Printf("Querying %v's repo...\n", pkgName)
		repo = getNpmRepo(pkgName)
		fmt.Printf("Got %v\n", repo)
	}

	cmd := exec.Command("gh", "release", "view", "-R", repo)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

// For more examples of using go-gh, see:
// https://github.com/cli/go-gh/blob/trunk/example_gh_test.go

func isNpm(name string) bool {
	matched, _ := regexp.MatchString("^(@[a-z\\d-~][a-z\\d-._~]*/)?[a-z\\d-~][a-z\\d-._~]*$", name)
	return matched
}

func getRepoFromUrl(url string) string {
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

type NpmRepository struct {
	Type string `json:"type"`
	Url  string `json:"url"`
}

type NpmPackage struct {
	Homepage   string        `json:"homepage"`
	Repository NpmRepository `json:"repository"`
}

func getNpmRepo(name string) string {
	out, err := exec.Command("npm", "view", name, "--json").Output()
	if err != nil {
		panic(err)
	}

	var pkg NpmPackage
	err = json.Unmarshal(out, &pkg)
	if err != nil {
		panic(err)
	}
	repo := getRepoFromUrl(pkg.Repository.Url)
	if repo != "" {
		return repo
	}
	repo = getRepoFromUrl(pkg.Homepage)
	if repo != "" {
		return repo
	}

	panic(fmt.Sprintf("Cannot get repo from npm package, %#v\n", pkg))
}
