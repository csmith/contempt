package sources

import (
	"context"
	"fmt"
	"github.com/csmith/latest"
	"sync"
	"text/template"
)

func AlpineReleaseFuncs(bom *map[string]string, mirror string) template.FuncMap {
	var version, url, checksum string
	once := sync.Once{}
	check := func() error {
		var err error
		once.Do(func() {
			version, url, checksum, err = latest.AlpineRelease(context.Background(), &latest.AlpineReleaseOptions{
				Mirror:  mirror,
				Flavour: "minirootfs",
			})
		})
		return err
	}

	return template.FuncMap{
		"alpine_url": func() (string, error) {
			if err := check(); err != nil {
				return "", err
			}

			(*bom)["alpine"] = version
			return url, nil
		},
		"alpine_checksum": func() (string, error) {
			if err := check(); err != nil {
				return "", err
			}

			(*bom)["alpine"] = version
			return checksum, nil
		},
	}
}

func GoReleaseFuncs(bom *map[string]string) template.FuncMap {
	var version, url, checksum string
	once := sync.Once{}
	check := func() error {
		var err error
		once.Do(func() {
			version, url, checksum, err = latest.GoRelease(context.Background(), nil)
		})
		return err
	}

	return template.FuncMap{
		"golang_url": func() (string, error) {
			if err := check(); err != nil {
				return "", err
			}

			(*bom)["golang"] = version
			return url, nil
		},
		"golang_checksum": func() (string, error) {
			if err := check(); err != nil {
				return "", err
			}

			(*bom)["golang"] = version
			return checksum, nil
		},
	}
}

func PostgresReleaseFuncs(bom *map[string]string) template.FuncMap {
	var res = make(template.FuncMap)

	dynamic := postgresDynamicReleaseFuncs(bom)
	for k := range dynamic {
		res[k] = dynamic[k]
	}

	for i := 13; i <= 17; i++ {
		specific := postgresSpecificReleaseFuncs(bom, i)
		for k := range specific {
			res[k] = specific[k]
		}
	}

	return res
}

func postgresSpecificReleaseFuncs(bom *map[string]string, majorVersion int) template.FuncMap {
	var version, url, checksum string
	once := sync.Once{}
	check := func() error {
		var err error
		once.Do(func() {
			version, url, checksum, err = latest.PostgresRelease(context.Background(), &latest.TagOptions{
				MajorVersionMax: majorVersion,
			})
		})
		return err
	}

	return template.FuncMap{
		fmt.Sprintf("postgres%d_url", majorVersion): func() (string, error) {
			if err := check(); err != nil {
				return "", err
			}

			(*bom)[fmt.Sprintf("postgres%d", majorVersion)] = version
			return url, nil
		},
		fmt.Sprintf("postgres%d_checksum", majorVersion): func() (string, error) {
			if err := check(); err != nil {
				return "", err
			}

			(*bom)[fmt.Sprintf("postgres%d", majorVersion)] = version
			return checksum, nil
		},
	}
}

func postgresDynamicReleaseFuncs(bom *map[string]string) template.FuncMap {
	type details struct {
		version  string
		url      string
		checksum string
	}
	var cache = make(map[int]details)
	query := func(majorVersion int) (*details, error) {
		if cache[majorVersion].version == "" {
			version, url, checksum, err := latest.PostgresRelease(context.Background(), &latest.TagOptions{
				MajorVersionMax: majorVersion,
			})
			if err != nil {
				return nil, err
			}
			cache[majorVersion] = details{
				version:  version,
				url:      url,
				checksum: checksum,
			}
		}
		d := cache[majorVersion]
		return &d, nil
	}

	return template.FuncMap{
		"postgres_url": func(version int) (string, error) {
			d, err := query(version)
			if err != nil {
				return "", err
			}

			(*bom)[fmt.Sprintf("postgres%d", version)] = d.version
			return d.url, nil
		},
		"postgres_checksum": func(version int) (string, error) {
			d, err := query(version)
			if err != nil {
				return "", err
			}

			(*bom)[fmt.Sprintf("postgres%d", version)] = d.version
			return d.checksum, nil
		},
	}
}
