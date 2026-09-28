package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/abits/viber/internal/doctor"
)

// jsonResult is the per-check shape written when --json is requested.
type jsonResult struct {
	Name        string `json:"name"`
	Required    bool   `json:"required"`
	Status      string `json:"status"`
	Version     string `json:"version,omitempty"`
	Remediation string `json:"remediation,omitempty"`
}

func newDoctorCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check that required and optional tools are installed.",
		// text lives in docs/doctor.txt
		Long: doctorLong,
		Args: usageArgs(cobra.NoArgs),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runDoctor(cmd, asJSON)
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print results as a JSON array")
	return cmd
}

// runDoctor probes each tool and renders the results to cmd's output writer.
// It returns doctor.ErrRequiredFailed when at least one required check failed.
func runDoctor(cmd *cobra.Command, asJSON bool) error {
	results := doctor.Run(cmd.Context(), doctor.DefaultChecks())

	if asJSON {
		return renderJSON(cmd.OutOrStdout(), results)
	}
	return renderHuman(cmd.OutOrStdout(), results)
}

// renderHuman writes one line per result plus a summary line.
func renderHuman(w io.Writer, results []doctor.Result) error {
	var errCount, warnCount int
	for _, r := range results {
		switch r.Status {
		case doctor.StatusError:
			errCount++
		case doctor.StatusWarning:
			warnCount++
		}

		// Third column: version when ok, problem otherwise.
		third := r.Version
		if r.Status != doctor.StatusOK {
			third = r.Problem
		}

		fmt.Fprintf(w, "%-7s %-20s %s\n", string(r.Status), r.Check.Name, third)

		// Remediation line for non-ok, non-skipped.
		if r.Status != doctor.StatusOK && r.Status != doctor.StatusSkipped && r.Check.Remediation != "" {
			fmt.Fprintf(w, "  %s\n", r.Check.Remediation)
		}
	}

	errWord := "errors"
	if errCount == 1 {
		errWord = "error"
	}
	warnWord := "warnings"
	if warnCount == 1 {
		warnWord = "warning"
	}
	fmt.Fprintf(w, "%d %s, %d %s\n", errCount, errWord, warnCount, warnWord)

	if errCount > 0 {
		return doctor.ErrRequiredFailed
	}
	return nil
}

// renderJSON marshals results to a JSON array and writes it with a trailing newline.
func renderJSON(w io.Writer, results []doctor.Result) error {
	out := make([]jsonResult, len(results))
	var errCount int
	for i, r := range results {
		jr := jsonResult{
			Name:     r.Check.Name,
			Required: r.Check.Required,
			Status:   string(r.Status),
		}
		if r.Status == doctor.StatusOK {
			jr.Version = r.Version
		}
		if r.Status != doctor.StatusOK && r.Status != doctor.StatusSkipped {
			jr.Remediation = r.Check.Remediation
		}
		if r.Status == doctor.StatusError {
			errCount++
		}
		out[i] = jr
	}

	data, err := json.Marshal(out)
	if err != nil {
		return fmt.Errorf("marshal doctor results: %w", err)
	}
	fmt.Fprintf(w, "%s\n", data)

	if errCount > 0 {
		return doctor.ErrRequiredFailed
	}
	return nil
}
