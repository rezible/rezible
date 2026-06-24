-- create "agent_runs" table
CREATE TABLE "agent_runs" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "attempt" bigint NOT NULL, "status" character varying NOT NULL DEFAULT 'queued', "started_at" timestamptz NULL, "finished_at" timestamptz NULL, "error_message" text NULL, "tenant_id" bigint NOT NULL, "agent_task_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "agentrun_tenant_id" to table: "agent_runs"
CREATE INDEX "agentrun_tenant_id" ON "agent_runs" ("tenant_id");
-- create index "agentrun_tenant_id_agent_task_id_attempt" to table: "agent_runs"
CREATE UNIQUE INDEX "agentrun_tenant_id_agent_task_id_attempt" ON "agent_runs" ("tenant_id", "agent_task_id", "attempt");
-- create index "agentrun_tenant_id_status_created_at" to table: "agent_runs"
CREATE INDEX "agentrun_tenant_id_status_created_at" ON "agent_runs" ("tenant_id", "status", "created_at");
-- create index "agentrun_tenant_id_updated_at" to table: "agent_runs"
CREATE INDEX "agentrun_tenant_id_updated_at" ON "agent_runs" ("tenant_id", "updated_at");
-- create "agent_run_citations" table
CREATE TABLE "agent_run_citations" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "citation_kind" character varying NOT NULL, "domain_entity_type" character varying NULL, "domain_entity_id" uuid NULL, "summary" text NOT NULL, "snapshot" jsonb NULL, "tenant_id" bigint NOT NULL, "agent_run_id" uuid NOT NULL, "knowledge_entity_id" uuid NULL, "knowledge_relationship_id" uuid NULL, "knowledge_evidence_id" uuid NULL, "agent_task_id" uuid NULL, "agent_run_tool_call_id" uuid NULL, PRIMARY KEY ("id"));
-- create index "agentruncitation_tenant_id" to table: "agent_run_citations"
CREATE INDEX "agentruncitation_tenant_id" ON "agent_run_citations" ("tenant_id");
-- create index "agentruncitation_tenant_id_agent_run_id_created_at" to table: "agent_run_citations"
CREATE INDEX "agentruncitation_tenant_id_agent_run_id_created_at" ON "agent_run_citations" ("tenant_id", "agent_run_id", "created_at");
-- create index "agentruncitation_tenant_id_citation_kind" to table: "agent_run_citations"
CREATE INDEX "agentruncitation_tenant_id_citation_kind" ON "agent_run_citations" ("tenant_id", "citation_kind");
-- create index "agentruncitation_tenant_id_domain_entity_type_domain_entity_id" to table: "agent_run_citations"
CREATE INDEX "agentruncitation_tenant_id_domain_entity_type_domain_entity_id" ON "agent_run_citations" ("tenant_id", "domain_entity_type", "domain_entity_id");
-- create index "agentruncitation_tenant_id_knowledge_entity_id" to table: "agent_run_citations"
CREATE INDEX "agentruncitation_tenant_id_knowledge_entity_id" ON "agent_run_citations" ("tenant_id", "knowledge_entity_id");
-- create index "agentruncitation_tenant_id_knowledge_relationship_id" to table: "agent_run_citations"
CREATE INDEX "agentruncitation_tenant_id_knowledge_relationship_id" ON "agent_run_citations" ("tenant_id", "knowledge_relationship_id");
-- create index "agentruncitation_tenant_id_knowledge_evidence_id" to table: "agent_run_citations"
CREATE INDEX "agentruncitation_tenant_id_knowledge_evidence_id" ON "agent_run_citations" ("tenant_id", "knowledge_evidence_id");
-- create index "agentruncitation_tenant_id_agent_task_id" to table: "agent_run_citations"
CREATE INDEX "agentruncitation_tenant_id_agent_task_id" ON "agent_run_citations" ("tenant_id", "agent_task_id");
-- create index "agentruncitation_tenant_id_agent_run_tool_call_id" to table: "agent_run_citations"
CREATE INDEX "agentruncitation_tenant_id_agent_run_tool_call_id" ON "agent_run_citations" ("tenant_id", "agent_run_tool_call_id");
-- create "agent_run_findings" table
CREATE TABLE "agent_run_findings" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "sequence" bigint NOT NULL, "finding_kind" character varying NOT NULL, "content" text NOT NULL, "tenant_id" bigint NOT NULL, "agent_run_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "agentrunfinding_tenant_id" to table: "agent_run_findings"
CREATE INDEX "agentrunfinding_tenant_id" ON "agent_run_findings" ("tenant_id");
-- create index "agentrunfinding_tenant_id_agent_run_id_created_at" to table: "agent_run_findings"
CREATE INDEX "agentrunfinding_tenant_id_agent_run_id_created_at" ON "agent_run_findings" ("tenant_id", "agent_run_id", "created_at");
-- create index "agentrunfinding_tenant_id_agent_run_id_sequence" to table: "agent_run_findings"
CREATE UNIQUE INDEX "agentrunfinding_tenant_id_agent_run_id_sequence" ON "agent_run_findings" ("tenant_id", "agent_run_id", "sequence");
-- create index "agentrunfinding_tenant_id_finding_kind" to table: "agent_run_findings"
CREATE INDEX "agentrunfinding_tenant_id_finding_kind" ON "agent_run_findings" ("tenant_id", "finding_kind");
-- create "agent_run_finding_citations" table
CREATE TABLE "agent_run_finding_citations" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "support_kind" character varying NOT NULL, "tenant_id" bigint NOT NULL, "agent_run_finding_id" uuid NOT NULL, "agent_run_citation_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "agentrunfindingcitation_tenant_id" to table: "agent_run_finding_citations"
CREATE INDEX "agentrunfindingcitation_tenant_id" ON "agent_run_finding_citations" ("tenant_id");
-- create index "agentrunfindingcitation_tenant_id_agent_run_finding_id" to table: "agent_run_finding_citations"
CREATE INDEX "agentrunfindingcitation_tenant_id_agent_run_finding_id" ON "agent_run_finding_citations" ("tenant_id", "agent_run_finding_id");
-- create index "agentrunfindingcitation_tenant_id_agent_run_citation_id" to table: "agent_run_finding_citations"
CREATE INDEX "agentrunfindingcitation_tenant_id_agent_run_citation_id" ON "agent_run_finding_citations" ("tenant_id", "agent_run_citation_id");
-- create index "agentrunfindingcitation_agent__043bc294dcf83a3b3084f97c43ceb39f" to table: "agent_run_finding_citations"
CREATE UNIQUE INDEX "agentrunfindingcitation_agent__043bc294dcf83a3b3084f97c43ceb39f" ON "agent_run_finding_citations" ("agent_run_finding_id", "agent_run_citation_id", "support_kind");
-- create "agent_run_results" table
CREATE TABLE "agent_run_results" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "content" text NOT NULL, "data" jsonb NULL, "tenant_id" bigint NOT NULL, "agent_run_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "agentrunresult_tenant_id" to table: "agent_run_results"
CREATE INDEX "agentrunresult_tenant_id" ON "agent_run_results" ("tenant_id");
-- create index "agentrunresult_tenant_id_agent_run_id" to table: "agent_run_results"
CREATE UNIQUE INDEX "agentrunresult_tenant_id_agent_run_id" ON "agent_run_results" ("tenant_id", "agent_run_id");
-- create "agent_run_tool_calls" table
CREATE TABLE "agent_run_tool_calls" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "tool_name" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'requested', "tool_params" jsonb NULL, "result" jsonb NULL, "error_message" text NULL, "started_at" timestamptz NULL, "finished_at" timestamptz NULL, "tenant_id" bigint NOT NULL, "agent_run_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "agentruntoolcall_tenant_id" to table: "agent_run_tool_calls"
CREATE INDEX "agentruntoolcall_tenant_id" ON "agent_run_tool_calls" ("tenant_id");
-- create index "agentruntoolcall_tenant_id_agent_run_id_created_at" to table: "agent_run_tool_calls"
CREATE INDEX "agentruntoolcall_tenant_id_agent_run_id_created_at" ON "agent_run_tool_calls" ("tenant_id", "agent_run_id", "created_at");
-- create index "agentruntoolcall_tenant_id_tool_name_created_at" to table: "agent_run_tool_calls"
CREATE INDEX "agentruntoolcall_tenant_id_tool_name_created_at" ON "agent_run_tool_calls" ("tenant_id", "tool_name", "created_at");
-- create index "agentruntoolcall_tenant_id_status_created_at" to table: "agent_run_tool_calls"
CREATE INDEX "agentruntoolcall_tenant_id_status_created_at" ON "agent_run_tool_calls" ("tenant_id", "status", "created_at");
-- create "agent_tasks" table
CREATE TABLE "agent_tasks" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "workflow_kind" character varying NOT NULL, "workflow_input" jsonb NULL, "trigger_kind" character varying NOT NULL, "trigger_payload" jsonb NULL, "tenant_id" bigint NOT NULL, "owner_user_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "agenttask_tenant_id" to table: "agent_tasks"
CREATE INDEX "agenttask_tenant_id" ON "agent_tasks" ("tenant_id");
-- create index "agenttask_tenant_id_owner_user_id_created_at" to table: "agent_tasks"
CREATE INDEX "agenttask_tenant_id_owner_user_id_created_at" ON "agent_tasks" ("tenant_id", "owner_user_id", "created_at");
-- create index "agenttask_tenant_id_workflow_kind_created_at" to table: "agent_tasks"
CREATE INDEX "agenttask_tenant_id_workflow_kind_created_at" ON "agent_tasks" ("tenant_id", "workflow_kind", "created_at");
-- create "alerts" table
CREATE TABLE "alerts" ("id" uuid NOT NULL, "title" character varying NOT NULL, "description" character varying NULL, "definition" character varying NULL, "tenant_id" bigint NOT NULL, "knowledge_entity_id" uuid NULL, "roster_id" uuid NULL, PRIMARY KEY ("id"));
-- create index "alert_tenant_id" to table: "alerts"
CREATE INDEX "alert_tenant_id" ON "alerts" ("tenant_id");
-- create index "alert_tenant_id_knowledge_entity_id" to table: "alerts"
CREATE UNIQUE INDEX "alert_tenant_id_knowledge_entity_id" ON "alerts" ("tenant_id", "knowledge_entity_id");
-- create "alert_feedbacks" table
CREATE TABLE "alert_feedbacks" ("id" uuid NOT NULL, "actionable" boolean NOT NULL, "accurate" character varying NOT NULL, "documentation_available" boolean NOT NULL, "documentation_needs_update" boolean NOT NULL, "tenant_id" bigint NOT NULL, "alert_id" uuid NOT NULL, "alert_instance_id" uuid NULL, "normalized_event_alert_feedback" uuid NULL, PRIMARY KEY ("id"));
-- create index "alertfeedback_tenant_id" to table: "alert_feedbacks"
CREATE INDEX "alertfeedback_tenant_id" ON "alert_feedbacks" ("tenant_id");
-- create "documents" table
CREATE TABLE "documents" ("id" uuid NOT NULL, "content" bytea NOT NULL, "access_restricted" boolean NOT NULL DEFAULT false, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "document_tenant_id" to table: "documents"
CREATE INDEX "document_tenant_id" ON "documents" ("tenant_id");
-- create "document_accesses" table
CREATE TABLE "document_accesses" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "can_view" boolean NOT NULL DEFAULT false, "can_edit" boolean NOT NULL DEFAULT false, "can_manage" boolean NOT NULL DEFAULT false, "tenant_id" bigint NOT NULL, "document_id" uuid NOT NULL, "user_id" uuid NULL, "team_id" uuid NULL, PRIMARY KEY ("id"));
-- create index "documentaccess_tenant_id" to table: "document_accesses"
CREATE INDEX "documentaccess_tenant_id" ON "document_accesses" ("tenant_id");
-- create "event_annotations" table
CREATE TABLE "event_annotations" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "minutes_occupied" bigint NOT NULL, "notes" text NOT NULL, "tags" jsonb NOT NULL, "tenant_id" bigint NOT NULL, "event_id" uuid NOT NULL, "creator_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "eventannotation_tenant_id" to table: "event_annotations"
CREATE INDEX "eventannotation_tenant_id" ON "event_annotations" ("tenant_id");
-- create "incidents" table
CREATE TABLE "incidents" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "slug" character varying NOT NULL, "title" character varying NOT NULL, "summary" character varying NULL, "chat_channel_id" character varying NULL, "opened_at" timestamptz NOT NULL, "tenant_id" bigint NOT NULL, "knowledge_entity_id" uuid NULL, "severity_id" uuid NOT NULL, "type_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "incidents_slug_key" to table: "incidents"
CREATE UNIQUE INDEX "incidents_slug_key" ON "incidents" ("slug");
-- create index "incident_tenant_id" to table: "incidents"
CREATE INDEX "incident_tenant_id" ON "incidents" ("tenant_id");
-- create index "incident_tenant_id_knowledge_entity_id" to table: "incidents"
CREATE UNIQUE INDEX "incident_tenant_id_knowledge_entity_id" ON "incidents" ("tenant_id", "knowledge_entity_id");
-- create "incident_debriefs" table
CREATE TABLE "incident_debriefs" ("id" uuid NOT NULL, "required" boolean NOT NULL, "started" boolean NOT NULL, "incident_id" uuid NOT NULL, "tenant_id" bigint NOT NULL, "user_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "incidentdebrief_tenant_id" to table: "incident_debriefs"
CREATE INDEX "incidentdebrief_tenant_id" ON "incident_debriefs" ("tenant_id");
-- create "incident_debrief_messages" table
CREATE TABLE "incident_debrief_messages" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "type" character varying NOT NULL, "requested_tool" character varying NULL, "body" text NOT NULL, "debrief_id" uuid NOT NULL, "tenant_id" bigint NOT NULL, "question_id" uuid NULL, PRIMARY KEY ("id"));
-- create index "incidentdebriefmessage_tenant_id" to table: "incident_debrief_messages"
CREATE INDEX "incidentdebriefmessage_tenant_id" ON "incident_debrief_messages" ("tenant_id");
-- create "incident_debrief_questions" table
CREATE TABLE "incident_debrief_questions" ("id" uuid NOT NULL, "content" text NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "incidentdebriefquestion_tenant_id" to table: "incident_debrief_questions"
CREATE INDEX "incidentdebriefquestion_tenant_id" ON "incident_debrief_questions" ("tenant_id");
-- create "incident_debrief_suggestions" table
CREATE TABLE "incident_debrief_suggestions" ("id" uuid NOT NULL, "content" text NOT NULL, "incident_debrief_suggestions" uuid NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "incidentdebriefsuggestion_tenant_id" to table: "incident_debrief_suggestions"
CREATE INDEX "incidentdebriefsuggestion_tenant_id" ON "incident_debrief_suggestions" ("tenant_id");
-- create "incident_fields" table
CREATE TABLE "incident_fields" ("id" uuid NOT NULL, "archive_time" timestamptz NULL, "name" character varying NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "incidentfield_tenant_id" to table: "incident_fields"
CREATE INDEX "incidentfield_tenant_id" ON "incident_fields" ("tenant_id");
-- create "incident_field_options" table
CREATE TABLE "incident_field_options" ("id" uuid NOT NULL, "archive_time" timestamptz NULL, "type" character varying NOT NULL, "value" character varying NOT NULL, "incident_field_id" uuid NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "incidentfieldoption_tenant_id" to table: "incident_field_options"
CREATE INDEX "incidentfieldoption_tenant_id" ON "incident_field_options" ("tenant_id");
-- create "incident_impacts" table
CREATE TABLE "incident_impacts" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "source" character varying NULL, "note" text NULL, "tenant_id" bigint NOT NULL, "incident_id" uuid NOT NULL, "knowledge_entity_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "incidentimpact_tenant_id" to table: "incident_impacts"
CREATE INDEX "incidentimpact_tenant_id" ON "incident_impacts" ("tenant_id");
-- create index "incidentimpact_tenant_id_incident_id" to table: "incident_impacts"
CREATE INDEX "incidentimpact_tenant_id_incident_id" ON "incident_impacts" ("tenant_id", "incident_id");
-- create index "incidentimpact_tenant_id_knowledge_entity_id" to table: "incident_impacts"
CREATE INDEX "incidentimpact_tenant_id_knowledge_entity_id" ON "incident_impacts" ("tenant_id", "knowledge_entity_id");
-- create index "incidentimpact_tenant_id_incident_id_knowledge_entity_id" to table: "incident_impacts"
CREATE UNIQUE INDEX "incidentimpact_tenant_id_incident_id_knowledge_entity_id" ON "incident_impacts" ("tenant_id", "incident_id", "knowledge_entity_id");
-- create "incident_links" table
CREATE TABLE "incident_links" ("id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY, "link_type" character varying NOT NULL, "description" character varying NULL, "tenant_id" bigint NOT NULL, "incident_id" uuid NOT NULL, "linked_incident_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "incidentlink_tenant_id" to table: "incident_links"
CREATE INDEX "incidentlink_tenant_id" ON "incident_links" ("tenant_id");
-- create index "incidentlink_incident_id_linked_incident_id" to table: "incident_links"
CREATE UNIQUE INDEX "incidentlink_incident_id_linked_incident_id" ON "incident_links" ("incident_id", "linked_incident_id");
-- create "incident_milestones" table
CREATE TABLE "incident_milestones" ("id" uuid NOT NULL, "kind" character varying NOT NULL, "timestamp" timestamptz NOT NULL, "description" character varying NULL, "source" character varying NULL, "metadata" jsonb NULL, "incident_id" uuid NOT NULL, "tenant_id" bigint NOT NULL, "user_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "incidentmilestone_tenant_id" to table: "incident_milestones"
CREATE INDEX "incidentmilestone_tenant_id" ON "incident_milestones" ("tenant_id");
-- create "incident_roles" table
CREATE TABLE "incident_roles" ("id" uuid NOT NULL, "archive_time" timestamptz NULL, "name" character varying NOT NULL, "required" boolean NOT NULL DEFAULT false, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "incidentrole_tenant_id" to table: "incident_roles"
CREATE INDEX "incidentrole_tenant_id" ON "incident_roles" ("tenant_id");
-- create "incident_role_assignments" table
CREATE TABLE "incident_role_assignments" ("id" uuid NOT NULL, "tenant_id" bigint NOT NULL, "incident_id" uuid NOT NULL, "user_id" uuid NOT NULL, "role_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "incidentroleassignment_tenant_id" to table: "incident_role_assignments"
CREATE INDEX "incidentroleassignment_tenant_id" ON "incident_role_assignments" ("tenant_id");
-- create index "incidentroleassignment_user_id_incident_id" to table: "incident_role_assignments"
CREATE UNIQUE INDEX "incidentroleassignment_user_id_incident_id" ON "incident_role_assignments" ("user_id", "incident_id");
-- create "incident_severities" table
CREATE TABLE "incident_severities" ("id" uuid NOT NULL, "archive_time" timestamptz NULL, "name" character varying NOT NULL, "rank" bigint NOT NULL, "color" character varying NULL, "description" character varying NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "incidentseverity_tenant_id" to table: "incident_severities"
CREATE INDEX "incidentseverity_tenant_id" ON "incident_severities" ("tenant_id");
-- create index "incidentseverity_tenant_id_name" to table: "incident_severities"
CREATE UNIQUE INDEX "incidentseverity_tenant_id_name" ON "incident_severities" ("tenant_id", "name");
-- create "incident_tags" table
CREATE TABLE "incident_tags" ("id" uuid NOT NULL, "archive_time" timestamptz NULL, "key" character varying NOT NULL, "value" character varying NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "incidenttag_tenant_id" to table: "incident_tags"
CREATE INDEX "incidenttag_tenant_id" ON "incident_tags" ("tenant_id");
-- create "incident_timeline_events" table
CREATE TABLE "incident_timeline_events" ("id" uuid NOT NULL, "timestamp" timestamptz NOT NULL, "kind" character varying NOT NULL, "title" character varying NOT NULL, "description" text NULL, "is_key" boolean NOT NULL DEFAULT false, "sequence" bigint NOT NULL DEFAULT 0, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "incident_id" uuid NOT NULL, "tenant_id" bigint NOT NULL, "event_id" uuid NULL, PRIMARY KEY ("id"));
-- create index "incidenttimelineevent_tenant_id" to table: "incident_timeline_events"
CREATE INDEX "incidenttimelineevent_tenant_id" ON "incident_timeline_events" ("tenant_id");
-- create index "incidenttimelineevent_kind" to table: "incident_timeline_events"
CREATE INDEX "incidenttimelineevent_kind" ON "incident_timeline_events" ("kind");
-- create "incident_timeline_event_contexts" table
CREATE TABLE "incident_timeline_event_contexts" ("id" uuid NOT NULL, "system_state" text NULL, "decision_options" jsonb NULL, "decision_rationale" text NULL, "involved_personnel" jsonb NULL, "created_at" timestamptz NOT NULL, "incident_timeline_event_context" uuid NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "incident_timeline_event_contexts_incident_timeline_event_context_key" to table: "incident_timeline_event_contexts"
CREATE UNIQUE INDEX "incident_timeline_event_contexts_incident_timeline_event_context_key" ON "incident_timeline_event_contexts" ("incident_timeline_event_context");
-- create index "incidenttimelineeventcontext_tenant_id" to table: "incident_timeline_event_contexts"
CREATE INDEX "incidenttimelineeventcontext_tenant_id" ON "incident_timeline_event_contexts" ("tenant_id");
-- create "incident_timeline_event_contributing_factors" table
CREATE TABLE "incident_timeline_event_contributing_factors" ("id" uuid NOT NULL, "factor_type" character varying NOT NULL, "description" text NULL, "created_at" timestamptz NOT NULL, "incident_timeline_event_factors" uuid NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "incidenttimelineeventcontributingfactor_tenant_id" to table: "incident_timeline_event_contributing_factors"
CREATE INDEX "incidenttimelineeventcontributingfactor_tenant_id" ON "incident_timeline_event_contributing_factors" ("tenant_id");
-- create "incident_timeline_event_evidences" table
CREATE TABLE "incident_timeline_event_evidences" ("id" uuid NOT NULL, "evidence_type" character varying NOT NULL, "url" character varying NOT NULL, "title" character varying NOT NULL, "description" text NULL, "created_at" timestamptz NOT NULL, "incident_timeline_event_evidence" uuid NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "incidenttimelineeventevidence_tenant_id" to table: "incident_timeline_event_evidences"
CREATE INDEX "incidenttimelineeventevidence_tenant_id" ON "incident_timeline_event_evidences" ("tenant_id");
-- create "incident_timeline_event_topology_contexts" table
CREATE TABLE "incident_timeline_event_topology_contexts" ("id" uuid NOT NULL, "relationship" character varying NOT NULL, "created_at" timestamptz NOT NULL, "tenant_id" bigint NOT NULL, "incident_event_id" uuid NOT NULL, "knowledge_entity_id" uuid NULL, "snapshot_entity_id" uuid NULL, PRIMARY KEY ("id"));
-- create index "incidenttimelineeventtopologycontext_tenant_id" to table: "incident_timeline_event_topology_contexts"
CREATE INDEX "incidenttimelineeventtopologycontext_tenant_id" ON "incident_timeline_event_topology_contexts" ("tenant_id");
-- create "incident_types" table
CREATE TABLE "incident_types" ("id" uuid NOT NULL, "archive_time" timestamptz NULL, "name" character varying NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "incidenttype_tenant_id" to table: "incident_types"
CREATE INDEX "incidenttype_tenant_id" ON "incident_types" ("tenant_id");
-- create index "incidenttype_tenant_id_name" to table: "incident_types"
CREATE UNIQUE INDEX "incidenttype_tenant_id_name" ON "incident_types" ("tenant_id", "name");
-- create "integrations" table
CREATE TABLE "integrations" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "integration_name" character varying NOT NULL, "external_provider_ref" character varying NOT NULL, "installation_config" jsonb NOT NULL, "user_settings" jsonb NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "integration_tenant_id" to table: "integrations"
CREATE INDEX "integration_tenant_id" ON "integrations" ("tenant_id");
-- create index "integration_tenant_id_integration_name" to table: "integrations"
CREATE INDEX "integration_tenant_id_integration_name" ON "integrations" ("tenant_id", "integration_name");
-- create index "integration_tenant_id_integration_name_external_provider_ref" to table: "integrations"
CREATE UNIQUE INDEX "integration_tenant_id_integration_name_external_provider_ref" ON "integrations" ("tenant_id", "integration_name", "external_provider_ref");
-- create "integration_event_sync_cursors" table
CREATE TABLE "integration_event_sync_cursors" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "provider_source" character varying NOT NULL, "cursor" character varying NULL, "last_synced_at" timestamptz NOT NULL, "tenant_id" bigint NOT NULL, "integration_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "integrationeventsynccursor_tenant_id" to table: "integration_event_sync_cursors"
CREATE INDEX "integrationeventsynccursor_tenant_id" ON "integration_event_sync_cursors" ("tenant_id");
-- create index "integrationeventsynccursor_ten_914d3d8b389cb5d930bdf0bb43683869" to table: "integration_event_sync_cursors"
CREATE UNIQUE INDEX "integrationeventsynccursor_ten_914d3d8b389cb5d930bdf0bb43683869" ON "integration_event_sync_cursors" ("tenant_id", "integration_id", "provider_source");
-- create "integration_event_sync_runs" table
CREATE TABLE "integration_event_sync_runs" ("id" uuid NOT NULL, "source_cursors" jsonb NULL, "sync_reason" character varying NOT NULL DEFAULT 'manual', "started_at" timestamptz NOT NULL, "finished_at" timestamptz NULL, "status" character varying NOT NULL, "events_pulled" bigint NOT NULL DEFAULT 0, "events_ingested" bigint NOT NULL DEFAULT 0, "duplicates" bigint NOT NULL DEFAULT 0, "failure_message" character varying NULL, "tenant_id" bigint NOT NULL, "integration_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "integrationeventsyncrun_tenant_id" to table: "integration_event_sync_runs"
CREATE INDEX "integrationeventsyncrun_tenant_id" ON "integration_event_sync_runs" ("tenant_id");
-- create index "integrationeventsyncrun_tenant_id_integration_id_started_at" to table: "integration_event_sync_runs"
CREATE INDEX "integrationeventsyncrun_tenant_id_integration_id_started_at" ON "integration_event_sync_runs" ("tenant_id", "integration_id", "started_at");
-- create index "integrationeventsyncrun_tenant_id_status_started_at" to table: "integration_event_sync_runs"
CREATE INDEX "integrationeventsyncrun_tenant_id_status_started_at" ON "integration_event_sync_runs" ("tenant_id", "status", "started_at");
-- create "integration_user_install_states" table
CREATE TABLE "integration_user_install_states" ("id" uuid NOT NULL, "integration_name" character varying NOT NULL, "oauth_state" character varying NULL, "installation_targets" jsonb NULL, "expires_at" timestamptz NOT NULL, "tenant_id" bigint NOT NULL, "user_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "integrationuserinstallstate_tenant_id" to table: "integration_user_install_states"
CREATE INDEX "integrationuserinstallstate_tenant_id" ON "integration_user_install_states" ("tenant_id");
-- create index "integrationuserinstallstate_tenant_id_user_id_integration_name" to table: "integration_user_install_states"
CREATE UNIQUE INDEX "integrationuserinstallstate_tenant_id_user_id_integration_name" ON "integration_user_install_states" ("tenant_id", "user_id", "integration_name");
-- create "knowledge_entities" table
CREATE TABLE "knowledge_entities" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "kind" character varying NOT NULL, "display_name" character varying NOT NULL, "description" text NULL, "first_observed_at" timestamptz NULL, "last_observed_at" timestamptz NULL, "deleted_at" timestamptz NULL, "properties" jsonb NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "knowledgeentity_tenant_id" to table: "knowledge_entities"
CREATE INDEX "knowledgeentity_tenant_id" ON "knowledge_entities" ("tenant_id");
-- create index "knowledgeentity_tenant_id_kind" to table: "knowledge_entities"
CREATE INDEX "knowledgeentity_tenant_id_kind" ON "knowledge_entities" ("tenant_id", "kind");
-- create index "knowledgeentity_tenant_id_updated_at" to table: "knowledge_entities"
CREATE INDEX "knowledgeentity_tenant_id_updated_at" ON "knowledge_entities" ("tenant_id", "updated_at");
-- create index "knowledgeentity_tenant_id_kind_last_observed_at" to table: "knowledge_entities"
CREATE INDEX "knowledgeentity_tenant_id_kind_last_observed_at" ON "knowledge_entities" ("tenant_id", "kind", "last_observed_at");
-- create index "knowledgeentity_tenant_id_kind_deleted_at" to table: "knowledge_entities"
CREATE INDEX "knowledgeentity_tenant_id_kind_deleted_at" ON "knowledge_entities" ("tenant_id", "kind", "deleted_at");
-- create "knowledge_entity_alias" table
CREATE TABLE "knowledge_entity_alias" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "display_name" character varying NULL, "provider" character varying NOT NULL, "provider_subject_ref" character varying NOT NULL, "tenant_id" bigint NOT NULL, "entity_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "knowledgeentityalias_tenant_id" to table: "knowledge_entity_alias"
CREATE INDEX "knowledgeentityalias_tenant_id" ON "knowledge_entity_alias" ("tenant_id");
-- create index "knowledgeentityalias_tenant_id_entity_id" to table: "knowledge_entity_alias"
CREATE INDEX "knowledgeentityalias_tenant_id_entity_id" ON "knowledge_entity_alias" ("tenant_id", "entity_id");
-- create index "knowledgeentityalias_tenant_id_provider_provider_subject_ref" to table: "knowledge_entity_alias"
CREATE UNIQUE INDEX "knowledgeentityalias_tenant_id_provider_provider_subject_ref" ON "knowledge_entity_alias" ("tenant_id", "provider", "provider_subject_ref");
-- create "knowledge_evidences" table
CREATE TABLE "knowledge_evidences" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "subject_type" character varying NOT NULL, "assertion" character varying NOT NULL, "evidence_kind" character varying NOT NULL, "observed_at" timestamptz NOT NULL, "effective_at" timestamptz NULL, "tenant_id" bigint NOT NULL, "entity_id" uuid NULL, "relationship_id" uuid NULL, "alias_id" uuid NULL, "event_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "knowledgeevidence_tenant_id" to table: "knowledge_evidences"
CREATE INDEX "knowledgeevidence_tenant_id" ON "knowledge_evidences" ("tenant_id");
-- create index "knowledgeevidence_tenant_id_en_f1b9bb3b8ee45b8e08fb3b4cf03cfec9" to table: "knowledge_evidences"
CREATE UNIQUE INDEX "knowledgeevidence_tenant_id_en_f1b9bb3b8ee45b8e08fb3b4cf03cfec9" ON "knowledge_evidences" ("tenant_id", "entity_id", "event_id", "subject_type", "relationship_id", "alias_id");
-- create index "knowledgeevidence_tenant_id_entity_id" to table: "knowledge_evidences"
CREATE INDEX "knowledgeevidence_tenant_id_entity_id" ON "knowledge_evidences" ("tenant_id", "entity_id");
-- create index "knowledgeevidence_tenant_id_relationship_id" to table: "knowledge_evidences"
CREATE INDEX "knowledgeevidence_tenant_id_relationship_id" ON "knowledge_evidences" ("tenant_id", "relationship_id");
-- create index "knowledgeevidence_tenant_id_alias_id" to table: "knowledge_evidences"
CREATE INDEX "knowledgeevidence_tenant_id_alias_id" ON "knowledge_evidences" ("tenant_id", "alias_id");
-- create index "knowledgeevidence_tenant_id_event_id" to table: "knowledge_evidences"
CREATE INDEX "knowledgeevidence_tenant_id_event_id" ON "knowledge_evidences" ("tenant_id", "event_id");
-- create "knowledge_relationships" table
CREATE TABLE "knowledge_relationships" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "kind" character varying NOT NULL, "display_name" character varying NULL, "description" text NULL, "first_observed_at" timestamptz NULL, "last_observed_at" timestamptz NULL, "deleted_at" timestamptz NULL, "properties" jsonb NULL, "tenant_id" bigint NOT NULL, "source_entity_id" uuid NOT NULL, "target_entity_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "knowledgerelationship_tenant_id" to table: "knowledge_relationships"
CREATE INDEX "knowledgerelationship_tenant_id" ON "knowledge_relationships" ("tenant_id");
-- create index "knowledgerelationship_tenant_i_c2e180b6bf727a089ab234a0504ce8ba" to table: "knowledge_relationships"
CREATE UNIQUE INDEX "knowledgerelationship_tenant_i_c2e180b6bf727a089ab234a0504ce8ba" ON "knowledge_relationships" ("tenant_id", "kind", "source_entity_id", "target_entity_id");
-- create index "knowledgerelationship_tenant_id_kind" to table: "knowledge_relationships"
CREATE INDEX "knowledgerelationship_tenant_id_kind" ON "knowledge_relationships" ("tenant_id", "kind");
-- create index "knowledgerelationship_tenant_id_source_entity_id" to table: "knowledge_relationships"
CREATE INDEX "knowledgerelationship_tenant_id_source_entity_id" ON "knowledge_relationships" ("tenant_id", "source_entity_id");
-- create index "knowledgerelationship_tenant_id_target_entity_id" to table: "knowledge_relationships"
CREATE INDEX "knowledgerelationship_tenant_id_target_entity_id" ON "knowledge_relationships" ("tenant_id", "target_entity_id");
-- create index "knowledgerelationship_tenant_id_updated_at" to table: "knowledge_relationships"
CREATE INDEX "knowledgerelationship_tenant_id_updated_at" ON "knowledge_relationships" ("tenant_id", "updated_at");
-- create index "knowledgerelationship_tenant_id_kind_last_observed_at" to table: "knowledge_relationships"
CREATE INDEX "knowledgerelationship_tenant_id_kind_last_observed_at" ON "knowledge_relationships" ("tenant_id", "kind", "last_observed_at");
-- create index "knowledgerelationship_tenant_id_kind_deleted_at" to table: "knowledge_relationships"
CREATE INDEX "knowledgerelationship_tenant_id_kind_deleted_at" ON "knowledge_relationships" ("tenant_id", "kind", "deleted_at");
-- create "meeting_schedules" table
CREATE TABLE "meeting_schedules" ("id" uuid NOT NULL, "archive_time" timestamptz NULL, "name" character varying NOT NULL, "description" character varying NULL, "begin_minute" bigint NOT NULL, "duration_minutes" bigint NOT NULL, "start_date" timestamptz NOT NULL, "repeats" character varying NOT NULL, "repetition_step" bigint NOT NULL DEFAULT 1, "week_days" jsonb NULL, "monthly_on" character varying NULL, "until_date" timestamptz NULL, "num_repetitions" bigint NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "meetingschedule_tenant_id" to table: "meeting_schedules"
CREATE INDEX "meetingschedule_tenant_id" ON "meeting_schedules" ("tenant_id");
-- create "meeting_sessions" table
CREATE TABLE "meeting_sessions" ("id" uuid NOT NULL, "title" character varying NOT NULL, "started_at" timestamptz NOT NULL, "ended_at" timestamptz NULL, "document_name" character varying NOT NULL, "tenant_id" bigint NOT NULL, "meeting_session_schedule" uuid NULL, PRIMARY KEY ("id"));
-- create index "meetingsession_tenant_id" to table: "meeting_sessions"
CREATE INDEX "meetingsession_tenant_id" ON "meeting_sessions" ("tenant_id");
-- create "normalized_events" table
CREATE TABLE "normalized_events" ("id" uuid NOT NULL, "activity_kind" character varying NOT NULL, "provider" character varying NOT NULL, "provider_source" character varying NOT NULL, "provider_event_ref" character varying NOT NULL, "provider_subject_ref" character varying NOT NULL, "subject_kind" character varying NOT NULL, "attributes" jsonb NOT NULL, "created_at" timestamptz NOT NULL, "occurred_at" timestamptz NOT NULL, "received_at" timestamptz NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "normalizedevent_tenant_id" to table: "normalized_events"
CREATE INDEX "normalizedevent_tenant_id" ON "normalized_events" ("tenant_id");
-- create index "normalizedevent_tenant_id_prov_2fbcf05a5722a73691feb72471c5e433" to table: "normalized_events"
CREATE UNIQUE INDEX "normalizedevent_tenant_id_prov_2fbcf05a5722a73691feb72471c5e433" ON "normalized_events" ("tenant_id", "provider", "provider_source", "provider_event_ref", "provider_subject_ref");
-- create index "normalizedevent_tenant_id_provider_provider_source_occurred_at" to table: "normalized_events"
CREATE INDEX "normalizedevent_tenant_id_provider_provider_source_occurred_at" ON "normalized_events" ("tenant_id", "provider", "provider_source", "occurred_at");
-- create "normalized_event_projection_status" table
CREATE TABLE "normalized_event_projection_status" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "handler_name" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'pending', "last_error" character varying NULL, "last_attempted_at" timestamptz NULL, "succeeded_at" timestamptz NULL, "failed_at" timestamptz NULL, "tenant_id" bigint NOT NULL, "normalized_event_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "normalizedeventprojectionstatus_tenant_id" to table: "normalized_event_projection_status"
CREATE INDEX "normalizedeventprojectionstatus_tenant_id" ON "normalized_event_projection_status" ("tenant_id");
-- create index "normalizedeventprojectionstatu_26223016baedaca5556963f76f513be8" to table: "normalized_event_projection_status"
CREATE UNIQUE INDEX "normalizedeventprojectionstatu_26223016baedaca5556963f76f513be8" ON "normalized_event_projection_status" ("tenant_id", "normalized_event_id", "handler_name");
-- create index "normalizedeventprojectionstatus_tenant_id_status_updated_at" to table: "normalized_event_projection_status"
CREATE INDEX "normalizedeventprojectionstatus_tenant_id_status_updated_at" ON "normalized_event_projection_status" ("tenant_id", "status", "updated_at");
-- create "oncall_handover_templates" table
CREATE TABLE "oncall_handover_templates" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "contents" bytea NOT NULL, "is_default" boolean NOT NULL DEFAULT false, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "oncallhandovertemplate_tenant_id" to table: "oncall_handover_templates"
CREATE INDEX "oncallhandovertemplate_tenant_id" ON "oncall_handover_templates" ("tenant_id");
-- create "oncall_rosters" table
CREATE TABLE "oncall_rosters" ("id" uuid NOT NULL, "archive_time" timestamptz NULL, "name" character varying NOT NULL, "slug" character varying NOT NULL, "timezone" character varying NULL, "chat_handle" character varying NULL, "chat_channel_id" character varying NULL, "handover_template_id" uuid NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "oncall_rosters_slug_key" to table: "oncall_rosters"
CREATE UNIQUE INDEX "oncall_rosters_slug_key" ON "oncall_rosters" ("slug");
-- create index "oncallroster_tenant_id" to table: "oncall_rosters"
CREATE INDEX "oncallroster_tenant_id" ON "oncall_rosters" ("tenant_id");
-- create "oncall_roster_metrics" table
CREATE TABLE "oncall_roster_metrics" ("id" uuid NOT NULL, "tenant_id" bigint NOT NULL, "roster_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "oncallrostermetrics_tenant_id" to table: "oncall_roster_metrics"
CREATE INDEX "oncallrostermetrics_tenant_id" ON "oncall_roster_metrics" ("tenant_id");
-- create "oncall_schedules" table
CREATE TABLE "oncall_schedules" ("id" uuid NOT NULL, "archive_time" timestamptz NULL, "name" character varying NOT NULL, "timezone" character varying NULL, "roster_id" uuid NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "oncallschedule_tenant_id" to table: "oncall_schedules"
CREATE INDEX "oncallschedule_tenant_id" ON "oncall_schedules" ("tenant_id");
-- create "oncall_schedule_participants" table
CREATE TABLE "oncall_schedule_participants" ("id" uuid NOT NULL, "index" bigint NOT NULL, "schedule_id" uuid NOT NULL, "tenant_id" bigint NOT NULL, "user_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "oncallscheduleparticipant_tenant_id" to table: "oncall_schedule_participants"
CREATE INDEX "oncallscheduleparticipant_tenant_id" ON "oncall_schedule_participants" ("tenant_id");
-- create "oncall_shifts" table
CREATE TABLE "oncall_shifts" ("id" uuid NOT NULL, "role" character varying NULL DEFAULT 'primary', "start_at" timestamptz NOT NULL, "end_at" timestamptz NOT NULL, "tenant_id" bigint NOT NULL, "user_id" uuid NOT NULL, "roster_id" uuid NOT NULL, "primary_shift_id" uuid NULL, PRIMARY KEY ("id"), CONSTRAINT "oncall_shifts_oncall_shifts_primary_shift" FOREIGN KEY ("primary_shift_id") REFERENCES "oncall_shifts" ("id") ON DELETE SET NULL);
-- create index "oncall_shifts_primary_shift_id_key" to table: "oncall_shifts"
CREATE UNIQUE INDEX "oncall_shifts_primary_shift_id_key" ON "oncall_shifts" ("primary_shift_id");
-- create index "oncallshift_tenant_id" to table: "oncall_shifts"
CREATE INDEX "oncallshift_tenant_id" ON "oncall_shifts" ("tenant_id");
-- create "oncall_shift_handovers" table
CREATE TABLE "oncall_shift_handovers" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "reminder_sent" boolean NOT NULL DEFAULT false, "updated_at" timestamptz NOT NULL, "sent_at" timestamptz NULL, "contents" bytea NOT NULL, "shift_id" uuid NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "oncall_shift_handovers_shift_id_key" to table: "oncall_shift_handovers"
CREATE UNIQUE INDEX "oncall_shift_handovers_shift_id_key" ON "oncall_shift_handovers" ("shift_id");
-- create index "oncallshifthandover_tenant_id" to table: "oncall_shift_handovers"
CREATE INDEX "oncallshifthandover_tenant_id" ON "oncall_shift_handovers" ("tenant_id");
-- create "oncall_shift_metrics" table
CREATE TABLE "oncall_shift_metrics" ("id" uuid NOT NULL, "updated_at" timestamptz NOT NULL, "burden_score" real NOT NULL, "event_frequency" real NOT NULL, "life_impact" real NOT NULL, "time_impact" real NOT NULL, "response_requirements" real NOT NULL, "isolation" real NOT NULL, "incidents_total" real NOT NULL, "incident_response_time" real NOT NULL, "events_total" real NOT NULL, "alerts_total" real NOT NULL, "interrupts_total" real NOT NULL, "interrupts_night" real NOT NULL, "interrupts_business_hours" real NOT NULL, "shift_id" uuid NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "oncall_shift_metrics_shift_id_key" to table: "oncall_shift_metrics"
CREATE UNIQUE INDEX "oncall_shift_metrics_shift_id_key" ON "oncall_shift_metrics" ("shift_id");
-- create index "oncallshiftmetrics_tenant_id" to table: "oncall_shift_metrics"
CREATE INDEX "oncallshiftmetrics_tenant_id" ON "oncall_shift_metrics" ("tenant_id");
-- create "organizations" table
CREATE TABLE "organizations" ("id" uuid NOT NULL, "auth_provider_id" character varying NOT NULL, "name" character varying NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "organization_tenant_id" to table: "organizations"
CREATE INDEX "organization_tenant_id" ON "organizations" ("tenant_id");
-- create index "organization_tenant_id_auth_provider_id" to table: "organizations"
CREATE UNIQUE INDEX "organization_tenant_id_auth_provider_id" ON "organizations" ("tenant_id", "auth_provider_id");
-- create "organization_preferences" table
CREATE TABLE "organization_preferences" ("id" uuid NOT NULL, "initial_setup_at" timestamptz NULL, "enable_incident_management" boolean NOT NULL DEFAULT false, "organization_id" uuid NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "organization_preferences_organization_id_key" to table: "organization_preferences"
CREATE UNIQUE INDEX "organization_preferences_organization_id_key" ON "organization_preferences" ("organization_id");
-- create index "organizationpreferences_tenant_id" to table: "organization_preferences"
CREATE INDEX "organizationpreferences_tenant_id" ON "organization_preferences" ("tenant_id");
-- create "organization_roles" table
CREATE TABLE "organization_roles" ("id" uuid NOT NULL, "role" character varying NOT NULL, "tenant_id" bigint NOT NULL, "organization_id" uuid NOT NULL, "user_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "organization_roles_user_id_key" to table: "organization_roles"
CREATE UNIQUE INDEX "organization_roles_user_id_key" ON "organization_roles" ("user_id");
-- create index "organizationrole_tenant_id" to table: "organization_roles"
CREATE INDEX "organizationrole_tenant_id" ON "organization_roles" ("tenant_id");
-- create index "organizationrole_tenant_id_organization_id_user_id" to table: "organization_roles"
CREATE UNIQUE INDEX "organizationrole_tenant_id_organization_id_user_id" ON "organization_roles" ("tenant_id", "organization_id", "user_id");
-- create "playbooks" table
CREATE TABLE "playbooks" ("id" uuid NOT NULL, "title" character varying NOT NULL, "content" bytea NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "playbook_tenant_id" to table: "playbooks"
CREATE INDEX "playbook_tenant_id" ON "playbooks" ("tenant_id");
-- create "retrospectives" table
CREATE TABLE "retrospectives" ("id" uuid NOT NULL, "kind" character varying NOT NULL, "state" character varying NOT NULL, "document_id" uuid NOT NULL, "incident_id" uuid NOT NULL, "tenant_id" bigint NOT NULL, "system_analysis_id" uuid NULL, PRIMARY KEY ("id"));
-- create index "retrospectives_document_id_key" to table: "retrospectives"
CREATE UNIQUE INDEX "retrospectives_document_id_key" ON "retrospectives" ("document_id");
-- create index "retrospectives_incident_id_key" to table: "retrospectives"
CREATE UNIQUE INDEX "retrospectives_incident_id_key" ON "retrospectives" ("incident_id");
-- create index "retrospectives_system_analysis_id_key" to table: "retrospectives"
CREATE UNIQUE INDEX "retrospectives_system_analysis_id_key" ON "retrospectives" ("system_analysis_id");
-- create index "retrospective_tenant_id" to table: "retrospectives"
CREATE INDEX "retrospective_tenant_id" ON "retrospectives" ("tenant_id");
-- create "retrospective_comments" table
CREATE TABLE "retrospective_comments" ("id" uuid NOT NULL, "content" bytea NOT NULL, "tenant_id" bigint NOT NULL, "retrospective_id" uuid NOT NULL, "user_id" uuid NOT NULL, "retrospective_review_id" uuid NULL, "parent_reply_id" uuid NULL, PRIMARY KEY ("id"), CONSTRAINT "retrospective_comments_retrospective_comments_replies" FOREIGN KEY ("parent_reply_id") REFERENCES "retrospective_comments" ("id") ON DELETE SET NULL);
-- create index "retrospectivecomment_tenant_id" to table: "retrospective_comments"
CREATE INDEX "retrospectivecomment_tenant_id" ON "retrospective_comments" ("tenant_id");
-- create "retrospective_reviews" table
CREATE TABLE "retrospective_reviews" ("id" uuid NOT NULL, "state" character varying NOT NULL, "tenant_id" bigint NOT NULL, "retrospective_id" uuid NOT NULL, "requester_id" uuid NOT NULL, "reviewer_id" uuid NOT NULL, "comment_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "retrospectivereview_tenant_id" to table: "retrospective_reviews"
CREATE INDEX "retrospectivereview_tenant_id" ON "retrospective_reviews" ("tenant_id");
-- create "system_analyses" table
CREATE TABLE "system_analyses" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "tenant_id" bigint NOT NULL, "topology_snapshot_id" uuid NULL, PRIMARY KEY ("id"));
-- create index "systemanalysis_tenant_id" to table: "system_analyses"
CREATE INDEX "systemanalysis_tenant_id" ON "system_analyses" ("tenant_id");
-- create "system_analysis_topology_edges" table
CREATE TABLE "system_analysis_topology_edges" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "description" text NULL, "tenant_id" bigint NOT NULL, "analysis_id" uuid NOT NULL, "snapshot_relationship_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "systemanalysistopologyedge_tenant_id" to table: "system_analysis_topology_edges"
CREATE INDEX "systemanalysistopologyedge_tenant_id" ON "system_analysis_topology_edges" ("tenant_id");
-- create "system_analysis_topology_nodes" table
CREATE TABLE "system_analysis_topology_nodes" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "description" text NULL, "pos_x" double precision NOT NULL DEFAULT 0, "pos_y" double precision NOT NULL DEFAULT 0, "tenant_id" bigint NOT NULL, "analysis_id" uuid NOT NULL, "snapshot_entity_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "systemanalysistopologynode_tenant_id" to table: "system_analysis_topology_nodes"
CREATE INDEX "systemanalysistopologynode_tenant_id" ON "system_analysis_topology_nodes" ("tenant_id");
-- create "system_topology_snapshots" table
CREATE TABLE "system_topology_snapshots" ("id" uuid NOT NULL, "as_of" timestamptz NOT NULL, "name" character varying NULL, "scope" character varying NOT NULL DEFAULT 'all', "scope_properties" jsonb NULL, "created_at" timestamptz NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "systemtopologysnapshot_tenant_id" to table: "system_topology_snapshots"
CREATE INDEX "systemtopologysnapshot_tenant_id" ON "system_topology_snapshots" ("tenant_id");
-- create index "systemtopologysnapshot_tenant_id_as_of" to table: "system_topology_snapshots"
CREATE INDEX "systemtopologysnapshot_tenant_id_as_of" ON "system_topology_snapshots" ("tenant_id", "as_of");
-- create index "systemtopologysnapshot_tenant_id_created_at" to table: "system_topology_snapshots"
CREATE INDEX "systemtopologysnapshot_tenant_id_created_at" ON "system_topology_snapshots" ("tenant_id", "created_at");
-- create "system_topology_snapshot_entities" table
CREATE TABLE "system_topology_snapshot_entities" ("id" uuid NOT NULL, "entity_kind" character varying NOT NULL, "display_name" character varying NOT NULL, "description" text NULL, "properties" jsonb NULL, "aliases" jsonb NULL, "created_at" timestamptz NOT NULL, "tenant_id" bigint NOT NULL, "snapshot_id" uuid NOT NULL, "knowledge_entity_id" uuid NULL, PRIMARY KEY ("id"));
-- create index "systemtopologysnapshotentity_tenant_id" to table: "system_topology_snapshot_entities"
CREATE INDEX "systemtopologysnapshotentity_tenant_id" ON "system_topology_snapshot_entities" ("tenant_id");
-- create index "systemtopologysnapshotentity_tenant_id_snapshot_id" to table: "system_topology_snapshot_entities"
CREATE INDEX "systemtopologysnapshotentity_tenant_id_snapshot_id" ON "system_topology_snapshot_entities" ("tenant_id", "snapshot_id");
-- create index "systemtopologysnapshotentity_tenant_id_knowledge_entity_id" to table: "system_topology_snapshot_entities"
CREATE INDEX "systemtopologysnapshotentity_tenant_id_knowledge_entity_id" ON "system_topology_snapshot_entities" ("tenant_id", "knowledge_entity_id");
-- create index "systemtopologysnapshotentity_t_f5c4f8cfa84671bf6a92d9f49c2f214d" to table: "system_topology_snapshot_entities"
CREATE UNIQUE INDEX "systemtopologysnapshotentity_t_f5c4f8cfa84671bf6a92d9f49c2f214d" ON "system_topology_snapshot_entities" ("tenant_id", "snapshot_id", "knowledge_entity_id");
-- create "system_topology_snapshot_relationships" table
CREATE TABLE "system_topology_snapshot_relationships" ("id" uuid NOT NULL, "relationship_kind" character varying NOT NULL, "display_name" character varying NULL, "description" text NULL, "properties" jsonb NULL, "created_at" timestamptz NOT NULL, "tenant_id" bigint NOT NULL, "knowledge_relationship_id" uuid NULL, "snapshot_id" uuid NOT NULL, "source_snapshot_entity_id" uuid NOT NULL, "target_snapshot_entity_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "systemtopologysnapshotrelationship_tenant_id" to table: "system_topology_snapshot_relationships"
CREATE INDEX "systemtopologysnapshotrelationship_tenant_id" ON "system_topology_snapshot_relationships" ("tenant_id");
-- create index "systemtopologysnapshotrelationship_tenant_id_snapshot_id" to table: "system_topology_snapshot_relationships"
CREATE INDEX "systemtopologysnapshotrelationship_tenant_id_snapshot_id" ON "system_topology_snapshot_relationships" ("tenant_id", "snapshot_id");
-- create index "systemtopologysnapshotrelation_a06f34ed83db45fd13d05244cf667720" to table: "system_topology_snapshot_relationships"
CREATE INDEX "systemtopologysnapshotrelation_a06f34ed83db45fd13d05244cf667720" ON "system_topology_snapshot_relationships" ("tenant_id", "knowledge_relationship_id");
-- create index "systemtopologysnapshotrelation_ce56930e9b9fd6d129909352010bc447" to table: "system_topology_snapshot_relationships"
CREATE INDEX "systemtopologysnapshotrelation_ce56930e9b9fd6d129909352010bc447" ON "system_topology_snapshot_relationships" ("tenant_id", "source_snapshot_entity_id");
-- create index "systemtopologysnapshotrelation_8bfb42b7b88bb99b0dab90f97e7c8e86" to table: "system_topology_snapshot_relationships"
CREATE INDEX "systemtopologysnapshotrelation_8bfb42b7b88bb99b0dab90f97e7c8e86" ON "system_topology_snapshot_relationships" ("tenant_id", "target_snapshot_entity_id");
-- create index "systemtopologysnapshotrelation_96315e4d75ceda9b93720e1ad418db13" to table: "system_topology_snapshot_relationships"
CREATE UNIQUE INDEX "systemtopologysnapshotrelation_96315e4d75ceda9b93720e1ad418db13" ON "system_topology_snapshot_relationships" ("tenant_id", "snapshot_id", "knowledge_relationship_id");
-- create "tasks" table
CREATE TABLE "tasks" ("id" uuid NOT NULL, "type" character varying NOT NULL, "title" character varying NOT NULL, "incident_id" uuid NULL, "tenant_id" bigint NOT NULL, "assignee_id" uuid NULL, "creator_id" uuid NULL, PRIMARY KEY ("id"));
-- create index "task_tenant_id" to table: "tasks"
CREATE INDEX "task_tenant_id" ON "tasks" ("tenant_id");
-- create "teams" table
CREATE TABLE "teams" ("id" uuid NOT NULL, "slug" character varying NOT NULL, "name" character varying NOT NULL, "chat_channel_id" character varying NULL, "timezone" character varying NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "teams_slug_key" to table: "teams"
CREATE UNIQUE INDEX "teams_slug_key" ON "teams" ("slug");
-- create index "team_tenant_id" to table: "teams"
CREATE INDEX "team_tenant_id" ON "teams" ("tenant_id");
-- create "team_memberships" table
CREATE TABLE "team_memberships" ("id" uuid NOT NULL, "role" character varying NOT NULL DEFAULT 'member', "tenant_id" bigint NOT NULL, "team_id" uuid NOT NULL, "user_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "teammembership_tenant_id" to table: "team_memberships"
CREATE INDEX "teammembership_tenant_id" ON "team_memberships" ("tenant_id");
-- create index "teammembership_team_id_user_id" to table: "team_memberships"
CREATE UNIQUE INDEX "teammembership_team_id_user_id" ON "team_memberships" ("team_id", "user_id");
-- create "tenants" table
CREATE TABLE "tenants" ("id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY, PRIMARY KEY ("id"));
-- create "tickets" table
CREATE TABLE "tickets" ("id" uuid NOT NULL, "title" character varying NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "ticket_tenant_id" to table: "tickets"
CREATE INDEX "ticket_tenant_id" ON "tickets" ("tenant_id");
-- create "users" table
CREATE TABLE "users" ("id" uuid NOT NULL, "email" character varying NOT NULL, "name" character varying NOT NULL DEFAULT '', "chat_id" character varying NULL, "timezone" character varying NULL, "auth_provider_id" character varying NULL, "tenant_id" bigint NOT NULL, "knowledge_entity_id" uuid NULL, PRIMARY KEY ("id"));
-- create index "user_tenant_id" to table: "users"
CREATE INDEX "user_tenant_id" ON "users" ("tenant_id");
-- create index "user_tenant_id_knowledge_entity_id" to table: "users"
CREATE UNIQUE INDEX "user_tenant_id_knowledge_entity_id" ON "users" ("tenant_id", "knowledge_entity_id");
-- create index "user_auth_provider_id" to table: "users"
CREATE UNIQUE INDEX "user_auth_provider_id" ON "users" ("auth_provider_id");
-- create index "user_tenant_id_email" to table: "users"
CREATE UNIQUE INDEX "user_tenant_id_email" ON "users" ("tenant_id", "email");
-- create "user_auth_sessions" table
CREATE TABLE "user_auth_sessions" ("id" uuid NOT NULL, "expires_at" timestamptz NOT NULL, "scopes" jsonb NULL, "tenant_id" bigint NOT NULL, "user_id" uuid NOT NULL, "organization_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "userauthsession_tenant_id" to table: "user_auth_sessions"
CREATE INDEX "userauthsession_tenant_id" ON "user_auth_sessions" ("tenant_id");
-- create "video_conferences" table
CREATE TABLE "video_conferences" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "provider" character varying NOT NULL, "external_id" character varying NULL, "join_url" character varying NOT NULL, "host_url" character varying NULL, "dial_in" character varying NULL, "passcode" character varying NULL, "status" character varying NOT NULL DEFAULT 'creating', "metadata" jsonb NULL, "created_by_integration" character varying NULL, "incident_id" uuid NULL, "meeting_session_id" uuid NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "video_conferences_meeting_session_id_key" to table: "video_conferences"
CREATE UNIQUE INDEX "video_conferences_meeting_session_id_key" ON "video_conferences" ("meeting_session_id");
-- create index "videoconference_tenant_id" to table: "video_conferences"
CREATE INDEX "videoconference_tenant_id" ON "video_conferences" ("tenant_id");
-- create index "videoconference_incident_id_status" to table: "video_conferences"
CREATE INDEX "videoconference_incident_id_status" ON "video_conferences" ("incident_id", "status");
-- create index "videoconference_meeting_session_id_status" to table: "video_conferences"
CREATE INDEX "videoconference_meeting_session_id_status" ON "video_conferences" ("meeting_session_id", "status");
-- create "incident_field_selections" table
CREATE TABLE "incident_field_selections" ("incident_id" uuid NOT NULL, "incident_field_option_id" uuid NOT NULL, PRIMARY KEY ("incident_id", "incident_field_option_id"));
-- create "incident_tag_assignments" table
CREATE TABLE "incident_tag_assignments" ("incident_id" uuid NOT NULL, "incident_tag_id" uuid NOT NULL, PRIMARY KEY ("incident_id", "incident_tag_id"));
-- create "incident_review_sessions" table
CREATE TABLE "incident_review_sessions" ("incident_id" uuid NOT NULL, "meeting_session_id" uuid NOT NULL, PRIMARY KEY ("incident_id", "meeting_session_id"));
-- create "incident_debrief_question_incident_fields" table
CREATE TABLE "incident_debrief_question_incident_fields" ("incident_debrief_question_id" uuid NOT NULL, "incident_field_id" uuid NOT NULL, PRIMARY KEY ("incident_debrief_question_id", "incident_field_id"));
-- create "incident_debrief_question_incident_roles" table
CREATE TABLE "incident_debrief_question_incident_roles" ("incident_debrief_question_id" uuid NOT NULL, "incident_role_id" uuid NOT NULL, PRIMARY KEY ("incident_debrief_question_id", "incident_role_id"));
-- create "incident_debrief_question_incident_severities" table
CREATE TABLE "incident_debrief_question_incident_severities" ("incident_debrief_question_id" uuid NOT NULL, "incident_severity_id" uuid NOT NULL, PRIMARY KEY ("incident_debrief_question_id", "incident_severity_id"));
-- create "incident_debrief_question_incident_tags" table
CREATE TABLE "incident_debrief_question_incident_tags" ("incident_debrief_question_id" uuid NOT NULL, "incident_tag_id" uuid NOT NULL, PRIMARY KEY ("incident_debrief_question_id", "incident_tag_id"));
-- create "incident_debrief_question_incident_types" table
CREATE TABLE "incident_debrief_question_incident_types" ("incident_debrief_question_id" uuid NOT NULL, "incident_type_id" uuid NOT NULL, PRIMARY KEY ("incident_debrief_question_id", "incident_type_id"));
-- create "meeting_schedule_owning_team" table
CREATE TABLE "meeting_schedule_owning_team" ("meeting_schedule_id" uuid NOT NULL, "team_id" uuid NOT NULL, PRIMARY KEY ("meeting_schedule_id", "team_id"));
-- create "oncall_shift_handover_pinned_annotations" table
CREATE TABLE "oncall_shift_handover_pinned_annotations" ("oncall_shift_handover_id" uuid NOT NULL, "event_annotation_id" uuid NOT NULL, PRIMARY KEY ("oncall_shift_handover_id", "event_annotation_id"));
-- create "playbook_alerts" table
CREATE TABLE "playbook_alerts" ("playbook_id" uuid NOT NULL, "alert_id" uuid NOT NULL, PRIMARY KEY ("playbook_id", "alert_id"));
-- create "task_tickets" table
CREATE TABLE "task_tickets" ("task_id" uuid NOT NULL, "ticket_id" uuid NOT NULL, PRIMARY KEY ("task_id", "ticket_id"));
-- create "team_oncall_rosters" table
CREATE TABLE "team_oncall_rosters" ("team_id" uuid NOT NULL, "oncall_roster_id" uuid NOT NULL, PRIMARY KEY ("team_id", "oncall_roster_id"));
-- create "user_watched_oncall_rosters" table
CREATE TABLE "user_watched_oncall_rosters" ("user_id" uuid NOT NULL, "oncall_roster_id" uuid NOT NULL, PRIMARY KEY ("user_id", "oncall_roster_id"));
-- modify "agent_runs" table
ALTER TABLE "agent_runs" ADD CONSTRAINT "agent_runs_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "agent_runs_agent_tasks_agent_task" FOREIGN KEY ("agent_task_id") REFERENCES "agent_tasks" ("id") ON DELETE NO ACTION;
-- modify "agent_run_citations" table
ALTER TABLE "agent_run_citations" ADD CONSTRAINT "agent_run_citations_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "agent_run_citations_agent_runs_agent_run" FOREIGN KEY ("agent_run_id") REFERENCES "agent_runs" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "agent_run_citations_knowledge_entities_knowledge_entity" FOREIGN KEY ("knowledge_entity_id") REFERENCES "knowledge_entities" ("id") ON DELETE SET NULL, ADD CONSTRAINT "agent_run_citations_knowledge__2174aff27b8bc3fd37438d763b4bd03b" FOREIGN KEY ("knowledge_relationship_id") REFERENCES "knowledge_relationships" ("id") ON DELETE SET NULL, ADD CONSTRAINT "agent_run_citations_knowledge_evidences_knowledge_evidence" FOREIGN KEY ("knowledge_evidence_id") REFERENCES "knowledge_evidences" ("id") ON DELETE SET NULL, ADD CONSTRAINT "agent_run_citations_agent_tasks_agent_task" FOREIGN KEY ("agent_task_id") REFERENCES "agent_tasks" ("id") ON DELETE SET NULL, ADD CONSTRAINT "agent_run_citations_agent_run_tool_calls_agent_run_tool_call" FOREIGN KEY ("agent_run_tool_call_id") REFERENCES "agent_run_tool_calls" ("id") ON DELETE SET NULL;
-- modify "agent_run_findings" table
ALTER TABLE "agent_run_findings" ADD CONSTRAINT "agent_run_findings_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "agent_run_findings_agent_runs_agent_run" FOREIGN KEY ("agent_run_id") REFERENCES "agent_runs" ("id") ON DELETE NO ACTION;
-- modify "agent_run_finding_citations" table
ALTER TABLE "agent_run_finding_citations" ADD CONSTRAINT "agent_run_finding_citations_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "agent_run_finding_citations_ag_9a3b43a8d817a43c8de87f621827096a" FOREIGN KEY ("agent_run_finding_id") REFERENCES "agent_run_findings" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "agent_run_finding_citations_ag_012686cc6fb703a2437581a00eaa6ea0" FOREIGN KEY ("agent_run_citation_id") REFERENCES "agent_run_citations" ("id") ON DELETE NO ACTION;
-- modify "agent_run_results" table
ALTER TABLE "agent_run_results" ADD CONSTRAINT "agent_run_results_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "agent_run_results_agent_runs_agent_run" FOREIGN KEY ("agent_run_id") REFERENCES "agent_runs" ("id") ON DELETE NO ACTION;
-- modify "agent_run_tool_calls" table
ALTER TABLE "agent_run_tool_calls" ADD CONSTRAINT "agent_run_tool_calls_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "agent_run_tool_calls_agent_runs_agent_run" FOREIGN KEY ("agent_run_id") REFERENCES "agent_runs" ("id") ON DELETE NO ACTION;
-- modify "agent_tasks" table
ALTER TABLE "agent_tasks" ADD CONSTRAINT "agent_tasks_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "agent_tasks_users_owner_user" FOREIGN KEY ("owner_user_id") REFERENCES "users" ("id") ON DELETE NO ACTION;
-- modify "alerts" table
ALTER TABLE "alerts" ADD CONSTRAINT "alerts_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "alerts_knowledge_entities_knowledge_entity" FOREIGN KEY ("knowledge_entity_id") REFERENCES "knowledge_entities" ("id") ON DELETE SET NULL, ADD CONSTRAINT "alerts_oncall_rosters_alerts" FOREIGN KEY ("roster_id") REFERENCES "oncall_rosters" ("id") ON DELETE SET NULL;
-- modify "alert_feedbacks" table
ALTER TABLE "alert_feedbacks" ADD CONSTRAINT "alert_feedbacks_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "alert_feedbacks_alerts_alert" FOREIGN KEY ("alert_id") REFERENCES "alerts" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "alert_feedbacks_normalized_events_alert_instance" FOREIGN KEY ("alert_instance_id") REFERENCES "normalized_events" ("id") ON DELETE SET NULL, ADD CONSTRAINT "alert_feedbacks_normalized_events_alert_feedback" FOREIGN KEY ("normalized_event_alert_feedback") REFERENCES "normalized_events" ("id") ON DELETE SET NULL;
-- modify "documents" table
ALTER TABLE "documents" ADD CONSTRAINT "documents_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "document_accesses" table
ALTER TABLE "document_accesses" ADD CONSTRAINT "document_accesses_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "document_accesses_documents_document" FOREIGN KEY ("document_id") REFERENCES "documents" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "document_accesses_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE SET NULL, ADD CONSTRAINT "document_accesses_teams_team" FOREIGN KEY ("team_id") REFERENCES "teams" ("id") ON DELETE SET NULL;
-- modify "event_annotations" table
ALTER TABLE "event_annotations" ADD CONSTRAINT "event_annotations_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "event_annotations_normalized_events_event" FOREIGN KEY ("event_id") REFERENCES "normalized_events" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "event_annotations_users_creator" FOREIGN KEY ("creator_id") REFERENCES "users" ("id") ON DELETE NO ACTION;
-- modify "incidents" table
ALTER TABLE "incidents" ADD CONSTRAINT "incidents_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "incidents_knowledge_entities_knowledge_entity" FOREIGN KEY ("knowledge_entity_id") REFERENCES "knowledge_entities" ("id") ON DELETE SET NULL, ADD CONSTRAINT "incidents_incident_severities_severity" FOREIGN KEY ("severity_id") REFERENCES "incident_severities" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "incidents_incident_types_type" FOREIGN KEY ("type_id") REFERENCES "incident_types" ("id") ON DELETE NO ACTION;
-- modify "incident_debriefs" table
ALTER TABLE "incident_debriefs" ADD CONSTRAINT "incident_debriefs_incidents_debriefs" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "incident_debriefs_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "incident_debriefs_users_incident_debriefs" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE NO ACTION;
-- modify "incident_debrief_messages" table
ALTER TABLE "incident_debrief_messages" ADD CONSTRAINT "incident_debrief_messages_incident_debriefs_messages" FOREIGN KEY ("debrief_id") REFERENCES "incident_debriefs" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "incident_debrief_messages_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "incident_debrief_messages_inci_0d1f0b105ef851edb04b442b34a0d17f" FOREIGN KEY ("question_id") REFERENCES "incident_debrief_questions" ("id") ON DELETE SET NULL;
-- modify "incident_debrief_questions" table
ALTER TABLE "incident_debrief_questions" ADD CONSTRAINT "incident_debrief_questions_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "incident_debrief_suggestions" table
ALTER TABLE "incident_debrief_suggestions" ADD CONSTRAINT "incident_debrief_suggestions_incident_debriefs_suggestions" FOREIGN KEY ("incident_debrief_suggestions") REFERENCES "incident_debriefs" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "incident_debrief_suggestions_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "incident_fields" table
ALTER TABLE "incident_fields" ADD CONSTRAINT "incident_fields_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "incident_field_options" table
ALTER TABLE "incident_field_options" ADD CONSTRAINT "incident_field_options_incident_fields_options" FOREIGN KEY ("incident_field_id") REFERENCES "incident_fields" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "incident_field_options_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "incident_impacts" table
ALTER TABLE "incident_impacts" ADD CONSTRAINT "incident_impacts_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "incident_impacts_incidents_incident" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "incident_impacts_knowledge_entities_knowledge_entity" FOREIGN KEY ("knowledge_entity_id") REFERENCES "knowledge_entities" ("id") ON DELETE NO ACTION;
-- modify "incident_links" table
ALTER TABLE "incident_links" ADD CONSTRAINT "incident_links_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "incident_links_incidents_incident" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "incident_links_incidents_linked_incident" FOREIGN KEY ("linked_incident_id") REFERENCES "incidents" ("id") ON DELETE NO ACTION;
-- modify "incident_milestones" table
ALTER TABLE "incident_milestones" ADD CONSTRAINT "incident_milestones_incidents_milestones" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "incident_milestones_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "incident_milestones_users_incident_milestones" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE NO ACTION;
-- modify "incident_roles" table
ALTER TABLE "incident_roles" ADD CONSTRAINT "incident_roles_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "incident_role_assignments" table
ALTER TABLE "incident_role_assignments" ADD CONSTRAINT "incident_role_assignments_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "incident_role_assignments_incidents_incident" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "incident_role_assignments_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "incident_role_assignments_incident_roles_role" FOREIGN KEY ("role_id") REFERENCES "incident_roles" ("id") ON DELETE NO ACTION;
-- modify "incident_severities" table
ALTER TABLE "incident_severities" ADD CONSTRAINT "incident_severities_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "incident_tags" table
ALTER TABLE "incident_tags" ADD CONSTRAINT "incident_tags_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "incident_timeline_events" table
ALTER TABLE "incident_timeline_events" ADD CONSTRAINT "incident_timeline_events_incidents_timeline_events" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "incident_timeline_events_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "incident_timeline_events_normalized_events_event" FOREIGN KEY ("event_id") REFERENCES "normalized_events" ("id") ON DELETE SET NULL;
-- modify "incident_timeline_event_contexts" table
ALTER TABLE "incident_timeline_event_contexts" ADD CONSTRAINT "incident_timeline_event_contex_5ac24bfc474fb61b9351fb5cba7c87cb" FOREIGN KEY ("incident_timeline_event_context") REFERENCES "incident_timeline_events" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "incident_timeline_event_contexts_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "incident_timeline_event_contributing_factors" table
ALTER TABLE "incident_timeline_event_contributing_factors" ADD CONSTRAINT "incident_timeline_event_contri_0aecb2f20121e2d628c1402580fe5d71" FOREIGN KEY ("incident_timeline_event_factors") REFERENCES "incident_timeline_events" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "incident_timeline_event_contributing_factors_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "incident_timeline_event_evidences" table
ALTER TABLE "incident_timeline_event_evidences" ADD CONSTRAINT "incident_timeline_event_eviden_37786b98ea2184b38a27c223bdf28160" FOREIGN KEY ("incident_timeline_event_evidence") REFERENCES "incident_timeline_events" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "incident_timeline_event_evidences_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "incident_timeline_event_topology_contexts" table
ALTER TABLE "incident_timeline_event_topology_contexts" ADD CONSTRAINT "incident_timeline_event_topology_contexts_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "incident_timeline_event_topolo_9d336da69e411af955f7f9ddff677001" FOREIGN KEY ("incident_event_id") REFERENCES "incident_timeline_events" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "incident_timeline_event_topolo_71fa443670cce4d56bfb9e7041730983" FOREIGN KEY ("knowledge_entity_id") REFERENCES "knowledge_entities" ("id") ON DELETE SET NULL, ADD CONSTRAINT "incident_timeline_event_topolo_c9aecaf9d13cd231fe8f3c66ff0e4671" FOREIGN KEY ("snapshot_entity_id") REFERENCES "system_topology_snapshot_entities" ("id") ON DELETE SET NULL;
-- modify "incident_types" table
ALTER TABLE "incident_types" ADD CONSTRAINT "incident_types_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "integrations" table
ALTER TABLE "integrations" ADD CONSTRAINT "integrations_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "integration_event_sync_cursors" table
ALTER TABLE "integration_event_sync_cursors" ADD CONSTRAINT "integration_event_sync_cursors_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "integration_event_sync_cursors_integrations_integration" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON DELETE NO ACTION;
-- modify "integration_event_sync_runs" table
ALTER TABLE "integration_event_sync_runs" ADD CONSTRAINT "integration_event_sync_runs_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "integration_event_sync_runs_integrations_integration" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON DELETE NO ACTION;
-- modify "integration_user_install_states" table
ALTER TABLE "integration_user_install_states" ADD CONSTRAINT "integration_user_install_states_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "integration_user_install_states_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE NO ACTION;
-- modify "knowledge_entities" table
ALTER TABLE "knowledge_entities" ADD CONSTRAINT "knowledge_entities_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "knowledge_entity_alias" table
ALTER TABLE "knowledge_entity_alias" ADD CONSTRAINT "knowledge_entity_alias_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "knowledge_entity_alias_knowledge_entities_entity" FOREIGN KEY ("entity_id") REFERENCES "knowledge_entities" ("id") ON DELETE NO ACTION;
-- modify "knowledge_evidences" table
ALTER TABLE "knowledge_evidences" ADD CONSTRAINT "knowledge_evidences_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "knowledge_evidences_knowledge_entities_entity" FOREIGN KEY ("entity_id") REFERENCES "knowledge_entities" ("id") ON DELETE SET NULL, ADD CONSTRAINT "knowledge_evidences_knowledge_relationships_relationship" FOREIGN KEY ("relationship_id") REFERENCES "knowledge_relationships" ("id") ON DELETE SET NULL, ADD CONSTRAINT "knowledge_evidences_knowledge_entity_alias_alias" FOREIGN KEY ("alias_id") REFERENCES "knowledge_entity_alias" ("id") ON DELETE SET NULL, ADD CONSTRAINT "knowledge_evidences_normalized_events_event" FOREIGN KEY ("event_id") REFERENCES "normalized_events" ("id") ON DELETE NO ACTION;
-- modify "knowledge_relationships" table
ALTER TABLE "knowledge_relationships" ADD CONSTRAINT "knowledge_relationships_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "knowledge_relationships_knowledge_entities_source_entity" FOREIGN KEY ("source_entity_id") REFERENCES "knowledge_entities" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "knowledge_relationships_knowledge_entities_target_entity" FOREIGN KEY ("target_entity_id") REFERENCES "knowledge_entities" ("id") ON DELETE NO ACTION;
-- modify "meeting_schedules" table
ALTER TABLE "meeting_schedules" ADD CONSTRAINT "meeting_schedules_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "meeting_sessions" table
ALTER TABLE "meeting_sessions" ADD CONSTRAINT "meeting_sessions_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "meeting_sessions_meeting_schedules_schedule" FOREIGN KEY ("meeting_session_schedule") REFERENCES "meeting_schedules" ("id") ON DELETE SET NULL;
-- modify "normalized_events" table
ALTER TABLE "normalized_events" ADD CONSTRAINT "normalized_events_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "normalized_event_projection_status" table
ALTER TABLE "normalized_event_projection_status" ADD CONSTRAINT "normalized_event_projection_status_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "normalized_event_projection_st_57b31f9b9ba804f03db1c8815e863e31" FOREIGN KEY ("normalized_event_id") REFERENCES "normalized_events" ("id") ON DELETE NO ACTION;
-- modify "oncall_handover_templates" table
ALTER TABLE "oncall_handover_templates" ADD CONSTRAINT "oncall_handover_templates_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "oncall_rosters" table
ALTER TABLE "oncall_rosters" ADD CONSTRAINT "oncall_rosters_oncall_handover_templates_roster" FOREIGN KEY ("handover_template_id") REFERENCES "oncall_handover_templates" ("id") ON DELETE SET NULL, ADD CONSTRAINT "oncall_rosters_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "oncall_roster_metrics" table
ALTER TABLE "oncall_roster_metrics" ADD CONSTRAINT "oncall_roster_metrics_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "oncall_roster_metrics_oncall_rosters_roster" FOREIGN KEY ("roster_id") REFERENCES "oncall_rosters" ("id") ON DELETE NO ACTION;
-- modify "oncall_schedules" table
ALTER TABLE "oncall_schedules" ADD CONSTRAINT "oncall_schedules_oncall_rosters_schedules" FOREIGN KEY ("roster_id") REFERENCES "oncall_rosters" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "oncall_schedules_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "oncall_schedule_participants" table
ALTER TABLE "oncall_schedule_participants" ADD CONSTRAINT "oncall_schedule_participants_oncall_schedules_participants" FOREIGN KEY ("schedule_id") REFERENCES "oncall_schedules" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "oncall_schedule_participants_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "oncall_schedule_participants_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE NO ACTION;
-- modify "oncall_shifts" table
ALTER TABLE "oncall_shifts" ADD CONSTRAINT "oncall_shifts_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "oncall_shifts_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "oncall_shifts_oncall_rosters_roster" FOREIGN KEY ("roster_id") REFERENCES "oncall_rosters" ("id") ON DELETE NO ACTION;
-- modify "oncall_shift_handovers" table
ALTER TABLE "oncall_shift_handovers" ADD CONSTRAINT "oncall_shift_handovers_oncall_shifts_handover" FOREIGN KEY ("shift_id") REFERENCES "oncall_shifts" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "oncall_shift_handovers_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "oncall_shift_metrics" table
ALTER TABLE "oncall_shift_metrics" ADD CONSTRAINT "oncall_shift_metrics_oncall_shifts_metrics" FOREIGN KEY ("shift_id") REFERENCES "oncall_shifts" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "oncall_shift_metrics_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "organizations" table
ALTER TABLE "organizations" ADD CONSTRAINT "organizations_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "organization_preferences" table
ALTER TABLE "organization_preferences" ADD CONSTRAINT "organization_preferences_organizations_preferences" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "organization_preferences_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "organization_roles" table
ALTER TABLE "organization_roles" ADD CONSTRAINT "organization_roles_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "organization_roles_organizations_organization" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "organization_roles_users_organization_role" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE NO ACTION;
-- modify "playbooks" table
ALTER TABLE "playbooks" ADD CONSTRAINT "playbooks_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "retrospectives" table
ALTER TABLE "retrospectives" ADD CONSTRAINT "retrospectives_documents_retrospective" FOREIGN KEY ("document_id") REFERENCES "documents" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "retrospectives_incidents_retrospective" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "retrospectives_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "retrospectives_system_analyses_retrospective" FOREIGN KEY ("system_analysis_id") REFERENCES "system_analyses" ("id") ON DELETE SET NULL;
-- modify "retrospective_comments" table
ALTER TABLE "retrospective_comments" ADD CONSTRAINT "retrospective_comments_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "retrospective_comments_retrospectives_retrospective" FOREIGN KEY ("retrospective_id") REFERENCES "retrospectives" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "retrospective_comments_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "retrospective_comments_retrospective_reviews_review" FOREIGN KEY ("retrospective_review_id") REFERENCES "retrospective_reviews" ("id") ON DELETE SET NULL;
-- modify "retrospective_reviews" table
ALTER TABLE "retrospective_reviews" ADD CONSTRAINT "retrospective_reviews_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "retrospective_reviews_retrospectives_retrospective" FOREIGN KEY ("retrospective_id") REFERENCES "retrospectives" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "retrospective_reviews_users_requester" FOREIGN KEY ("requester_id") REFERENCES "users" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "retrospective_reviews_users_reviewer" FOREIGN KEY ("reviewer_id") REFERENCES "users" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "retrospective_reviews_retrospective_comments_comment" FOREIGN KEY ("comment_id") REFERENCES "retrospective_comments" ("id") ON DELETE NO ACTION;
-- modify "system_analyses" table
ALTER TABLE "system_analyses" ADD CONSTRAINT "system_analyses_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "system_analyses_system_topology_snapshots_topology_snapshot" FOREIGN KEY ("topology_snapshot_id") REFERENCES "system_topology_snapshots" ("id") ON DELETE SET NULL;
-- modify "system_analysis_topology_edges" table
ALTER TABLE "system_analysis_topology_edges" ADD CONSTRAINT "system_analysis_topology_edges_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "system_analysis_topology_edges_system_analyses_analysis" FOREIGN KEY ("analysis_id") REFERENCES "system_analyses" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "system_analysis_topology_edges_c4b40b33a79f2054fdf3e995173eda08" FOREIGN KEY ("snapshot_relationship_id") REFERENCES "system_topology_snapshot_relationships" ("id") ON DELETE NO ACTION;
-- modify "system_analysis_topology_nodes" table
ALTER TABLE "system_analysis_topology_nodes" ADD CONSTRAINT "system_analysis_topology_nodes_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "system_analysis_topology_nodes_system_analyses_analysis" FOREIGN KEY ("analysis_id") REFERENCES "system_analyses" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "system_analysis_topology_nodes_1dea46edc4c4b6fa57943e6a9c4a3f02" FOREIGN KEY ("snapshot_entity_id") REFERENCES "system_topology_snapshot_entities" ("id") ON DELETE NO ACTION;
-- modify "system_topology_snapshots" table
ALTER TABLE "system_topology_snapshots" ADD CONSTRAINT "system_topology_snapshots_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "system_topology_snapshot_entities" table
ALTER TABLE "system_topology_snapshot_entities" ADD CONSTRAINT "system_topology_snapshot_entities_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "system_topology_snapshot_entit_624ed73670c571ebba5995e15ccf51fd" FOREIGN KEY ("snapshot_id") REFERENCES "system_topology_snapshots" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "system_topology_snapshot_entit_6acf350be9828a2b8ba4f3166002a8fd" FOREIGN KEY ("knowledge_entity_id") REFERENCES "knowledge_entities" ("id") ON DELETE SET NULL;
-- modify "system_topology_snapshot_relationships" table
ALTER TABLE "system_topology_snapshot_relationships" ADD CONSTRAINT "system_topology_snapshot_relationships_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "system_topology_snapshot_relat_5f2c55c7b6ad86bc36bd1003d1e2579f" FOREIGN KEY ("knowledge_relationship_id") REFERENCES "knowledge_relationships" ("id") ON DELETE SET NULL, ADD CONSTRAINT "system_topology_snapshot_relat_8618d8e705fc800d4b45f7ba21b7151f" FOREIGN KEY ("snapshot_id") REFERENCES "system_topology_snapshots" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "system_topology_snapshot_relat_020fdabf27aca33de3b8244fcbb40d4d" FOREIGN KEY ("source_snapshot_entity_id") REFERENCES "system_topology_snapshot_entities" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "system_topology_snapshot_relat_49bd6a99a61ed2571218b82bfd309a9d" FOREIGN KEY ("target_snapshot_entity_id") REFERENCES "system_topology_snapshot_entities" ("id") ON DELETE NO ACTION;
-- modify "tasks" table
ALTER TABLE "tasks" ADD CONSTRAINT "tasks_incidents_tasks" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON DELETE SET NULL, ADD CONSTRAINT "tasks_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "tasks_users_assigned_tasks" FOREIGN KEY ("assignee_id") REFERENCES "users" ("id") ON DELETE SET NULL, ADD CONSTRAINT "tasks_users_created_tasks" FOREIGN KEY ("creator_id") REFERENCES "users" ("id") ON DELETE SET NULL;
-- modify "teams" table
ALTER TABLE "teams" ADD CONSTRAINT "teams_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "team_memberships" table
ALTER TABLE "team_memberships" ADD CONSTRAINT "team_memberships_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "team_memberships_teams_team" FOREIGN KEY ("team_id") REFERENCES "teams" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "team_memberships_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE NO ACTION;
-- modify "tickets" table
ALTER TABLE "tickets" ADD CONSTRAINT "tickets_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "users" table
ALTER TABLE "users" ADD CONSTRAINT "users_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "users_knowledge_entities_knowledge_entity" FOREIGN KEY ("knowledge_entity_id") REFERENCES "knowledge_entities" ("id") ON DELETE SET NULL;
-- modify "user_auth_sessions" table
ALTER TABLE "user_auth_sessions" ADD CONSTRAINT "user_auth_sessions_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "user_auth_sessions_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE NO ACTION, ADD CONSTRAINT "user_auth_sessions_organizations_organization" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON DELETE NO ACTION;
-- modify "video_conferences" table
ALTER TABLE "video_conferences" ADD CONSTRAINT "video_conferences_incidents_video_conferences" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON DELETE SET NULL, ADD CONSTRAINT "video_conferences_meeting_sessions_video_conference" FOREIGN KEY ("meeting_session_id") REFERENCES "meeting_sessions" ("id") ON DELETE SET NULL, ADD CONSTRAINT "video_conferences_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON DELETE NO ACTION;
-- modify "incident_field_selections" table
ALTER TABLE "incident_field_selections" ADD CONSTRAINT "incident_field_selections_incident_id" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON DELETE CASCADE, ADD CONSTRAINT "incident_field_selections_incident_field_option_id" FOREIGN KEY ("incident_field_option_id") REFERENCES "incident_field_options" ("id") ON DELETE CASCADE;
-- modify "incident_tag_assignments" table
ALTER TABLE "incident_tag_assignments" ADD CONSTRAINT "incident_tag_assignments_incident_id" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON DELETE CASCADE, ADD CONSTRAINT "incident_tag_assignments_incident_tag_id" FOREIGN KEY ("incident_tag_id") REFERENCES "incident_tags" ("id") ON DELETE CASCADE;
-- modify "incident_review_sessions" table
ALTER TABLE "incident_review_sessions" ADD CONSTRAINT "incident_review_sessions_incident_id" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON DELETE CASCADE, ADD CONSTRAINT "incident_review_sessions_meeting_session_id" FOREIGN KEY ("meeting_session_id") REFERENCES "meeting_sessions" ("id") ON DELETE CASCADE;
-- modify "incident_debrief_question_incident_fields" table
ALTER TABLE "incident_debrief_question_incident_fields" ADD CONSTRAINT "incident_debrief_question_inci_44abe8f51887ab1da22a39603e050506" FOREIGN KEY ("incident_debrief_question_id") REFERENCES "incident_debrief_questions" ("id") ON DELETE CASCADE, ADD CONSTRAINT "incident_debrief_question_incident_fields_incident_field_id" FOREIGN KEY ("incident_field_id") REFERENCES "incident_fields" ("id") ON DELETE CASCADE;
-- modify "incident_debrief_question_incident_roles" table
ALTER TABLE "incident_debrief_question_incident_roles" ADD CONSTRAINT "incident_debrief_question_inci_88623030b1280506f5687158ce17d47b" FOREIGN KEY ("incident_debrief_question_id") REFERENCES "incident_debrief_questions" ("id") ON DELETE CASCADE, ADD CONSTRAINT "incident_debrief_question_incident_roles_incident_role_id" FOREIGN KEY ("incident_role_id") REFERENCES "incident_roles" ("id") ON DELETE CASCADE;
-- modify "incident_debrief_question_incident_severities" table
ALTER TABLE "incident_debrief_question_incident_severities" ADD CONSTRAINT "incident_debrief_question_inci_97954e08b0b3c4887d54c7b57902b851" FOREIGN KEY ("incident_debrief_question_id") REFERENCES "incident_debrief_questions" ("id") ON DELETE CASCADE, ADD CONSTRAINT "incident_debrief_question_inci_31f81d29bd5de0dbc30bab30ee6a1ce2" FOREIGN KEY ("incident_severity_id") REFERENCES "incident_severities" ("id") ON DELETE CASCADE;
-- modify "incident_debrief_question_incident_tags" table
ALTER TABLE "incident_debrief_question_incident_tags" ADD CONSTRAINT "incident_debrief_question_inci_5246f21c867173836779684ae23c83a5" FOREIGN KEY ("incident_debrief_question_id") REFERENCES "incident_debrief_questions" ("id") ON DELETE CASCADE, ADD CONSTRAINT "incident_debrief_question_incident_tags_incident_tag_id" FOREIGN KEY ("incident_tag_id") REFERENCES "incident_tags" ("id") ON DELETE CASCADE;
-- modify "incident_debrief_question_incident_types" table
ALTER TABLE "incident_debrief_question_incident_types" ADD CONSTRAINT "incident_debrief_question_inci_4140c1ef65a5d052c29594bc82faae77" FOREIGN KEY ("incident_debrief_question_id") REFERENCES "incident_debrief_questions" ("id") ON DELETE CASCADE, ADD CONSTRAINT "incident_debrief_question_incident_types_incident_type_id" FOREIGN KEY ("incident_type_id") REFERENCES "incident_types" ("id") ON DELETE CASCADE;
-- modify "meeting_schedule_owning_team" table
ALTER TABLE "meeting_schedule_owning_team" ADD CONSTRAINT "meeting_schedule_owning_team_meeting_schedule_id" FOREIGN KEY ("meeting_schedule_id") REFERENCES "meeting_schedules" ("id") ON DELETE CASCADE, ADD CONSTRAINT "meeting_schedule_owning_team_team_id" FOREIGN KEY ("team_id") REFERENCES "teams" ("id") ON DELETE CASCADE;
-- modify "oncall_shift_handover_pinned_annotations" table
ALTER TABLE "oncall_shift_handover_pinned_annotations" ADD CONSTRAINT "oncall_shift_handover_pinned_a_ea6451c95975edb633f05ea5a22d6958" FOREIGN KEY ("oncall_shift_handover_id") REFERENCES "oncall_shift_handovers" ("id") ON DELETE CASCADE, ADD CONSTRAINT "oncall_shift_handover_pinned_annotations_event_annotation_id" FOREIGN KEY ("event_annotation_id") REFERENCES "event_annotations" ("id") ON DELETE CASCADE;
-- modify "playbook_alerts" table
ALTER TABLE "playbook_alerts" ADD CONSTRAINT "playbook_alerts_playbook_id" FOREIGN KEY ("playbook_id") REFERENCES "playbooks" ("id") ON DELETE CASCADE, ADD CONSTRAINT "playbook_alerts_alert_id" FOREIGN KEY ("alert_id") REFERENCES "alerts" ("id") ON DELETE CASCADE;
-- modify "task_tickets" table
ALTER TABLE "task_tickets" ADD CONSTRAINT "task_tickets_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON DELETE CASCADE, ADD CONSTRAINT "task_tickets_ticket_id" FOREIGN KEY ("ticket_id") REFERENCES "tickets" ("id") ON DELETE CASCADE;
-- modify "team_oncall_rosters" table
ALTER TABLE "team_oncall_rosters" ADD CONSTRAINT "team_oncall_rosters_team_id" FOREIGN KEY ("team_id") REFERENCES "teams" ("id") ON DELETE CASCADE, ADD CONSTRAINT "team_oncall_rosters_oncall_roster_id" FOREIGN KEY ("oncall_roster_id") REFERENCES "oncall_rosters" ("id") ON DELETE CASCADE;
-- modify "user_watched_oncall_rosters" table
ALTER TABLE "user_watched_oncall_rosters" ADD CONSTRAINT "user_watched_oncall_rosters_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE, ADD CONSTRAINT "user_watched_oncall_rosters_oncall_roster_id" FOREIGN KEY ("oncall_roster_id") REFERENCES "oncall_rosters" ("id") ON DELETE CASCADE;
