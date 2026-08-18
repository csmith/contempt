package sources

import (
	"context"
	"flag"
	"fmt"
	"regexp"
	"strings"
	tt "text/template"

	"github.com/csmith/contempt/pkg/template"
	"github.com/csmith/latest/v3"
)

var (
	gitTagUser = flag.String("git-tag-user", "", "Username to use when querying git tags")
	gitTagPass = flag.String("git-tag-pass", "", "Password to use when querying git tags")
)

var tagNumbers = regexp.MustCompile(`[0-9]+`)

// cleanTag builds a semver-comparable version from tags that don't use dots
// as separators, e.g. curl's `8_13_0` becomes `8.13.0`. Anything after the
// first hyphen is kept as a pre-release, so `8_13_0-1` becomes `8.13.0-1`
// and `1.2.3-rc.1` is left alone.
func cleanTag(tag string) string {
	first, rest, found := strings.Cut(tag, "-")
	version := strings.Join(tagNumbers.FindAllString(first, -1), ".")
	if version == "" || !found {
		return version
	}
	return version + "-" + rest
}

func buildTagOptions(opts []string) (latest.TagOptions, error) {
	options := latest.TagOptions{
		IgnoreDates:      true,
		IgnoreErrors:     true,
		IgnorePreRelease: true,
	}

	for _, opt := range opts {
		switch {
		case opt == "clean":
			options.Transform = cleanTag
		case opt == "unreleased":
			options.IgnorePreRelease = false
		case strings.HasPrefix(opt, "prefix="):
			if prefix := strings.TrimPrefix(opt, "prefix="); prefix != "" {
				options.TrimPrefixes = append(options.TrimPrefixes, prefix)
			} else {
				return options, fmt.Errorf("empty prefix in tag option %q", opt)
			}
		default:
			return options, fmt.Errorf("unknown tag option %q (valid options: clean, unreleased, prefix=<string>)", opt)
		}
	}

	return options, nil
}

func gitTag(writer template.BomWriter, bomPrefix, repo string, opts ...string) (string, error) {
	options, err := buildTagOptions(opts)
	if err != nil {
		return "", err
	}

	tag, _, err := latest.GitTag(
		context.Background(),
		repo,
		&latest.GitTagOptions{
			Username:   *gitTagUser,
			Password:   *gitTagPass,
			TagOptions: options,
		},
	)
	if err != nil {
		return "", err
	}

	// latest returns the raw tag; derive the canonical version for the BOM
	// by applying the same trimming and transformation it uses to parse tags.
	canonical := tag
	for _, prefix := range options.TrimPrefixes {
		canonical = strings.TrimPrefix(canonical, prefix)
	}
	if options.Transform != nil {
		canonical = options.Transform(canonical)
	}

	writer.Write(fmt.Sprintf("%s:%s", bomPrefix, repo), canonical)
	return tag, nil
}

func GitSource() template.FunctionSource {
	return func(writer template.BomWriter) tt.FuncMap {
		return tt.FuncMap{
			"git_tag": func(repo string, opts ...string) (string, error) {
				return gitTag(writer, "git", repo, opts...)
			},

			"github_tag": func(repo string, opts ...string) (string, error) {
				return gitTag(writer, "github", fmt.Sprintf("https://github.com/%s", repo), opts...)
			},

			// Deprecated: use git_tag with an "unreleased" option instead.
			"unreleased_git_tag": func(repo string, opts ...string) (string, error) {
				return gitTag(writer, "git", repo, append([]string{"unreleased"}, opts...)...)
			},

			// Deprecated: use git_tag with a "prefix=" option instead.
			"prefixed_git_tag": func(repo, prefix string) (string, error) {
				return gitTag(writer, "git", repo, "prefix="+prefix)
			},

			// Deprecated: use github_tag with a "prefix=" option instead.
			"prefixed_github_tag": func(repo, prefix string) (string, error) {
				return gitTag(writer, "github", fmt.Sprintf("https://github.com/%s", repo), "prefix="+prefix)
			},

			// Deprecated: use unreleased_git_tag with a "prefix=" option instead.
			"prefixed_unreleased_git_tag": func(repo, prefix string) (string, error) {
				return gitTag(writer, "git", repo, "prefix="+prefix, "unreleased")
			},
		}
	}
}
