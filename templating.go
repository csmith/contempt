package contempt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/csmith/contempt/pkg/materials"
	"github.com/csmith/contempt/pkg/template"
	"github.com/csmith/contempt/pkg/template/sources"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

var engine *template.Engine

func InitTemplates(imageRegistry, alpineMirror string, includes fs.FS) {
	engine = template.NewEngine(
		slog.New(slog.NewTextHandler(os.Stdout, nil)),
		includes,
	)

	engine.Register(sources.AlpinePackagesSource(alpineMirror))
	engine.Register(sources.ImageSource(imageRegistry))
	engine.Register(sources.GitSource())
	engine.Register(sources.HttpSource())
	engine.Register(sources.AlpineReleaseSource(alpineMirror))
	engine.Register(sources.GoReleaseSource())
	engine.Register(sources.PostgresReleaseSource())
	engine.Register(sources.UtilSource())
}

func Generate(sourceLink, inBase, inRelativePath, outFile string) ([]materials.Change, error) {
	oldMaterials := materials.Read(outFile)
	inFile := filepath.Join(inBase, inRelativePath)

	writer := &bytes.Buffer{}
	newMaterials, err := engine.Execute(writer, inFile)
	if err != nil {
		return nil, fmt.Errorf("unable to render template file %s: %v", outFile, err)
	}

	bom, _ := json.Marshal(newMaterials)
	header := fmt.Sprintf("# Generated from %s%s\n# BOM: %s\n\n", sourceLink, inRelativePath, bom)

	content := append([]byte(header), writer.Bytes()...)
	if err := os.WriteFile(outFile, content, os.FileMode(0600)); err != nil {
		return nil, fmt.Errorf("unable to write container file to %s: %v", outFile, err)
	}

	return materials.Diff(oldMaterials, newMaterials), nil
}
