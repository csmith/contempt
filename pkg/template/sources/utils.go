package sources

import (
	"errors"
	"github.com/csmith/contempt/pkg/template"
	tt "text/template"
)

func UtilSource() template.FunctionSource {
	return func(_ template.BomWriter) tt.FuncMap {
		return tt.FuncMap{
			"increment_int": func(x int) int {
				return x + 1
			},
			"map": func(pairs ...any) (map[string]interface{}, error) {
				if len(pairs)%2 != 0 {
					return nil, errors.New("incorrect number of arguments")
				}

				m := make(map[string]interface{}, len(pairs)/2)
				for i := 0; i < len(pairs); i += 2 {
					k, ok := pairs[i].(string)
					if !ok {
						return nil, errors.New("map keys must be strings")
					}
					m[k] = pairs[i+1]
				}

				return m, nil
			},
			"arr": func(elements ...any) []interface{} {
				return elements
			},
		}
	}
}
