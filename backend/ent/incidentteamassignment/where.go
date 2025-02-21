// Code generated by ent, DO NOT EDIT.

package incidentteamassignment

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id int) predicate.IncidentTeamAssignment {
	return predicate.IncidentTeamAssignment(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id int) predicate.IncidentTeamAssignment {
	return predicate.IncidentTeamAssignment(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id int) predicate.IncidentTeamAssignment {
	return predicate.IncidentTeamAssignment(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...int) predicate.IncidentTeamAssignment {
	return predicate.IncidentTeamAssignment(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...int) predicate.IncidentTeamAssignment {
	return predicate.IncidentTeamAssignment(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id int) predicate.IncidentTeamAssignment {
	return predicate.IncidentTeamAssignment(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id int) predicate.IncidentTeamAssignment {
	return predicate.IncidentTeamAssignment(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id int) predicate.IncidentTeamAssignment {
	return predicate.IncidentTeamAssignment(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id int) predicate.IncidentTeamAssignment {
	return predicate.IncidentTeamAssignment(sql.FieldLTE(FieldID, id))
}

// IncidentID applies equality check predicate on the "incident_id" field. It's identical to IncidentIDEQ.
func IncidentID(v uuid.UUID) predicate.IncidentTeamAssignment {
	return predicate.IncidentTeamAssignment(sql.FieldEQ(FieldIncidentID, v))
}

// TeamID applies equality check predicate on the "team_id" field. It's identical to TeamIDEQ.
func TeamID(v uuid.UUID) predicate.IncidentTeamAssignment {
	return predicate.IncidentTeamAssignment(sql.FieldEQ(FieldTeamID, v))
}

// IncidentIDEQ applies the EQ predicate on the "incident_id" field.
func IncidentIDEQ(v uuid.UUID) predicate.IncidentTeamAssignment {
	return predicate.IncidentTeamAssignment(sql.FieldEQ(FieldIncidentID, v))
}

// IncidentIDNEQ applies the NEQ predicate on the "incident_id" field.
func IncidentIDNEQ(v uuid.UUID) predicate.IncidentTeamAssignment {
	return predicate.IncidentTeamAssignment(sql.FieldNEQ(FieldIncidentID, v))
}

// IncidentIDIn applies the In predicate on the "incident_id" field.
func IncidentIDIn(vs ...uuid.UUID) predicate.IncidentTeamAssignment {
	return predicate.IncidentTeamAssignment(sql.FieldIn(FieldIncidentID, vs...))
}

// IncidentIDNotIn applies the NotIn predicate on the "incident_id" field.
func IncidentIDNotIn(vs ...uuid.UUID) predicate.IncidentTeamAssignment {
	return predicate.IncidentTeamAssignment(sql.FieldNotIn(FieldIncidentID, vs...))
}

// TeamIDEQ applies the EQ predicate on the "team_id" field.
func TeamIDEQ(v uuid.UUID) predicate.IncidentTeamAssignment {
	return predicate.IncidentTeamAssignment(sql.FieldEQ(FieldTeamID, v))
}

// TeamIDNEQ applies the NEQ predicate on the "team_id" field.
func TeamIDNEQ(v uuid.UUID) predicate.IncidentTeamAssignment {
	return predicate.IncidentTeamAssignment(sql.FieldNEQ(FieldTeamID, v))
}

// TeamIDIn applies the In predicate on the "team_id" field.
func TeamIDIn(vs ...uuid.UUID) predicate.IncidentTeamAssignment {
	return predicate.IncidentTeamAssignment(sql.FieldIn(FieldTeamID, vs...))
}

// TeamIDNotIn applies the NotIn predicate on the "team_id" field.
func TeamIDNotIn(vs ...uuid.UUID) predicate.IncidentTeamAssignment {
	return predicate.IncidentTeamAssignment(sql.FieldNotIn(FieldTeamID, vs...))
}

// HasIncident applies the HasEdge predicate on the "incident" edge.
func HasIncident() predicate.IncidentTeamAssignment {
	return predicate.IncidentTeamAssignment(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, IncidentTable, IncidentColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasIncidentWith applies the HasEdge predicate on the "incident" edge with a given conditions (other predicates).
func HasIncidentWith(preds ...predicate.Incident) predicate.IncidentTeamAssignment {
	return predicate.IncidentTeamAssignment(func(s *sql.Selector) {
		step := newIncidentStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasTeam applies the HasEdge predicate on the "team" edge.
func HasTeam() predicate.IncidentTeamAssignment {
	return predicate.IncidentTeamAssignment(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, TeamTable, TeamColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasTeamWith applies the HasEdge predicate on the "team" edge with a given conditions (other predicates).
func HasTeamWith(preds ...predicate.Team) predicate.IncidentTeamAssignment {
	return predicate.IncidentTeamAssignment(func(s *sql.Selector) {
		step := newTeamStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.IncidentTeamAssignment) predicate.IncidentTeamAssignment {
	return predicate.IncidentTeamAssignment(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.IncidentTeamAssignment) predicate.IncidentTeamAssignment {
	return predicate.IncidentTeamAssignment(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.IncidentTeamAssignment) predicate.IncidentTeamAssignment {
	return predicate.IncidentTeamAssignment(sql.NotPredicates(p))
}
