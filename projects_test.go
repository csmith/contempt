package contempt

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindProjects(t *testing.T) {
	dir := t.TempDir()
	includes := filepath.Join(dir, "_includes")
	if err := os.MkdirAll(includes, 0755); err != nil {
		t.Fatalf("Failed to create includes dir: %v", err)
	}

	write := func(project, template, content string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Join(dir, project), 0755); err != nil {
			t.Fatalf("Failed to create project dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, project, template), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write template: %v", err)
		}
	}

	write("base", "Dockerfile.gotpl", "FROM {{image \"docker.io/library/alpine\"}}\n")
	write("middle", "Dockerfile.gotpl", "FROM {{image \"base\"}}\nFROM {{image \"base\"}}\nCOPY --from={{image \"middle\"}} /src /dst\n")
	write("top", "Containerfile.gotpl", "FROM {{image \"middle\"}}\nFROM {{image \"docker.io/library/nginx\"}}\n")
	write("ignored", "Dockerfile.gotpl", "FROM {{image \"base\"}}\n")
	if err := os.WriteFile(filepath.Join(dir, "ignored", "IGNORE"), nil, 0644); err != nil {
		t.Fatalf("Failed to write IGNORE file: %v", err)
	}

	InitTemplates("reg.example.com", "", os.DirFS(includes))

	projects, err := FindProjects(dir, "Dockerfile.gotpl", "Containerfile.gotpl")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(projects) != 3 {
		t.Fatalf("Expected 3 projects, got %d: %#v", len(projects), projects)
	}

	expected := []Project{
		{Name: "base", Template: "Dockerfile.gotpl", Needed: []string{}},
		{Name: "middle", Template: "Dockerfile.gotpl", Needed: []string{"base"}},
		{Name: "top", Template: "Containerfile.gotpl", Needed: []string{"middle"}},
	}
	for i := range expected {
		if projects[i].Name != expected[i].Name {
			t.Errorf("Project %d: expected name %s, got %s", i, expected[i].Name, projects[i].Name)
		}
		if projects[i].Template != expected[i].Template {
			t.Errorf("Project %d: expected template %s, got %s", i, expected[i].Template, projects[i].Template)
		}
		if got, want := len(projects[i].Needed), len(expected[i].Needed); got != want {
			t.Errorf("Project %s: expected %d dependencies, got %d: %#v", projects[i].Name, want, got, projects[i].Needed)
		} else {
			for j := range expected[i].Needed {
				if projects[i].Needed[j] != expected[i].Needed[j] {
					t.Errorf("Project %s: expected dependency %d to be %s, got %s", projects[i].Name, j, expected[i].Needed[j], projects[i].Needed[j])
				}
			}
		}
	}
}
