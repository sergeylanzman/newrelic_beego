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
