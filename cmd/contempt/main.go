package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/csmith/contempt"
	"github.com/csmith/contempt/pkg/materials"
	"github.com/csmith/envflag/v2"
)

var (
	templateExplicit bool
	outputExplicit   bool
	templateName     = flag.String("template", "Dockerfile.gotpl", "The name of the template files")
	outputName       = flag.String("output", "Dockerfile", "The name of the output files")
	filter           = flag.String(
		"project",
		"",
		"A comma-separated list of projects to generate, instead of all detected ones",
	)
	sourceLink = flag.String(
		"source-link",
		"https://github.com/example/repo/blob/master/",
		"Link to a browsable version of the source repo",
	)
	commit           = flag.Bool("commit", false, "Whether to automatically git commit each changed file")
	build            = flag.Bool("build", false, "Whether to automatically build on successful commit")
	forceBuild       = flag.Bool("force-build", false, "Whether to build projects regardless of changes")
	push             = flag.Bool("push", false, "Whether to automatically push on successful commit")
	pushRetries      = flag.Int("push-retries", 2, "How many times to retry pushing an image if it fails")
	workflowCommands = flag.Bool(
		"workflow-commands",
		true,
		"Whether to output GitHub Actions workflow commands to format logs",
	)
	registry     = flag.String("registry", "reg.c5h.io", "Registry to use for pushes and pulls")
	alpineMirror = flag.String(
		"alpine-mirror",
		"https://dl-cdn.alpinelinux.org/alpine/",
		"Base URL of the Alpine mirror to use to query version and package info",
	)
	includesDir = flag.String("includes", "_includes", "Folder of template files to include")
)

func main() {
	envflag.Parse()

	flag.Visit(
		func(f *flag.Flag) {
			if f.Name == "template" {
				templateExplicit = true
			}
			if f.Name == "output" {
				outputExplicit = true
			}
		},
	)

	if flag.NArg() != 2 {
		_, _ = fmt.Fprintf(os.Stderr, "Required arguments missing: <input dir> <output dir>\n")
		flag.Usage()
		os.Exit(2)
	}

	contempt.InitTemplates(*registry, *alpineMirror, os.DirFS(*includesDir))

	projectDir, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		log.Fatalf("Failed to resolve project directory: %v", err)
	}

	templateNames := []string{*templateName}
	if !templateExplicit && !outputExplicit {
		templateNames = []string{"Dockerfile.gotpl", "Containerfile.gotpl"}
	}

	projects, err := contempt.FindProjects(projectDir, templateNames...)
	if err != nil {
		log.Fatalf("Failed to find projects: %v", err)
	}

	ctx := context.Background()
	checkExternalDependencies(ctx)

	filtered := strings.Split(*filter, ",")

	for i := range projects {
		if *filter != "" && !slices.Contains(filtered, projects[i].Name) {
			continue
		}

		if *workflowCommands {
			fmt.Printf("::group::%s\n", projects[i].Name)
		}
		processProject(ctx, projects[i])
		if *workflowCommands {
			fmt.Printf("::endgroup::\n")
		}
	}
}

func processProject(ctx context.Context, project contempt.Project) {
	log.Printf("Checking project %s", project.Name)
	outputForProject := *outputName
	if project.Template == "Containerfile.gotpl" {
		outputForProject = "Containerfile"
	}

	changes, err := contempt.Generate(
		*sourceLink,
		flag.Arg(0),
		filepath.Join(project.Name, project.Template),
		filepath.Join(flag.Arg(1), project.Name, outputForProject),
	)
	if err != nil {
		log.Fatalf("Failed to generate project %s: %v", project.Name, err)
	}

	if *commit {
		if err := doCommit(ctx, project.Name, outputForProject, changes); err != nil {
			log.Printf("Failed to commit %s: %v", project.Name, err)
			return
		}
	}

	if (*commit && *build) || *forceBuild {
		buildAndPush(ctx, project.Name)
	}
}

func buildAndPush(ctx context.Context, project string) {
	imageName := fmt.Sprintf("%s/%s", *registry, project)
	if err := runBuildahCommand(
		ctx,
		"bud",
		"--timestamp",
		"0",
		"--layers",
		"--tag",
		imageName,
		filepath.Join(flag.Arg(1), project),
	); err != nil {
		log.Fatalf("Failed to build %s: %v", project, err)
	}

	if !*push {
		return
	}

	for r := 0; r <= *pushRetries; r++ {
		err := runBuildahCommand(ctx, "push", imageName)
		if err == nil {
			return
		}
		log.Printf("Failed to push %s [attempt %d/%d]: %v", project, r+1, *pushRetries+1, err)
	}
	log.Fatalf("Failed to push %s after %d attempts", project, *pushRetries+1)
}

func doCommit(ctx context.Context, project, outputForProject string, changes []materials.Change) error {
	if err := runGitCommand(
		ctx,
		"-C",
		flag.Arg(1),
		"add",
		filepath.Join(project, outputForProject),
	); err != nil {
		return err
	}

	if err := runGitCommand(
		ctx,
		"-C",
		flag.Arg(1),
		"commit",
		"--no-gpg-sign",
		"-m",
		fmt.Sprintf("[%s] %s", project, formatChanges(changes)),
		filepath.Join(project, outputForProject),
	); err != nil {
		return err
	}
	return nil
}

func runGitCommand(ctx context.Context, args ...string) error {
	return runCommand(exec.CommandContext(ctx, "git", args...))
}

func runBuildahCommand(ctx context.Context, args ...string) error {
	return runCommand(exec.CommandContext(ctx, "/usr/bin/buildah", args...))
}

func runCommand(cmd *exec.Cmd) error {
	log.Printf("Running \"%s\"", strings.Join(cmd.Args, "\" \""))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func checkExternalDependencies(ctx context.Context) {
	if *build || *forceBuild {
		if err := runBuildahCommand(ctx, "--version"); err != nil {
			log.Fatalf("Contempt is configured to build, but buildah doesn't seem to be working: %v", err)
		}
	}

	if *commit {
		if err := runGitCommand(ctx, "--version"); err != nil {
			log.Fatalf("Contempt is configured to commit, but git doesn't seem to be working: %v", err)
		}
	}
}

func formatChanges(changes []materials.Change) string {
	if len(changes) == 0 {
		return "no detected changes"
	}

	builder := strings.Builder{}

	if len(changes) > 1 {
		builder.WriteString(fmt.Sprintf("%d changes\n", len(changes)))

		sort.Slice(
			changes,
			func(i, j int) bool {
				return changes[i].Material < changes[j].Material
			},
		)
	}

	for i := range changes {
		oldVersion := changes[i].Old
		newVersion := changes[i].New
		if oldVersion == "" && newVersion == "" {
			builder.WriteString(fmt.Sprintf("\n%s unknown changes", changes[i].Material))
		} else if oldVersion == "" {
			builder.WriteString(fmt.Sprintf("\n%s (unknown)->%.8s", changes[i].Material, newVersion))
		} else if newVersion == "" {
			builder.WriteString(fmt.Sprintf("\n%s %.8s->(unknown)", changes[i].Material, oldVersion))
		} else {
			builder.WriteString(fmt.Sprintf("\n%s %.12s->%.12s", changes[i].Material, oldVersion, newVersion))
		}
	}

	return strings.TrimPrefix(builder.String(), "\n")
}
