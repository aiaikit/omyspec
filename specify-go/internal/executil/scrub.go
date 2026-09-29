package executil

import "strings"

// IsGitHubCredentialKey reports whether an env var key looks like a
// GitHub credential (broad prefix GH_/GITHUB_ or substring _GITHUB_
// followed by credential-shaped suffix).
//
// Mirrors Python `_is_github_credential_env_key` in src/specify_cli/_version.py.
func IsGitHubCredentialKey(key string) bool {
	if strings.HasPrefix(key, "GH_") || strings.HasPrefix(key, "GITHUB_") {
		return true
	}
	// Narrow: any "_GITHUB_" substring + credential suffix
	if strings.Contains(key, "_GITHUB_") {
		for _, suf := range []string{"TOKEN", "SECRET", "KEY", "PASSWORD"} {
			if strings.HasSuffix(key, suf) {
				return true
			}
		}
	}
	return false
}

// ScrubEnv returns a copy of env with GitHub-credential-shaped entries removed.
// Phase 1 does not modify the values of remaining keys.
func ScrubEnv(env []string) []string {
	out := make([]string, 0, len(env))
	for _, kv := range env {
		i := strings.IndexByte(kv, '=')
		if i < 0 {
			out = append(out, kv)
			continue
		}
		if IsGitHubCredentialKey(kv[:i]) {
			continue
		}
		out = append(out, kv)
	}
	return out
}
