package sources

import (
	"context"
	tt "text/template"

	"github.com/csmith/apkutils/v2"
	"github.com/csmith/contempt/pkg/template"
	"github.com/csmith/latest/v3"
)

func AlpinePackagesSource(mirror string) template.FunctionSource {
	return func(writer template.BomWriter) tt.FuncMap {
		return tt.FuncMap{
			"alpine_packages": func(packages ...string) (map[string]string, error) {
				res, err := latestAlpinePackages(mirror, packages...)
				if err != nil {
					return nil, err
				}
				for i := range res {
					writer("apk:"+i, res[i])
				}
				return res, nil
			},
		}
	}
}

// latestAlpinePackages returns a map of packages to their latest version. The result will include all the provided
// package names, plus all of their direct and transitive dependencies.
func latestAlpinePackages(mirror string, names ...string) (map[string]string, error) {
	packages, err := apkPackageInfos(mirror)
	if err != nil {
		return nil, err
	}
	return resolveAlpinePackages(packages, names...)
}

func resolveAlpinePackages(packages map[string]*latest.AlpinePackageInfo, names ...string) (map[string]string, error) {
	flattened, err := apkutils.FlattenDependencies(packages, names...)
	if err != nil {
		return nil, err
	}

	res := make(map[string]string, len(flattened))
	for name, p := range flattened {
		res[name] = p.Version
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
			info, err := latest.AlpinePackages(
				context.Background(),
				&latest.AlpinePackagesOptions{
					Mirror:     mirror,
					Repository: repo,
				},
			)
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
