package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

func TestInitHelpMentionsAllFlags(t *testing.T) {
	cmd := newInitCmd()
	var missing []string
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if !strings.Contains(initLong, "--"+f.Name) {
			missing = append(missing, f.Name)
		}
	})
	if len(missing) > 0 {
		t.Fatalf("docs/init.txt missing flags: %v", missing)
	}
}
