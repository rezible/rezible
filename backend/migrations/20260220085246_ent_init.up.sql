-- create "alerts" table
CREATE TABLE "alerts" ("id" uuid NOT NULL, "external_id" character varying NULL, "title" character varying NOT NULL, "description" character varying NULL, "definition" character varying NULL, "tenant_id" bigint NOT NULL, "roster_id" uuid NULL, PRIMARY KEY ("id"));
-- create index "alert_tenant_id" to table: "alerts"
CREATE INDEX "alert_tenant_id" ON "alerts" ("tenant_id");
-- create "alert_feedbacks" table
CREATE TABLE "alert_feedbacks" ("id" uuid NOT NULL, "actionable" boolean NOT NULL, "accurate" character varying NOT NULL, "documentation_available" boolean NOT NULL, "documentation_needs_update" boolean NOT NULL, "tenant_id" bigint NOT NULL, "alert_instance_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "alertfeedback_tenant_id" to table: "alert_feedbacks"
CREATE INDEX "alertfeedback_tenant_id" ON "alert_feedbacks" ("tenant_id");
-- create "alert_instances" table
CREATE TABLE "alert_instances" ("id" uuid NOT NULL, "external_id" character varying NULL, "acknowledged_at" timestamptz NULL, "alert_instances" uuid NULL, "tenant_id" bigint NOT NULL, "alert_id" uuid NOT NULL, "event_id" uuid NOT NULL, "alert_instance_feedback" uuid NULL, PRIMARY KEY ("id"));
-- create index "alertinstance_tenant_id" to table: "alert_instances"
CREATE INDEX "alertinstance_tenant_id" ON "alert_instances" ("tenant_id");
-- create "documents" table
CREATE TABLE "documents" ("id" uuid NOT NULL, "content" bytea NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "document_tenant_id" to table: "documents"
CREATE INDEX "document_tenant_id" ON "documents" ("tenant_id");
-- create "events" table
CREATE TABLE "events" ("id" uuid NOT NULL, "external_id" character varying NULL, "timestamp" timestamptz NOT NULL, "kind" character varying NOT NULL, "title" character varying NOT NULL, "description" character varying NOT NULL, "source" character varying NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "event_tenant_id" to table: "events"
CREATE INDEX "event_tenant_id" ON "events" ("tenant_id");
-- create "event_annotations" table
CREATE TABLE "event_annotations" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "minutes_occupied" bigint NOT NULL, "notes" text NOT NULL, "tags" jsonb NOT NULL, "tenant_id" bigint NOT NULL, "event_id" uuid NOT NULL, "creator_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "eventannotation_tenant_id" to table: "event_annotations"
CREATE INDEX "eventannotation_tenant_id" ON "event_annotations" ("tenant_id");
-- create "incidents" table
CREATE TABLE "incidents" ("id" uuid NOT NULL, "external_id" character varying NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "slug" character varying NOT NULL, "title" character varying NOT NULL, "summary" character varying NULL, "chat_channel_id" character varying NULL, "opened_at" timestamptz NOT NULL, "tenant_id" bigint NOT NULL, "severity_id" uuid NOT NULL, "type_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "incident_tenant_id" to table: "incidents"
CREATE INDEX "incident_tenant_id" ON "incidents" ("tenant_id");
-- create index "incidents_slug_key" to table: "incidents"
CREATE UNIQUE INDEX "incidents_slug_key" ON "incidents" ("slug");
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
-- create "incident_events" table
CREATE TABLE "incident_events" ("id" uuid NOT NULL, "timestamp" timestamptz NOT NULL, "kind" character varying NOT NULL, "title" character varying NOT NULL, "description" text NULL, "is_key" boolean NOT NULL DEFAULT false, "sequence" bigint NOT NULL DEFAULT 0, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "incident_id" uuid NOT NULL, "tenant_id" bigint NOT NULL, "event_id" uuid NULL, PRIMARY KEY ("id"));
-- create index "incidentevent_kind" to table: "incident_events"
CREATE INDEX "incidentevent_kind" ON "incident_events" ("kind");
-- create index "incidentevent_tenant_id" to table: "incident_events"
CREATE INDEX "incidentevent_tenant_id" ON "incident_events" ("tenant_id");
-- create "incident_event_contexts" table
CREATE TABLE "incident_event_contexts" ("id" uuid NOT NULL, "system_state" text NULL, "decision_options" jsonb NULL, "decision_rationale" text NULL, "involved_personnel" jsonb NULL, "created_at" timestamptz NOT NULL, "incident_event_context" uuid NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "incident_event_contexts_incident_event_context_key" to table: "incident_event_contexts"
CREATE UNIQUE INDEX "incident_event_contexts_incident_event_context_key" ON "incident_event_contexts" ("incident_event_context");
-- create index "incidenteventcontext_tenant_id" to table: "incident_event_contexts"
CREATE INDEX "incidenteventcontext_tenant_id" ON "incident_event_contexts" ("tenant_id");
-- create "incident_event_contributing_factors" table
CREATE TABLE "incident_event_contributing_factors" ("id" uuid NOT NULL, "factor_type" character varying NOT NULL, "description" text NULL, "created_at" timestamptz NOT NULL, "incident_event_factors" uuid NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "incidenteventcontributingfactor_tenant_id" to table: "incident_event_contributing_factors"
CREATE INDEX "incidenteventcontributingfactor_tenant_id" ON "incident_event_contributing_factors" ("tenant_id");
-- create "incident_event_evidences" table
CREATE TABLE "incident_event_evidences" ("id" uuid NOT NULL, "evidence_type" character varying NOT NULL, "url" character varying NOT NULL, "title" character varying NOT NULL, "description" text NULL, "created_at" timestamptz NOT NULL, "incident_event_evidence" uuid NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "incidenteventevidence_tenant_id" to table: "incident_event_evidences"
CREATE INDEX "incidenteventevidence_tenant_id" ON "incident_event_evidences" ("tenant_id");
-- create "incident_event_system_components" table
CREATE TABLE "incident_event_system_components" ("id" uuid NOT NULL, "relationship" character varying NOT NULL, "created_at" timestamptz NOT NULL, "tenant_id" bigint NOT NULL, "incident_event_id" uuid NULL, "system_component_id" uuid NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "incident_event_system_componen_47cae3c6c94b8ac00bd374afd6aa371f" FOREIGN KEY ("incident_event_id") REFERENCES "incident_event_system_components" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "incident_event_system_components_incident_event_id_key" to table: "incident_event_system_components"
CREATE UNIQUE INDEX "incident_event_system_components_incident_event_id_key" ON "incident_event_system_components" ("incident_event_id");
-- create index "incidenteventsystemcomponent_i_68243ec125e2acebf985bc112b82147a" to table: "incident_event_system_components"
CREATE UNIQUE INDEX "incidenteventsystemcomponent_i_68243ec125e2acebf985bc112b82147a" ON "incident_event_system_components" ("incident_event_id", "system_component_id");
-- create index "incidenteventsystemcomponent_tenant_id" to table: "incident_event_system_components"
CREATE INDEX "incidenteventsystemcomponent_tenant_id" ON "incident_event_system_components" ("tenant_id");
-- create "incident_fields" table
CREATE TABLE "incident_fields" ("id" uuid NOT NULL, "archive_time" timestamptz NULL, "name" character varying NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "incidentfield_tenant_id" to table: "incident_fields"
CREATE INDEX "incidentfield_tenant_id" ON "incident_fields" ("tenant_id");
-- create "incident_field_options" table
CREATE TABLE "incident_field_options" ("id" uuid NOT NULL, "archive_time" timestamptz NULL, "type" character varying NOT NULL, "value" character varying NOT NULL, "incident_field_id" uuid NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "incidentfieldoption_tenant_id" to table: "incident_field_options"
CREATE INDEX "incidentfieldoption_tenant_id" ON "incident_field_options" ("tenant_id");
-- create "incident_links" table
CREATE TABLE "incident_links" ("id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY, "description" character varying NULL, "link_type" character varying NOT NULL, "tenant_id" bigint NOT NULL, "incident_id" uuid NOT NULL, "linked_incident_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "incidentlink_incident_id_linked_incident_id" to table: "incident_links"
CREATE UNIQUE INDEX "incidentlink_incident_id_linked_incident_id" ON "incident_links" ("incident_id", "linked_incident_id");
-- create index "incidentlink_tenant_id" to table: "incident_links"
CREATE INDEX "incidentlink_tenant_id" ON "incident_links" ("tenant_id");
-- create "incident_milestones" table
CREATE TABLE "incident_milestones" ("id" uuid NOT NULL, "kind" character varying NOT NULL, "timestamp" timestamptz NOT NULL, "description" character varying NULL, "source" character varying NULL, "metadata" jsonb NULL, "incident_id" uuid NOT NULL, "tenant_id" bigint NOT NULL, "user_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "incidentmilestone_tenant_id" to table: "incident_milestones"
CREATE INDEX "incidentmilestone_tenant_id" ON "incident_milestones" ("tenant_id");
-- create "incident_roles" table
CREATE TABLE "incident_roles" ("id" uuid NOT NULL, "archive_time" timestamptz NULL, "external_id" character varying NULL, "name" character varying NOT NULL, "required" boolean NOT NULL DEFAULT false, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "incidentrole_tenant_id" to table: "incident_roles"
CREATE INDEX "incidentrole_tenant_id" ON "incident_roles" ("tenant_id");
-- create "incident_role_assignments" table
CREATE TABLE "incident_role_assignments" ("id" uuid NOT NULL, "tenant_id" bigint NOT NULL, "incident_id" uuid NOT NULL, "user_id" uuid NOT NULL, "role_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "incidentroleassignment_tenant_id" to table: "incident_role_assignments"
CREATE INDEX "incidentroleassignment_tenant_id" ON "incident_role_assignments" ("tenant_id");
-- create index "incidentroleassignment_user_id_incident_id" to table: "incident_role_assignments"
CREATE UNIQUE INDEX "incidentroleassignment_user_id_incident_id" ON "incident_role_assignments" ("user_id", "incident_id");
-- create "incident_severities" table
CREATE TABLE "incident_severities" ("id" uuid NOT NULL, "archive_time" timestamptz NULL, "external_id" character varying NULL, "name" character varying NOT NULL, "rank" bigint NOT NULL, "color" character varying NULL, "description" character varying NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "incidentseverity_tenant_id" to table: "incident_severities"
CREATE INDEX "incidentseverity_tenant_id" ON "incident_severities" ("tenant_id");
-- create "incident_tags" table
CREATE TABLE "incident_tags" ("id" uuid NOT NULL, "archive_time" timestamptz NULL, "key" character varying NOT NULL, "value" character varying NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "incidenttag_tenant_id" to table: "incident_tags"
CREATE INDEX "incidenttag_tenant_id" ON "incident_tags" ("tenant_id");
-- create "incident_types" table
CREATE TABLE "incident_types" ("id" uuid NOT NULL, "archive_time" timestamptz NULL, "name" character varying NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "incidenttype_tenant_id" to table: "incident_types"
CREATE INDEX "incidenttype_tenant_id" ON "incident_types" ("tenant_id");
-- create "integrations" table
CREATE TABLE "integrations" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "name" character varying NOT NULL, "config" jsonb NOT NULL, "user_preferences" jsonb NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "integration_tenant_id" to table: "integrations"
CREATE INDEX "integration_tenant_id" ON "integrations" ("tenant_id");
-- create index "integration_tenant_id_name" to table: "integrations"
CREATE UNIQUE INDEX "integration_tenant_id_name" ON "integrations" ("tenant_id", "name");
-- create "meeting_schedules" table
CREATE TABLE "meeting_schedules" ("id" uuid NOT NULL, "archive_time" timestamptz NULL, "name" character varying NOT NULL, "description" character varying NULL, "begin_minute" bigint NOT NULL, "duration_minutes" bigint NOT NULL, "start_date" timestamptz NOT NULL, "repeats" character varying NOT NULL, "repetition_step" bigint NOT NULL DEFAULT 1, "week_days" jsonb NULL, "monthly_on" character varying NULL, "until_date" timestamptz NULL, "num_repetitions" bigint NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "meetingschedule_tenant_id" to table: "meeting_schedules"
CREATE INDEX "meetingschedule_tenant_id" ON "meeting_schedules" ("tenant_id");
-- create "meeting_sessions" table
CREATE TABLE "meeting_sessions" ("id" uuid NOT NULL, "title" character varying NOT NULL, "started_at" timestamptz NOT NULL, "ended_at" timestamptz NULL, "document_name" character varying NOT NULL, "tenant_id" bigint NOT NULL, "meeting_session_schedule" uuid NULL, PRIMARY KEY ("id"));
-- create index "meetingsession_tenant_id" to table: "meeting_sessions"
CREATE INDEX "meetingsession_tenant_id" ON "meeting_sessions" ("tenant_id");
-- create "oncall_handover_templates" table
CREATE TABLE "oncall_handover_templates" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "contents" bytea NOT NULL, "is_default" boolean NOT NULL DEFAULT false, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "oncallhandovertemplate_tenant_id" to table: "oncall_handover_templates"
CREATE INDEX "oncallhandovertemplate_tenant_id" ON "oncall_handover_templates" ("tenant_id");
-- create "oncall_rosters" table
CREATE TABLE "oncall_rosters" ("id" uuid NOT NULL, "archive_time" timestamptz NULL, "external_id" character varying NULL, "name" character varying NOT NULL, "slug" character varying NOT NULL, "timezone" character varying NULL, "chat_handle" character varying NULL, "chat_channel_id" character varying NULL, "handover_template_id" uuid NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "oncall_rosters_slug_key" to table: "oncall_rosters"
CREATE UNIQUE INDEX "oncall_rosters_slug_key" ON "oncall_rosters" ("slug");
-- create index "oncallroster_tenant_id" to table: "oncall_rosters"
CREATE INDEX "oncallroster_tenant_id" ON "oncall_rosters" ("tenant_id");
-- create "oncall_roster_metrics" table
CREATE TABLE "oncall_roster_metrics" ("id" uuid NOT NULL, "tenant_id" bigint NOT NULL, "roster_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "oncallrostermetrics_tenant_id" to table: "oncall_roster_metrics"
CREATE INDEX "oncallrostermetrics_tenant_id" ON "oncall_roster_metrics" ("tenant_id");
-- create "oncall_schedules" table
CREATE TABLE "oncall_schedules" ("id" uuid NOT NULL, "archive_time" timestamptz NULL, "external_id" character varying NULL, "name" character varying NOT NULL, "timezone" character varying NULL, "roster_id" uuid NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "oncallschedule_tenant_id" to table: "oncall_schedules"
CREATE INDEX "oncallschedule_tenant_id" ON "oncall_schedules" ("tenant_id");
-- create "oncall_schedule_participants" table
CREATE TABLE "oncall_schedule_participants" ("id" uuid NOT NULL, "index" bigint NOT NULL, "schedule_id" uuid NOT NULL, "tenant_id" bigint NOT NULL, "user_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "oncallscheduleparticipant_tenant_id" to table: "oncall_schedule_participants"
CREATE INDEX "oncallscheduleparticipant_tenant_id" ON "oncall_schedule_participants" ("tenant_id");
-- create "oncall_shifts" table
CREATE TABLE "oncall_shifts" ("id" uuid NOT NULL, "external_id" character varying NULL, "role" character varying NULL DEFAULT 'primary', "start_at" timestamptz NOT NULL, "end_at" timestamptz NOT NULL, "tenant_id" bigint NOT NULL, "user_id" uuid NOT NULL, "roster_id" uuid NOT NULL, "primary_shift_id" uuid NULL, PRIMARY KEY ("id"), CONSTRAINT "oncall_shifts_oncall_shifts_primary_shift" FOREIGN KEY ("primary_shift_id") REFERENCES "oncall_shifts" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
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
CREATE TABLE "organizations" ("id" uuid NOT NULL, "external_id" character varying NOT NULL, "name" character varying NOT NULL, "initial_setup_at" timestamptz NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "organization_tenant_id" to table: "organizations"
CREATE INDEX "organization_tenant_id" ON "organizations" ("tenant_id");
-- create "playbooks" table
CREATE TABLE "playbooks" ("id" uuid NOT NULL, "external_id" character varying NULL, "title" character varying NOT NULL, "content" bytea NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "playbook_tenant_id" to table: "playbooks"
CREATE INDEX "playbook_tenant_id" ON "playbooks" ("tenant_id");
-- create "provider_sync_histories" table
CREATE TABLE "provider_sync_histories" ("id" uuid NOT NULL, "data_type" character varying NOT NULL, "started_at" timestamptz NOT NULL, "finished_at" timestamptz NOT NULL, "num_mutations" bigint NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "providersynchistory_tenant_id" to table: "provider_sync_histories"
CREATE INDEX "providersynchistory_tenant_id" ON "provider_sync_histories" ("tenant_id");
-- create "retrospectives" table
CREATE TABLE "retrospectives" ("id" uuid NOT NULL, "type" character varying NOT NULL, "state" character varying NOT NULL, "document_id" uuid NOT NULL, "incident_id" uuid NOT NULL, "tenant_id" bigint NOT NULL, "system_analysis_id" uuid NULL, PRIMARY KEY ("id"));
-- create index "retrospective_tenant_id" to table: "retrospectives"
CREATE INDEX "retrospective_tenant_id" ON "retrospectives" ("tenant_id");
-- create index "retrospectives_document_id_key" to table: "retrospectives"
CREATE UNIQUE INDEX "retrospectives_document_id_key" ON "retrospectives" ("document_id");
-- create index "retrospectives_incident_id_key" to table: "retrospectives"
CREATE UNIQUE INDEX "retrospectives_incident_id_key" ON "retrospectives" ("incident_id");
-- create index "retrospectives_system_analysis_id_key" to table: "retrospectives"
CREATE UNIQUE INDEX "retrospectives_system_analysis_id_key" ON "retrospectives" ("system_analysis_id");
-- create "retrospective_comments" table
CREATE TABLE "retrospective_comments" ("id" uuid NOT NULL, "content" bytea NOT NULL, "tenant_id" bigint NOT NULL, "retrospective_id" uuid NOT NULL, "user_id" uuid NOT NULL, "retrospective_review_id" uuid NULL, "parent_reply_id" uuid NULL, PRIMARY KEY ("id"), CONSTRAINT "retrospective_comments_retrospective_comments_replies" FOREIGN KEY ("parent_reply_id") REFERENCES "retrospective_comments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "retrospectivecomment_tenant_id" to table: "retrospective_comments"
CREATE INDEX "retrospectivecomment_tenant_id" ON "retrospective_comments" ("tenant_id");
-- create "retrospective_reviews" table
CREATE TABLE "retrospective_reviews" ("id" uuid NOT NULL, "state" character varying NOT NULL, "tenant_id" bigint NOT NULL, "retrospective_id" uuid NOT NULL, "requester_id" uuid NOT NULL, "reviewer_id" uuid NOT NULL, "comment_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "retrospectivereview_tenant_id" to table: "retrospective_reviews"
CREATE INDEX "retrospectivereview_tenant_id" ON "retrospective_reviews" ("tenant_id");
-- create "system_analyses" table
CREATE TABLE "system_analyses" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "systemanalysis_tenant_id" to table: "system_analyses"
CREATE INDEX "systemanalysis_tenant_id" ON "system_analyses" ("tenant_id");
-- create "system_analysis_components" table
CREATE TABLE "system_analysis_components" ("id" uuid NOT NULL, "description" text NULL, "pos_x" double precision NOT NULL DEFAULT 0, "pos_y" double precision NOT NULL DEFAULT 0, "created_at" timestamptz NOT NULL, "tenant_id" bigint NOT NULL, "analysis_id" uuid NOT NULL, "component_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "systemanalysiscomponent_component_id_analysis_id" to table: "system_analysis_components"
CREATE UNIQUE INDEX "systemanalysiscomponent_component_id_analysis_id" ON "system_analysis_components" ("component_id", "analysis_id");
-- create index "systemanalysiscomponent_tenant_id" to table: "system_analysis_components"
CREATE INDEX "systemanalysiscomponent_tenant_id" ON "system_analysis_components" ("tenant_id");
-- create "system_analysis_relationships" table
CREATE TABLE "system_analysis_relationships" ("id" uuid NOT NULL, "description" text NULL, "created_at" timestamptz NOT NULL, "tenant_id" bigint NOT NULL, "analysis_id" uuid NOT NULL, "relationship_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "systemanalysisrelationship_tenant_id" to table: "system_analysis_relationships"
CREATE INDEX "systemanalysisrelationship_tenant_id" ON "system_analysis_relationships" ("tenant_id");
-- create "system_components" table
CREATE TABLE "system_components" ("id" uuid NOT NULL, "external_id" character varying NULL, "name" character varying NOT NULL, "description" text NULL, "properties" jsonb NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "tenant_id" bigint NOT NULL, "kind_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "systemcomponent_tenant_id" to table: "system_components"
CREATE INDEX "systemcomponent_tenant_id" ON "system_components" ("tenant_id");
-- create "system_component_constraints" table
CREATE TABLE "system_component_constraints" ("id" uuid NOT NULL, "label" text NOT NULL, "description" text NULL, "created_at" timestamptz NOT NULL, "tenant_id" bigint NOT NULL, "component_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "systemcomponentconstraint_tenant_id" to table: "system_component_constraints"
CREATE INDEX "systemcomponentconstraint_tenant_id" ON "system_component_constraints" ("tenant_id");
-- create "system_component_controls" table
CREATE TABLE "system_component_controls" ("id" uuid NOT NULL, "label" text NOT NULL, "description" text NULL, "created_at" timestamptz NOT NULL, "tenant_id" bigint NOT NULL, "component_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "systemcomponentcontrol_tenant_id" to table: "system_component_controls"
CREATE INDEX "systemcomponentcontrol_tenant_id" ON "system_component_controls" ("tenant_id");
-- create "system_component_kinds" table
CREATE TABLE "system_component_kinds" ("id" uuid NOT NULL, "external_id" character varying NULL, "label" text NOT NULL, "description" text NULL, "created_at" timestamptz NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "systemcomponentkind_tenant_id" to table: "system_component_kinds"
CREATE INDEX "systemcomponentkind_tenant_id" ON "system_component_kinds" ("tenant_id");
-- create "system_component_relationships" table
CREATE TABLE "system_component_relationships" ("id" uuid NOT NULL, "external_id" character varying NULL, "description" text NULL, "created_at" timestamptz NOT NULL, "tenant_id" bigint NOT NULL, "source_id" uuid NOT NULL, "target_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "systemcomponentrelationship_source_id_target_id" to table: "system_component_relationships"
CREATE UNIQUE INDEX "systemcomponentrelationship_source_id_target_id" ON "system_component_relationships" ("source_id", "target_id");
-- create index "systemcomponentrelationship_tenant_id" to table: "system_component_relationships"
CREATE INDEX "systemcomponentrelationship_tenant_id" ON "system_component_relationships" ("tenant_id");
-- create "system_component_signals" table
CREATE TABLE "system_component_signals" ("id" uuid NOT NULL, "label" text NOT NULL, "description" text NULL, "created_at" timestamptz NOT NULL, "tenant_id" bigint NOT NULL, "component_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "systemcomponentsignal_tenant_id" to table: "system_component_signals"
CREATE INDEX "systemcomponentsignal_tenant_id" ON "system_component_signals" ("tenant_id");
-- create "system_hazards" table
CREATE TABLE "system_hazards" ("id" uuid NOT NULL, "external_id" character varying NULL, "name" character varying NOT NULL, "description" text NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "systemhazard_tenant_id" to table: "system_hazards"
CREATE INDEX "systemhazard_tenant_id" ON "system_hazards" ("tenant_id");
-- create "system_relationship_control_actions" table
CREATE TABLE "system_relationship_control_actions" ("id" uuid NOT NULL, "name" character varying NOT NULL, "description" text NULL, "created_at" timestamptz NOT NULL, "tenant_id" bigint NOT NULL, "relationship_id" uuid NOT NULL, "control_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "systemrelationshipcontrolaction_relationship_id_control_id" to table: "system_relationship_control_actions"
CREATE UNIQUE INDEX "systemrelationshipcontrolaction_relationship_id_control_id" ON "system_relationship_control_actions" ("relationship_id", "control_id");
-- create index "systemrelationshipcontrolaction_tenant_id" to table: "system_relationship_control_actions"
CREATE INDEX "systemrelationshipcontrolaction_tenant_id" ON "system_relationship_control_actions" ("tenant_id");
-- create "system_relationship_feedback_signals" table
CREATE TABLE "system_relationship_feedback_signals" ("id" uuid NOT NULL, "name" character varying NOT NULL, "description" text NULL, "created_at" timestamptz NOT NULL, "tenant_id" bigint NOT NULL, "relationship_id" uuid NOT NULL, "signal_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "systemrelationshipfeedbacksignal_relationship_id_signal_id" to table: "system_relationship_feedback_signals"
CREATE UNIQUE INDEX "systemrelationshipfeedbacksignal_relationship_id_signal_id" ON "system_relationship_feedback_signals" ("relationship_id", "signal_id");
-- create index "systemrelationshipfeedbacksignal_tenant_id" to table: "system_relationship_feedback_signals"
CREATE INDEX "systemrelationshipfeedbacksignal_tenant_id" ON "system_relationship_feedback_signals" ("tenant_id");
-- create "tasks" table
CREATE TABLE "tasks" ("id" uuid NOT NULL, "type" character varying NOT NULL, "title" character varying NOT NULL, "incident_id" uuid NULL, "tenant_id" bigint NOT NULL, "assignee_id" uuid NULL, "creator_id" uuid NULL, PRIMARY KEY ("id"));
-- create index "task_tenant_id" to table: "tasks"
CREATE INDEX "task_tenant_id" ON "tasks" ("tenant_id");
-- create "teams" table
CREATE TABLE "teams" ("id" uuid NOT NULL, "external_id" character varying NULL, "slug" character varying NOT NULL, "name" character varying NOT NULL, "chat_channel_id" character varying NULL, "timezone" character varying NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "team_tenant_id" to table: "teams"
CREATE INDEX "team_tenant_id" ON "teams" ("tenant_id");
-- create index "teams_slug_key" to table: "teams"
CREATE UNIQUE INDEX "teams_slug_key" ON "teams" ("slug");
-- create "team_memberships" table
CREATE TABLE "team_memberships" ("id" uuid NOT NULL, "role" character varying NOT NULL DEFAULT 'member', "tenant_id" bigint NOT NULL, "team_id" uuid NOT NULL, "user_id" uuid NOT NULL, PRIMARY KEY ("id"));
-- create index "teammembership_team_id_user_id" to table: "team_memberships"
CREATE UNIQUE INDEX "teammembership_team_id_user_id" ON "team_memberships" ("team_id", "user_id");
-- create index "teammembership_tenant_id" to table: "team_memberships"
CREATE INDEX "teammembership_tenant_id" ON "team_memberships" ("tenant_id");
-- create "tenants" table
CREATE TABLE "tenants" ("id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY, PRIMARY KEY ("id"));
-- create "tickets" table
CREATE TABLE "tickets" ("id" uuid NOT NULL, "external_id" character varying NULL, "title" character varying NOT NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "ticket_tenant_id" to table: "tickets"
CREATE INDEX "ticket_tenant_id" ON "tickets" ("tenant_id");
-- create "users" table
CREATE TABLE "users" ("id" uuid NOT NULL, "auth_provider_id" character varying NULL, "email" character varying NOT NULL, "name" character varying NULL DEFAULT '', "is_org_admin" boolean NOT NULL DEFAULT false, "chat_id" character varying NULL, "timezone" character varying NULL, "confirmed" boolean NOT NULL DEFAULT false, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "user_tenant_id" to table: "users"
CREATE INDEX "user_tenant_id" ON "users" ("tenant_id");
-- create "video_conferences" table
CREATE TABLE "video_conferences" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "provider" character varying NOT NULL, "external_id" character varying NULL, "join_url" character varying NOT NULL, "host_url" character varying NULL, "dial_in" character varying NULL, "passcode" character varying NULL, "status" character varying NOT NULL DEFAULT 'creating', "metadata" jsonb NULL, "created_by_integration" character varying NULL, "incident_id" uuid NULL, "meeting_session_id" uuid NULL, "tenant_id" bigint NOT NULL, PRIMARY KEY ("id"));
-- create index "video_conferences_meeting_session_id_key" to table: "video_conferences"
CREATE UNIQUE INDEX "video_conferences_meeting_session_id_key" ON "video_conferences" ("meeting_session_id");
-- create index "videoconference_incident_id_status" to table: "video_conferences"
CREATE INDEX "videoconference_incident_id_status" ON "video_conferences" ("incident_id", "status");
-- create index "videoconference_meeting_session_id_status" to table: "video_conferences"
CREATE INDEX "videoconference_meeting_session_id_status" ON "video_conferences" ("meeting_session_id", "status");
-- create index "videoconference_tenant_id" to table: "video_conferences"
CREATE INDEX "videoconference_tenant_id" ON "video_conferences" ("tenant_id");
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
-- create "system_hazard_components" table
CREATE TABLE "system_hazard_components" ("system_hazard_id" uuid NOT NULL, "system_component_id" uuid NOT NULL, PRIMARY KEY ("system_hazard_id", "system_component_id"));
-- create "system_hazard_constraints" table
CREATE TABLE "system_hazard_constraints" ("system_hazard_id" uuid NOT NULL, "system_component_constraint_id" uuid NOT NULL, PRIMARY KEY ("system_hazard_id", "system_component_constraint_id"));
-- create "system_hazard_relationships" table
CREATE TABLE "system_hazard_relationships" ("system_hazard_id" uuid NOT NULL, "system_component_relationship_id" uuid NOT NULL, PRIMARY KEY ("system_hazard_id", "system_component_relationship_id"));
-- create "task_tickets" table
CREATE TABLE "task_tickets" ("task_id" uuid NOT NULL, "ticket_id" uuid NOT NULL, PRIMARY KEY ("task_id", "ticket_id"));
-- create "team_oncall_rosters" table
CREATE TABLE "team_oncall_rosters" ("team_id" uuid NOT NULL, "oncall_roster_id" uuid NOT NULL, PRIMARY KEY ("team_id", "oncall_roster_id"));
-- create "user_watched_oncall_rosters" table
CREATE TABLE "user_watched_oncall_rosters" ("user_id" uuid NOT NULL, "oncall_roster_id" uuid NOT NULL, PRIMARY KEY ("user_id", "oncall_roster_id"));
-- modify "alerts" table
ALTER TABLE "alerts" ADD CONSTRAINT "alerts_oncall_rosters_alerts" FOREIGN KEY ("roster_id") REFERENCES "oncall_rosters" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "alerts_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "alert_feedbacks" table
ALTER TABLE "alert_feedbacks" ADD CONSTRAINT "alert_feedbacks_alert_instances_alert_instance" FOREIGN KEY ("alert_instance_id") REFERENCES "alert_instances" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "alert_feedbacks_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "alert_instances" table
ALTER TABLE "alert_instances" ADD CONSTRAINT "alert_instances_alert_feedbacks_feedback" FOREIGN KEY ("alert_instance_feedback") REFERENCES "alert_feedbacks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "alert_instances_alerts_alert" FOREIGN KEY ("alert_id") REFERENCES "alerts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "alert_instances_alerts_instances" FOREIGN KEY ("alert_instances") REFERENCES "alerts" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "alert_instances_events_event" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "alert_instances_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "documents" table
ALTER TABLE "documents" ADD CONSTRAINT "documents_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "events" table
ALTER TABLE "events" ADD CONSTRAINT "events_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "event_annotations" table
ALTER TABLE "event_annotations" ADD CONSTRAINT "event_annotations_events_event" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "event_annotations_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "event_annotations_users_creator" FOREIGN KEY ("creator_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "incidents" table
ALTER TABLE "incidents" ADD CONSTRAINT "incidents_incident_severities_severity" FOREIGN KEY ("severity_id") REFERENCES "incident_severities" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "incidents_incident_types_type" FOREIGN KEY ("type_id") REFERENCES "incident_types" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "incidents_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "incident_debriefs" table
ALTER TABLE "incident_debriefs" ADD CONSTRAINT "incident_debriefs_incidents_debriefs" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "incident_debriefs_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "incident_debriefs_users_incident_debriefs" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "incident_debrief_messages" table
ALTER TABLE "incident_debrief_messages" ADD CONSTRAINT "incident_debrief_messages_inci_0d1f0b105ef851edb04b442b34a0d17f" FOREIGN KEY ("question_id") REFERENCES "incident_debrief_questions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "incident_debrief_messages_incident_debriefs_messages" FOREIGN KEY ("debrief_id") REFERENCES "incident_debriefs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "incident_debrief_messages_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "incident_debrief_questions" table
ALTER TABLE "incident_debrief_questions" ADD CONSTRAINT "incident_debrief_questions_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "incident_debrief_suggestions" table
ALTER TABLE "incident_debrief_suggestions" ADD CONSTRAINT "incident_debrief_suggestions_incident_debriefs_suggestions" FOREIGN KEY ("incident_debrief_suggestions") REFERENCES "incident_debriefs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "incident_debrief_suggestions_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "incident_events" table
ALTER TABLE "incident_events" ADD CONSTRAINT "incident_events_events_event" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "incident_events_incidents_events" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "incident_events_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "incident_event_contexts" table
ALTER TABLE "incident_event_contexts" ADD CONSTRAINT "incident_event_contexts_incident_events_context" FOREIGN KEY ("incident_event_context") REFERENCES "incident_events" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "incident_event_contexts_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "incident_event_contributing_factors" table
ALTER TABLE "incident_event_contributing_factors" ADD CONSTRAINT "incident_event_contributing_factors_incident_events_factors" FOREIGN KEY ("incident_event_factors") REFERENCES "incident_events" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "incident_event_contributing_factors_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "incident_event_evidences" table
ALTER TABLE "incident_event_evidences" ADD CONSTRAINT "incident_event_evidences_incident_events_evidence" FOREIGN KEY ("incident_event_evidence") REFERENCES "incident_events" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "incident_event_evidences_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "incident_event_system_components" table
ALTER TABLE "incident_event_system_components" ADD CONSTRAINT "incident_event_system_componen_426e7b9f5e52750ab9f88715a403e203" FOREIGN KEY ("system_component_id") REFERENCES "system_components" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "incident_event_system_components_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "incident_fields" table
ALTER TABLE "incident_fields" ADD CONSTRAINT "incident_fields_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "incident_field_options" table
ALTER TABLE "incident_field_options" ADD CONSTRAINT "incident_field_options_incident_fields_options" FOREIGN KEY ("incident_field_id") REFERENCES "incident_fields" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "incident_field_options_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "incident_links" table
ALTER TABLE "incident_links" ADD CONSTRAINT "incident_links_incidents_incident" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "incident_links_incidents_linked_incident" FOREIGN KEY ("linked_incident_id") REFERENCES "incidents" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "incident_links_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "incident_milestones" table
ALTER TABLE "incident_milestones" ADD CONSTRAINT "incident_milestones_incidents_milestones" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "incident_milestones_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "incident_milestones_users_incident_milestones" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "incident_roles" table
ALTER TABLE "incident_roles" ADD CONSTRAINT "incident_roles_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "incident_role_assignments" table
ALTER TABLE "incident_role_assignments" ADD CONSTRAINT "incident_role_assignments_incident_roles_role" FOREIGN KEY ("role_id") REFERENCES "incident_roles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "incident_role_assignments_incidents_incident" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "incident_role_assignments_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "incident_role_assignments_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "incident_severities" table
ALTER TABLE "incident_severities" ADD CONSTRAINT "incident_severities_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "incident_tags" table
ALTER TABLE "incident_tags" ADD CONSTRAINT "incident_tags_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "incident_types" table
ALTER TABLE "incident_types" ADD CONSTRAINT "incident_types_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "integrations" table
ALTER TABLE "integrations" ADD CONSTRAINT "integrations_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "meeting_schedules" table
ALTER TABLE "meeting_schedules" ADD CONSTRAINT "meeting_schedules_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "meeting_sessions" table
ALTER TABLE "meeting_sessions" ADD CONSTRAINT "meeting_sessions_meeting_schedules_schedule" FOREIGN KEY ("meeting_session_schedule") REFERENCES "meeting_schedules" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "meeting_sessions_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "oncall_handover_templates" table
ALTER TABLE "oncall_handover_templates" ADD CONSTRAINT "oncall_handover_templates_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "oncall_rosters" table
ALTER TABLE "oncall_rosters" ADD CONSTRAINT "oncall_rosters_oncall_handover_templates_roster" FOREIGN KEY ("handover_template_id") REFERENCES "oncall_handover_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "oncall_rosters_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "oncall_roster_metrics" table
ALTER TABLE "oncall_roster_metrics" ADD CONSTRAINT "oncall_roster_metrics_oncall_rosters_roster" FOREIGN KEY ("roster_id") REFERENCES "oncall_rosters" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "oncall_roster_metrics_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "oncall_schedules" table
ALTER TABLE "oncall_schedules" ADD CONSTRAINT "oncall_schedules_oncall_rosters_schedules" FOREIGN KEY ("roster_id") REFERENCES "oncall_rosters" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "oncall_schedules_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "oncall_schedule_participants" table
ALTER TABLE "oncall_schedule_participants" ADD CONSTRAINT "oncall_schedule_participants_oncall_schedules_participants" FOREIGN KEY ("schedule_id") REFERENCES "oncall_schedules" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "oncall_schedule_participants_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "oncall_schedule_participants_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "oncall_shifts" table
ALTER TABLE "oncall_shifts" ADD CONSTRAINT "oncall_shifts_oncall_rosters_roster" FOREIGN KEY ("roster_id") REFERENCES "oncall_rosters" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "oncall_shifts_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "oncall_shifts_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "oncall_shift_handovers" table
ALTER TABLE "oncall_shift_handovers" ADD CONSTRAINT "oncall_shift_handovers_oncall_shifts_handover" FOREIGN KEY ("shift_id") REFERENCES "oncall_shifts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "oncall_shift_handovers_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "oncall_shift_metrics" table
ALTER TABLE "oncall_shift_metrics" ADD CONSTRAINT "oncall_shift_metrics_oncall_shifts_metrics" FOREIGN KEY ("shift_id") REFERENCES "oncall_shifts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "oncall_shift_metrics_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "organizations" table
ALTER TABLE "organizations" ADD CONSTRAINT "organizations_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "playbooks" table
ALTER TABLE "playbooks" ADD CONSTRAINT "playbooks_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "provider_sync_histories" table
ALTER TABLE "provider_sync_histories" ADD CONSTRAINT "provider_sync_histories_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "retrospectives" table
ALTER TABLE "retrospectives" ADD CONSTRAINT "retrospectives_documents_retrospective" FOREIGN KEY ("document_id") REFERENCES "documents" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "retrospectives_incidents_retrospective" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "retrospectives_system_analyses_retrospective" FOREIGN KEY ("system_analysis_id") REFERENCES "system_analyses" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "retrospectives_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "retrospective_comments" table
ALTER TABLE "retrospective_comments" ADD CONSTRAINT "retrospective_comments_retrospective_reviews_review" FOREIGN KEY ("retrospective_review_id") REFERENCES "retrospective_reviews" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "retrospective_comments_retrospectives_retrospective" FOREIGN KEY ("retrospective_id") REFERENCES "retrospectives" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "retrospective_comments_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "retrospective_comments_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "retrospective_reviews" table
ALTER TABLE "retrospective_reviews" ADD CONSTRAINT "retrospective_reviews_retrospective_comments_comment" FOREIGN KEY ("comment_id") REFERENCES "retrospective_comments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "retrospective_reviews_retrospectives_retrospective" FOREIGN KEY ("retrospective_id") REFERENCES "retrospectives" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "retrospective_reviews_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "retrospective_reviews_users_requester" FOREIGN KEY ("requester_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "retrospective_reviews_users_reviewer" FOREIGN KEY ("reviewer_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "system_analyses" table
ALTER TABLE "system_analyses" ADD CONSTRAINT "system_analyses_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "system_analysis_components" table
ALTER TABLE "system_analysis_components" ADD CONSTRAINT "system_analysis_components_system_analyses_analysis" FOREIGN KEY ("analysis_id") REFERENCES "system_analyses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "system_analysis_components_system_components_component" FOREIGN KEY ("component_id") REFERENCES "system_components" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "system_analysis_components_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "system_analysis_relationships" table
ALTER TABLE "system_analysis_relationships" ADD CONSTRAINT "system_analysis_relationships__78049507d24c4a0660356c42d5058c63" FOREIGN KEY ("relationship_id") REFERENCES "system_component_relationships" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "system_analysis_relationships_system_analyses_system_analysis" FOREIGN KEY ("analysis_id") REFERENCES "system_analyses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "system_analysis_relationships_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "system_components" table
ALTER TABLE "system_components" ADD CONSTRAINT "system_components_system_component_kinds_kind" FOREIGN KEY ("kind_id") REFERENCES "system_component_kinds" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "system_components_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "system_component_constraints" table
ALTER TABLE "system_component_constraints" ADD CONSTRAINT "system_component_constraints_system_components_component" FOREIGN KEY ("component_id") REFERENCES "system_components" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "system_component_constraints_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "system_component_controls" table
ALTER TABLE "system_component_controls" ADD CONSTRAINT "system_component_controls_system_components_component" FOREIGN KEY ("component_id") REFERENCES "system_components" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "system_component_controls_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "system_component_kinds" table
ALTER TABLE "system_component_kinds" ADD CONSTRAINT "system_component_kinds_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "system_component_relationships" table
ALTER TABLE "system_component_relationships" ADD CONSTRAINT "system_component_relationships_system_components_source" FOREIGN KEY ("source_id") REFERENCES "system_components" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "system_component_relationships_system_components_target" FOREIGN KEY ("target_id") REFERENCES "system_components" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "system_component_relationships_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "system_component_signals" table
ALTER TABLE "system_component_signals" ADD CONSTRAINT "system_component_signals_system_components_component" FOREIGN KEY ("component_id") REFERENCES "system_components" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "system_component_signals_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "system_hazards" table
ALTER TABLE "system_hazards" ADD CONSTRAINT "system_hazards_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "system_relationship_control_actions" table
ALTER TABLE "system_relationship_control_actions" ADD CONSTRAINT "system_relationship_control_ac_12aa32796847b4e25eb84b90c585973c" FOREIGN KEY ("control_id") REFERENCES "system_component_controls" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "system_relationship_control_ac_742bb3020af53fe20337f3c129a7d01d" FOREIGN KEY ("relationship_id") REFERENCES "system_component_relationships" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "system_relationship_control_actions_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "system_relationship_feedback_signals" table
ALTER TABLE "system_relationship_feedback_signals" ADD CONSTRAINT "system_relationship_feedback_s_39c0e6fe83e88c85b812ba9411e28182" FOREIGN KEY ("relationship_id") REFERENCES "system_component_relationships" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "system_relationship_feedback_s_6f55fe5b92f1a576065bd207d9dce675" FOREIGN KEY ("signal_id") REFERENCES "system_component_signals" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "system_relationship_feedback_signals_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "tasks" table
ALTER TABLE "tasks" ADD CONSTRAINT "tasks_incidents_tasks" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "tasks_users_assigned_tasks" FOREIGN KEY ("assignee_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "tasks_users_created_tasks" FOREIGN KEY ("creator_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "teams" table
ALTER TABLE "teams" ADD CONSTRAINT "teams_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "team_memberships" table
ALTER TABLE "team_memberships" ADD CONSTRAINT "team_memberships_teams_team" FOREIGN KEY ("team_id") REFERENCES "teams" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "team_memberships_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "team_memberships_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "tickets" table
ALTER TABLE "tickets" ADD CONSTRAINT "tickets_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "users" table
ALTER TABLE "users" ADD CONSTRAINT "users_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "video_conferences" table
ALTER TABLE "video_conferences" ADD CONSTRAINT "video_conferences_incidents_video_conferences" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "video_conferences_meeting_sessions_video_conference" FOREIGN KEY ("meeting_session_id") REFERENCES "meeting_sessions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "video_conferences_tenants_tenant" FOREIGN KEY ("tenant_id") REFERENCES "tenants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- modify "incident_field_selections" table
ALTER TABLE "incident_field_selections" ADD CONSTRAINT "incident_field_selections_incident_field_option_id" FOREIGN KEY ("incident_field_option_id") REFERENCES "incident_field_options" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "incident_field_selections_incident_id" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "incident_tag_assignments" table
ALTER TABLE "incident_tag_assignments" ADD CONSTRAINT "incident_tag_assignments_incident_id" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "incident_tag_assignments_incident_tag_id" FOREIGN KEY ("incident_tag_id") REFERENCES "incident_tags" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "incident_review_sessions" table
ALTER TABLE "incident_review_sessions" ADD CONSTRAINT "incident_review_sessions_incident_id" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "incident_review_sessions_meeting_session_id" FOREIGN KEY ("meeting_session_id") REFERENCES "meeting_sessions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "incident_debrief_question_incident_fields" table
ALTER TABLE "incident_debrief_question_incident_fields" ADD CONSTRAINT "incident_debrief_question_inci_44abe8f51887ab1da22a39603e050506" FOREIGN KEY ("incident_debrief_question_id") REFERENCES "incident_debrief_questions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "incident_debrief_question_incident_fields_incident_field_id" FOREIGN KEY ("incident_field_id") REFERENCES "incident_fields" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "incident_debrief_question_incident_roles" table
ALTER TABLE "incident_debrief_question_incident_roles" ADD CONSTRAINT "incident_debrief_question_inci_88623030b1280506f5687158ce17d47b" FOREIGN KEY ("incident_debrief_question_id") REFERENCES "incident_debrief_questions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "incident_debrief_question_incident_roles_incident_role_id" FOREIGN KEY ("incident_role_id") REFERENCES "incident_roles" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "incident_debrief_question_incident_severities" table
ALTER TABLE "incident_debrief_question_incident_severities" ADD CONSTRAINT "incident_debrief_question_inci_31f81d29bd5de0dbc30bab30ee6a1ce2" FOREIGN KEY ("incident_severity_id") REFERENCES "incident_severities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "incident_debrief_question_inci_97954e08b0b3c4887d54c7b57902b851" FOREIGN KEY ("incident_debrief_question_id") REFERENCES "incident_debrief_questions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "incident_debrief_question_incident_tags" table
ALTER TABLE "incident_debrief_question_incident_tags" ADD CONSTRAINT "incident_debrief_question_inci_5246f21c867173836779684ae23c83a5" FOREIGN KEY ("incident_debrief_question_id") REFERENCES "incident_debrief_questions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "incident_debrief_question_incident_tags_incident_tag_id" FOREIGN KEY ("incident_tag_id") REFERENCES "incident_tags" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "incident_debrief_question_incident_types" table
ALTER TABLE "incident_debrief_question_incident_types" ADD CONSTRAINT "incident_debrief_question_inci_4140c1ef65a5d052c29594bc82faae77" FOREIGN KEY ("incident_debrief_question_id") REFERENCES "incident_debrief_questions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "incident_debrief_question_incident_types_incident_type_id" FOREIGN KEY ("incident_type_id") REFERENCES "incident_types" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "meeting_schedule_owning_team" table
ALTER TABLE "meeting_schedule_owning_team" ADD CONSTRAINT "meeting_schedule_owning_team_meeting_schedule_id" FOREIGN KEY ("meeting_schedule_id") REFERENCES "meeting_schedules" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "meeting_schedule_owning_team_team_id" FOREIGN KEY ("team_id") REFERENCES "teams" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "oncall_shift_handover_pinned_annotations" table
ALTER TABLE "oncall_shift_handover_pinned_annotations" ADD CONSTRAINT "oncall_shift_handover_pinned_a_ea6451c95975edb633f05ea5a22d6958" FOREIGN KEY ("oncall_shift_handover_id") REFERENCES "oncall_shift_handovers" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "oncall_shift_handover_pinned_annotations_event_annotation_id" FOREIGN KEY ("event_annotation_id") REFERENCES "event_annotations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "playbook_alerts" table
ALTER TABLE "playbook_alerts" ADD CONSTRAINT "playbook_alerts_alert_id" FOREIGN KEY ("alert_id") REFERENCES "alerts" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "playbook_alerts_playbook_id" FOREIGN KEY ("playbook_id") REFERENCES "playbooks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "system_hazard_components" table
ALTER TABLE "system_hazard_components" ADD CONSTRAINT "system_hazard_components_system_component_id" FOREIGN KEY ("system_component_id") REFERENCES "system_components" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "system_hazard_components_system_hazard_id" FOREIGN KEY ("system_hazard_id") REFERENCES "system_hazards" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "system_hazard_constraints" table
ALTER TABLE "system_hazard_constraints" ADD CONSTRAINT "system_hazard_constraints_system_component_constraint_id" FOREIGN KEY ("system_component_constraint_id") REFERENCES "system_component_constraints" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "system_hazard_constraints_system_hazard_id" FOREIGN KEY ("system_hazard_id") REFERENCES "system_hazards" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "system_hazard_relationships" table
ALTER TABLE "system_hazard_relationships" ADD CONSTRAINT "system_hazard_relationships_system_component_relationship_id" FOREIGN KEY ("system_component_relationship_id") REFERENCES "system_component_relationships" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "system_hazard_relationships_system_hazard_id" FOREIGN KEY ("system_hazard_id") REFERENCES "system_hazards" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "task_tickets" table
ALTER TABLE "task_tickets" ADD CONSTRAINT "task_tickets_task_id" FOREIGN KEY ("task_id") REFERENCES "tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "task_tickets_ticket_id" FOREIGN KEY ("ticket_id") REFERENCES "tickets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "team_oncall_rosters" table
ALTER TABLE "team_oncall_rosters" ADD CONSTRAINT "team_oncall_rosters_oncall_roster_id" FOREIGN KEY ("oncall_roster_id") REFERENCES "oncall_rosters" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "team_oncall_rosters_team_id" FOREIGN KEY ("team_id") REFERENCES "teams" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "user_watched_oncall_rosters" table
ALTER TABLE "user_watched_oncall_rosters" ADD CONSTRAINT "user_watched_oncall_rosters_oncall_roster_id" FOREIGN KEY ("oncall_roster_id") REFERENCES "oncall_rosters" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "user_watched_oncall_rosters_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
