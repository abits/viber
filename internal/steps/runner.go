// Package steps turns the side-effecting parts of `viber init` into an ordered
// list that can be driven by either a TUI spinner or plain stdout, without
// either presentation needing to know what a step actually does.
//
// Steps report progress as [Msg] values on a channel rather than printing, so
// that error text is rendered exactly once, by the caller of viber's Execute.
package steps

import "context"

// Step is a single named, side-effecting unit of work in a scaffold run.
type Step interface {
	Name() string
	Run(ctx context.Context) error
}

// Msg is a progress event emitted by a [Runner]. The set of implementations is
// closed; a presentation may switch over them exhaustively.
type Msg interface{ isMsg() }

// StartedMsg is emitted immediately before a step runs.
type StartedMsg struct{ Name string }

// DoneMsg is emitted after a step returns successfully.
type DoneMsg struct{ Name string }

// FailedMsg is emitted when a step fails or the context is cancelled. It is
// always the last message of a run.
type FailedMsg struct {
	Name string
	Err  error
}

// AllDoneMsg is emitted after the final step succeeds.
type AllDoneMsg struct{}

func (StartedMsg) isMsg() {}
func (DoneMsg) isMsg()    {}
func (FailedMsg) isMsg()  {}
func (AllDoneMsg) isMsg() {}

// Runner executes Steps in order, stopping at the first failure.
type Runner struct {
	Steps []Step
}

// Start runs the steps on a background goroutine and streams progress on the
// returned channel, which is closed when the run ends.
//
// Cancellation is checked between steps, so a cancelled context stops the run
// even for steps that do not themselves observe ctx. Callers must drain the
// channel until it closes, or the goroutine blocks: see RunSpinner.
func (r *Runner) Start(ctx context.Context) <-chan Msg {
	ch := make(chan Msg, 4)
	go func() {
		defer close(ch)
		for _, s := range r.Steps {
			if err := ctx.Err(); err != nil {
				ch <- FailedMsg{Name: s.Name(), Err: err}
				return
			}
			ch <- StartedMsg{Name: s.Name()}
			if err := s.Run(ctx); err != nil {
				ch <- FailedMsg{Name: s.Name(), Err: err}
				return
			}
			ch <- DoneMsg{Name: s.Name()}
		}
		ch <- AllDoneMsg{}
	}()
	return ch
}
