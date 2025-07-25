// Code generated by ent, DO NOT EDIT.

package task

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.Task {
	return predicate.Task(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.Task {
	return predicate.Task(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.Task {
	return predicate.Task(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.Task {
	return predicate.Task(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.Task {
	return predicate.Task(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.Task {
	return predicate.Task(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.Task {
	return predicate.Task(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.Task {
	return predicate.Task(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.Task {
	return predicate.Task(sql.FieldLTE(FieldID, id))
}

// Title applies equality check predicate on the "title" field. It's identical to TitleEQ.
func Title(v string) predicate.Task {
	return predicate.Task(sql.FieldEQ(FieldTitle, v))
}

// IncidentID applies equality check predicate on the "incident_id" field. It's identical to IncidentIDEQ.
func IncidentID(v uuid.UUID) predicate.Task {
	return predicate.Task(sql.FieldEQ(FieldIncidentID, v))
}

// AssigneeID applies equality check predicate on the "assignee_id" field. It's identical to AssigneeIDEQ.
func AssigneeID(v uuid.UUID) predicate.Task {
	return predicate.Task(sql.FieldEQ(FieldAssigneeID, v))
}

// CreatorID applies equality check predicate on the "creator_id" field. It's identical to CreatorIDEQ.
func CreatorID(v uuid.UUID) predicate.Task {
	return predicate.Task(sql.FieldEQ(FieldCreatorID, v))
}

// TypeEQ applies the EQ predicate on the "type" field.
func TypeEQ(v Type) predicate.Task {
	return predicate.Task(sql.FieldEQ(FieldType, v))
}

// TypeNEQ applies the NEQ predicate on the "type" field.
func TypeNEQ(v Type) predicate.Task {
	return predicate.Task(sql.FieldNEQ(FieldType, v))
}

// TypeIn applies the In predicate on the "type" field.
func TypeIn(vs ...Type) predicate.Task {
	return predicate.Task(sql.FieldIn(FieldType, vs...))
}

// TypeNotIn applies the NotIn predicate on the "type" field.
func TypeNotIn(vs ...Type) predicate.Task {
	return predicate.Task(sql.FieldNotIn(FieldType, vs...))
}

// TitleEQ applies the EQ predicate on the "title" field.
func TitleEQ(v string) predicate.Task {
	return predicate.Task(sql.FieldEQ(FieldTitle, v))
}

// TitleNEQ applies the NEQ predicate on the "title" field.
func TitleNEQ(v string) predicate.Task {
	return predicate.Task(sql.FieldNEQ(FieldTitle, v))
}

// TitleIn applies the In predicate on the "title" field.
func TitleIn(vs ...string) predicate.Task {
	return predicate.Task(sql.FieldIn(FieldTitle, vs...))
}

// TitleNotIn applies the NotIn predicate on the "title" field.
func TitleNotIn(vs ...string) predicate.Task {
	return predicate.Task(sql.FieldNotIn(FieldTitle, vs...))
}

// TitleGT applies the GT predicate on the "title" field.
func TitleGT(v string) predicate.Task {
	return predicate.Task(sql.FieldGT(FieldTitle, v))
}

// TitleGTE applies the GTE predicate on the "title" field.
func TitleGTE(v string) predicate.Task {
	return predicate.Task(sql.FieldGTE(FieldTitle, v))
}

// TitleLT applies the LT predicate on the "title" field.
func TitleLT(v string) predicate.Task {
	return predicate.Task(sql.FieldLT(FieldTitle, v))
}

// TitleLTE applies the LTE predicate on the "title" field.
func TitleLTE(v string) predicate.Task {
	return predicate.Task(sql.FieldLTE(FieldTitle, v))
}

// TitleContains applies the Contains predicate on the "title" field.
func TitleContains(v string) predicate.Task {
	return predicate.Task(sql.FieldContains(FieldTitle, v))
}

// TitleHasPrefix applies the HasPrefix predicate on the "title" field.
func TitleHasPrefix(v string) predicate.Task {
	return predicate.Task(sql.FieldHasPrefix(FieldTitle, v))
}

// TitleHasSuffix applies the HasSuffix predicate on the "title" field.
func TitleHasSuffix(v string) predicate.Task {
	return predicate.Task(sql.FieldHasSuffix(FieldTitle, v))
}

// TitleEqualFold applies the EqualFold predicate on the "title" field.
func TitleEqualFold(v string) predicate.Task {
	return predicate.Task(sql.FieldEqualFold(FieldTitle, v))
}

// TitleContainsFold applies the ContainsFold predicate on the "title" field.
func TitleContainsFold(v string) predicate.Task {
	return predicate.Task(sql.FieldContainsFold(FieldTitle, v))
}

// IncidentIDEQ applies the EQ predicate on the "incident_id" field.
func IncidentIDEQ(v uuid.UUID) predicate.Task {
	return predicate.Task(sql.FieldEQ(FieldIncidentID, v))
}

// IncidentIDNEQ applies the NEQ predicate on the "incident_id" field.
func IncidentIDNEQ(v uuid.UUID) predicate.Task {
	return predicate.Task(sql.FieldNEQ(FieldIncidentID, v))
}

// IncidentIDIn applies the In predicate on the "incident_id" field.
func IncidentIDIn(vs ...uuid.UUID) predicate.Task {
	return predicate.Task(sql.FieldIn(FieldIncidentID, vs...))
}

// IncidentIDNotIn applies the NotIn predicate on the "incident_id" field.
func IncidentIDNotIn(vs ...uuid.UUID) predicate.Task {
	return predicate.Task(sql.FieldNotIn(FieldIncidentID, vs...))
}

// IncidentIDIsNil applies the IsNil predicate on the "incident_id" field.
func IncidentIDIsNil() predicate.Task {
	return predicate.Task(sql.FieldIsNull(FieldIncidentID))
}

// IncidentIDNotNil applies the NotNil predicate on the "incident_id" field.
func IncidentIDNotNil() predicate.Task {
	return predicate.Task(sql.FieldNotNull(FieldIncidentID))
}

// AssigneeIDEQ applies the EQ predicate on the "assignee_id" field.
func AssigneeIDEQ(v uuid.UUID) predicate.Task {
	return predicate.Task(sql.FieldEQ(FieldAssigneeID, v))
}

// AssigneeIDNEQ applies the NEQ predicate on the "assignee_id" field.
func AssigneeIDNEQ(v uuid.UUID) predicate.Task {
	return predicate.Task(sql.FieldNEQ(FieldAssigneeID, v))
}

// AssigneeIDIn applies the In predicate on the "assignee_id" field.
func AssigneeIDIn(vs ...uuid.UUID) predicate.Task {
	return predicate.Task(sql.FieldIn(FieldAssigneeID, vs...))
}

// AssigneeIDNotIn applies the NotIn predicate on the "assignee_id" field.
func AssigneeIDNotIn(vs ...uuid.UUID) predicate.Task {
	return predicate.Task(sql.FieldNotIn(FieldAssigneeID, vs...))
}

// AssigneeIDIsNil applies the IsNil predicate on the "assignee_id" field.
func AssigneeIDIsNil() predicate.Task {
	return predicate.Task(sql.FieldIsNull(FieldAssigneeID))
}

// AssigneeIDNotNil applies the NotNil predicate on the "assignee_id" field.
func AssigneeIDNotNil() predicate.Task {
	return predicate.Task(sql.FieldNotNull(FieldAssigneeID))
}

// CreatorIDEQ applies the EQ predicate on the "creator_id" field.
func CreatorIDEQ(v uuid.UUID) predicate.Task {
	return predicate.Task(sql.FieldEQ(FieldCreatorID, v))
}

// CreatorIDNEQ applies the NEQ predicate on the "creator_id" field.
func CreatorIDNEQ(v uuid.UUID) predicate.Task {
	return predicate.Task(sql.FieldNEQ(FieldCreatorID, v))
}

// CreatorIDIn applies the In predicate on the "creator_id" field.
func CreatorIDIn(vs ...uuid.UUID) predicate.Task {
	return predicate.Task(sql.FieldIn(FieldCreatorID, vs...))
}

// CreatorIDNotIn applies the NotIn predicate on the "creator_id" field.
func CreatorIDNotIn(vs ...uuid.UUID) predicate.Task {
	return predicate.Task(sql.FieldNotIn(FieldCreatorID, vs...))
}

// CreatorIDIsNil applies the IsNil predicate on the "creator_id" field.
func CreatorIDIsNil() predicate.Task {
	return predicate.Task(sql.FieldIsNull(FieldCreatorID))
}

// CreatorIDNotNil applies the NotNil predicate on the "creator_id" field.
func CreatorIDNotNil() predicate.Task {
	return predicate.Task(sql.FieldNotNull(FieldCreatorID))
}

// HasTickets applies the HasEdge predicate on the "tickets" edge.
func HasTickets() predicate.Task {
	return predicate.Task(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, false, TicketsTable, TicketsPrimaryKey...),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasTicketsWith applies the HasEdge predicate on the "tickets" edge with a given conditions (other predicates).
func HasTicketsWith(preds ...predicate.Ticket) predicate.Task {
	return predicate.Task(func(s *sql.Selector) {
		step := newTicketsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasIncident applies the HasEdge predicate on the "incident" edge.
func HasIncident() predicate.Task {
	return predicate.Task(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, IncidentTable, IncidentColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasIncidentWith applies the HasEdge predicate on the "incident" edge with a given conditions (other predicates).
func HasIncidentWith(preds ...predicate.Incident) predicate.Task {
	return predicate.Task(func(s *sql.Selector) {
		step := newIncidentStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasAssignee applies the HasEdge predicate on the "assignee" edge.
func HasAssignee() predicate.Task {
	return predicate.Task(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, AssigneeTable, AssigneeColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasAssigneeWith applies the HasEdge predicate on the "assignee" edge with a given conditions (other predicates).
func HasAssigneeWith(preds ...predicate.User) predicate.Task {
	return predicate.Task(func(s *sql.Selector) {
		step := newAssigneeStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasCreator applies the HasEdge predicate on the "creator" edge.
func HasCreator() predicate.Task {
	return predicate.Task(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, CreatorTable, CreatorColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasCreatorWith applies the HasEdge predicate on the "creator" edge with a given conditions (other predicates).
func HasCreatorWith(preds ...predicate.User) predicate.Task {
	return predicate.Task(func(s *sql.Selector) {
		step := newCreatorStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.Task) predicate.Task {
	return predicate.Task(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.Task) predicate.Task {
	return predicate.Task(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.Task) predicate.Task {
	return predicate.Task(sql.NotPredicates(p))
}
