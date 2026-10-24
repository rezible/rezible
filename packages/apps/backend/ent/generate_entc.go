//go:build ignore

package main

import (
	"log"
	"text/template"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"

	"github.com/rezible/rezible/ent/schema/knowledgerefs"
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

// tenantReferencesTemplate generates CheckTenantReferences on the mutation of every tenant-owned type. It
// counts the rows each set or added edge ID names through the mutation's client, whose privacy filter scopes
// the query to the context's tenant, and reports a missing or cross-tenant target as not found. Under
// schema.ReportMissingReferences, a missing target is left to its foreign key instead (checkTenantReference).
var tenantReferencesTemplate = gen.MustParse(gen.NewTemplate("tenant_references").
	Funcs(template.FuncMap{"hasTenantID": hasTenantID}).
	Parse(`
{{ define "tenant_references" }}
	{{ template "header" $ }}
	import (
		"context"
		{{ $seen := dict }}
		{{- range $n := $.Nodes }}{{ if hasTenantID $n }}{{ range $e := $n.EdgesWithID }}
			{{- if and (hasTenantID $e.Type) (not (hasKey $seen $e.Type.Name)) }}
				{{ $e.Type.PackageAlias }} "{{ $.Config.Package }}/{{ $e.Type.PackageDir }}"
				{{- $seen = set $seen $e.Type.Name true }}
			{{- end }}
		{{- end }}{{ end }}{{ end }}
	)

	{{ range $n := $.Nodes }}{{ if hasTenantID $n }}
		// CheckTenantReferences returns a not found error if an edge ID set or added by the mutation names a row
		// that is missing or belongs to another tenant.
		func (m *{{ $n.MutationName }}) CheckTenantReferences(ctx context.Context) error {
			{{- range $e := $n.EdgesWithID }}{{ if hasTenantID $e.Type }}
				if ids := m.{{ $e.StructField }}IDs(); len(ids) > 0 {
					if err := checkTenantReference(ctx, {{ $e.Type.Package }}.Label, len(ids), m.Client().{{ $e.Type.Name }}.Query().Where({{ $e.Type.Package }}.IDIn(ids...))); err != nil {
						return err
					}
				}
			{{- end }}{{ end }}
			return nil
		}
	{{ end }}{{ end }}
{{ end }}
`))

func hasTenantID(t *gen.Type) bool {
	for _, f := range t.Fields {
		if f.Name == "tenant_id" {
			return true
		}
	}
	return false
}

func main() {
	cfg := &gen.Config{
		Features: []gen.Feature{
			gen.FeatureSnapshot,
			gen.FeatureVersionedMigration,
			gen.FeatureEntQL,
			gen.FeatureIntercept,
			gen.FeatureUpsert,
			gen.FeaturePrivacy,
			gen.FeatureSchemaConfig,
			gen.FeatureLock,
		},
		Templates: []*gen.Template{
			debugTemplate,
			enumValuesTemplate,
			tenantReferencesTemplate,
		},
	}
	opts := []entc.Option{
		entc.Extensions(knowledgerefs.NewExtension()),
	}
	if genErr := entc.Generate("./schema", cfg, opts...); genErr != nil {
		log.Fatalln("running ent codegen:", genErr)
	}
}
