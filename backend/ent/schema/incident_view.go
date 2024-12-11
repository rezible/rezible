package schema

/*
import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)
type IncidentView struct {
	ent.View
}

const incidentViewSql = `
	SELECT
		i.id,
		i.slug,
		i.title,
		i.private,
		i.summary,
		retros.id AS retrospective_id,
		t.id AS incident_type_id,
		t.name AS incident_type_name,
		sev.id AS incident_severity_id,
		sev.name AS incident_severity_name
	FROM incidents i
		JOIN retrospectives retros
			ON i.id = retros.incident_retrospective
		JOIN incident_types t
			ON i.incident_type = t.id
		JOIN incident_severities sev
			ON i.incident_severity = sev.id
		JOIN incident_field_selections f_sel
			ON i.id = f_sel.incident_id
			JOIN public.incident_field_options ifo
				ON f_sel.incident_field_option_id = ifo.id`

func (IncidentView) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.View(`SELECT id`),
	}
}

func (IncidentView) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}),
	}
}
*/
