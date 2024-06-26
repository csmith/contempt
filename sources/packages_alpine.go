package sources

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var alpineMirror = flag.String("alpine-mirror", "https://dl-cdn.alpinelinux.org/alpine/", "Base URL of the Alpine mirror to use to query version and package info")

const apkIndexPath = "latest-stable/%s/x86_64/APKINDEX.tar.gz"

// LatestAlpinePackages returns a map of packages to their latest version. The result will include all the provided
// package names, plus all of their direct and transitive dependencies.
func LatestAlpinePackages(names ...string) (map[string]string, error) {
	packages, err := apkPackageInfos()
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

var apkPackageCache map[string]*packageInfo

// apkPackageInfos returns a map of all apk packages and their latest info.
func apkPackageInfos() (map[string]*packageInfo, error) {
	if apkPackageCache != nil {
		return apkPackageCache, nil
	}

	apkPackageCache = make(map[string]*packageInfo)
	for _, repo := range []string{"community", "main"} {
		err := func() error {
			u, err := url.JoinPath(*alpineMirror, fmt.Sprintf(apkIndexPath, repo))
			if err != nil {
				return err
			}
			res, err := http.Get(u)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			info, err := readApkIndex(res.Body)
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

// readApkIndex reads a .tar.gz archive containing an APKINDEX file, returning the packages within.
func readApkIndex(reader io.Reader) (map[string]*packageInfo, error) {
	gz, err := gzip.NewReader(reader)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*packageInfo)
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}

		if header.Typeflag == tar.TypeReg && header.Name == "APKINDEX" {
			return readApkIndexContent(tr)
		}
	}
	return res, nil
}

// readApkIndexContent reads an APKINDEX file, parsing out the contained packages.
func readApkIndexContent(reader io.Reader) (map[string]*packageInfo, error) {
	res := make(map[string]*packageInfo)
	scanner := bufio.NewScanner(reader)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	current := &packageInfo{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			res[current.Name] = current

			for i := range current.Provides {
				// Don't overwrite real packages with provides info
				if _, ok := res[current.Provides[i]]; !ok {
					res[current.Provides[i]] = current
				}
			}

			current = &packageInfo{}
		} else if strings.HasPrefix(line, "P:") {
			current.Name = strings.TrimPrefix(line, "P:")
		} else if strings.HasPrefix(line, "D:") {
			d := strings.Fields(strings.TrimPrefix(line, "D:"))
			for i := range d {
				current.Dependencies = append(current.Dependencies, stripVersion(d[i]))
			}
		} else if strings.HasPrefix(line, "p:") {
			p := strings.Fields(strings.TrimPrefix(line, "p:"))
			for i := range p {
				current.Provides = append(current.Provides, stripVersion(p[i]))
			}
		} else if strings.HasPrefix(line, "V:") {
			current.Version = strings.TrimPrefix(line, "V:")
		}
	}

	if scanner.Err() != nil {
		return nil, fmt.Errorf("unable to read index: %v", scanner.Err())
	}

	return res, nil
}

// packageInfo describes a package available in a repository.
type packageInfo struct {
	Name         string
	Version      string
	Dependencies []string
	Provides     []string
}

// stripVersion removes version qualifiers from a package name such as `foo>=1.2`.
func stripVersion(name string) string {
	i := strings.IndexAny(name, ">=<~")
	if i > -1 {
		return name[0:i]
	}
	return name
}
