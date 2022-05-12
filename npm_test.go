package main

import (
	"testing"
)

func TestIsNpm(t *testing.T) {
	if isNpm("vue") != true {
		t.Error()
	}
	if isNpm("@vue/cli") != true {
		t.Error()
	}
	if isNpm("vuejs/core") != false {
		t.Error()
	}
}

// git+https://github.com/vuejs/core.git
func TestGetRepoFromUrl(t *testing.T) {
	if getNpmRepoFromUrl("git+https://github.com/vuejs/core.git") != "vuejs/core" {
		t.Error()
	}
	if getNpmRepoFromUrl("https://vuejs.github.io/core") != "vuejs/core" {
		t.Error()
	}
}

func TestGetNpmRepo(t *testing.T) {
	if getNpmRepo("vue") != "vuejs/core" {
		t.Error()
	}
}
