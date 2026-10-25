package extensions

import (
	"errors"
	"slices"
	"strings"
	"text/template"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

type splitMutations struct {
	entc.DefaultExtension
}

// SplitMutations splits ent's mutation.go, which outgrows editor file size limits, into a <type>_mutation.go
// per type. Each file renders ent's own "mutation" template for a graph holding only that type, without the
// operation and node type constants the template emits for its whole graph. mutation.go keeps those constants.
func SplitMutations() entc.Extension {
	return splitMutations{}
}

// Hooks implements entc.Extension.
func (s splitMutations) Hooks() []gen.Hook {
	replaceMutationTemplate := func(next gen.Generator) gen.Generator {
		return gen.GenerateFunc(func(g *gen.Graph) error {
			mutationIdx := slices.IndexFunc(gen.GraphTemplates, func(t gen.GraphTemplate) bool {
				return t.Name == "mutation"
			})
			if mutationIdx < 0 {
				return errors.New("split mutations: ent has no mutation graph template")
			}
			gen.GraphTemplates[mutationIdx].Name = "helper/mutation_constants"
			gen.Templates = append(gen.Templates, gen.TypeTemplate{
				Name:   "helper/mutation_type",
				Cond:   func(t *gen.Type) bool { return !t.IsView() },
				Format: func(t *gen.Type) string { return t.PackageDir() + "_mutation.go" },
			})
			return next.Generate(g)
		})
	}
	return []gen.Hook{replaceMutationTemplate}
}

// Templates implements entc.Extension.
func (splitMutations) Templates() []*gen.Template {
	return []*gen.Template{gen.MustParse(gen.NewTemplate("mutation_split").
		Funcs(template.FuncMap{
			"typeGraph": func(t *gen.Type) *gen.Graph {
				return &gen.Graph{Config: t.Config, Nodes: []*gen.Type{t}}
			},
			"stripConstants": func(src string) (string, error) {
				if start := strings.Index(src, "const ("); start >= 0 {
					if end := strings.Index(src[start:], "\n)"); end >= 0 {
						return src[:start] + src[start+end+len("\n)"):], nil
					}
				}
				return "", errors.New("invalid ent mutation template constants block")
			},
		}).
		Parse(splitMutationsTemplate))}
}

var splitMutationsTemplate = `
{{ define "helper/mutation_constants" }}
	{{ template "header" $ }}
	import "entgo.io/ent"

	const (
		// Operation types.
		OpCreate    = ent.OpCreate
		OpDelete    = ent.OpDelete
		OpDeleteOne = ent.OpDeleteOne
		OpUpdate    = ent.OpUpdate
		OpUpdateOne = ent.OpUpdateOne

		// Node types.
		{{- range $n := $.Nodes }}
			{{ $n.TypeName }} = "{{ $n.Name }}"
		{{- end }}
	)
{{ end }}

{{ define "helper/mutation_type" }}
	{{- stripConstants (xtemplate "mutation" (typeGraph $)) }}
{{ end }}
`
