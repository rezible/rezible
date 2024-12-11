package schema

/*
// LadderAncestry holds the schema definition for the LadderAncestry entity.
type LadderAncestry struct {
	ent.Schema
}

// Fields of the LadderAncestry.
func (LadderAncestry) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("ladder_id", uuid.UUID{}),
		field.UUID("ancestor_id", uuid.UUID{}),
		field.UUID("intermediary_id", uuid.UUID{}).Optional(),
		field.Int("distance"),
		field.Int("distance_to_intermediary").Optional(),
	}
}

func (LadderAncestry) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("ladder", "ancestor").
			Unique(),
		index.Fields("ladder_id", "ancestor_id", "distance").
			Unique(),
		index.Edges("ancestor"),
		index.Edges("intermediary"),
	}
}

// Edges of the LadderAncestry.
func (LadderAncestry) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("ladder", Ladder.Type).
			Ref("ancestry").
			Field("ladder_id").
			Unique().
			Required(),

		edge.From("ancestor", Ladder.Type).
			Ref("ancestors").
			Field("ancestor_id").
			Unique().
			Required(),

		edge.From("intermediary", Ladder.Type).
			Ref("ancestry_intermediaries").
			Field("intermediary_id").
			Unique(),
	}
}

func (LadderAncestry) Annotations() []schema.Annotation {
	return []schema.Annotation{
		&entsql.Annotation{
			Checks: map[string]string{
				"ancestry_closure_check":      `intermediary_id IS NOT NULL OR ladder_id = ancestor_id`,
				"closure_distance_check":      "distance = 0 OR ladder_id <> ancestor_id",
				"intermediary_distance_check": "(distance_to_intermediary IS NULL AND distance = 0) OR distance_to_intermediary = (distance - 1)",
			},
		},
	}
}
*/
