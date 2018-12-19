package newrelic_beego

import (
	"testing"
)

var testData = []struct {
	pattern  string
	expected string
}{
	{"/api/admin", "api/admin"},
	{"/api/entity/:id", "api/entity/:id"},
	{"/api/entity/?:id", "api/entity/?:id"},
	{"/api/entity/:id:int", "api/entity/:id"},
	{"/api/entity/:id:string/status", "api/entity/:id/status"},
	{"/api/entity/:id([0-9]+)/status", "api/entity/:id/status"},
	{"/api/entity/:id([0-9]+)/status/:statusId(.+)", "api/entity/:id/status/:statusId"},
	{"/api/entity/:id([0-9]+)_:name", "api/entity/:id_:name"},
	{"/cms_:id_:page.html", "cms_:id_:page.html"},
	{"cms_:id(.+)_:page.html", "cms_:id_:page.html"},
}

func TestGeneratePath(t *testing.T) {
	for _, item := range testData {
		actual := generatePath(item.pattern)
		if actual != item.expected {
			t.Errorf("Error: %s != %s", actual, item.expected)
		}
	}
}
func TestParseSkipPathsEmpty(t *testing.T) {
	paths := parseSkipPaths("        ")

	if len(paths) != 0 {
		t.Errorf("Skip paths should be empty")
	}
}

func TestParseSkipPathsSingle(t *testing.T) {
	paths := parseSkipPaths(`,     ,   /handle 
							,,,  ,
							,    ,`)

	if len(paths) != 1 {
		t.Errorf("Skip paths should contain 1 value")
	}

	if paths["/handle"] != true {
		t.Errorf("Skip paths should contain /handle=true")
	}
}

func TestParseSkipPathsMulti(t *testing.T) {
	expected := []string{"/alive", "/handle"}
	paths := parseSkipPaths(`
						,,, ,, 
					/AlIve,
						  /handlE, `)

	if len(paths) != 2 {
		t.Errorf("Skip paths should contain %d values", len(expected))
	}

	for _, path := range expected {
		if paths[path] != true {
			t.Errorf("Skip paths should contain %s", path)
		}
	}
}

func TestShouldSkip_Wildcard(t *testing.T) {
	if !shouldSkip(map[string]bool{"/handle/": true}, "/handle/123") {
		t.Errorf("Wilcard is used, but suffix is not matched")
	}
}

func TestShouldSkip_ExactMatch(t *testing.T) {
	if !shouldSkip(map[string]bool{"/handle/": true}, "/handle/") {
		t.Errorf("Wildcard is not used, but suffixed is not matched")
	}
}

func TestShouldSkip_NoMatch(t *testing.T) {
	if shouldSkip(map[string]bool{"/handle/": true}, "/alive") {
		t.Errorf("Path is not present, but wildcard was matched")
	}
}

func TestShouldSkip_WildcardMultiplePaths(t *testing.T) {
	if !shouldSkip(map[string]bool{"/handle/": true, "/alive": true},
		"/handle/123") {
		t.Errorf("Wildcard is used, but suffixed is not matched")
	}
}
