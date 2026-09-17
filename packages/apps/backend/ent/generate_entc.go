//go:build ignore

package main

import (
	"log"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

var enumValuesTemplate = gen.MustParse(gen.NewTemplate("enum_values").Parse(`
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

var debugTemplate = gen.MustParse(gen.NewTemplate("debug").Parse(`
{{ define "debug" }}
	{{/* A template that adds the functionality for running each client <T> in debug mode */}}
	{{ $pkg := base $.Config.Package }}
	{{ template "header" $ }}
	{{ range $n := $.Nodes }}
		{{ $client := print $n.Name "Client" }}
		func (c *{{ $client }}) Debug() *{{ $client }} {
			if c.debug {
				return c
			}
			cfg := config{driver: dialect.Debug(c.driver, c.log), log: c.log, debug: true, hooks: c.hooks, inters: c.inters}
			return &{{ $client }}{config: cfg}
		}
	{{ end }}
{{ end }}
`))

func main() {
	cfg := &gen.Config{
		Features: []gen.Feature{
			gen.FeatureSnapshot,
			gen.FeatureVersionedMigration,
			gen.FeatureEntQL,
			gen.FeatureIntercept,
			gen.FeatureUpsert,
			gen.FeatureModifier,
			gen.FeaturePrivacy,
			gen.FeatureSchemaConfig,
		},
		Templates: []*gen.Template{
			debugTemplate,
			enumValuesTemplate,
		},
	}
	if genErr := entc.Generate("./schema", cfg); genErr != nil {
		log.Fatalln("running ent codegen:", genErr)
	}
}
