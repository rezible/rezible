// Code generated by ent, DO NOT EDIT.

package incidenttype

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldLTE(FieldID, id))
}

// ArchiveTime applies equality check predicate on the "archive_time" field. It's identical to ArchiveTimeEQ.
func ArchiveTime(v time.Time) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldEQ(FieldArchiveTime, v))
}

// Name applies equality check predicate on the "name" field. It's identical to NameEQ.
func Name(v string) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldEQ(FieldName, v))
}

// ArchiveTimeEQ applies the EQ predicate on the "archive_time" field.
func ArchiveTimeEQ(v time.Time) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldEQ(FieldArchiveTime, v))
}

// ArchiveTimeNEQ applies the NEQ predicate on the "archive_time" field.
func ArchiveTimeNEQ(v time.Time) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldNEQ(FieldArchiveTime, v))
}

// ArchiveTimeIn applies the In predicate on the "archive_time" field.
func ArchiveTimeIn(vs ...time.Time) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldIn(FieldArchiveTime, vs...))
}

// ArchiveTimeNotIn applies the NotIn predicate on the "archive_time" field.
func ArchiveTimeNotIn(vs ...time.Time) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldNotIn(FieldArchiveTime, vs...))
}

// ArchiveTimeGT applies the GT predicate on the "archive_time" field.
func ArchiveTimeGT(v time.Time) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldGT(FieldArchiveTime, v))
}

// ArchiveTimeGTE applies the GTE predicate on the "archive_time" field.
func ArchiveTimeGTE(v time.Time) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldGTE(FieldArchiveTime, v))
}

// ArchiveTimeLT applies the LT predicate on the "archive_time" field.
func ArchiveTimeLT(v time.Time) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldLT(FieldArchiveTime, v))
}

// ArchiveTimeLTE applies the LTE predicate on the "archive_time" field.
func ArchiveTimeLTE(v time.Time) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldLTE(FieldArchiveTime, v))
}

// ArchiveTimeIsNil applies the IsNil predicate on the "archive_time" field.
func ArchiveTimeIsNil() predicate.IncidentType {
	return predicate.IncidentType(sql.FieldIsNull(FieldArchiveTime))
}

// ArchiveTimeNotNil applies the NotNil predicate on the "archive_time" field.
func ArchiveTimeNotNil() predicate.IncidentType {
	return predicate.IncidentType(sql.FieldNotNull(FieldArchiveTime))
}

// NameEQ applies the EQ predicate on the "name" field.
func NameEQ(v string) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldEQ(FieldName, v))
}

// NameNEQ applies the NEQ predicate on the "name" field.
func NameNEQ(v string) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldNEQ(FieldName, v))
}

// NameIn applies the In predicate on the "name" field.
func NameIn(vs ...string) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldIn(FieldName, vs...))
}

// NameNotIn applies the NotIn predicate on the "name" field.
func NameNotIn(vs ...string) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldNotIn(FieldName, vs...))
}

// NameGT applies the GT predicate on the "name" field.
func NameGT(v string) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldGT(FieldName, v))
}

// NameGTE applies the GTE predicate on the "name" field.
func NameGTE(v string) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldGTE(FieldName, v))
}

// NameLT applies the LT predicate on the "name" field.
func NameLT(v string) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldLT(FieldName, v))
}

// NameLTE applies the LTE predicate on the "name" field.
func NameLTE(v string) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldLTE(FieldName, v))
}

// NameContains applies the Contains predicate on the "name" field.
func NameContains(v string) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldContains(FieldName, v))
}

// NameHasPrefix applies the HasPrefix predicate on the "name" field.
func NameHasPrefix(v string) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldHasPrefix(FieldName, v))
}

// NameHasSuffix applies the HasSuffix predicate on the "name" field.
func NameHasSuffix(v string) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldHasSuffix(FieldName, v))
}

// NameEqualFold applies the EqualFold predicate on the "name" field.
func NameEqualFold(v string) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldEqualFold(FieldName, v))
}

// NameContainsFold applies the ContainsFold predicate on the "name" field.
func NameContainsFold(v string) predicate.IncidentType {
	return predicate.IncidentType(sql.FieldContainsFold(FieldName, v))
}

// HasIncidents applies the HasEdge predicate on the "incidents" edge.
func HasIncidents() predicate.IncidentType {
	return predicate.IncidentType(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, IncidentsTable, IncidentsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasIncidentsWith applies the HasEdge predicate on the "incidents" edge with a given conditions (other predicates).
func HasIncidentsWith(preds ...predicate.Incident) predicate.IncidentType {
	return predicate.IncidentType(func(s *sql.Selector) {
		step := newIncidentsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasDebriefQuestions applies the HasEdge predicate on the "debrief_questions" edge.
func HasDebriefQuestions() predicate.IncidentType {
	return predicate.IncidentType(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, DebriefQuestionsTable, DebriefQuestionsPrimaryKey...),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasDebriefQuestionsWith applies the HasEdge predicate on the "debrief_questions" edge with a given conditions (other predicates).
func HasDebriefQuestionsWith(preds ...predicate.IncidentDebriefQuestion) predicate.IncidentType {
	return predicate.IncidentType(func(s *sql.Selector) {
		step := newDebriefQuestionsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.IncidentType) predicate.IncidentType {
	return predicate.IncidentType(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.IncidentType) predicate.IncidentType {
	return predicate.IncidentType(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.IncidentType) predicate.IncidentType {
	return predicate.IncidentType(sql.NotPredicates(p))
}
