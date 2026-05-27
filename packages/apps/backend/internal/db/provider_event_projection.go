package db

import "github.com/rezible/rezible/integrations/projections"

func RegisterEventProcessors(reg *projections.EventProjectionHandlerRegistry) {
	reg.RegisterHandler("knowledge", HandleKnowledgeEntityEventProjection,
		projections.SubjectKindCodeForge,
		projections.SubjectKindCodeChange,
		projections.SubjectKindSystemComponent,
		projections.SubjectKindSystemRelationship,
	)
	reg.RegisterHandler("users", HandleUserEventProjection, projections.SubjectKindUser)
	reg.RegisterHandler("incidents", HandleIncidentEventProjection, projections.SubjectKindIncident)
	reg.RegisterHandler("alerts", HandleAlertEventProjection, projections.SubjectKindAlert)

}
