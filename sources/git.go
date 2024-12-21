package sources

import (
	"context"
	"fmt"
	"github.com/csmith/latest"
	"strings"
	"text/template"
)

func GitFuncs(bom *map[string]string) template.FuncMap {
	return template.FuncMap{
		"git_tag": func(repo string) (string, error) {
			tag, err := latest.GitTag(
				context.Background(),
				repo,
				&latest.TagOptions{
					IgnoreDates:      true,
					IgnoreErrors:     true,
					IgnorePreRelease: true,
				},
			)
			if err != nil {
				return "", err
			}
			(*bom)[fmt.Sprintf("git:%s", repo)] = tag
			return tag, nil
		},

		"prefixed_git_tag": func(repo, prefix string) (string, error) {
			tag, err := latest.GitTag(
				context.Background(),
				repo,
				&latest.TagOptions{
					IgnoreDates:      true,
					IgnoreErrors:     true,
					IgnorePreRelease: true,
					TrimPrefixes:     []string{prefix},
				},
			)
			if err != nil {
				return "", err
			}
			(*bom)[fmt.Sprintf("git:%s", repo)] = strings.TrimPrefix(tag, prefix)
			return tag, nil
		},

		"github_tag": func(repo string) (string, error) {
			tag, err := latest.GitTag(
				context.Background(),
				fmt.Sprintf("https://github.com/%s", repo),
				&latest.TagOptions{
					IgnoreDates:      true,
					IgnoreErrors:     true,
					IgnorePreRelease: true,
				},
			)
			if err != nil {
				return "", err
			}
			(*bom)[fmt.Sprintf("github:%s", repo)] = tag
			return tag, nil
		},

		"prefixed_github_tag": func(repo, prefix string) (string, error) {
			tag, err := latest.GitTag(
				context.Background(),
				fmt.Sprintf("https://github.com/%s", repo),
				&latest.TagOptions{
					IgnoreDates:      true,
					IgnoreErrors:     true,
					IgnorePreRelease: true,
					TrimPrefixes:     []string{prefix},
				},
			)
			if err != nil {
				return "", err
			}
			(*bom)[fmt.Sprintf("github:%s", repo)] = strings.TrimPrefix(tag, prefix)
			return tag, nil
		},
	}
}
