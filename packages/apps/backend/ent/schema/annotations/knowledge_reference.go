// Package annotations holds schema annotations read by Ent code generation. It imports no generated Ent code, so
// the generator can use it while the generated package is being rewritten.
package annotations

// EntityMergeDeletesDuplicates marks a reference edge whose rows an entity merge deletes, rather than re-points,
// where the re-pointed row would repeat one the survivor already has under the column's unique key. Without it, a
// merge re-points every row and such a repeat fails the merge. Code generation refuses it on a column outside a
// unique key.
type EntityMergeDeletesDuplicates struct{}

// Name implements entgo.io/ent/schema.Annotation.
func (EntityMergeDeletesDuplicates) Name() string {
	return "EntityMergeDeletesDuplicates"
}

// EntityMergeDeletesDuplicateRows annotates a reference edge with EntityMergeDeletesDuplicates.
func EntityMergeDeletesDuplicateRows() EntityMergeDeletesDuplicates {
	return EntityMergeDeletesDuplicates{}
}
