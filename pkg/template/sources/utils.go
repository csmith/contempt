package sources

import (
	"github.com/csmith/contempt/pkg/template"
	tt "text/template"
)

func UtilSource() template.FunctionSource {
	return func(_ template.BomWriter) tt.FuncMap {
		return tt.FuncMap{
			"increment_int": func(x int) int {
				return x + 1
			},
		}
	}
}
