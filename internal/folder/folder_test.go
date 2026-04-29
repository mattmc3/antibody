package folder_test

import (
	"testing"

	"github.com/mattmc3/antibody/internal/folder"
)

var data = []struct {
	url, folder string
}{
	{
		"http://google.com",
		"http-COLON--SLASH--SLASH-google.com",
	},
	{
		"git@github.com:mattmc3/antibody.git",
		"git-AT-github.com-COLON-mattmc3-SLASH-antibody.git",
	},
	{
		"https://github.com/mattmc3/folder",
		"https-COLON--SLASH--SLASH-github.com-SLASH-mattmc3-SLASH-folder",
	},
}

func TestFolder(t *testing.T) {
	for _, d := range data {
		if d.folder != folder.FromURL(d.url) {
			t.Error(d.folder, "!=", folder.FromURL(d.url))
		}
		if d.url != folder.ToURL(d.folder) {
			t.Error(d.url, "!=", folder.ToURL(d.folder))
		}
	}
}
