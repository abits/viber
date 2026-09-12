package steps

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"strings"
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
	if !slices.Equal(got, want) {
		t.Fatalf("got=%v want=%v", got, want)
	}
}

// TestRunnerStopsOnCancelledContext pins that a cancelled context aborts the
// run even when the individual steps never look at ctx themselves.
func TestRunnerStopsOnCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	r := &Runner{Steps: []Step{fakeStep{name: "a"}, fakeStep{name: "b"}}}
	got := drain(r.Start(ctx))

	assertEq(t, got, []string{"Failed:a"})
}

// TestRunnerCancelsMidRun checks the between-steps cancellation check: the
// first step cancels, so the second must never start.
func TestRunnerCancelsMidRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	r := &Runner{Steps: []Step{
		stepFn{name: "a", run: func(context.Context) error { cancel(); return nil }},
		fakeStep{name: "b"},
	}}
	got := drain(r.Start(ctx))

	assertEq(t, got, []string{"Started:a", "Done:a", "Failed:b"})
}

// TestRunPlainDoesNotPrintErrorText pins that RunPlain reports which step
// failed but leaves the error text to the single printer in package cmd.
func TestRunPlainDoesNotPrintErrorText(t *testing.T) {
	var buf bytes.Buffer
	boom := errors.New("a very distinctive failure")
	err := RunPlain(context.Background(), []Step{fakeStep{name: "a", err: boom}}, &buf)

	if !errors.Is(err, boom) {
		t.Fatalf("RunPlain err = %v, want %v", err, boom)
	}
	if strings.Contains(buf.String(), boom.Error()) {
		t.Errorf("RunPlain printed the error text; cmd.execute would double it:\n%s", &buf)
	}
	if !strings.Contains(buf.String(), "FAILED") {
		t.Errorf("RunPlain did not report the failure at all:\n%s", &buf)
	}
}

func TestRunPlainReportsProgress(t *testing.T) {
	var buf bytes.Buffer
	steps := []Step{fakeStep{name: "first"}, fakeStep{name: "second"}}
	if err := RunPlain(context.Background(), steps, &buf); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"[1/2] first", "[2/2] second", "ok"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("output missing %q:\n%s", want, &buf)
		}
	}
}
