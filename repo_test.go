package cheerio

import (
	"testing"
)

func TestFetchSourceRepoURL(t *testing.T) {
	tests := []struct {
		pkg         string
		wantRepoURL string
	}{
		{"flask_cm", "https://github.com/futuregrid/flask_cm"},
		{"zipaccess", "https://github.com/iki/zipaccess"},
	}

	for _, test := range tests {
		repoURL, err := DefaultPyPI.FetchSourceRepoURL(test.pkg)
		if err != nil {
			t.Error("FetchSourceRepoURL error:", err)
			continue
		}
		if test.wantRepoURL != repoURL {
			t.Errorf("%s: want repoURL == %q, got %q", test.pkg, test.wantRepoURL, repoURL)
		}
	}
}
