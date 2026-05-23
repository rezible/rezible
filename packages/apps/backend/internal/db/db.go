package db

import "github.com/rezible/rezible/integrations/projections"

func RegisterEventProcessors() {
	projections.RegisterHandler("knowledge", handleKnowledgeEntityEventProjection,
		projections.SubjectKindCodeForge,
		projections.SubjectKindCodeChange,
		projections.SubjectKindSystemComponent,
		projections.SubjectKindSystemRelationship,
	)
	projections.RegisterHandler("users", handleUserEventProjection, projections.SubjectKindUser)
	projections.RegisterHandler("incidents", handleIncidentEventProjection, projections.SubjectKindIncident)
	projections.RegisterHandler("alerts", handleAlertEventProjection, projections.SubjectKindAlert)
}
