package dlsec

import "testing"

func TestValidateURL(t *testing.T) {
	cases := []struct {
		url     string
		wantErr bool
	}{
		{"https://api.github.com/repos/foo/bar/releases/latest", false},
		{"http://localhost:8080/catalog.json", false},
		{"http://127.0.0.1:8080/catalog.json", false},
		{"http://[::1]:8080/catalog.json", false},
		{"http://example.com/foo", true},
		{"https://example.com/foo", false},
		{"ftp://example.com/foo", true},
		{"file:///etc/passwd", true},
		{"", true},
		{"not-a-url", true},
	}
	for _, c := range cases {
		err := ValidateURL(c.url)
		if (err != nil) != c.wantErr {
			t.Errorf("ValidateURL(%q): err = %v, wantErr = %v", c.url, err, c.wantErr)
		}
	}
}
