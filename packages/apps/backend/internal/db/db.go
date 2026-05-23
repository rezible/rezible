package db

import "github.com/rezible/rezible/integrations/projections"

func RegisterEventProcessors() {
	projections.RegisterHandler("knowledge", handleKnowledgeEntityEventProjection)
	projections.RegisterHandler("users", handleUserEventProjection)
	projections.RegisterHandler("incidents", handleIncidentEventProjection)
	projections.RegisterHandler("alerts", handleAlertEventProjection)
}
