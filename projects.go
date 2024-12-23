package contempt

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FindProjects returns a slice of all images that can be built from this repo, sorted such that images are positioned
// after all of their dependencies.
func FindProjects(dir, templateName string) ([]string, error) {
	deps := make(map[string][]string)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}

		if d.Name() == templateName {
			project := filepath.Dir(path)
			if _, err := os.Stat(filepath.Join(project, "IGNORE")); errors.Is(err, os.ErrNotExist) {
				deps[filepath.Base(project)] = dependencies(project, templateName)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var res []string
	satisfied := func(reqs []string) bool {
		found := 0
		for i := range reqs {
			for j := range res {
				if res[j] == reqs[i] {
					found++
					break
				}
			}
		}
		return found == len(reqs)
	}

	for len(deps) > 0 {
		var batch []string
		for d := range deps {
			if satisfied(deps[d]) {
				batch = append(batch, d)
				delete(deps, d)
			}
		}
		if len(batch) == 0 {
			return nil, fmt.Errorf("could not fully resolve dependencies: %#v", deps)
		}

		sort.Strings(batch)
		res = append(res, batch...)
	}

	return res, nil
}

func dependencies(dir, templateName string) []string {
	templatePath := filepath.Join(dir, templateName)

	calls, err := engine.DryRun(templatePath)
	if err != nil {
		panic(err)
	}

	var res []string
	for i := range calls["image"] {
		dep := calls["image"][i][0]
		if index := strings.IndexByte(dep, '.'); index == -1 || index > strings.IndexByte(dep, '/') {
			res = append(res, dep)
		}
	}
	
	return res
}
