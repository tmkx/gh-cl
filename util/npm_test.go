package util

import (
	"testing"
)

func TestIsNpm(t *testing.T) {
	if IsNpm("vue") != true {
		t.Error()
	}
	if IsNpm("@vue/cli") != true {
		t.Error()
	}
	if IsNpm("vuejs/core") != false {
		t.Error()
	}
}

// git+https://github.com/vuejs/core.git
func TestGetRepoFromUrl(t *testing.T) {
	if GetNpmRepoFromUrl("git+https://github.com/vuejs/core.git") != "vuejs/core" {
		t.Error()
	}
	if GetNpmRepoFromUrl("git+https://github.com/vercel/next.js.git") != "vercel/next.js" {
		t.Error()
	}
	if GetNpmRepoFromUrl("https://vuejs.github.io/core") != "vuejs/core" {
		t.Error()
	}
}

//func TestGetNpmRepo(t *testing.T) {
//	if GetNpmRepo("vue") != "vuejs/core" {
//		t.Error()
//	}
//}
