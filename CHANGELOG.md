# Unreleased

# 1.11.2

- Update liquid library to allow better whitespace handling in
  orchestrator

# 1.11.1

- Fix compiler error in orchestrator (thanks @greboid)

# 1.11.0

- Fixed issue resolving dependencies if functions were passed non-string
  arguments, or returned ints.
- Add `{{map}}` and `{{arr}}` utility template functions.
- Fix postgres sources not working.

# 1.10.0

- Refactor templating to use the [latest](https://github.com/csmith/latest) library
  instead of Contempt knowing how to check the latest version of everything
- Added generic `{{postgres_url <version>}}` and `{{postgres_checksum <version>}}`
  functions. The older version specific functions are deprecated, but will
  not be removed in the near future.
- Support for including other templates. The option "includes" specifies a
  directory containing templates (default: `_includes`) which can then be
  called using Go's standard `{{template "name.gotpl"}}` convention.

# 1.9.0

- Add support for retrieving releases of postgres 16/17 (thanks @greboid)

# 1.8.1

- If a scanner error occurs while reading the Alpine APK index, it is now
  propagated up instead of ignored
- Fix `{{alpine_packages}}` failing for Alpine v3.20 due to extremely long lines
  in the package index

# 1.8.0

- Add support for setting the alpine mirror via the `--alpine-mirror` flag,
  and change default to a working mirror.
- Dependency updates

# 1.7.4

- Revert `--identity-label=false` change as it's not supported by the version
  of Buildah currently deployed on GitHub Actions.

# 1.7.3

- Add logging for commands that are executed on the host, and automatically
  query the version of buildah and git when running with `--commit` or `--build`

# 1.7.2

- Further fix for randomly choosing between multiple tags with the same semver.
- Add `increment_int` template function (thanks @Greboid).
- Pass `--identity-label=false` to Buildah to make builds more reproducible (thanks @Greboid).

# 1.7.1

- Update gitrefs dependency, which fixes issue where contempt will randomly
  choose between tags if they're all the same semver (e.g. `v1.2.3` and `1.2.3`)

# v1.7.0

- Add orchestrator binary, for generating configs based on the dependencies between projects.
  This can be used to (for example) generate a GitHub Actions workflow file that contains a
  separate job for each project, with the dependencies properly expressed between them. A
  future version of contempt will better support this single-project-at-a-time usecase.

# v1.6.1

- Ignore `~x.x` version selectors in apk dependencies (thanks @greboid)

# v1.6.0

- Added `{{postgres15_url}}` and `{{postgres15_checksum}}` template functions

# v1.5.1

- Fixed issue where all projects are ignored if the path is given as `.` (like in
  the examples in the README...)

# v1.5.0

- The `--project` flag can now contain multiple projects separated by commas.
- No longer recurses into directories which start with a `.` (e.g. `.git`) when finding projects.
- Now properly reports errors when finding projects, instead of panicing.
- Added `regex_url_content` template function (thanks @Greboid)

# v1.4.1

- Fix dependency resolution when using fully-qualified names in the `image` template function.

# v1.4.0

- When multiple materials change the commit message is now summarised as "N changes",
  the details are spread over multiple lines, and sorted alphabetically.
- The `image` template function now accepts fully-qualified image names, and will not
  pre-pend the default registry.
- Added `git_tag` and `prefixed_git_tag` template functions.
- Fixed `-push-retries` flag including the original push attempt in the count (i.e.,
  a value of `2` would retry once; a value of `0` would fail without trying.)

# v1.3.1

- `prefixed_github_tag` no longer includes the stripped prefix in the bill of materials.

# v1.3.0

- Added `-push-retries` flag to specify how many times a failed push should be retried. Defaults to 2.

# v1.2.0

- Added option (enabled by default) to print workflow commands for GitHub Actions to group log output.
- Skip 'conflicts with' dependencies when resolving Alpine packages (thanks @Greboid)

# v1.1.0

- Make the project build order deterministic.
- LatestGitHubTag now uses [gitrefs](https://github.com/csmith/gitrefs) to get the latest tag instead of the GitHub API.
- Fix "Generated from" header when running with absolute paths

# v1.0.4

- Fix infinite loop if running with paths other than ".", or if projects are nested multiple directories deep.
- Explicitly error if dependencies can't be resolved.
- Improved error message if GitHub tags couldn't be resolved.

# v1.0.3

- Fixed image digests being truncated to "sha256:" and a single character in commit messages
- Increased size of versions shown in commit messages from 8 to 12

# v1.0.2

- Fixed commit messages showing old/new versions the wrong way around

# v1.0.1

- Flags can be specified as env vars

# v1.0.0

- Initial version
