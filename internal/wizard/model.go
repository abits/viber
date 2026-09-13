package wizard

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model is the Bubble Tea model backing the interactive init wizard. It walks
// the prompts in [fields] one at a time, validating each before advancing.
type Model struct {
	fields    []field
	input     textinput.Model
	idx       int
	ans       Answers
	err       error
	cancelled bool
	done      bool
}

// New builds a wizard pre-filled with seed; seeded values become the default
// for their prompt.
func New(seed Answers) Model {
	m := Model{fields: fields, ans: seed}
	m.input = textinput.New()
	m.input.Prompt = "> "
	m.primeInput()
	return m
}

func (m *Model) primeInput() {
	f := m.fields[m.idx]
	def := f.getDefault(m.ans)
	m.input.SetValue(def)
	m.input.CursorEnd()
	m.input.Placeholder = def
	m.input.Focus()
	m.err = nil
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return textinput.Blink }

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			v := strings.TrimSpace(m.input.Value())
			f := m.fields[m.idx]
			if v == "" && f.required {
				m.err = fmt.Errorf("%s is required", f.prompt)
				return m, nil
			}
			if err := f.validate(v); err != nil {
				m.err = err
				return m, nil
			}
			f.set(&m.ans, v)
			m.idx++
			if m.idx >= len(m.fields) {
				m.done = true
				return m, tea.Quit
			}
			m.primeInput()
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

var (
	titleStyle = lipgloss.NewStyle().Bold(true)
	hintStyle  = lipgloss.NewStyle().Faint(true)
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
)

// View implements tea.Model.
func (m Model) View() string {
	if m.done || m.cancelled {
		return ""
	}
	f := m.fields[m.idx]
	var b strings.Builder
	fmt.Fprintf(&b, "%s (%d/%d)\n", titleStyle.Render(f.prompt), m.idx+1, len(m.fields))
	if f.hint != "" {
		fmt.Fprintf(&b, "%s\n", hintStyle.Render(f.hint))
	}
	fmt.Fprintln(&b, m.input.View())
	if m.err != nil {
		fmt.Fprintf(&b, "%s\n", errStyle.Render(m.err.Error()))
	}
	return b.String()
}

// ErrCancelled is returned when the user aborts the wizard with Ctrl-C or Esc.
var ErrCancelled = errors.New("wizard cancelled")

// Run prompts for the answers still missing from seed and returns the completed
// set. It reports ErrCancelled if the user aborts, or ctx.Err() if ctx is
// cancelled.
func Run(ctx context.Context, seed Answers) (Answers, error) {
	prog := tea.NewProgram(New(seed), tea.WithContext(ctx))
	final, err := prog.Run()
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return Answers{}, ctxErr
		}
		return Answers{}, err
	}
	fm, _ := final.(Model)
	if fm.cancelled {
		return Answers{}, ErrCancelled
	}
	fm.ans.Dir = filepath.Clean(fm.ans.Dir)
	return fm.ans, nil
}
