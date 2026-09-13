package steps

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ErrCancelled is returned when the user aborts a run with Ctrl-C.
var ErrCancelled = errors.New("cancelled")

type stepResult struct {
	name   string
	failed bool
}

type spinModel struct {
	spinner  spinner.Model
	ch       <-chan Msg
	results  []stepResult
	current  string
	finished bool
	err      error
}

func waitMsg(ch <-chan Msg) tea.Cmd {
	return func() tea.Msg {
		m, ok := <-ch
		if !ok {
			return AllDoneMsg{}
		}
		return m
	}
}

func (m spinModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, waitMsg(m.ch))
}

func (m spinModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case StartedMsg:
		m.current = msg.Name
		return m, waitMsg(m.ch)
	case DoneMsg:
		m.results = append(m.results, stepResult{name: msg.Name})
		m.current = ""
		return m, waitMsg(m.ch)
	case FailedMsg:
		m.results = append(m.results, stepResult{name: msg.Name, failed: true})
		m.current = ""
		m.finished = true
		m.err = msg.Err
		return m, tea.Quit
	case AllDoneMsg:
		m.finished = true
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.finished = true
			m.err = ErrCancelled
			return m, tea.Quit
		}
	}
	return m, nil
}

var (
	okStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	failStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	dimStyle  = lipgloss.NewStyle().Faint(true)
)

// View renders step status only. The error text belongs to the caller of
// Execute, which prints it once; repeating it here would double it up.
func (m spinModel) View() string {
	var b strings.Builder
	for _, r := range m.results {
		mark, style := "✓", okStyle
		if r.failed {
			mark, style = "✗", failStyle
		}
		fmt.Fprintf(&b, "  %s %s\n", style.Render(mark), r.name)
	}
	if m.current != "" {
		fmt.Fprintf(&b, "  %s %s\n", m.spinner.View(), dimStyle.Render(m.current))
	}
	return b.String()
}

// RunSpinner executes steps under a Bubble Tea spinner. It returns the first
// step error, ErrCancelled on Ctrl-C, or ctx.Err() if ctx is cancelled.
func RunSpinner(ctx context.Context, steps []Step) error {
	r := &Runner{Steps: steps}
	ch := r.Start(ctx)
	// Once the TUI exits (Ctrl-C, or a killed program) nothing reads ch, and
	// the runner goroutine would block forever on its next send. Draining in
	// the background lets it observe cancellation and shut down cleanly.
	defer func() {
		go func() {
			for msg := range ch {
				_ = msg
			}
		}()
	}()

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	prog := tea.NewProgram(spinModel{spinner: sp, ch: ch}, tea.WithContext(ctx))
	final, err := prog.Run()
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		return err
	}
	if fm, ok := final.(spinModel); ok && fm.err != nil {
		return fm.err
	}
	return nil
}

// RunPlain executes steps writing one status line per step to w. It reports
// which step failed but not why: the error is returned so that the caller of
// Execute can print it exactly once.
func RunPlain(ctx context.Context, steps []Step, w io.Writer) error {
	r := &Runner{Steps: steps}
	n := len(steps)
	i := 0
	lineOpen := false
	for msg := range r.Start(ctx) {
		switch m := msg.(type) {
		case StartedMsg:
			i++
			fmt.Fprintf(w, "[%d/%d] %s...", i, n, m.Name)
			lineOpen = true
		case DoneMsg:
			fmt.Fprintln(w, " ok")
			lineOpen = false
		case FailedMsg:
			// A cancelled context fails a step that never started, so there
			// may be no open status line to terminate.
			if lineOpen {
				fmt.Fprintln(w, " FAILED")
			} else {
				fmt.Fprintf(w, "%s: FAILED\n", m.Name)
			}
			return m.Err
		case AllDoneMsg:
			return nil
		}
	}
	return nil
}
