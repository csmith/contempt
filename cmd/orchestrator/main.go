package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/csmith/contempt"
	"github.com/csmith/envflag/v2"
	"gopkg.in/osteele/liquid.v1"
)

var (
	flags    = flag.NewFlagSet("orchestrator", flag.ExitOnError)
	registry = flags.String("registry", "", "The name of the registry that images are pushed to")
	template = flags.String("template", "", "Path of the template to read")
	output   = flags.String("output", "", "Path to output the generated file")
	includes = flags.String("includes", "_includes", "Folder of template files to include")
)

func main() {
	envflag.Parse(envflag.WithFlagSet(flags))

	if flags.NArg() != 1 {
		_, _ = fmt.Fprintf(os.Stderr, "Expected one positional argument: <input dir>\n")
		flags.Usage()
		os.Exit(2)
	}

	flags.VisitAll(func(f *flag.Flag) {
		if f.Value.String() == "" {
			_, _ = fmt.Fprintf(os.Stderr, "Missing required flag: %s\n", f.Name)
			flags.Usage()
			os.Exit(2)
		}
	})

	// Templates are only ever dry-run to determine their dependencies, so the
	// template functions that talk to the network are never invoked and the
	// Alpine mirror is irrelevant.
	contempt.InitTemplates(*registry, "", os.DirFS(*includes))

	projects, err := contempt.FindProjects(flags.Arg(0), "Dockerfile.gotpl", "Containerfile.gotpl")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to find projects in '%s': %v\n", flags.Arg(0), err)
		os.Exit(3)
	}

	targets := make([]target, 0, len(projects))
	for i := range projects {
		targets = append(targets, target{projects[i].Name, projects[i].Needed})
	}

	tpl, err := os.ReadFile(*template)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to load template from '%s': %v\n", *template, err)
		os.Exit(6)
	}

	engine := liquid.NewEngine()
	out, err := engine.ParseAndRender(tpl, liquid.Bindings{
		"targets": targets,
	})
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to generate output: %v\n", err)
		os.Exit(7)
	}

	err = os.WriteFile(*output, out, os.FileMode(0644))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to write output to '%s': %v\n", *output, err)
		os.Exit(8)
	}
}

type target struct {
	Name   string   `liquid:"name"`
	Needed []string `liquid:"needed"`
}
