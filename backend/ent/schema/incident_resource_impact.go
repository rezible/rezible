package schema

/*
import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// IncidentResourceImpact holds the schema definition for the IncidentResourceImpact entity.
type IncidentResourceImpact struct {
	ent.Schema
}

// Fields of the IncidentResourceImpact.
func (IncidentResourceImpact) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Default(uuid.New),
		field.UUID("incident_id", uuid.UUID{}),
		field.UUID("service_id", uuid.UUID{}).Optional(),
		field.UUID("functionality_id", uuid.UUID{}).Optional(),
	}
}

func (IncidentResourceImpact) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("incident"),
	}
}

// Edges of the IncidentResourceImpact.
func (IncidentResourceImpact) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("incident", Incident.Type).Ref("impacted_resources").
			Unique().Required().Field("incident_id"),
		edge.From("service", Service.Type).Ref("incidents").
			Unique().Field("service_id"),
		edge.From("functionality", Functionality.Type).Ref("incidents").
			Unique().Field("functionality_id"),
		edge.From("resulting_incidents", IncidentLink.Type).
			Ref("resource_impact"),
	}
}
*/
