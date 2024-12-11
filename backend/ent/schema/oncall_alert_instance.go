package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// OncallAlertInstance holds the schema definition for the OncallAlertInstance entity.
type OncallAlertInstance struct {
	ent.Schema
}

// Fields of the OncallAlertInstance.
func (OncallAlertInstance) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("alert_id", uuid.UUID{}),
		field.Time("created_at"),
		field.Time("acked_at"),
		field.UUID("receiver_user_id", uuid.UUID{}),
	}
}

// Edges of the OncallAlertInstance.
func (OncallAlertInstance) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("alert", OncallAlert.Type).
			Ref("instances").Unique().Required().
			Field("alert_id"),

		edge.To("receiver", User.Type).
			Unique().Required().
			Field("receiver_user_id"),
	}
}
