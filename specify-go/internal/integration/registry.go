package integration

import "errors"

var ErrUnknownIntegration = errors.New("unknown integration")

type IntegrationConfig struct {
	Key         string
	Name        string
	RequiresCLI bool
	InstallURL  string
}

var INTEGRATION_REGISTRY = map[string]IntegrationConfig{
	"claude":  {Key: "claude", Name: "Claude Code", RequiresCLI: true, InstallURL: "https://docs.anthropic.com/en/docs/claude-code"},
	"copilot": {Key: "copilot", Name: "GitHub Copilot", RequiresCLI: false, InstallURL: ""},
	"codex":   {Key: "codex", Name: "OpenAI Codex", RequiresCLI: true, InstallURL: "https://openai.com/index/openai-codex/"},
	"generic": {Key: "generic", Name: "Generic Agent", RequiresCLI: false, InstallURL: ""},
}

func GetIntegration(key string) (IntegrationConfig, error) {
	cfg, ok := INTEGRATION_REGISTRY[key]
	if !ok {
		return IntegrationConfig{}, ErrUnknownIntegration
	}
	return cfg, nil
}
