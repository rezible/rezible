package extensions

import (
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

type enumValues struct {
	entc.DefaultExtension
}

// EnumValues generates a <Field>Values slice of the permitted values for every enum field without a Go type.
func EnumValues() entc.Extension {
	return enumValues{}
}

// Templates implements entc.Extension.
func (enumValues) Templates() []*gen.Template {
	return []*gen.Template{enumValuesTemplate}
}

var enumValuesTemplate = gen.MustParse(gen.NewTemplate("enum_values").
	Parse(`
{{ define "meta/additional/enum-values" }}
	{{ range $field := $.EnumFields }}
		{{ if not $field.HasGoType }}
			// {{ $field.StructField }}Values contains all permitted values. Treat this slice as read-only.
			var {{ $field.StructField }}Values = []string{
				{{ range $field.Enums }}
					{{- printf "%q" .Value }},
				{{ end }}
			}
		{{ end }}
	{{ end }}
{{ end }}
`))
