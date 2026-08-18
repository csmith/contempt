# Contempt

Contempt is a tool to generate Dockerfiles (or Containerfiles) from templates.

It comes with support for various useful functions for getting the latest versions
of software, expanding out the dependency trees of package managers, and so on.

The most basic invocation of contempt takes a source directory and a destination
directory (these can be the same). Within these directories, it is expected that
each container image is defined in its own subdir, e.g.:

```
input
↳ image1
  ↳ Dockerfile.gotpl
  ↳ something-or-other.patch
↳ image2
  ↳ Dockerfile.gotpl

output
↳ image1
  ↳ Dockerfile
↳ image2
  ↳ Dockerfile
```

Contempt's job is to take those template files and generate the plain version
in the output directory:

```shell
go install github.com/csmith/contempt/cmd/contempt@latest
contempt input_dir output_dir
```

Templates are named `Dockerfile.gotpl` by default; `Containerfile.gotpl` is also
detected automatically (and generates a `Containerfile`). Directories containing
an `IGNORE` file are skipped.

Each generated file starts with a header that records where it was generated
from, along with a "bill of materials" (BOM) listing every upstream version that
was baked into it:

```dockerfile
# Generated from https://github.com/csmith/dockerfiles/blob/master/alpine/Dockerfile.gotpl
# BOM: {"alpine":"3.24.1"}

FROM reg.c5h.io/alpine AS verify
...
```

Contempt compares the old and new BOMs to detect what changed, and uses that as
the commit message when committing:

```shell
contempt -commit . .
git log --oneline
[alpine] alpine 3.24.0->3.24.1
```

Contempt also has options to make a git commit every time an output file changes,
build the corresponding image using buildah, and push it to a registry:

```shell
contempt -commit -build -push . .
```

You can also limit contempt to a single project:

```shell
contempt -project=image1 . .
```

Other miscellaneous options are available:

```
Usage of contempt:
  -alpine-mirror string
    [ALPINE_MIRROR] Base URL of the Alpine mirror to use to query version and package info (default "https://dl-cdn.alpinelinux.org/alpine/")
  -build
    [BUILD] Whether to automatically build on successful commit
  -commit
    [COMMIT] Whether to automatically git commit each changed file
  -force-build
    [FORCE_BUILD] Whether to build projects regardless of changes
  -git-tag-pass string
    [GIT_TAG_PASS] Password to use when querying git tags
  -git-tag-user string
    [GIT_TAG_USER] Username to use when querying git tags
  -includes string
    [INCLUDES] Folder of template files to include (default "_includes")
  -output string
    [OUTPUT] The name of the output files (default "Dockerfile")
  -project string
    [PROJECT] A comma-separated list of projects to generate, instead of all detected ones
  -push
    [PUSH] Whether to automatically push on successful commit
  -push-retries int
    [PUSH_RETRIES] How many times to retry pushing an image if it fails (default 2)
  -registry string
    [REGISTRY] Registry to use for pushes and pulls (default "reg.c5h.io")
  -registry-pass string
    [REGISTRY_PASS] Password to use when querying the container registry
  -registry-user string
    [REGISTRY_USER] Username to use when querying the container registry
  -source-link string
    [SOURCE_LINK] Link to a browsable version of the source repo (default "https://github.com/example/repo/blob/master/")
  -template string
    [TEMPLATE] The name of the template files (default "Dockerfile.gotpl")
  -workflow-commands
    [WORKFLOW_COMMANDS] Whether to output GitHub Actions workflow commands to format logs (default true)
```

In practice, you will probably want to set the `-registry` and `-source-link` parameters to point
at the correct place along with the `commit`/`build`/`push` options as required.

When run without `-project`, projects are processed in dependency order: an image
that builds `FROM` another image in the same repository is always generated (and
therefore built) after its dependencies.

## Template functions

Contempt uses Go's built-in [text/template](https://golang.org/pkg/text/template/) package,
and provides the following functions:

### Image

```gotemplate
{{image "alpine"}}
```

Fetches the latest digest for the given image from the configured registry, and returns the fully-qualified
name with the digest (e.g. `reg.c5h.io/alpine@sha256:abcd...........`).

If the image name includes a registry, then it is used as-is regardless of the `registry` flag value:

```gotemplate
{{image "docker.io/library/hello-world"}}
```

Note: see below for information on passing credentials when using more than one registry.

### Registry

```gotemplate
{{registry}}
```

Returns the registry configured with the `-registry` flag.

### Alpine release

```gotemplate
{{alpine_url}}
{{alpine_checksum}}
```

Returns the URL and checksum for the latest release of Alpine.

### Golang release

```gotemplate
{{golang_url}}
{{golang_checksum}}
```

Returns the URL and checksum for the latest release of Golang.

### Postgres release

```gotemplate
# Deprecated - use {{postgres_url 13}} and {{postgres_checksum 13}}
{{postgres13_url}}
{{postgres13_checksum}}

# Deprecated - use {{postgres_url 14}} and {{postgres_checksum 14}}
{{postgres14_url}}
{{postgres14_checksum}}

# Deprecated - use {{postgres_url 15}} and {{postgres_checksum 15}}
{{postgres15_url}}
{{postgres15_checksum}}

# Deprecated - use {{postgres_url 16}} and {{postgres_checksum 16}}
{{postgres16_url}}
{{postgres16_checksum}}

# Deprecated - use {{postgres_url 17}} and {{postgres_checksum 17}}
{{postgres17_url}}
{{postgres17_checksum}}

{{postgres_url 17}}
{{postgres_checksum 17}}
```

Returns the URL and checksum for the latest release of a specific major version of Postgres.

### Alpine packages

```gotemplate
RUN apk add --no-cache \
        {{range $key, $value := alpine_packages "ca-certificates" "musl" "tzdata" "rsync" -}}
        {{$key}}={{$value}} \
        {{end}};
```

Given one or more Alpine packages, resolves all of their dependencies and returns a flattened
list of all packages pinned to their current versions.

### GitHub tag

```gotemplate
{{github_tag "csmith/contempt"}}
{{prefixed_github_tag "csmith/contempt" "release-"}}
```

Returns the latest semver tag of the given repository. The "prefixed" variant will discard
the given prefix from tag names before comparing them using semver.

Use the `-git-tag-user` and `-git-tag-pass` flags if authentication is required.

### Git tag

```gotemplate
{{git_tag "https://git.sr.ht/~csmith/example"}}
{{prefixed_git_tag "https://git.sr.ht/~csmith/example" "release-"}}
```

Returns the latest semver tag of the given repository. The "prefixed" variant will discard
the given prefix from tag names before comparing them using semver.

The `unreleased_git_tag` and `prefixed_unreleased_git_tag` variants work in the same way,
but will also consider pre-release versions when determining the latest tag.

Use the `-git-tag-user` and `-git-tag-pass` flags if authentication is required.

### Regex URL content

```gotemplate
{{regex_url_content "google_button" "https://www.google.com/" "I'm feeling (L[a-z]+)"}}
```

Requests the given URL over HTTP and attempts to match the regular expression.
Returns the text captured by the first capturing group in the regex.
The first argument is a friendly name used for logging and BOM tracking.

### Increment int

```gotemplate
{{increment_int 3}}
```

Returns the given integer incremented by one.

### Create map

```gotemplate
{{map "key1" 1 "key2" .SomeData}}
```

Creates a map. Argument list must be even, and all keys must be strings.
Useful for passing data to other templates.

### Create array

```gotemplate
{{arr "elem1" "elem2" .SomeData}}
```

Creates an array (slice). Useful for passing data to other templates.

### Included templates

Templates in the includes directory (configured with the `-includes` flag, default
`_includes`) can be invoked from other templates using Go's standard template
inclusion syntax:

```gotemplate
{{template "install-apk.gotpl" (map "Packages" (arr "curl"))}}
```

## The orchestrator

The orchestrator is a companion command that generates configuration files based
on the projects in a repository and the dependencies between them. Its most
common use is generating a CI workflow that contains a separate job for each
project, with the dependencies between them properly expressed — allowing the
images to be built in parallel while still respecting build order.

```shell
go install github.com/csmith/contempt/cmd/orchestrator@latest
orchestrator -template workflow.yml.tpl -output workflow.yml .
```

The orchestrator discovers projects in the same way as contempt (including
honouring `IGNORE` files), and determines their dependencies by dry-running each
template: the template is executed with all its functions replaced by stubs that
record their arguments, so no network access takes place. Any project whose name
is passed to the `{{image}}` function is treated as a dependency.

Dependencies are then rendered into a template of your choice, written using the
[Liquid](https://github.com/osteele/liquid) syntax. The template is given a
single variable, `targets`, which is a list of all projects in dependency order.
Each target has:

* `name` — the name of the project (its directory name)
* `needed` — the names of any other projects this one depends on

For example, given a repository with an `alpine` image and a `postgres` image
built `FROM {{image "alpine"}}`, this template:

```liquid
{% for target in targets %}
  {{ target.name }}:
    needs:
{%- for dep in target.needed %}
      - {{ dep }}
{%- endfor %}
{% endfor %}
```

renders as:

```yaml
  alpine:
    needs:
  postgres:
    needs:
      - alpine
```

A complete, practical template for generating a GitHub Actions workflow — including
the `{% raw %}` blocks needed to stop Liquid from consuming `${{ }}` expressions —
can be found in the [examples](examples/) directory.

The orchestrator has the following options, all of which can also be set as
environment variables:

```
Usage of orchestrator:
  -includes string
    [INCLUDES] Folder of template files to include (default "_includes")
  -output string
    [OUTPUT] Path to output the generated file
  -registry string
    [REGISTRY] The name of the registry that images are pushed to
  -template string
    [TEMPLATE] Path of the template to read
```

The `-registry`, `-template` and `-output` flags are required. Like contempt, the
orchestrator takes a single positional argument: the directory containing the
projects.

## Dealing with registry credentials

There are two cases in which contempt requires credentials: checking the latest digest for an image in a non-public
registry (when the `{{image}}` template function is used), and pushing built images (when the `-push` flag is used).

### Checking digests

You can supply a single set of credentials to use for checking digests using the `-registry-user` and `-registry-pass`
flags (or associated environment variables). If these options aren't passed and the registry is not public, then
credentials will be read from `~/.docker/config.json` if it exists, else `${XDG_RUNTIME_DIR}/containers/auth.json`.

### Pushing

For pushes, contempt expects `buildah` to handle authentication for it. To that end, you will probably want to call
`buildah login` before running contempt. Buildah will also read from `~/.docker/config.json` so a `docker login`
will also suffice.

### GitHub Actions

If you are running contempt using GitHub Actions (or possibly other CI tooling) and need to supply multiple sets
of credentials for the `{{image}}` function, you may encounter a number of inconvenient issues:

- The `XDG_RUNTIME_DIR` env var is not set, and the `/run/user` directory is not writable, meaning `buildah login`
  stores its credentials in `/var/tmp/containers-user-1001/containers/containers/auth.json`. Contempt will not
  read from this location when trying to find credentials for the `{{image}}` function.
- The default actions image comes pre-supplied with a `~/.docker/config.json` with credentials for Docker Hub.
  Because this file exists contempt won't even attempt to read `${XDG_RUNTIME_DIR}/containers/auth.json`, even
  if you've set the environment variable to a sensible value.

The simplest way to deal with this situation is to use `docker login` to write credentials to Docker's config file.

## Examples

The [examples](examples/) directory contains ready-to-use GitHub Actions workflows
for running contempt against a repository of templates: a simple
[single-job workflow](examples/contempt/) that lets contempt process everything
itself, and a [two-part setup](examples/orchestrator/) that uses the orchestrator
to generate a workflow with one job per project.

Check out [csmith/dockerfiles](https://github.com/csmith/dockerfiles) for a live collection of
templates and outputs generated using contempt and the orchestrator.
