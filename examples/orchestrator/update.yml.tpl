# Rendered by `contempt orchestrator` into .github/workflows/update.yml — one
# job per project, with `needs:` expressing the dependencies between images.
on:
  workflow_dispatch:
  schedule:
    - cron: '33 3 * * *'
  push:
    paths-ignore:
      - '**/Containerfile'
      - '**/Dockerfile'
      - '.github/**'
name: update
concurrency: update
jobs:
{% for target in targets %}
  {{ target.name }}:
    name: Build {{ target.name }}
    runs-on: ubuntu-latest
{% if target.needed.size > 0 %}
    needs:
{%- for dep in target.needed %}
      - {{ dep }}
{%- endfor -%}
{%- endif -%}
{%- raw %}
    steps:
      - name: Checkout source
        uses: actions/checkout@v4
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: stable
          cache: false
      - name: Update
        env:
          BUILDAH_ISOLATION: chroot
          REGISTRY: reg.example.com
          REGISTRY_USER: ${{ secrets.REGISTRY_USER }}
          REGISTRY_PASS: ${{ secrets.REGISTRY_PASS }}
          SOURCE_LINK: https://github.com/your-name/your-repo-of-templates/blob/master/
        run:  |
          go install github.com/csmith/contempt/cmd/contempt@latest
          git config user.name "${{ secrets.GIT_USERNAME }}"
          git config user.email "${{ secrets.GIT_EMAIL }}"
          buildah login -u "$REGISTRY_USER" -p "$REGISTRY_PASS" "$REGISTRY"
{%- endraw %}
          contempt --commit --build --push --project {{ target.name }} . .
          # The jobs run in parallel, so several of them may try to push to the
          # same branch at once: retry with a rebase until it goes through.
          retries=0
          until git push
          do
            if (( ++retries > 5 )); then
                echo "Failed to push after 5 retries"
                exit 1
            fi
            echo "Git push failed, pulling and retrying"
            git pull --rebase
          done
{% endfor %}
