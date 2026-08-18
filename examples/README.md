# Examples

Ready-to-use [GitHub Actions](https://docs.github.com/en/actions) workflows for
keeping a repository of contempt templates up to date. Copy the relevant files
into your `.github/workflows/` directory and adjust the registry, source link
and secrets to match your setup.

Both examples expect the following secrets to be configured:

| Secret            | Purpose                                                    |
|-------------------|------------------------------------------------------------|
| `REGISTRY_USER`   | Username for the container registry                        |
| `REGISTRY_PASS`   | Password for the container registry                        |
| `GIT_USERNAME`    | Name to use for commits made by the workflow               |
| `GIT_EMAIL`       | Email address to use for commits made by the workflow      |

## [contempt](contempt/) on its own for serial execution

A single workflow with a single job: install contempt and let it process every
project in the repository itself. Contempt processes projects in dependency
order, so images are built in the right order, just one at a time.

This is the easiest setup, and fine for small repositories.

## [orchestrator](orchestrator/) for parallel jobs

A two-part setup that uses the [orchestrator](../README.md#the-orchestrator)
to generate a workflow with a separate job for each project:

* `update.yml.tpl` is a Liquid template describing what a per-project job
  should look like. The orchestrator renders it once per project, filling in
  the job's `needs:` section so that GitHub Actions builds images in parallel
  while still respecting the order they depend on each other in.
* `update-workflow.yml` is a normal workflow that runs the orchestrator
  whenever the repository changes, regenerating `.github/workflows/update.yml`
  (the rendered result) and committing it back. GitHub Actions only picks up
  changes to workflow files from commits, so this step is what makes new or
  removed projects appear in (or disappear from) the build.

Two Liquid details worth knowing when adapting the template:

* GitHub Actions' `${{ }}` expressions clash with Liquid's `{{ }}`, so the bulk
  of the job is wrapped in a `{% raw %}`...`{% endraw %}` block to stop Liquid
  from trying to interpret them.
* The `contempt --project {{ target.name }}` line sits *outside* the raw block,
  so that Liquid *does* interpolate the project name there.

Because the generated jobs run in parallel, several of them may try to `git
push` to the same branch at the same time; the jobs retry with a rebase until
their push goes through.

For a live example of this setup (plus a large collection of templates), see
[csmith/dockerfiles](https://github.com/csmith/dockerfiles).
