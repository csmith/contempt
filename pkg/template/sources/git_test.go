package sources

import (
	"reflect"
	"strings"
	"testing"
	tt "text/template"

	"github.com/csmith/latest/v3"
)

// TestGitTagTemplateOptions checks that options pass through the template
// engine to the option parser. An invalid option is used so that the call
// fails before any network access happens.
func TestGitTagTemplateOptions(t *testing.T) {
	funcs := GitSource()(nil)

	for _, tmpl := range []string{
		`{{git_tag "https://example.com/repo" "nope"}}`,
		`{{github_tag "owner/repo" "nope"}}`,
		`{{unreleased_git_tag "https://example.com/repo" "nope"}}`,
	} {
		parsed, err := tt.New("test").Funcs(funcs).Parse(tmpl)
		if err != nil {
			t.Fatalf("Failed to parse %q: %v", tmpl, err)
		}

		var out strings.Builder
		if err := parsed.Execute(&out, nil); err == nil {
			t.Errorf("Expected %q to fail, but it succeeded", tmpl)
		} else if !strings.Contains(err.Error(), "unknown tag option") {
			t.Errorf("Expected unknown option error from %q, got: %v", tmpl, err)
		}
	}
}

func TestBuildTagOptions(t *testing.T) {
	defaults := latest.TagOptions{
		IgnoreDates:      true,
		IgnoreErrors:     true,
		IgnorePreRelease: true,
	}

	tests := []struct {
		name     string
		opts     []string
		expected latest.TagOptions
		err      string
	}{
		{
			name:     "no options",
			opts:     nil,
			expected: defaults,
		},
		{
			name: "clean",
			opts: []string{"clean"},
			expected: func() latest.TagOptions {
				o := defaults
				o.Transform = cleanTag
				return o
			}(),
		},
		{
			name: "unreleased",
			opts: []string{"unreleased"},
			expected: func() latest.TagOptions {
				o := defaults
				o.IgnorePreRelease = false
				return o
			}(),
		},
		{
			name: "prefix",
			opts: []string{"prefix=release-"},
			expected: func() latest.TagOptions {
				o := defaults
				o.TrimPrefixes = []string{"release-"}
				return o
			}(),
		},
		{
			name: "multiple prefixes in order",
			opts: []string{"prefix=v", "prefix=release-"},
			expected: func() latest.TagOptions {
				o := defaults
				o.TrimPrefixes = []string{"v", "release-"}
				return o
			}(),
		},
		{
			name: "combined options",
			opts: []string{"prefix=curl-", "clean", "unreleased"},
			expected: func() latest.TagOptions {
				o := defaults
				o.TrimPrefixes = []string{"curl-"}
				o.Transform = cleanTag
				o.IgnorePreRelease = false
				return o
			}(),
		},
		{
			name: "empty prefix",
			opts: []string{"prefix="},
			err:  `empty prefix in tag option "prefix="`,
		},
		{
			name: "unknown option",
			opts: []string{"cleanish"},
			err:  `unknown tag option "cleanish" (valid options: clean, unreleased, prefix=<string>)`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := buildTagOptions(test.opts)
			if test.err != "" {
				if err == nil || err.Error() != test.err {
					t.Fatalf("Expected error %q, got: %v", test.err, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			// DeepEqual treats non-nil funcs as unequal, so compare them by
			// behaviour and everything else by value.
			actualTransform, expectedTransform := actual.Transform, test.expected.Transform
			actual.Transform, test.expected.Transform = nil, nil
			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("Expected %+v, got %+v", test.expected, actual)
			}
			if (actualTransform == nil) != (expectedTransform == nil) {
				t.Errorf("Expected transform nil: %v, got nil: %v", expectedTransform == nil, actualTransform == nil)
			} else if actualTransform != nil && actualTransform("8_13_0") != expectedTransform("8_13_0") {
				t.Errorf("Expected transform(%q) == %q, got %q", "8_13_0", expectedTransform("8_13_0"), actualTransform("8_13_0"))
			}
		})
	}
}

func TestCleanTag(t *testing.T) {
	tests := []struct {
		tag      string
		expected string
	}{
		{tag: "8_13_0", expected: "8.13.0"},
		{tag: "8_13_0-1", expected: "8.13.0-1"},
		{tag: "8_13_0beta1", expected: "8.13.0.1"}, // letters between digits act as separators
		{tag: "1.2.3", expected: "1.2.3"},
		{tag: "v1.2.3", expected: "1.2.3"},
		{tag: "1.2.3-rc.1", expected: "1.2.3-rc.1"},
		{tag: "curl-8_13_0", expected: ""}, // prefix must be stripped before cleaning
	}

	for _, test := range tests {
		t.Run(test.tag, func(t *testing.T) {
			if actual := cleanTag(test.tag); actual != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, actual)
			}
		})
	}
}
