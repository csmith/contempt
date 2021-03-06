package sources

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/go-version"
)

func LatestPostgresRelease(majorVersion string) func() (latest string, url string, checksum string) {
	return func() (latest string, url string, checksum string) {
		const (
			postgresReleaseIndex = "https://ftp.postgresql.org/pub/source/"
			postgresDownloadUrl  = postgresReleaseIndex + "v%[1]s/postgresql-%[1]s.tar.bz2"
			postgresChecksumUrl  = postgresReleaseIndex + "v%[1]s/postgresql-%[1]s.tar.bz2.sha256"
		)

		versions, err := FindInHtml(postgresReleaseIndex, fmt.Sprintf(`a[href*="v%s."]`, majorVersion))
		if err != nil {
			log.Fatalf("Couldn't find releases: %v", err)
		}

		best := version.Must(version.NewVersion("0.0.0"))
		for i := range versions {
			v := version.Must(version.NewVersion(strings.TrimSuffix(versions[i], "/")))
			if (best == nil || v.GreaterThanOrEqual(best)) && v.Prerelease() == "" {
				best = v
				latest = strings.TrimPrefix(strings.TrimSuffix(versions[i], "/"), "v")
			}
		}

		if latest == "" {
			log.Fatalf("Couldn't find candidate version from postgres releases: %v", versions)
		}

		url = fmt.Sprintf(postgresDownloadUrl, latest)
		checksum, err = DownloadHash(fmt.Sprintf(postgresChecksumUrl, latest))
		if err != nil {
			log.Fatalf("Couldn't get checksum for postgres releases: %v", versions)
		}

		return
	}
}
