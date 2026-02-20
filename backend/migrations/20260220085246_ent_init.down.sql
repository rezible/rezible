-- reverse: modify "user_watched_oncall_rosters" table
ALTER TABLE "user_watched_oncall_rosters" DROP CONSTRAINT "user_watched_oncall_rosters_user_id", DROP CONSTRAINT "user_watched_oncall_rosters_oncall_roster_id";
-- reverse: modify "team_oncall_rosters" table
ALTER TABLE "team_oncall_rosters" DROP CONSTRAINT "team_oncall_rosters_team_id", DROP CONSTRAINT "team_oncall_rosters_oncall_roster_id";
-- reverse: modify "task_tickets" table
ALTER TABLE "task_tickets" DROP CONSTRAINT "task_tickets_ticket_id", DROP CONSTRAINT "task_tickets_task_id";
-- reverse: modify "system_hazard_relationships" table
ALTER TABLE "system_hazard_relationships" DROP CONSTRAINT "system_hazard_relationships_system_hazard_id", DROP CONSTRAINT "system_hazard_relationships_system_component_relationship_id";
-- reverse: modify "system_hazard_constraints" table
ALTER TABLE "system_hazard_constraints" DROP CONSTRAINT "system_hazard_constraints_system_hazard_id", DROP CONSTRAINT "system_hazard_constraints_system_component_constraint_id";
-- reverse: modify "system_hazard_components" table
ALTER TABLE "system_hazard_components" DROP CONSTRAINT "system_hazard_components_system_hazard_id", DROP CONSTRAINT "system_hazard_components_system_component_id";
-- reverse: modify "playbook_alerts" table
ALTER TABLE "playbook_alerts" DROP CONSTRAINT "playbook_alerts_playbook_id", DROP CONSTRAINT "playbook_alerts_alert_id";
-- reverse: modify "oncall_shift_handover_pinned_annotations" table
ALTER TABLE "oncall_shift_handover_pinned_annotations" DROP CONSTRAINT "oncall_shift_handover_pinned_annotations_event_annotation_id", DROP CONSTRAINT "oncall_shift_handover_pinned_a_ea6451c95975edb633f05ea5a22d6958";
-- reverse: modify "meeting_schedule_owning_team" table
ALTER TABLE "meeting_schedule_owning_team" DROP CONSTRAINT "meeting_schedule_owning_team_team_id", DROP CONSTRAINT "meeting_schedule_owning_team_meeting_schedule_id";
-- reverse: modify "incident_debrief_question_incident_types" table
ALTER TABLE "incident_debrief_question_incident_types" DROP CONSTRAINT "incident_debrief_question_incident_types_incident_type_id", DROP CONSTRAINT "incident_debrief_question_inci_4140c1ef65a5d052c29594bc82faae77";
-- reverse: modify "incident_debrief_question_incident_tags" table
ALTER TABLE "incident_debrief_question_incident_tags" DROP CONSTRAINT "incident_debrief_question_incident_tags_incident_tag_id", DROP CONSTRAINT "incident_debrief_question_inci_5246f21c867173836779684ae23c83a5";
-- reverse: modify "incident_debrief_question_incident_severities" table
ALTER TABLE "incident_debrief_question_incident_severities" DROP CONSTRAINT "incident_debrief_question_inci_97954e08b0b3c4887d54c7b57902b851", DROP CONSTRAINT "incident_debrief_question_inci_31f81d29bd5de0dbc30bab30ee6a1ce2";
-- reverse: modify "incident_debrief_question_incident_roles" table
ALTER TABLE "incident_debrief_question_incident_roles" DROP CONSTRAINT "incident_debrief_question_incident_roles_incident_role_id", DROP CONSTRAINT "incident_debrief_question_inci_88623030b1280506f5687158ce17d47b";
-- reverse: modify "incident_debrief_question_incident_fields" table
ALTER TABLE "incident_debrief_question_incident_fields" DROP CONSTRAINT "incident_debrief_question_incident_fields_incident_field_id", DROP CONSTRAINT "incident_debrief_question_inci_44abe8f51887ab1da22a39603e050506";
-- reverse: modify "incident_review_sessions" table
ALTER TABLE "incident_review_sessions" DROP CONSTRAINT "incident_review_sessions_meeting_session_id", DROP CONSTRAINT "incident_review_sessions_incident_id";
-- reverse: modify "incident_tag_assignments" table
ALTER TABLE "incident_tag_assignments" DROP CONSTRAINT "incident_tag_assignments_incident_tag_id", DROP CONSTRAINT "incident_tag_assignments_incident_id";
-- reverse: modify "incident_field_selections" table
ALTER TABLE "incident_field_selections" DROP CONSTRAINT "incident_field_selections_incident_id", DROP CONSTRAINT "incident_field_selections_incident_field_option_id";
-- reverse: modify "video_conferences" table
ALTER TABLE "video_conferences" DROP CONSTRAINT "video_conferences_tenants_tenant", DROP CONSTRAINT "video_conferences_meeting_sessions_video_conference", DROP CONSTRAINT "video_conferences_incidents_video_conferences";
-- reverse: modify "users" table
ALTER TABLE "users" DROP CONSTRAINT "users_tenants_tenant";
-- reverse: modify "tickets" table
ALTER TABLE "tickets" DROP CONSTRAINT "tickets_tenants_tenant";
-- reverse: modify "team_memberships" table
ALTER TABLE "team_memberships" DROP CONSTRAINT "team_memberships_users_user", DROP CONSTRAINT "team_memberships_tenants_tenant", DROP CONSTRAINT "team_memberships_teams_team";
-- reverse: modify "teams" table
ALTER TABLE "teams" DROP CONSTRAINT "teams_tenants_tenant";
-- reverse: modify "tasks" table
ALTER TABLE "tasks" DROP CONSTRAINT "tasks_users_created_tasks", DROP CONSTRAINT "tasks_users_assigned_tasks", DROP CONSTRAINT "tasks_tenants_tenant", DROP CONSTRAINT "tasks_incidents_tasks";
-- reverse: modify "system_relationship_feedback_signals" table
ALTER TABLE "system_relationship_feedback_signals" DROP CONSTRAINT "system_relationship_feedback_signals_tenants_tenant", DROP CONSTRAINT "system_relationship_feedback_s_6f55fe5b92f1a576065bd207d9dce675", DROP CONSTRAINT "system_relationship_feedback_s_39c0e6fe83e88c85b812ba9411e28182";
-- reverse: modify "system_relationship_control_actions" table
ALTER TABLE "system_relationship_control_actions" DROP CONSTRAINT "system_relationship_control_actions_tenants_tenant", DROP CONSTRAINT "system_relationship_control_ac_742bb3020af53fe20337f3c129a7d01d", DROP CONSTRAINT "system_relationship_control_ac_12aa32796847b4e25eb84b90c585973c";
-- reverse: modify "system_hazards" table
ALTER TABLE "system_hazards" DROP CONSTRAINT "system_hazards_tenants_tenant";
-- reverse: modify "system_component_signals" table
ALTER TABLE "system_component_signals" DROP CONSTRAINT "system_component_signals_tenants_tenant", DROP CONSTRAINT "system_component_signals_system_components_component";
-- reverse: modify "system_component_relationships" table
ALTER TABLE "system_component_relationships" DROP CONSTRAINT "system_component_relationships_tenants_tenant", DROP CONSTRAINT "system_component_relationships_system_components_target", DROP CONSTRAINT "system_component_relationships_system_components_source";
-- reverse: modify "system_component_kinds" table
ALTER TABLE "system_component_kinds" DROP CONSTRAINT "system_component_kinds_tenants_tenant";
-- reverse: modify "system_component_controls" table
ALTER TABLE "system_component_controls" DROP CONSTRAINT "system_component_controls_tenants_tenant", DROP CONSTRAINT "system_component_controls_system_components_component";
-- reverse: modify "system_component_constraints" table
ALTER TABLE "system_component_constraints" DROP CONSTRAINT "system_component_constraints_tenants_tenant", DROP CONSTRAINT "system_component_constraints_system_components_component";
-- reverse: modify "system_components" table
ALTER TABLE "system_components" DROP CONSTRAINT "system_components_tenants_tenant", DROP CONSTRAINT "system_components_system_component_kinds_kind";
-- reverse: modify "system_analysis_relationships" table
ALTER TABLE "system_analysis_relationships" DROP CONSTRAINT "system_analysis_relationships_tenants_tenant", DROP CONSTRAINT "system_analysis_relationships_system_analyses_system_analysis", DROP CONSTRAINT "system_analysis_relationships__78049507d24c4a0660356c42d5058c63";
-- reverse: modify "system_analysis_components" table
ALTER TABLE "system_analysis_components" DROP CONSTRAINT "system_analysis_components_tenants_tenant", DROP CONSTRAINT "system_analysis_components_system_components_component", DROP CONSTRAINT "system_analysis_components_system_analyses_analysis";
-- reverse: modify "system_analyses" table
ALTER TABLE "system_analyses" DROP CONSTRAINT "system_analyses_tenants_tenant";
-- reverse: modify "retrospective_reviews" table
ALTER TABLE "retrospective_reviews" DROP CONSTRAINT "retrospective_reviews_users_reviewer", DROP CONSTRAINT "retrospective_reviews_users_requester", DROP CONSTRAINT "retrospective_reviews_tenants_tenant", DROP CONSTRAINT "retrospective_reviews_retrospectives_retrospective", DROP CONSTRAINT "retrospective_reviews_retrospective_comments_comment";
-- reverse: modify "retrospective_comments" table
ALTER TABLE "retrospective_comments" DROP CONSTRAINT "retrospective_comments_users_user", DROP CONSTRAINT "retrospective_comments_tenants_tenant", DROP CONSTRAINT "retrospective_comments_retrospectives_retrospective", DROP CONSTRAINT "retrospective_comments_retrospective_reviews_review";
-- reverse: modify "retrospectives" table
ALTER TABLE "retrospectives" DROP CONSTRAINT "retrospectives_tenants_tenant", DROP CONSTRAINT "retrospectives_system_analyses_retrospective", DROP CONSTRAINT "retrospectives_incidents_retrospective", DROP CONSTRAINT "retrospectives_documents_retrospective";
-- reverse: modify "provider_sync_histories" table
ALTER TABLE "provider_sync_histories" DROP CONSTRAINT "provider_sync_histories_tenants_tenant";
-- reverse: modify "playbooks" table
ALTER TABLE "playbooks" DROP CONSTRAINT "playbooks_tenants_tenant";
-- reverse: modify "organizations" table
ALTER TABLE "organizations" DROP CONSTRAINT "organizations_tenants_tenant";
-- reverse: modify "oncall_shift_metrics" table
ALTER TABLE "oncall_shift_metrics" DROP CONSTRAINT "oncall_shift_metrics_tenants_tenant", DROP CONSTRAINT "oncall_shift_metrics_oncall_shifts_metrics";
-- reverse: modify "oncall_shift_handovers" table
ALTER TABLE "oncall_shift_handovers" DROP CONSTRAINT "oncall_shift_handovers_tenants_tenant", DROP CONSTRAINT "oncall_shift_handovers_oncall_shifts_handover";
-- reverse: modify "oncall_shifts" table
ALTER TABLE "oncall_shifts" DROP CONSTRAINT "oncall_shifts_users_user", DROP CONSTRAINT "oncall_shifts_tenants_tenant", DROP CONSTRAINT "oncall_shifts_oncall_rosters_roster";
-- reverse: modify "oncall_schedule_participants" table
ALTER TABLE "oncall_schedule_participants" DROP CONSTRAINT "oncall_schedule_participants_users_user", DROP CONSTRAINT "oncall_schedule_participants_tenants_tenant", DROP CONSTRAINT "oncall_schedule_participants_oncall_schedules_participants";
-- reverse: modify "oncall_schedules" table
ALTER TABLE "oncall_schedules" DROP CONSTRAINT "oncall_schedules_tenants_tenant", DROP CONSTRAINT "oncall_schedules_oncall_rosters_schedules";
-- reverse: modify "oncall_roster_metrics" table
ALTER TABLE "oncall_roster_metrics" DROP CONSTRAINT "oncall_roster_metrics_tenants_tenant", DROP CONSTRAINT "oncall_roster_metrics_oncall_rosters_roster";
-- reverse: modify "oncall_rosters" table
ALTER TABLE "oncall_rosters" DROP CONSTRAINT "oncall_rosters_tenants_tenant", DROP CONSTRAINT "oncall_rosters_oncall_handover_templates_roster";
-- reverse: modify "oncall_handover_templates" table
ALTER TABLE "oncall_handover_templates" DROP CONSTRAINT "oncall_handover_templates_tenants_tenant";
-- reverse: modify "meeting_sessions" table
ALTER TABLE "meeting_sessions" DROP CONSTRAINT "meeting_sessions_tenants_tenant", DROP CONSTRAINT "meeting_sessions_meeting_schedules_schedule";
-- reverse: modify "meeting_schedules" table
ALTER TABLE "meeting_schedules" DROP CONSTRAINT "meeting_schedules_tenants_tenant";
-- reverse: modify "integrations" table
ALTER TABLE "integrations" DROP CONSTRAINT "integrations_tenants_tenant";
-- reverse: modify "incident_types" table
ALTER TABLE "incident_types" DROP CONSTRAINT "incident_types_tenants_tenant";
-- reverse: modify "incident_tags" table
ALTER TABLE "incident_tags" DROP CONSTRAINT "incident_tags_tenants_tenant";
-- reverse: modify "incident_severities" table
ALTER TABLE "incident_severities" DROP CONSTRAINT "incident_severities_tenants_tenant";
-- reverse: modify "incident_role_assignments" table
ALTER TABLE "incident_role_assignments" DROP CONSTRAINT "incident_role_assignments_users_user", DROP CONSTRAINT "incident_role_assignments_tenants_tenant", DROP CONSTRAINT "incident_role_assignments_incidents_incident", DROP CONSTRAINT "incident_role_assignments_incident_roles_role";
-- reverse: modify "incident_roles" table
ALTER TABLE "incident_roles" DROP CONSTRAINT "incident_roles_tenants_tenant";
-- reverse: modify "incident_milestones" table
ALTER TABLE "incident_milestones" DROP CONSTRAINT "incident_milestones_users_incident_milestones", DROP CONSTRAINT "incident_milestones_tenants_tenant", DROP CONSTRAINT "incident_milestones_incidents_milestones";
-- reverse: modify "incident_links" table
ALTER TABLE "incident_links" DROP CONSTRAINT "incident_links_tenants_tenant", DROP CONSTRAINT "incident_links_incidents_linked_incident", DROP CONSTRAINT "incident_links_incidents_incident";
-- reverse: modify "incident_field_options" table
ALTER TABLE "incident_field_options" DROP CONSTRAINT "incident_field_options_tenants_tenant", DROP CONSTRAINT "incident_field_options_incident_fields_options";
-- reverse: modify "incident_fields" table
ALTER TABLE "incident_fields" DROP CONSTRAINT "incident_fields_tenants_tenant";
-- reverse: modify "incident_event_system_components" table
ALTER TABLE "incident_event_system_components" DROP CONSTRAINT "incident_event_system_components_tenants_tenant", DROP CONSTRAINT "incident_event_system_componen_426e7b9f5e52750ab9f88715a403e203";
-- reverse: modify "incident_event_evidences" table
ALTER TABLE "incident_event_evidences" DROP CONSTRAINT "incident_event_evidences_tenants_tenant", DROP CONSTRAINT "incident_event_evidences_incident_events_evidence";
-- reverse: modify "incident_event_contributing_factors" table
ALTER TABLE "incident_event_contributing_factors" DROP CONSTRAINT "incident_event_contributing_factors_tenants_tenant", DROP CONSTRAINT "incident_event_contributing_factors_incident_events_factors";
-- reverse: modify "incident_event_contexts" table
ALTER TABLE "incident_event_contexts" DROP CONSTRAINT "incident_event_contexts_tenants_tenant", DROP CONSTRAINT "incident_event_contexts_incident_events_context";
-- reverse: modify "incident_events" table
ALTER TABLE "incident_events" DROP CONSTRAINT "incident_events_tenants_tenant", DROP CONSTRAINT "incident_events_incidents_events", DROP CONSTRAINT "incident_events_events_event";
-- reverse: modify "incident_debrief_suggestions" table
ALTER TABLE "incident_debrief_suggestions" DROP CONSTRAINT "incident_debrief_suggestions_tenants_tenant", DROP CONSTRAINT "incident_debrief_suggestions_incident_debriefs_suggestions";
-- reverse: modify "incident_debrief_questions" table
ALTER TABLE "incident_debrief_questions" DROP CONSTRAINT "incident_debrief_questions_tenants_tenant";
-- reverse: modify "incident_debrief_messages" table
ALTER TABLE "incident_debrief_messages" DROP CONSTRAINT "incident_debrief_messages_tenants_tenant", DROP CONSTRAINT "incident_debrief_messages_incident_debriefs_messages", DROP CONSTRAINT "incident_debrief_messages_inci_0d1f0b105ef851edb04b442b34a0d17f";
-- reverse: modify "incident_debriefs" table
ALTER TABLE "incident_debriefs" DROP CONSTRAINT "incident_debriefs_users_incident_debriefs", DROP CONSTRAINT "incident_debriefs_tenants_tenant", DROP CONSTRAINT "incident_debriefs_incidents_debriefs";
-- reverse: modify "incidents" table
ALTER TABLE "incidents" DROP CONSTRAINT "incidents_tenants_tenant", DROP CONSTRAINT "incidents_incident_types_type", DROP CONSTRAINT "incidents_incident_severities_severity";
-- reverse: modify "event_annotations" table
ALTER TABLE "event_annotations" DROP CONSTRAINT "event_annotations_users_creator", DROP CONSTRAINT "event_annotations_tenants_tenant", DROP CONSTRAINT "event_annotations_events_event";
-- reverse: modify "events" table
ALTER TABLE "events" DROP CONSTRAINT "events_tenants_tenant";
-- reverse: modify "documents" table
ALTER TABLE "documents" DROP CONSTRAINT "documents_tenants_tenant";
-- reverse: modify "alert_instances" table
ALTER TABLE "alert_instances" DROP CONSTRAINT "alert_instances_tenants_tenant", DROP CONSTRAINT "alert_instances_events_event", DROP CONSTRAINT "alert_instances_alerts_instances", DROP CONSTRAINT "alert_instances_alerts_alert", DROP CONSTRAINT "alert_instances_alert_feedbacks_feedback";
-- reverse: modify "alert_feedbacks" table
ALTER TABLE "alert_feedbacks" DROP CONSTRAINT "alert_feedbacks_tenants_tenant", DROP CONSTRAINT "alert_feedbacks_alert_instances_alert_instance";
-- reverse: modify "alerts" table
ALTER TABLE "alerts" DROP CONSTRAINT "alerts_tenants_tenant", DROP CONSTRAINT "alerts_oncall_rosters_alerts";
-- reverse: create "user_watched_oncall_rosters" table
DROP TABLE "user_watched_oncall_rosters";
-- reverse: create "team_oncall_rosters" table
DROP TABLE "team_oncall_rosters";
-- reverse: create "task_tickets" table
DROP TABLE "task_tickets";
-- reverse: create "system_hazard_relationships" table
DROP TABLE "system_hazard_relationships";
-- reverse: create "system_hazard_constraints" table
DROP TABLE "system_hazard_constraints";
-- reverse: create "system_hazard_components" table
DROP TABLE "system_hazard_components";
-- reverse: create "playbook_alerts" table
DROP TABLE "playbook_alerts";
-- reverse: create "oncall_shift_handover_pinned_annotations" table
DROP TABLE "oncall_shift_handover_pinned_annotations";
-- reverse: create "meeting_schedule_owning_team" table
DROP TABLE "meeting_schedule_owning_team";
-- reverse: create "incident_debrief_question_incident_types" table
DROP TABLE "incident_debrief_question_incident_types";
-- reverse: create "incident_debrief_question_incident_tags" table
DROP TABLE "incident_debrief_question_incident_tags";
-- reverse: create "incident_debrief_question_incident_severities" table
DROP TABLE "incident_debrief_question_incident_severities";
-- reverse: create "incident_debrief_question_incident_roles" table
DROP TABLE "incident_debrief_question_incident_roles";
-- reverse: create "incident_debrief_question_incident_fields" table
DROP TABLE "incident_debrief_question_incident_fields";
-- reverse: create "incident_review_sessions" table
DROP TABLE "incident_review_sessions";
-- reverse: create "incident_tag_assignments" table
DROP TABLE "incident_tag_assignments";
-- reverse: create "incident_field_selections" table
DROP TABLE "incident_field_selections";
-- reverse: create index "videoconference_tenant_id" to table: "video_conferences"
DROP INDEX "videoconference_tenant_id";
-- reverse: create index "videoconference_meeting_session_id_status" to table: "video_conferences"
DROP INDEX "videoconference_meeting_session_id_status";
-- reverse: create index "videoconference_incident_id_status" to table: "video_conferences"
DROP INDEX "videoconference_incident_id_status";
-- reverse: create index "video_conferences_meeting_session_id_key" to table: "video_conferences"
DROP INDEX "video_conferences_meeting_session_id_key";
-- reverse: create "video_conferences" table
DROP TABLE "video_conferences";
-- reverse: create index "user_tenant_id" to table: "users"
DROP INDEX "user_tenant_id";
-- reverse: create "users" table
DROP TABLE "users";
-- reverse: create index "ticket_tenant_id" to table: "tickets"
DROP INDEX "ticket_tenant_id";
-- reverse: create "tickets" table
DROP TABLE "tickets";
-- reverse: create "tenants" table
DROP TABLE "tenants";
-- reverse: create index "teammembership_tenant_id" to table: "team_memberships"
DROP INDEX "teammembership_tenant_id";
-- reverse: create index "teammembership_team_id_user_id" to table: "team_memberships"
DROP INDEX "teammembership_team_id_user_id";
-- reverse: create "team_memberships" table
DROP TABLE "team_memberships";
-- reverse: create index "teams_slug_key" to table: "teams"
DROP INDEX "teams_slug_key";
-- reverse: create index "team_tenant_id" to table: "teams"
DROP INDEX "team_tenant_id";
-- reverse: create "teams" table
DROP TABLE "teams";
-- reverse: create index "task_tenant_id" to table: "tasks"
DROP INDEX "task_tenant_id";
-- reverse: create "tasks" table
DROP TABLE "tasks";
-- reverse: create index "systemrelationshipfeedbacksignal_tenant_id" to table: "system_relationship_feedback_signals"
DROP INDEX "systemrelationshipfeedbacksignal_tenant_id";
-- reverse: create index "systemrelationshipfeedbacksignal_relationship_id_signal_id" to table: "system_relationship_feedback_signals"
DROP INDEX "systemrelationshipfeedbacksignal_relationship_id_signal_id";
-- reverse: create "system_relationship_feedback_signals" table
DROP TABLE "system_relationship_feedback_signals";
-- reverse: create index "systemrelationshipcontrolaction_tenant_id" to table: "system_relationship_control_actions"
DROP INDEX "systemrelationshipcontrolaction_tenant_id";
-- reverse: create index "systemrelationshipcontrolaction_relationship_id_control_id" to table: "system_relationship_control_actions"
DROP INDEX "systemrelationshipcontrolaction_relationship_id_control_id";
-- reverse: create "system_relationship_control_actions" table
DROP TABLE "system_relationship_control_actions";
-- reverse: create index "systemhazard_tenant_id" to table: "system_hazards"
DROP INDEX "systemhazard_tenant_id";
-- reverse: create "system_hazards" table
DROP TABLE "system_hazards";
-- reverse: create index "systemcomponentsignal_tenant_id" to table: "system_component_signals"
DROP INDEX "systemcomponentsignal_tenant_id";
-- reverse: create "system_component_signals" table
DROP TABLE "system_component_signals";
-- reverse: create index "systemcomponentrelationship_tenant_id" to table: "system_component_relationships"
DROP INDEX "systemcomponentrelationship_tenant_id";
-- reverse: create index "systemcomponentrelationship_source_id_target_id" to table: "system_component_relationships"
DROP INDEX "systemcomponentrelationship_source_id_target_id";
-- reverse: create "system_component_relationships" table
DROP TABLE "system_component_relationships";
-- reverse: create index "systemcomponentkind_tenant_id" to table: "system_component_kinds"
DROP INDEX "systemcomponentkind_tenant_id";
-- reverse: create "system_component_kinds" table
DROP TABLE "system_component_kinds";
-- reverse: create index "systemcomponentcontrol_tenant_id" to table: "system_component_controls"
DROP INDEX "systemcomponentcontrol_tenant_id";
-- reverse: create "system_component_controls" table
DROP TABLE "system_component_controls";
-- reverse: create index "systemcomponentconstraint_tenant_id" to table: "system_component_constraints"
DROP INDEX "systemcomponentconstraint_tenant_id";
-- reverse: create "system_component_constraints" table
DROP TABLE "system_component_constraints";
-- reverse: create index "systemcomponent_tenant_id" to table: "system_components"
DROP INDEX "systemcomponent_tenant_id";
-- reverse: create "system_components" table
DROP TABLE "system_components";
-- reverse: create index "systemanalysisrelationship_tenant_id" to table: "system_analysis_relationships"
DROP INDEX "systemanalysisrelationship_tenant_id";
-- reverse: create "system_analysis_relationships" table
DROP TABLE "system_analysis_relationships";
-- reverse: create index "systemanalysiscomponent_tenant_id" to table: "system_analysis_components"
DROP INDEX "systemanalysiscomponent_tenant_id";
-- reverse: create index "systemanalysiscomponent_component_id_analysis_id" to table: "system_analysis_components"
DROP INDEX "systemanalysiscomponent_component_id_analysis_id";
-- reverse: create "system_analysis_components" table
DROP TABLE "system_analysis_components";
-- reverse: create index "systemanalysis_tenant_id" to table: "system_analyses"
DROP INDEX "systemanalysis_tenant_id";
-- reverse: create "system_analyses" table
DROP TABLE "system_analyses";
-- reverse: create index "retrospectivereview_tenant_id" to table: "retrospective_reviews"
DROP INDEX "retrospectivereview_tenant_id";
-- reverse: create "retrospective_reviews" table
DROP TABLE "retrospective_reviews";
-- reverse: create index "retrospectivecomment_tenant_id" to table: "retrospective_comments"
DROP INDEX "retrospectivecomment_tenant_id";
-- reverse: create "retrospective_comments" table
DROP TABLE "retrospective_comments";
-- reverse: create index "retrospectives_system_analysis_id_key" to table: "retrospectives"
DROP INDEX "retrospectives_system_analysis_id_key";
-- reverse: create index "retrospectives_incident_id_key" to table: "retrospectives"
DROP INDEX "retrospectives_incident_id_key";
-- reverse: create index "retrospectives_document_id_key" to table: "retrospectives"
DROP INDEX "retrospectives_document_id_key";
-- reverse: create index "retrospective_tenant_id" to table: "retrospectives"
DROP INDEX "retrospective_tenant_id";
-- reverse: create "retrospectives" table
DROP TABLE "retrospectives";
-- reverse: create index "providersynchistory_tenant_id" to table: "provider_sync_histories"
DROP INDEX "providersynchistory_tenant_id";
-- reverse: create "provider_sync_histories" table
DROP TABLE "provider_sync_histories";
-- reverse: create index "playbook_tenant_id" to table: "playbooks"
DROP INDEX "playbook_tenant_id";
-- reverse: create "playbooks" table
DROP TABLE "playbooks";
-- reverse: create index "organization_tenant_id" to table: "organizations"
DROP INDEX "organization_tenant_id";
-- reverse: create "organizations" table
DROP TABLE "organizations";
-- reverse: create index "oncallshiftmetrics_tenant_id" to table: "oncall_shift_metrics"
DROP INDEX "oncallshiftmetrics_tenant_id";
-- reverse: create index "oncall_shift_metrics_shift_id_key" to table: "oncall_shift_metrics"
DROP INDEX "oncall_shift_metrics_shift_id_key";
-- reverse: create "oncall_shift_metrics" table
DROP TABLE "oncall_shift_metrics";
-- reverse: create index "oncallshifthandover_tenant_id" to table: "oncall_shift_handovers"
DROP INDEX "oncallshifthandover_tenant_id";
-- reverse: create index "oncall_shift_handovers_shift_id_key" to table: "oncall_shift_handovers"
DROP INDEX "oncall_shift_handovers_shift_id_key";
-- reverse: create "oncall_shift_handovers" table
DROP TABLE "oncall_shift_handovers";
-- reverse: create index "oncallshift_tenant_id" to table: "oncall_shifts"
DROP INDEX "oncallshift_tenant_id";
-- reverse: create index "oncall_shifts_primary_shift_id_key" to table: "oncall_shifts"
DROP INDEX "oncall_shifts_primary_shift_id_key";
-- reverse: create "oncall_shifts" table
DROP TABLE "oncall_shifts";
-- reverse: create index "oncallscheduleparticipant_tenant_id" to table: "oncall_schedule_participants"
DROP INDEX "oncallscheduleparticipant_tenant_id";
-- reverse: create "oncall_schedule_participants" table
DROP TABLE "oncall_schedule_participants";
-- reverse: create index "oncallschedule_tenant_id" to table: "oncall_schedules"
DROP INDEX "oncallschedule_tenant_id";
-- reverse: create "oncall_schedules" table
DROP TABLE "oncall_schedules";
-- reverse: create index "oncallrostermetrics_tenant_id" to table: "oncall_roster_metrics"
DROP INDEX "oncallrostermetrics_tenant_id";
-- reverse: create "oncall_roster_metrics" table
DROP TABLE "oncall_roster_metrics";
-- reverse: create index "oncallroster_tenant_id" to table: "oncall_rosters"
DROP INDEX "oncallroster_tenant_id";
-- reverse: create index "oncall_rosters_slug_key" to table: "oncall_rosters"
DROP INDEX "oncall_rosters_slug_key";
-- reverse: create "oncall_rosters" table
DROP TABLE "oncall_rosters";
-- reverse: create index "oncallhandovertemplate_tenant_id" to table: "oncall_handover_templates"
DROP INDEX "oncallhandovertemplate_tenant_id";
-- reverse: create "oncall_handover_templates" table
DROP TABLE "oncall_handover_templates";
-- reverse: create index "meetingsession_tenant_id" to table: "meeting_sessions"
DROP INDEX "meetingsession_tenant_id";
-- reverse: create "meeting_sessions" table
DROP TABLE "meeting_sessions";
-- reverse: create index "meetingschedule_tenant_id" to table: "meeting_schedules"
DROP INDEX "meetingschedule_tenant_id";
-- reverse: create "meeting_schedules" table
DROP TABLE "meeting_schedules";
-- reverse: create index "integration_tenant_id_name" to table: "integrations"
DROP INDEX "integration_tenant_id_name";
-- reverse: create index "integration_tenant_id" to table: "integrations"
DROP INDEX "integration_tenant_id";
-- reverse: create "integrations" table
DROP TABLE "integrations";
-- reverse: create index "incidenttype_tenant_id" to table: "incident_types"
DROP INDEX "incidenttype_tenant_id";
-- reverse: create "incident_types" table
DROP TABLE "incident_types";
-- reverse: create index "incidenttag_tenant_id" to table: "incident_tags"
DROP INDEX "incidenttag_tenant_id";
-- reverse: create "incident_tags" table
DROP TABLE "incident_tags";
-- reverse: create index "incidentseverity_tenant_id" to table: "incident_severities"
DROP INDEX "incidentseverity_tenant_id";
-- reverse: create "incident_severities" table
DROP TABLE "incident_severities";
-- reverse: create index "incidentroleassignment_user_id_incident_id" to table: "incident_role_assignments"
DROP INDEX "incidentroleassignment_user_id_incident_id";
-- reverse: create index "incidentroleassignment_tenant_id" to table: "incident_role_assignments"
DROP INDEX "incidentroleassignment_tenant_id";
-- reverse: create "incident_role_assignments" table
DROP TABLE "incident_role_assignments";
-- reverse: create index "incidentrole_tenant_id" to table: "incident_roles"
DROP INDEX "incidentrole_tenant_id";
-- reverse: create "incident_roles" table
DROP TABLE "incident_roles";
-- reverse: create index "incidentmilestone_tenant_id" to table: "incident_milestones"
DROP INDEX "incidentmilestone_tenant_id";
-- reverse: create "incident_milestones" table
DROP TABLE "incident_milestones";
-- reverse: create index "incidentlink_tenant_id" to table: "incident_links"
DROP INDEX "incidentlink_tenant_id";
-- reverse: create index "incidentlink_incident_id_linked_incident_id" to table: "incident_links"
DROP INDEX "incidentlink_incident_id_linked_incident_id";
-- reverse: create "incident_links" table
DROP TABLE "incident_links";
-- reverse: create index "incidentfieldoption_tenant_id" to table: "incident_field_options"
DROP INDEX "incidentfieldoption_tenant_id";
-- reverse: create "incident_field_options" table
DROP TABLE "incident_field_options";
-- reverse: create index "incidentfield_tenant_id" to table: "incident_fields"
DROP INDEX "incidentfield_tenant_id";
-- reverse: create "incident_fields" table
DROP TABLE "incident_fields";
-- reverse: create index "incidenteventsystemcomponent_tenant_id" to table: "incident_event_system_components"
DROP INDEX "incidenteventsystemcomponent_tenant_id";
-- reverse: create index "incidenteventsystemcomponent_i_68243ec125e2acebf985bc112b82147a" to table: "incident_event_system_components"
DROP INDEX "incidenteventsystemcomponent_i_68243ec125e2acebf985bc112b82147a";
-- reverse: create index "incident_event_system_components_incident_event_id_key" to table: "incident_event_system_components"
DROP INDEX "incident_event_system_components_incident_event_id_key";
-- reverse: create "incident_event_system_components" table
DROP TABLE "incident_event_system_components";
-- reverse: create index "incidenteventevidence_tenant_id" to table: "incident_event_evidences"
DROP INDEX "incidenteventevidence_tenant_id";
-- reverse: create "incident_event_evidences" table
DROP TABLE "incident_event_evidences";
-- reverse: create index "incidenteventcontributingfactor_tenant_id" to table: "incident_event_contributing_factors"
DROP INDEX "incidenteventcontributingfactor_tenant_id";
-- reverse: create "incident_event_contributing_factors" table
DROP TABLE "incident_event_contributing_factors";
-- reverse: create index "incidenteventcontext_tenant_id" to table: "incident_event_contexts"
DROP INDEX "incidenteventcontext_tenant_id";
-- reverse: create index "incident_event_contexts_incident_event_context_key" to table: "incident_event_contexts"
DROP INDEX "incident_event_contexts_incident_event_context_key";
-- reverse: create "incident_event_contexts" table
DROP TABLE "incident_event_contexts";
-- reverse: create index "incidentevent_tenant_id" to table: "incident_events"
DROP INDEX "incidentevent_tenant_id";
-- reverse: create index "incidentevent_kind" to table: "incident_events"
DROP INDEX "incidentevent_kind";
-- reverse: create "incident_events" table
DROP TABLE "incident_events";
-- reverse: create index "incidentdebriefsuggestion_tenant_id" to table: "incident_debrief_suggestions"
DROP INDEX "incidentdebriefsuggestion_tenant_id";
-- reverse: create "incident_debrief_suggestions" table
DROP TABLE "incident_debrief_suggestions";
-- reverse: create index "incidentdebriefquestion_tenant_id" to table: "incident_debrief_questions"
DROP INDEX "incidentdebriefquestion_tenant_id";
-- reverse: create "incident_debrief_questions" table
DROP TABLE "incident_debrief_questions";
-- reverse: create index "incidentdebriefmessage_tenant_id" to table: "incident_debrief_messages"
DROP INDEX "incidentdebriefmessage_tenant_id";
-- reverse: create "incident_debrief_messages" table
DROP TABLE "incident_debrief_messages";
-- reverse: create index "incidentdebrief_tenant_id" to table: "incident_debriefs"
DROP INDEX "incidentdebrief_tenant_id";
-- reverse: create "incident_debriefs" table
DROP TABLE "incident_debriefs";
-- reverse: create index "incidents_slug_key" to table: "incidents"
DROP INDEX "incidents_slug_key";
-- reverse: create index "incident_tenant_id" to table: "incidents"
DROP INDEX "incident_tenant_id";
-- reverse: create "incidents" table
DROP TABLE "incidents";
-- reverse: create index "eventannotation_tenant_id" to table: "event_annotations"
DROP INDEX "eventannotation_tenant_id";
-- reverse: create "event_annotations" table
DROP TABLE "event_annotations";
-- reverse: create index "event_tenant_id" to table: "events"
DROP INDEX "event_tenant_id";
-- reverse: create "events" table
DROP TABLE "events";
-- reverse: create index "document_tenant_id" to table: "documents"
DROP INDEX "document_tenant_id";
-- reverse: create "documents" table
DROP TABLE "documents";
-- reverse: create index "alertinstance_tenant_id" to table: "alert_instances"
DROP INDEX "alertinstance_tenant_id";
-- reverse: create "alert_instances" table
DROP TABLE "alert_instances";
-- reverse: create index "alertfeedback_tenant_id" to table: "alert_feedbacks"
DROP INDEX "alertfeedback_tenant_id";
-- reverse: create "alert_feedbacks" table
DROP TABLE "alert_feedbacks";
-- reverse: create index "alert_tenant_id" to table: "alerts"
DROP INDEX "alert_tenant_id";
-- reverse: create "alerts" table
DROP TABLE "alerts";
