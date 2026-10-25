package extensions

import (
	"cmp"
	"crypto/md5"
	"fmt"
	"slices"
	"text/template"

	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"

	"github.com/rezible/rezible/ent/schema/annotations"
)

type knowledgeReferences struct {
	entc.DefaultExtension
	refs *references
}

// KnowledgeReferences generates ent/knowledge_references.go, which lists every column referring to a knowledge
// entity or relationship, so an entity merge can re-point them. Generation fails on a foreign key to a knowledge
// subject that the list does not cover, on a reference shape the merge cannot re-point, and on an
// annotations.EntityMergeDeletesDuplicates the merge would not use. Use one per generation run: its hook
// collects the references the template then writes.
func KnowledgeReferences() entc.Extension {
	return &knowledgeReferences{}
}

// Hooks implements entc.Extension.
func (e *knowledgeReferences) Hooks() []gen.Hook {
	return []gen.Hook{e.collectHook}
}

// Templates implements entc.Extension.
func (e *knowledgeReferences) Templates() []*gen.Template {
	refsTemplate := gen.MustParse(gen.NewTemplate("knowledge_references").
		Funcs(template.FuncMap{
			"knowledgeReferences": func() *references {
				return e.refs
			},
		}).
		Parse(referencesTemplate))
	return []*gen.Template{refsTemplate}
}

func (e *knowledgeReferences) collectHook(next gen.Generator) gen.Generator {
	return gen.GenerateFunc(func(g *gen.Graph) error {
		refs, collectErr := collect(g, coreColumns)
		if collectErr != nil {
			return fmt.Errorf("knowledge references: %w", collectErr)
		}
		e.refs = refs
		return next.Generate(g)
	})
}

// subjectTypes are the schema types a knowledge reference points into, with the list each reference joins. Their
// tables are read from the graph, so renaming a table changes nothing, and a missing type fails generation.
var subjectTypes = map[string]string{
	"KnowledgeEntity":       "entity",
	"KnowledgeRelationship": "relationship",
}

// column is a column by its storage names.
type column struct {
	table  string
	column string
}

func (c column) String() string {
	return c.table + "." + c.column
}

// coreColumns are the knowledge graph's own references to its subjects, which an entity merge moves in steps of
// its own rather than through generated descriptors.
var coreColumns = map[column]bool{
	{"knowledge_entity_linking_attributes", "entity_id"}: true,
	{"knowledge_subject_alias", "entity_id"}:             true,
	{"knowledge_subject_alias", "relationship_id"}:       true,
	{"knowledge_relationships", "source_entity_id"}:      true,
	{"knowledge_relationships", "target_entity_id"}:      true,
}

// reference is a generated reference descriptor.
type reference struct {
	Table            string
	Column           string
	Node             string
	Setter           string
	DeleteDuplicates bool
	UniqueWith       []string
}

// references is the data of the knowledge_references template.
type references struct {
	Entities              []reference
	Relationships         []reference
	ForeignKeyConstraints []string
}

// list is a generated list of descriptors.
type list struct {
	Name       string
	Doc        string
	References []reference
}

// Lists are the generated lists of descriptors.
func (r references) Lists() []list {
	return []list{
		{
			Name:       "KnowledgeEntityReferences",
			Doc:        "KnowledgeEntityReferences are the columns, outside the knowledge graph, that refer to a knowledge entity.",
			References: r.Entities,
		},
		{
			Name:       "KnowledgeRelationshipReferences",
			Doc:        "KnowledgeRelationshipReferences are the columns, outside the knowledge graph, that refer to a knowledge relationship.",
			References: r.Relationships,
		},
	}
}

// referenceEdge is an edge whose own table holds a foreign key to a knowledge subject.
type referenceEdge struct {
	node *gen.Type
	edge *gen.Edge
}

func (e referenceEdge) String() string {
	return e.node.Name + "." + e.edge.Name
}

func (e referenceEdge) deletesDuplicates() bool {
	_, annotated := e.edge.Annotations[annotations.EntityMergeDeletesDuplicates{}.Name()]
	return annotated
}

// collect walks every foreign key to a knowledge subject. Each one is one of the core columns or a supported
// reference edge. Every subject type, and every core column, must exist, and each subject must have references:
// an empty list means the walk no longer matches the schema, which would silently stop merges re-pointing.
func collect(g *gen.Graph, core map[column]bool) (*references, error) {
	tables, tablesErr := g.Tables()
	if tablesErr != nil {
		return nil, tablesErr
	}
	subjectTables, subjectsErr := findSubjectTables(g)
	if subjectsErr != nil {
		return nil, subjectsErr
	}
	foundCore := make(map[column]bool, len(core))

	edges := make(map[column]referenceEdge)
	var annotated []referenceEdge
	for _, n := range g.Nodes {
		for _, e := range n.Edges {
			edge := referenceEdge{node: n, edge: e}
			if edge.deletesDuplicates() {
				annotated = append(annotated, edge)
			}
			if e.OwnFK() && subjectTables[e.Type.Table()] != "" {
				edges[column{e.Rel.Table, e.Rel.Column()}] = edge
			}
		}
	}

	refs := &references{}
	generated := make(map[*gen.Edge]bool)
	for _, table := range tables {
		for _, foreignKey := range table.ForeignKeys {
			subject := subjectTables[foreignKey.RefTable.Name]
			if subject == "" {
				continue
			}
			refs.ForeignKeyConstraints = append(refs.ForeignKeyConstraints, postgresSymbol(foreignKey.Symbol))
			if len(foreignKey.Columns) != 1 {
				return nil, fmt.Errorf("foreign key %s on %s to %s has %d columns; only single-column references are supported",
					foreignKey.Symbol, table.Name, foreignKey.RefTable.Name, len(foreignKey.Columns))
			}
			col := column{table.Name, foreignKey.Columns[0].Name}
			if core[col] {
				foundCore[col] = true
				continue
			}
			edge, isEdge := edges[col]
			if !isEdge {
				return nil, fmt.Errorf("%s has a foreign key to %s that is neither a core column nor an edge of its own table "+
					"(join tables and edges declared on the knowledge subject are not supported): declare it as an edge.To with Field() on the referring schema",
					col, foreignKey.RefTable.Name)
			}
			ref, refErr := newReference(table, col, edge)
			if refErr != nil {
				return nil, refErr
			}
			generated[edge.edge] = true
			if subject == "entity" {
				refs.Entities = append(refs.Entities, *ref)
			} else {
				refs.Relationships = append(refs.Relationships, *ref)
			}
		}
	}
	for _, edge := range annotated {
		if !generated[edge.edge] {
			return nil, fmt.Errorf("edge %s has annotations.EntityMergeDeletesDuplicates but is not a knowledge reference: remove the annotation", edge)
		}
	}

	for col := range core {
		if !foundCore[col] {
			return nil, fmt.Errorf("core column %s is no longer a foreign key to a knowledge subject: update coreColumns", col)
		}
	}
	if len(refs.Entities) == 0 || len(refs.Relationships) == 0 {
		return nil, fmt.Errorf("found %d entity and %d relationship references; each subject should have some, "+
			"so the walk no longer matches the schema", len(refs.Entities), len(refs.Relationships))
	}

	byColumn := func(a, b reference) int {
		return cmp.Or(cmp.Compare(a.Table, b.Table), cmp.Compare(a.Column, b.Column))
	}
	slices.SortFunc(refs.Entities, byColumn)
	slices.SortFunc(refs.Relationships, byColumn)
	slices.Sort(refs.ForeignKeyConstraints)
	return refs, nil
}

// findSubjectTables maps each subject type's table to its list, failing on a subject type the graph lacks.
func findSubjectTables(g *gen.Graph) (map[string]string, error) {
	tables := make(map[string]string, len(subjectTypes))
	for _, n := range g.Nodes {
		if subject, isSubject := subjectTypes[n.Name]; isSubject {
			tables[n.Table()] = subject
		}
	}
	for name := range subjectTypes {
		if !slices.ContainsFunc(g.Nodes, func(n *gen.Type) bool { return n.Name == name }) {
			return nil, fmt.Errorf("schema type %s is missing: update subjectTypes", name)
		}
	}
	return tables, nil
}

// newReference describes the reference edge holding col. A merge re-points every row, so a repeat under a unique
// key fails it, unless the edge deletes duplicates; then col must be in exactly one unique key, which is not
// partial.
func newReference(table *schema.Table, col column, edge referenceEdge) (*reference, error) {
	e := edge.edge
	f := e.Field()
	switch {
	case e.IsInverse():
		return nil, fmt.Errorf("%s is held by the inverse edge %s; declare the reference as an edge.To on its own schema", col, edge)
	case f == nil:
		return nil, fmt.Errorf("%s is held by edge %s, which has no Field(); declare the column as an edge field", col, edge)
	case f.Immutable || e.Immutable:
		return nil, fmt.Errorf("%s (edge %s) is Immutable(); a merge must update it, so drop Immutable()", col, edge)
	}
	for _, primaryKey := range table.PrimaryKey {
		if primaryKey.Name == col.column {
			return nil, fmt.Errorf("%s is part of the primary key of %s; a merge cannot re-point it", col, table.Name)
		}
	}

	ref := &reference{
		Table:  col.table,
		Column: col.column,
		Node:   edge.node.Name,
		Setter: f.MutationSet(),
	}
	if !edge.deletesDuplicates() {
		return ref, nil
	}

	var uniqueKeys [][]string
	if c, hasColumn := table.Column(col.column); hasColumn && c.Unique {
		uniqueKeys = append(uniqueKeys, []string{col.column})
	}
	for _, index := range table.Indexes {
		if !index.Unique || !slices.ContainsFunc(index.Columns, func(c *schema.Column) bool { return c.Name == col.column }) {
			continue
		}
		if index.Annotation != nil && index.Annotation.Where != "" {
			return nil, fmt.Errorf("edge %s deletes duplicates, but %s is part of the partial unique index %s; partial unique keys are not supported",
				edge, col, index.Name)
		}
		key := make([]string, 0, len(index.Columns))
		for _, c := range index.Columns {
			key = append(key, c.Name)
		}
		uniqueKeys = append(uniqueKeys, key)
	}
	switch len(uniqueKeys) {
	case 0:
		return nil, fmt.Errorf("edge %s has annotations.EntityMergeDeletesDuplicates, but %s is not part of a unique key: remove the annotation",
			edge, col)
	case 1:
	default:
		return nil, fmt.Errorf("edge %s deletes duplicates, but %s is part of %d unique keys %v; it supports one", edge, col, len(uniqueKeys), uniqueKeys)
	}
	ref.DeleteDuplicates = true
	for _, c := range uniqueKeys[0] {
		if c != col.column {
			ref.UniqueWith = append(ref.UniqueWith, c)
		}
	}
	return ref, nil
}

// postgresSymbol is the name Ent's migration gives a constraint in Postgres, which limits identifiers to 63
// bytes: a longer name becomes its first 30 bytes and its MD5 hash (see symbol in
// entgo.io/ent/dialect/sql/schema/atlas.go).
func postgresSymbol(name string) string {
	if len(name) <= 63 {
		return name
	}
	return fmt.Sprintf("%s_%x", name[:30], md5.Sum([]byte(name)))
}

// referencesTemplate writes the lists. Every top-level define in an Ent template emits a file, so it has none
// but its own.
const referencesTemplate = `
{{ define "knowledge_references" }}
	{{ template "header" $ }}
	{{ $refs := knowledgeReferences }}

	import (
		"context"

		"entgo.io/ent/dialect/sql"
		"github.com/google/uuid"

		"{{ $.Config.Package }}/predicate"
	)

	// KnowledgeReference is a column, outside the knowledge graph's own, that holds the ID of a knowledge entity or
	// relationship under a foreign key. Generated from the reference edges by ent/schema/extensions.
	type KnowledgeReference struct {
		// Table and Column are storage names.
		Table  string
		Column string
		// DeleteDuplicates says a merge deletes the other's row where, re-pointed, it would repeat the survivor's
		// under the unique key holding Column; otherwise such a repeat fails the merge.
		DeleteDuplicates bool
		// UniqueWith are the other columns, tenant_id among them, of that unique key; set with DeleteDuplicates.
		UniqueWith []string
		// Update sets Column to id on the rows matching where, through the generated update builder.
		Update func(ctx context.Context, client *Client, where func(*sql.Selector), id uuid.UUID) error
		// Delete deletes the rows matching where, through the generated delete builder; set with DeleteDuplicates.
		Delete func(ctx context.Context, client *Client, where func(*sql.Selector)) (int, error)
	}

	{{- range $list := $refs.Lists }}
		// {{ $list.Doc }}
		var {{ $list.Name }} = []KnowledgeReference{
			{{- range $ref := $list.References }}
				{
					Table:       {{ printf "%q" $ref.Table }},
					Column:      {{ printf "%q" $ref.Column }},
					{{- if $ref.DeleteDuplicates }}
						DeleteDuplicates: true,
						UniqueWith: []string{ {{- range $i, $c := $ref.UniqueWith }}{{ if $i }}, {{ end }}{{ printf "%q" $c }}{{ end -}} },
					{{- end }}
					Update: func(ctx context.Context, client *Client, where func(*sql.Selector), id uuid.UUID) error {
						update := client.{{ $ref.Node }}.Update().
							Where(predicate.{{ $ref.Node }}(where)).
							{{ $ref.Setter }}(id)
						return update.Exec(ctx)
					},
					{{- if $ref.DeleteDuplicates }}
						Delete: func(ctx context.Context, client *Client, where func(*sql.Selector)) (int, error) {
							deleteRows := client.{{ $ref.Node }}.Delete().
								Where(predicate.{{ $ref.Node }}(where))
							return deleteRows.Exec(ctx)
						},
					{{- end }}
				},
			{{- end }}
		}
	{{ end }}

	// KnowledgeForeignKeyConstraints are the names Postgres reports for every foreign key to knowledge entities or
	// relationships, the knowledge graph's own included.
	var KnowledgeForeignKeyConstraints = []string{
		{{- range $name := $refs.ForeignKeyConstraints }}
			{{ printf "%q" $name }},
		{{- end }}
	}
{{ end }}
`
