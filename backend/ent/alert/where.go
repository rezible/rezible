// Code generated by ent, DO NOT EDIT.

package alert

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.Alert {
	return predicate.Alert(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.Alert {
	return predicate.Alert(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.Alert {
	return predicate.Alert(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.Alert {
	return predicate.Alert(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.Alert {
	return predicate.Alert(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.Alert {
	return predicate.Alert(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.Alert {
	return predicate.Alert(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.Alert {
	return predicate.Alert(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.Alert {
	return predicate.Alert(sql.FieldLTE(FieldID, id))
}

// Title applies equality check predicate on the "title" field. It's identical to TitleEQ.
func Title(v string) predicate.Alert {
	return predicate.Alert(sql.FieldEQ(FieldTitle, v))
}

// ProviderID applies equality check predicate on the "provider_id" field. It's identical to ProviderIDEQ.
func ProviderID(v string) predicate.Alert {
	return predicate.Alert(sql.FieldEQ(FieldProviderID, v))
}

// TitleEQ applies the EQ predicate on the "title" field.
func TitleEQ(v string) predicate.Alert {
	return predicate.Alert(sql.FieldEQ(FieldTitle, v))
}

// TitleNEQ applies the NEQ predicate on the "title" field.
func TitleNEQ(v string) predicate.Alert {
	return predicate.Alert(sql.FieldNEQ(FieldTitle, v))
}

// TitleIn applies the In predicate on the "title" field.
func TitleIn(vs ...string) predicate.Alert {
	return predicate.Alert(sql.FieldIn(FieldTitle, vs...))
}

// TitleNotIn applies the NotIn predicate on the "title" field.
func TitleNotIn(vs ...string) predicate.Alert {
	return predicate.Alert(sql.FieldNotIn(FieldTitle, vs...))
}

// TitleGT applies the GT predicate on the "title" field.
func TitleGT(v string) predicate.Alert {
	return predicate.Alert(sql.FieldGT(FieldTitle, v))
}

// TitleGTE applies the GTE predicate on the "title" field.
func TitleGTE(v string) predicate.Alert {
	return predicate.Alert(sql.FieldGTE(FieldTitle, v))
}

// TitleLT applies the LT predicate on the "title" field.
func TitleLT(v string) predicate.Alert {
	return predicate.Alert(sql.FieldLT(FieldTitle, v))
}

// TitleLTE applies the LTE predicate on the "title" field.
func TitleLTE(v string) predicate.Alert {
	return predicate.Alert(sql.FieldLTE(FieldTitle, v))
}

// TitleContains applies the Contains predicate on the "title" field.
func TitleContains(v string) predicate.Alert {
	return predicate.Alert(sql.FieldContains(FieldTitle, v))
}

// TitleHasPrefix applies the HasPrefix predicate on the "title" field.
func TitleHasPrefix(v string) predicate.Alert {
	return predicate.Alert(sql.FieldHasPrefix(FieldTitle, v))
}

// TitleHasSuffix applies the HasSuffix predicate on the "title" field.
func TitleHasSuffix(v string) predicate.Alert {
	return predicate.Alert(sql.FieldHasSuffix(FieldTitle, v))
}

// TitleEqualFold applies the EqualFold predicate on the "title" field.
func TitleEqualFold(v string) predicate.Alert {
	return predicate.Alert(sql.FieldEqualFold(FieldTitle, v))
}

// TitleContainsFold applies the ContainsFold predicate on the "title" field.
func TitleContainsFold(v string) predicate.Alert {
	return predicate.Alert(sql.FieldContainsFold(FieldTitle, v))
}

// ProviderIDEQ applies the EQ predicate on the "provider_id" field.
func ProviderIDEQ(v string) predicate.Alert {
	return predicate.Alert(sql.FieldEQ(FieldProviderID, v))
}

// ProviderIDNEQ applies the NEQ predicate on the "provider_id" field.
func ProviderIDNEQ(v string) predicate.Alert {
	return predicate.Alert(sql.FieldNEQ(FieldProviderID, v))
}

// ProviderIDIn applies the In predicate on the "provider_id" field.
func ProviderIDIn(vs ...string) predicate.Alert {
	return predicate.Alert(sql.FieldIn(FieldProviderID, vs...))
}

// ProviderIDNotIn applies the NotIn predicate on the "provider_id" field.
func ProviderIDNotIn(vs ...string) predicate.Alert {
	return predicate.Alert(sql.FieldNotIn(FieldProviderID, vs...))
}

// ProviderIDGT applies the GT predicate on the "provider_id" field.
func ProviderIDGT(v string) predicate.Alert {
	return predicate.Alert(sql.FieldGT(FieldProviderID, v))
}

// ProviderIDGTE applies the GTE predicate on the "provider_id" field.
func ProviderIDGTE(v string) predicate.Alert {
	return predicate.Alert(sql.FieldGTE(FieldProviderID, v))
}

// ProviderIDLT applies the LT predicate on the "provider_id" field.
func ProviderIDLT(v string) predicate.Alert {
	return predicate.Alert(sql.FieldLT(FieldProviderID, v))
}

// ProviderIDLTE applies the LTE predicate on the "provider_id" field.
func ProviderIDLTE(v string) predicate.Alert {
	return predicate.Alert(sql.FieldLTE(FieldProviderID, v))
}

// ProviderIDContains applies the Contains predicate on the "provider_id" field.
func ProviderIDContains(v string) predicate.Alert {
	return predicate.Alert(sql.FieldContains(FieldProviderID, v))
}

// ProviderIDHasPrefix applies the HasPrefix predicate on the "provider_id" field.
func ProviderIDHasPrefix(v string) predicate.Alert {
	return predicate.Alert(sql.FieldHasPrefix(FieldProviderID, v))
}

// ProviderIDHasSuffix applies the HasSuffix predicate on the "provider_id" field.
func ProviderIDHasSuffix(v string) predicate.Alert {
	return predicate.Alert(sql.FieldHasSuffix(FieldProviderID, v))
}

// ProviderIDEqualFold applies the EqualFold predicate on the "provider_id" field.
func ProviderIDEqualFold(v string) predicate.Alert {
	return predicate.Alert(sql.FieldEqualFold(FieldProviderID, v))
}

// ProviderIDContainsFold applies the ContainsFold predicate on the "provider_id" field.
func ProviderIDContainsFold(v string) predicate.Alert {
	return predicate.Alert(sql.FieldContainsFold(FieldProviderID, v))
}

// HasMetrics applies the HasEdge predicate on the "metrics" edge.
func HasMetrics() predicate.Alert {
	return predicate.Alert(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, MetricsTable, MetricsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasMetricsWith applies the HasEdge predicate on the "metrics" edge with a given conditions (other predicates).
func HasMetricsWith(preds ...predicate.AlertMetrics) predicate.Alert {
	return predicate.Alert(func(s *sql.Selector) {
		step := newMetricsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasPlaybooks applies the HasEdge predicate on the "playbooks" edge.
func HasPlaybooks() predicate.Alert {
	return predicate.Alert(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, PlaybooksTable, PlaybooksPrimaryKey...),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasPlaybooksWith applies the HasEdge predicate on the "playbooks" edge with a given conditions (other predicates).
func HasPlaybooksWith(preds ...predicate.Playbook) predicate.Alert {
	return predicate.Alert(func(s *sql.Selector) {
		step := newPlaybooksStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasInstances applies the HasEdge predicate on the "instances" edge.
func HasInstances() predicate.Alert {
	return predicate.Alert(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, InstancesTable, InstancesColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasInstancesWith applies the HasEdge predicate on the "instances" edge with a given conditions (other predicates).
func HasInstancesWith(preds ...predicate.OncallEvent) predicate.Alert {
	return predicate.Alert(func(s *sql.Selector) {
		step := newInstancesStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.Alert) predicate.Alert {
	return predicate.Alert(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.Alert) predicate.Alert {
	return predicate.Alert(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.Alert) predicate.Alert {
	return predicate.Alert(sql.NotPredicates(p))
}
