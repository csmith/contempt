package sources

import (
	"context"
	"fmt"
	"sync"
	tt "text/template"

	"github.com/csmith/contempt/pkg/template"
	"github.com/csmith/latest/v2"
)

func AlpineReleaseSource(mirror string) template.FunctionSource {
	return func(writer template.BomWriter) tt.FuncMap {
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

		return tt.FuncMap{
			"alpine_url": func() (string, error) {
				if err := check(); err != nil {
					return "", err
				}

				writer.Write("alpine", version)
				return url, nil
			},
			"alpine_checksum": func() (string, error) {
				if err := check(); err != nil {
					return "", err
				}

				writer.Write("alpine", version)
				return checksum, nil
			},
		}
	}
}

func GoReleaseSource() template.FunctionSource {
	return func(writer template.BomWriter) tt.FuncMap {

		var version, url, checksum string
		once := sync.Once{}
		check := func() error {
			var err error
			once.Do(func() {
				version, url, checksum, err = latest.GoRelease(context.Background(), nil)
			})
			return err
		}

		return tt.FuncMap{
			"golang_url": func() (string, error) {
				if err := check(); err != nil {
					return "", err
				}

				writer.Write("golang", version)
				return url, nil
			},
			"golang_checksum": func() (string, error) {
				if err := check(); err != nil {
					return "", err
				}

				writer.Write("golang", version)
				return checksum, nil
			},
		}
	}
}

func PostgresReleaseSource() template.FunctionSource {
	return func(writer template.BomWriter) tt.FuncMap {
		var res = make(tt.FuncMap)

		dynamic := postgresDynamicReleaseFuncs(writer)
		for k := range dynamic {
			res[k] = dynamic[k]
		}

		for i := 13; i <= 17; i++ {
			specific := postgresSpecificReleaseFuncs(writer, i)
			for k := range specific {
				res[k] = specific[k]
			}
		}

		return res
	}
}

func postgresSpecificReleaseFuncs(writer template.BomWriter, majorVersion int) tt.FuncMap {
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

	return tt.FuncMap{
		fmt.Sprintf("postgres%d_url", majorVersion): func() (string, error) {
			if err := check(); err != nil {
				return "", err
			}

			writer.Write(fmt.Sprintf("postgres%d", majorVersion), version)
			return url, nil
		},

		fmt.Sprintf("postgres%d_checksum", majorVersion): func() (string, error) {
			if err := check(); err != nil {
				return "", err
			}

			writer.Write(fmt.Sprintf("postgres%d", majorVersion), version)
			return checksum, nil
		},
	}
}

func postgresDynamicReleaseFuncs(writer template.BomWriter) tt.FuncMap {
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

	return tt.FuncMap{
		"postgres_url": func(version int) (string, error) {
			d, err := query(version)
			if err != nil {
				return "", err
			}

			writer.Write(fmt.Sprintf("postgres%d", version), d.version)
			return d.url, nil
		},

		"postgres_checksum": func(version int) (string, error) {
			d, err := query(version)
			if err != nil {
				return "", err
			}

			writer.Write(fmt.Sprintf("postgres%d", version), d.version)
			return d.checksum, nil
		},
	}
}
