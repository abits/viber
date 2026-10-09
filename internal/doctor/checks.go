package doctor

// DefaultChecks returns the ordered list of tool checks viber doctor runs.
// Project checks are added separately when running inside a scaffold (see
// ProjectChecks and DetectProject).
func DefaultChecks() []Check {
	return []Check{
		{
			Name:        "git",
			Group:       GroupTools,
			Args:        []string{"git", "--version"},
			Required:    true,
			Remediation: "install git: https://git-scm.com/downloads",
		},
		{
			Name:        "openspec",
			Group:       GroupTools,
			Args:        []string{"openspec", "--version"},
			Required:    true,
			Remediation: "npm install -g @fission-ai/openspec",
		},
		{
			Name:        "gh",
			Group:       GroupTools,
			Args:        []string{"gh", "--version"},
			Required:    false,
			Remediation: "install gh: https://cli.github.com",
		},
		{
			Name:        "gh auth",
			Group:       GroupTools,
			Args:        []string{"gh", "auth", "status", "--hostname", "github.com"},
			Required:    false,
			DependsOn:   "gh",
			Remediation: "gh auth login",
		},
		{
			Name:        "jq",
			Group:       GroupTools,
			Args:        []string{"jq", "--version"},
			Required:    false,
			Remediation: "install jq: https://jqlang.github.io/jq/download/",
		},
		{
			Name:        "markdownlint-cli2",
			Group:       GroupTools,
			Args:        []string{"markdownlint-cli2", "--version"},
			Required:    false,
			Remediation: "npm install -g markdownlint-cli2",
		},
		{
			Name:        "claude",
			Group:       GroupTools,
			Args:        []string{"claude", "--version"},
			Required:    false,
			Remediation: "install claude: https://claude.ai/code",
		},
	}
}
