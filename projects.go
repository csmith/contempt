package contempt

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// Project describes a single image that can be built from a template.
type Project struct {
	// Name is the name of the directory containing the template.
	Name string
	// Template is the name of the template file the project is generated from.
	Template string
	// Needed is the names of the other projects that this one depends on.
	Needed []string
}

// FindProjects returns a slice of all images that can be built from this repo,
// sorted such that images are positioned after all of their dependencies.
//
// Dependencies are determined by dry-running each template: the template is
// executed with all of its functions replaced by stubs that record their
// arguments, so no network access takes place.
func FindProjects(dir string, templateNames ...string) ([]Project, error) {
	templates := make(map[string]string)
	deps := make(map[string][]string)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}

		for _, tn := range templateNames {
			if d.Name() == tn {
				project := filepath.Dir(path)
				if _, err := os.Stat(filepath.Join(project, "IGNORE")); errors.Is(err, os.ErrNotExist) {
					name := filepath.Base(project)
					needed, err := dependencies(project, tn)
					if err != nil {
						return fmt.Errorf("unable to determine dependencies of %s: %v", path, err)
					}
					templates[name] = tn
					deps[name] = needed
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var res []Project
	satisfied := func(reqs []string) bool {
		found := 0
		for i := range reqs {
			for j := range res {
				if res[j].Name == reqs[i] {
					found++
					break
				}
			}
		}
		return found == len(reqs)
	}

	pending := maps.Clone(deps)
	for len(pending) > 0 {
		var batch []string
		for d := range pending {
			if satisfied(pending[d]) {
				batch = append(batch, d)
				delete(pending, d)
			}
		}
		if len(batch) == 0 {
			return nil, fmt.Errorf("could not fully resolve dependencies: %#v", deps)
		}

		slices.Sort(batch)
		for i := range batch {
			res = append(res, Project{
				Name:     batch[i],
				Template: templates[batch[i]],
				Needed:   deps[batch[i]],
			})
		}
	}

	return res, nil
}

func dependencies(dir, templateName string) ([]string, error) {
	templatePath := filepath.Join(dir, templateName)

	calls, err := engine.DryRun(templatePath)
	if err != nil {
		return nil, err
	}

	// Ignore dependencies on yourself
	ownName := filepath.Base(dir)

	dependencies := make(map[string]bool)
	for i := range calls["image"] {
		dep, ok := calls["image"][i][0].(string)
		if !ok {
			continue
		}
		// Skip images from other registries: they have a hostname before the first slash
		if index := strings.IndexByte(dep, '.'); index != -1 && index < strings.IndexByte(dep, '/') {
			continue
		}
		if dep != ownName {
			dependencies[dep] = true
		}
	}

	res := maps.Keys(dependencies)
	slices.Sort(res)
	return res, nil
}
