package extensions

import (
	"slices"
	"text/template"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

type tenantReferences struct {
	entc.DefaultExtension
}

// TenantReferences generates CheckTenantReferences on the mutation of every tenant-owned type. It counts the
// rows each set or added edge ID names through the mutation's client, whose privacy filter scopes the query to
// the context's tenant, and reports a missing or cross-tenant target as not found. Under
// schema.ReportMissingReferences, a missing target is left to its foreign key instead (checkTenantReference).
func TenantReferences() entc.Extension {
	return tenantReferences{}
}

// Templates implements entc.Extension.
func (tenantReferences) Templates() []*gen.Template {
	return []*gen.Template{tenantReferencesTemplate}
}

func hasTenantID(t *gen.Type) bool {
	return slices.ContainsFunc(t.Fields, func(f *gen.Field) bool {
		return f.Name == "tenant_id"
	})
}

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
