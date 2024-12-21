package contempt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/csmith/contempt/sources"
)

var templateFuncs []template.FuncMap

// TODO: This is all a bit gross. Templating should be a separate package,
// and the BOM should be passed in to it rather than relying on random global state.

func InitTemplates(imageRegistry, alpineMirror string) {
	templateFuncs = append(
		templateFuncs, template.FuncMap{
			"regex_url_content": regexURLContent,
			"increment_int": func(x int) int {
				return x + 1
			},
		},
		sources.ImageFuncs(&materials, imageRegistry),
		sources.GitFuncs(&materials),
		sources.AlpineReleaseFuncs(&materials, alpineMirror),
		sources.GoReleaseFuncs(&materials),
		sources.PostgresReleaseFuncs(&materials),
		sources.AlpinePackagesFuncs(&materials, alpineMirror),
	)
}

func regexURLContent(name, url, regex string) string {
	res, err := sources.RegexURLContent(url, regex)
	if err != nil {
		log.Fatalf("Couldn't find regex in url '%s'", name)
	}
	materials[fmt.Sprintf("regexurl:%s", name)] = res
	return res
}

func Generate(sourceLink, inBase, inRelativePath, outFile string) ([]Change, error) {
	materials = make(map[string]string)
	oldMaterials := readBillOfMaterials(outFile)
	inFile := filepath.Join(inBase, inRelativePath)

	tpl := template.New(inFile)
	for i := range templateFuncs {
		tpl.Funcs(templateFuncs[i])
	}

	if _, err := tpl.ParseFiles(inFile); err != nil {
		return nil, fmt.Errorf("unable to parse template file %s: %v", inFile, err)
	}

	writer := &bytes.Buffer{}
	if err := tpl.ExecuteTemplate(writer, filepath.Base(inFile), nil); err != nil {
		return nil, fmt.Errorf("unable to render template file %s: %v", outFile, err)
	}

	bom, _ := json.Marshal(materials)
	header := fmt.Sprintf("# Generated from %s%s\n# BOM: %s\n\n", sourceLink, inRelativePath, bom)

	content := append([]byte(header), writer.Bytes()...)
	if err := os.WriteFile(outFile, content, os.FileMode(0600)); err != nil {
		return nil, fmt.Errorf("unable to write container file to %s: %v", outFile, err)
	}

	return diffMaterials(oldMaterials, materials), nil
}
