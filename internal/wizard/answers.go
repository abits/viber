// Package wizard collects the values `viber init` needs, either from an
// interactive Bubble Tea prompt or from flags and positional arguments.
package wizard

// Answers holds everything `viber init` needs to scaffold a project. Fields are
// filled from flags and positional arguments first; the wizard then prompts for
// whatever is still missing.
type Answers struct {
	// Name is the project name.
	Name string
	// Dir is the destination directory; defaults to ./<Name>.
	Dir string
	// Description is a one-line summary rendered into the scaffold.
	Description string
	// Remote, if set, is added to the new repository as 'origin'.
	Remote string
	// From selects a remote template set as owner/repo[@ref]; empty means the
	// template set embedded in the binary.
	From string
	// Force allows overwriting files in a non-empty destination.
	Force bool
}
