package sources

import (
	"context"
	"fmt"
	"github.com/csmith/latest"
	"strings"
	"text/template"
)

func AlpinePackagesFuncs(bom *map[string]string, mirror string) template.FuncMap {
	return template.FuncMap{
		"alpine_packages": func(packages ...string) (map[string]string, error) {
			res, err := latestAlpinePackages(mirror, packages...)
			if err != nil {
				return nil, err
			}
			for i := range res {
				(*bom)[fmt.Sprintf("apk:%s", i)] = res[i]
			}
			return res, nil
		},
	}
}

// latestAlpinePackages returns a map of packages to their latest version. The result will include all the provided
// package names, plus all of their direct and transitive dependencies.
func latestAlpinePackages(mirror string, names ...string) (map[string]string, error) {
	packages, err := apkPackageInfos(mirror)
	if err != nil {
		return nil, err
	}

	res := make(map[string]string)
	queue := append([]string{}, names...)

	for len(queue) > 0 {
		if _, ok := res[queue[0]]; ok {
			// We've already got a resolution for this package, skip it.
			queue = queue[1:]
			continue
		}

		if strings.HasPrefix(queue[0], "!") {
			//Package conflict, skip it
			queue = queue[1:]
			continue
		}

		p, ok := packages[queue[0]]
		if !ok {
			return nil, fmt.Errorf("package required but not found: %s", queue[0])
		}

		queue = append(queue[1:], p.Dependencies...)
		res[p.Name] = p.Version
	}

	return res, nil
}

var apkPackageCache map[string]*latest.AlpinePackageInfo

// apkPackageInfos returns a map of all apk packages and their latest info.
func apkPackageInfos(mirror string) (map[string]*latest.AlpinePackageInfo, error) {
	if apkPackageCache != nil {
		return apkPackageCache, nil
	}

	apkPackageCache = make(map[string]*latest.AlpinePackageInfo)
	for _, repo := range []string{"community", "main"} {
		err := func() error {
			info, err := latest.AlpinePackages(context.Background(), &latest.AlpinePackagesOptions{
				Mirror:     mirror,
				Repository: repo,
			})
			if err != nil {
				return err
			}
			for k := range info {
				apkPackageCache[k] = info[k]
			}
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}

	return apkPackageCache, nil
}
