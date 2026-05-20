package db

import "github.com/rezible/rezible/integrations/projections"

func RegisterEventProcessors() {
	projections.RegisterHandler("knowledge", knowledgeEntityEventProjectionHandler)
	projections.RegisterHandler("users", userEventProjectionHandler)
	projections.RegisterHandler("incidents", handleIncidentEventProjection)
	projections.RegisterHandler("alerts", alertEventProjectionHandler)
}
