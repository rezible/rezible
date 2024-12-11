package schema

/*
// Ladder holds the schema definition for the Ladder entity.
type Ladder struct {
	ent.Schema
}

// Fields of the Ladder.
func (Ladder) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("slug").Unique(),
		field.String("name"),
		field.Bool("private").Default(false),
		field.UUID("parent_id", uuid.UUID{}).Optional(),
	}
}

func (Ladder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "parent_id").
			Unique(),
	}
}

// Edges of the Ladder.
func (Ladder) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("users", User.Type),
		edge.To("teams", Team.Type),

		edge.To("services", Service.Type),

		edge.To("parent", Ladder.Type).
			Unique().
			Field("parent_id"),
		edge.From("children", Ladder.Type).
			Ref("parent"),

		edge.To("ancestry", LadderAncestry.Type),
		edge.To("ancestors", LadderAncestry.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("ancestry_intermediaries", LadderAncestry.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
*/
