package executil

import "testing"

func TestIsGitHubCredentialKey(t *testing.T) {
	cases := []struct {
		key  string
		want bool
	}{
		{"GH_TOKEN", true},
		{"GITHUB_TOKEN", true},
		{"GH_OTHER", true},
		{"GITHUB_USER", true},
		{"MY_GITHUB_TOKEN", true},
		{"FOO_TOKEN", false},
		{"PATH", false},
		{"HOME", false},
	}
	for _, c := range cases {
		if got := IsGitHubCredentialKey(c.key); got != c.want {
			t.Errorf("%q: got %v, want %v", c.key, got, c.want)
		}
	}
}

func TestScrubEnv(t *testing.T) {
	in := []string{
		"PATH=/usr/bin",
		"GH_TOKEN=secret",
		"GITHUB_TOKEN=secret",
		"HOME=/home/u",
	}
	got := ScrubEnv(in)
	want := []string{"PATH=/usr/bin", "HOME=/home/u"}
	if len(got) != len(want) {
		t.Fatalf("ScrubEnv length = %d, want %d (%v)", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("ScrubEnv[%d] = %q, want %q", i, got[i], w)
		}
	}
}
