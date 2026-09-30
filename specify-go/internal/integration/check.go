package integration

import (
	"github.com/aiaikit/speckit/internal/executil"
	"github.com/aiaikit/speckit/internal/ui"
)

// CheckTools checks all installed integrations from integration.json.
// Uses tracker to report each tool's status.
// Returns error only on catastrophic failure (not on missing tools).
func CheckTools(projectRoot string, tracker ui.Tracker) error {
	state, err := ReadState(projectRoot)
	if err != nil {
		return err
	}
	if state == nil {
		// No integration.json — not an error, just nothing to check
		return nil
	}
	for _, key := range state.InstalledIntegrations {
		cfg, err := GetIntegration(key)
		if err != nil {
			tracker.Add(key, key)
			tracker.Mark(key, ui.Error, "unknown integration")
			continue
		}
		tracker.Add(key, cfg.Name)
		if !cfg.RequiresCLI {
			tracker.Mark(key, ui.Skipped, "IDE-based")
			continue
		}
		if _, err := executil.LookPath(key); err == nil {
			tracker.Mark(key, ui.Done, "found")
		} else {
			tracker.Mark(key, ui.Error, "not found")
		}
	}
	return nil
}
