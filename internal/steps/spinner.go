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

type stepResult struct {
	name string
	err  error
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
		m.results = append(m.results, stepResult{name: msg.Name, err: msg.Err})
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
			m.err = errors.New("cancelled")
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

func (m spinModel) View() string {
	var b strings.Builder
	for _, r := range m.results {
		if r.err != nil {
			fmt.Fprintf(&b, "  %s %s — %v\n", failStyle.Render("✗"), r.name, r.err)
		} else {
			fmt.Fprintf(&b, "  %s %s\n", okStyle.Render("✓"), r.name)
		}
	}
	if m.current != "" {
		fmt.Fprintf(&b, "  %s %s\n", m.spinner.View(), dimStyle.Render(m.current))
	}
	return b.String()
}

func RunSpinner(ctx context.Context, steps []Step) error {
	r := &Runner{Steps: steps}
	ch := r.Start(ctx)
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	prog := tea.NewProgram(spinModel{spinner: sp, ch: ch})
	final, err := prog.Run()
	if err != nil {
		return err
	}
	if fm, ok := final.(spinModel); ok && fm.err != nil {
		return fm.err
	}
	return nil
}

func RunPlain(ctx context.Context, steps []Step, w io.Writer) error {
	r := &Runner{Steps: steps}
	ch := r.Start(ctx)
	n := len(steps)
	i := 0
	for msg := range ch {
		switch m := msg.(type) {
		case StartedMsg:
			i++
			fmt.Fprintf(w, "[%d/%d] %s...", i, n, m.Name)
		case DoneMsg:
			fmt.Fprintln(w, " ok")
		case FailedMsg:
			fmt.Fprintln(w, " FAILED")
			fmt.Fprintf(w, "        %v\n", m.Err)
			return m.Err
		case AllDoneMsg:
			return nil
		}
	}
	return nil
}
