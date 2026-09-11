package steps

import (
	"context"
	"errors"
	"testing"
)

type fakeStep struct {
	name string
	err  error
}

func (f fakeStep) Name() string                { return f.name }
func (f fakeStep) Run(_ context.Context) error { return f.err }

func TestRunnerHappyPath(t *testing.T) {
	r := &Runner{Steps: []Step{fakeStep{name: "a"}, fakeStep{name: "b"}}}
	got := drain(r.Start(context.Background()))
	want := []string{"Started:a", "Done:a", "Started:b", "Done:b", "AllDone"}
	assertEq(t, got, want)
}

func TestRunnerShortCircuitsOnFailure(t *testing.T) {
	boom := errors.New("boom")
	r := &Runner{Steps: []Step{
		fakeStep{name: "a"},
		fakeStep{name: "b", err: boom},
		fakeStep{name: "c"},
	}}
	got := drain(r.Start(context.Background()))
	want := []string{"Started:a", "Done:a", "Started:b", "Failed:b"}
	assertEq(t, got, want)
}

func drain(ch <-chan Msg) []string {
	var out []string
	for m := range ch {
		switch m := m.(type) {
		case StartedMsg:
			out = append(out, "Started:"+m.Name)
		case DoneMsg:
			out = append(out, "Done:"+m.Name)
		case FailedMsg:
			out = append(out, "Failed:"+m.Name)
		case AllDoneMsg:
			out = append(out, "AllDone")
		}
	}
	return out
}

func assertEq(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len got=%d want=%d\n  got=%v\n  want=%v", len(got), len(want), got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("msg[%d]=%q want %q", i, got[i], want[i])
		}
	}
}
