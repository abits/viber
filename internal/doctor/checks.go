package doctor

// DefaultChecks returns the ordered list of checks viber doctor runs.
func DefaultChecks() []Check {
	return []Check{
		{
			Name:        "git",
			Args:        []string{"git", "--version"},
			Required:    true,
			Remediation: "install git: https://git-scm.com/downloads",
		},
		{
			Name:        "openspec",
			Args:        []string{"openspec", "--version"},
			Required:    true,
			Remediation: "npm install -g @fission-ai/openspec",
		},
		{
			Name:        "gh",
			Args:        []string{"gh", "--version"},
			Required:    false,
			Remediation: "install gh: https://cli.github.com",
		},
		{
			Name:        "gh auth",
			Args:        []string{"gh", "auth", "status", "--hostname", "github.com"},
			Required:    false,
			DependsOn:   "gh",
			Remediation: "gh auth login",
		},
		{
			Name:        "jq",
			Args:        []string{"jq", "--version"},
			Required:    false,
			Remediation: "install jq: https://jqlang.github.io/jq/download/",
		},
		{
			Name:        "markdownlint-cli2",
			Args:        []string{"markdownlint-cli2", "--version"},
			Required:    false,
			Remediation: "npm install -g markdownlint-cli2",
		},
		{
			Name:        "claude",
			Args:        []string{"claude", "--version"},
			Required:    false,
			Remediation: "install claude: https://claude.ai/code",
		},
	}
}
