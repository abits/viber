package wizard

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdateHandlesFieldWithNilValidate(t *testing.T) {
	m := New(Answers{Name: "ok"})
	for _, f := range fields {
		if f.validate != nil {
			continue
		}
		m.idx = indexOf(t, f.prompt)
		m.primeInput()
		m.input.SetValue("anything")
		if _, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter}); m.err != nil {
			t.Fatalf("Update returned err %v on nil-validate field %q", m.err, f.prompt)
		}
	}
}

func indexOf(t *testing.T, prompt string) int {
	t.Helper()
	for i, f := range fields {
		if f.prompt == prompt {
			return i
		}
	}
	t.Fatalf("no wizard field with prompt %q", prompt)
	return -1
}

func TestProjectNameValidation(t *testing.T) {
	validate := fieldByPrompt(t, "Project name").validate
	cases := []struct {
		in string
		ok bool
	}{
		{"myproj", true},
		{"My.Proj-1_x", true},
		{"a", true},
		{"1proj", false},
		{".proj", false},
		{"-proj", false},
		{"my proj", false},
		{"my/proj", false},
		{"", false},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			if err := validate(tc.in); (err == nil) != tc.ok {
				t.Errorf("validate(%q) err = %v, want ok = %v", tc.in, err, tc.ok)
			}
		})
	}
}

func TestDirectoryDefaultsToDotSlashName(t *testing.T) {
	getDefault := fieldByPrompt(t, "Directory").getDefault
	cases := []struct {
		name string
		in   Answers
		want string
	}{
		{"derives from name", Answers{Name: "myproj"}, "./myproj"},
		{"explicit dir wins", Answers{Name: "myproj", Dir: "/tmp/x"}, "/tmp/x"},
		{"nothing to derive from", Answers{}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := getDefault(tc.in); got != tc.want {
				t.Errorf("getDefault(%+v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// TestOptionalFieldsHaveNoValidator documents that a nil validate means
// "accept anything"; Model.Update must not call it.
func TestOptionalFieldsHaveNoValidator(t *testing.T) {
	for _, f := range fields {
		if !f.required && f.validate != nil {
			t.Errorf("optional field %q carries a validator; expected nil", f.prompt)
		}
	}
}

func TestEverySetterWritesItsField(t *testing.T) {
	for _, f := range fields {
		t.Run(f.prompt, func(t *testing.T) {
			var a Answers
			f.set(&a, "sentinel")
			if a == (Answers{}) {
				t.Errorf("set for %q wrote nothing to Answers", f.prompt)
			}
			if got := f.getDefault(a); got != "sentinel" {
				t.Errorf("getDefault after set = %q, want %q (set/getDefault disagree)", got, "sentinel")
			}
		})
	}
}

func fieldByPrompt(t *testing.T, prompt string) field {
	t.Helper()
	for _, f := range fields {
		if f.prompt == prompt {
			return f
		}
	}
	t.Fatalf("no wizard field with prompt %q", prompt)
	return field{}
}
