// Code generated by ent, DO NOT EDIT.

package incidentlink

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/twohundreds/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id int) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id int) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id int) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...int) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...int) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id int) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id int) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id int) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id int) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldLTE(FieldID, id))
}

// IncidentID applies equality check predicate on the "incident_id" field. It's identical to IncidentIDEQ.
func IncidentID(v uuid.UUID) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldEQ(FieldIncidentID, v))
}

// LinkedIncidentID applies equality check predicate on the "linked_incident_id" field. It's identical to LinkedIncidentIDEQ.
func LinkedIncidentID(v uuid.UUID) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldEQ(FieldLinkedIncidentID, v))
}

// Description applies equality check predicate on the "description" field. It's identical to DescriptionEQ.
func Description(v string) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldEQ(FieldDescription, v))
}

// IncidentIDEQ applies the EQ predicate on the "incident_id" field.
func IncidentIDEQ(v uuid.UUID) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldEQ(FieldIncidentID, v))
}

// IncidentIDNEQ applies the NEQ predicate on the "incident_id" field.
func IncidentIDNEQ(v uuid.UUID) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldNEQ(FieldIncidentID, v))
}

// IncidentIDIn applies the In predicate on the "incident_id" field.
func IncidentIDIn(vs ...uuid.UUID) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldIn(FieldIncidentID, vs...))
}

// IncidentIDNotIn applies the NotIn predicate on the "incident_id" field.
func IncidentIDNotIn(vs ...uuid.UUID) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldNotIn(FieldIncidentID, vs...))
}

// LinkedIncidentIDEQ applies the EQ predicate on the "linked_incident_id" field.
func LinkedIncidentIDEQ(v uuid.UUID) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldEQ(FieldLinkedIncidentID, v))
}

// LinkedIncidentIDNEQ applies the NEQ predicate on the "linked_incident_id" field.
func LinkedIncidentIDNEQ(v uuid.UUID) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldNEQ(FieldLinkedIncidentID, v))
}

// LinkedIncidentIDIn applies the In predicate on the "linked_incident_id" field.
func LinkedIncidentIDIn(vs ...uuid.UUID) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldIn(FieldLinkedIncidentID, vs...))
}

// LinkedIncidentIDNotIn applies the NotIn predicate on the "linked_incident_id" field.
func LinkedIncidentIDNotIn(vs ...uuid.UUID) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldNotIn(FieldLinkedIncidentID, vs...))
}

// DescriptionEQ applies the EQ predicate on the "description" field.
func DescriptionEQ(v string) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldEQ(FieldDescription, v))
}

// DescriptionNEQ applies the NEQ predicate on the "description" field.
func DescriptionNEQ(v string) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldNEQ(FieldDescription, v))
}

// DescriptionIn applies the In predicate on the "description" field.
func DescriptionIn(vs ...string) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldIn(FieldDescription, vs...))
}

// DescriptionNotIn applies the NotIn predicate on the "description" field.
func DescriptionNotIn(vs ...string) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldNotIn(FieldDescription, vs...))
}

// DescriptionGT applies the GT predicate on the "description" field.
func DescriptionGT(v string) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldGT(FieldDescription, v))
}

// DescriptionGTE applies the GTE predicate on the "description" field.
func DescriptionGTE(v string) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldGTE(FieldDescription, v))
}

// DescriptionLT applies the LT predicate on the "description" field.
func DescriptionLT(v string) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldLT(FieldDescription, v))
}

// DescriptionLTE applies the LTE predicate on the "description" field.
func DescriptionLTE(v string) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldLTE(FieldDescription, v))
}

// DescriptionContains applies the Contains predicate on the "description" field.
func DescriptionContains(v string) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldContains(FieldDescription, v))
}

// DescriptionHasPrefix applies the HasPrefix predicate on the "description" field.
func DescriptionHasPrefix(v string) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldHasPrefix(FieldDescription, v))
}

// DescriptionHasSuffix applies the HasSuffix predicate on the "description" field.
func DescriptionHasSuffix(v string) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldHasSuffix(FieldDescription, v))
}

// DescriptionIsNil applies the IsNil predicate on the "description" field.
func DescriptionIsNil() predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldIsNull(FieldDescription))
}

// DescriptionNotNil applies the NotNil predicate on the "description" field.
func DescriptionNotNil() predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldNotNull(FieldDescription))
}

// DescriptionEqualFold applies the EqualFold predicate on the "description" field.
func DescriptionEqualFold(v string) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldEqualFold(FieldDescription, v))
}

// DescriptionContainsFold applies the ContainsFold predicate on the "description" field.
func DescriptionContainsFold(v string) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldContainsFold(FieldDescription, v))
}

// LinkTypeEQ applies the EQ predicate on the "link_type" field.
func LinkTypeEQ(v LinkType) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldEQ(FieldLinkType, v))
}

// LinkTypeNEQ applies the NEQ predicate on the "link_type" field.
func LinkTypeNEQ(v LinkType) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldNEQ(FieldLinkType, v))
}

// LinkTypeIn applies the In predicate on the "link_type" field.
func LinkTypeIn(vs ...LinkType) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldIn(FieldLinkType, vs...))
}

// LinkTypeNotIn applies the NotIn predicate on the "link_type" field.
func LinkTypeNotIn(vs ...LinkType) predicate.IncidentLink {
	return predicate.IncidentLink(sql.FieldNotIn(FieldLinkType, vs...))
}

// HasIncident applies the HasEdge predicate on the "incident" edge.
func HasIncident() predicate.IncidentLink {
	return predicate.IncidentLink(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, IncidentTable, IncidentColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasIncidentWith applies the HasEdge predicate on the "incident" edge with a given conditions (other predicates).
func HasIncidentWith(preds ...predicate.Incident) predicate.IncidentLink {
	return predicate.IncidentLink(func(s *sql.Selector) {
		step := newIncidentStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasLinkedIncident applies the HasEdge predicate on the "linked_incident" edge.
func HasLinkedIncident() predicate.IncidentLink {
	return predicate.IncidentLink(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, LinkedIncidentTable, LinkedIncidentColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasLinkedIncidentWith applies the HasEdge predicate on the "linked_incident" edge with a given conditions (other predicates).
func HasLinkedIncidentWith(preds ...predicate.Incident) predicate.IncidentLink {
	return predicate.IncidentLink(func(s *sql.Selector) {
		step := newLinkedIncidentStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasResourceImpact applies the HasEdge predicate on the "resource_impact" edge.
func HasResourceImpact() predicate.IncidentLink {
	return predicate.IncidentLink(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, ResourceImpactTable, ResourceImpactColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasResourceImpactWith applies the HasEdge predicate on the "resource_impact" edge with a given conditions (other predicates).
func HasResourceImpactWith(preds ...predicate.IncidentResourceImpact) predicate.IncidentLink {
	return predicate.IncidentLink(func(s *sql.Selector) {
		step := newResourceImpactStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.IncidentLink) predicate.IncidentLink {
	return predicate.IncidentLink(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.IncidentLink) predicate.IncidentLink {
	return predicate.IncidentLink(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.IncidentLink) predicate.IncidentLink {
	return predicate.IncidentLink(sql.NotPredicates(p))
}