package contempt

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"text/template"
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
	var res []string
	fakeFunks := template.FuncMap{}
	for i := range templateFuncs {
		for f := range templateFuncs[i] {
			out := reflect.ValueOf(templateFuncs[i][f]).Type().Out(0).Kind()
			if f == "image" {
				fakeFunks[f] = func(dep string) string {
					// Ignore fully-qualified images like "docker.io/library/alpine"
					if index := strings.IndexByte(dep, '.'); index == -1 || index > strings.IndexByte(dep, '/') {
						res = append(res, dep)
					}
					return ""
				}
			} else if out == reflect.Map {
				fakeFunks[f] = func(args ...string) map[string]string {
					return nil
				}
			} else if out == reflect.Slice {
				fakeFunks[f] = func(args ...string) []string {
					return nil
				}
			} else {
				fakeFunks[f] = func(args ...string) string {
					return ""
				}
			}
		}
	}

	templatePath := filepath.Join(dir, templateName)
	tpl := template.New(templatePath)
	tpl.Funcs(fakeFunks)
	_, _ = tpl.ParseFiles(templatePath)
	_ = tpl.ExecuteTemplate(io.Discard, templateName, nil)
	return res
}
