package steps

import "context"

type Step interface {
	Name() string
	Run(ctx context.Context) error
}

type Msg interface{ isMsg() }

type StartedMsg struct{ Name string }
type DoneMsg struct{ Name string }
type FailedMsg struct {
	Name string
	Err  error
}
type AllDoneMsg struct{}

func (StartedMsg) isMsg() {}
func (DoneMsg) isMsg()    {}
func (FailedMsg) isMsg()  {}
func (AllDoneMsg) isMsg() {}

type Runner struct {
	Steps []Step
}

func (r *Runner) Start(ctx context.Context) <-chan Msg {
	ch := make(chan Msg, 4)
	go func() {
		defer close(ch)
		for _, s := range r.Steps {
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
