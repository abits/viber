package doctor

import "testing"

func TestDefaultChecks(t *testing.T) {
	checks := DefaultChecks()

	// Assert exactly 7 entries.
	if len(checks) != 7 {
		t.Fatalf("len(DefaultChecks()) = %d, want 7", len(checks))
	}

	// Assert order and names.
	wantNames := []string{"git", "openspec", "gh", "gh auth", "jq", "markdownlint-cli2", "claude"}
	for i, want := range wantNames {
		if checks[i].Name != want {
			t.Errorf("checks[%d].Name = %q, want %q", i, checks[i].Name, want)
		}
	}

	// Assert Required flags: git and openspec are Required=true, rest false.
	for i, c := range checks {
		wantRequired := c.Name == "git" || c.Name == "openspec"
		if c.Required != wantRequired {
			t.Errorf("checks[%d] (%s).Required = %v, want %v", i, c.Name, c.Required, wantRequired)
		}
	}

	// Assert gh auth has DependsOn="gh".
	ghAuthIdx := -1
	for i, c := range checks {
		if c.Name == "gh auth" {
			ghAuthIdx = i
			break
		}
	}
	if ghAuthIdx < 0 {
		t.Fatal("no check named 'gh auth' found")
	}
	if checks[ghAuthIdx].DependsOn != "gh" {
		t.Errorf("'gh auth'.DependsOn = %q, want %q", checks[ghAuthIdx].DependsOn, "gh")
	}

	// Assert every entry has non-empty Remediation.
	for i, c := range checks {
		if c.Remediation == "" {
			t.Errorf("checks[%d] (%s) has empty Remediation", i, c.Name)
		}
	}
}
