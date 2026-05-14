package db

import "github.com/rezible/rezible/integrations/eventprojections"

func RegisterEventProcessors() {
	eventprojections.RegisterHandler("knowledge", knowledgeEntityEventProjectionHandler)
	eventprojections.RegisterHandler("users", userEventProjectionHandler)
	eventprojections.RegisterHandler("incidents", incidentEventProjectionHandler)
	eventprojections.RegisterHandler("alerts", alertEventProjectionHandler)
}
