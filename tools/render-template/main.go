// Command render-template renders viber's embedded template set into a
// directory without the side effects of `viber init` (no openspec, no git).
// The Makefile's dogfood targets use it to keep viber's own repository in
// step with the scaffold it generates.
//
// Usage:
//
//	go run ./tools/render-template -name NAME -desc DESC DIR
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/abits/viber/internal/templates"
)

func main() {
	name := flag.String("name", "", "project name")
	desc := flag.String("desc", "", "one-line description")
	flag.Parse()
	if *name == "" || flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: render-template -name NAME [-desc DESC] DIR")
		os.Exit(2)
	}
	dst := flag.Arg(0)

	src, err := templates.Default()
	if err != nil {
		fmt.Fprintf(os.Stderr, "render-template: load embedded templates: %v\n", err)
		os.Exit(1)
	}
	if err := templates.Render(src, dst, templates.Data{Name: *name, Description: *desc}, false); err != nil {
		fmt.Fprintf(os.Stderr, "render-template: render into %s: %v\n", dst, err)
		os.Exit(1)
	}
}
