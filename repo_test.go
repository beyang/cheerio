package cheerio

import (
	"testing"
)

func TestFetchSourceRepoURI(t *testing.T) {
	tests := []struct {
		pkg         string
		wantRepoURI string
	}{
		{"flask_cm", "https://github.com/futuregrid/flask_cm"},
		{"zipaccess", "https://github.com/iki/zipaccess"},
	}

	for _, test := range tests {
		repoURI, err := DefaultPyPI.FetchSourceRepoURI(test.pkg)
		if err != nil {
			t.Error("FetchSourceRepoURI error:", err)
			continue
		}
		if test.wantRepoURI != repoURI {
			t.Errorf("%s: want repoURI == %q, got %q", test.pkg, test.wantRepoURI, repoURI)
		}
	}
}
