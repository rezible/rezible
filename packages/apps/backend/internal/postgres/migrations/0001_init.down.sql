-- reverse: modify "user_watched_oncall_rosters" table
ALTER TABLE "user_watched_oncall_rosters" DROP CONSTRAINT "user_watched_oncall_rosters_oncall_roster_id", DROP CONSTRAINT "user_watched_oncall_rosters_user_id";
-- reverse: modify "team_oncall_rosters" table
ALTER TABLE "team_oncall_rosters" DROP CONSTRAINT "team_oncall_rosters_oncall_roster_id", DROP CONSTRAINT "team_oncall_rosters_team_id";
-- reverse: modify "task_tickets" table
ALTER TABLE "task_tickets" DROP CONSTRAINT "task_tickets_ticket_id", DROP CONSTRAINT "task_tickets_task_id";
-- reverse: modify "playbook_alerts" table
ALTER TABLE "playbook_alerts" DROP CONSTRAINT "playbook_alerts_alert_id", DROP CONSTRAINT "playbook_alerts_playbook_id";
-- reverse: modify "oncall_shift_handover_pinned_annotations" table
ALTER TABLE "oncall_shift_handover_pinned_annotations" DROP CONSTRAINT "oncall_shift_handover_pinned_annotations_event_annotation_id", DROP CONSTRAINT "oncall_shift_handover_pinned_a_ea6451c95975edb633f05ea5a22d6958";
-- reverse: modify "meeting_schedule_owning_team" table
ALTER TABLE "meeting_schedule_owning_team" DROP CONSTRAINT "meeting_schedule_owning_team_team_id", DROP CONSTRAINT "meeting_schedule_owning_team_meeting_schedule_id";
-- reverse: modify "incident_debrief_question_incident_types" table
ALTER TABLE "incident_debrief_question_incident_types" DROP CONSTRAINT "incident_debrief_question_incident_types_incident_type_id", DROP CONSTRAINT "incident_debrief_question_inci_4140c1ef65a5d052c29594bc82faae77";
-- reverse: modify "incident_debrief_question_incident_tags" table
ALTER TABLE "incident_debrief_question_incident_tags" DROP CONSTRAINT "incident_debrief_question_incident_tags_incident_tag_id", DROP CONSTRAINT "incident_debrief_question_inci_5246f21c867173836779684ae23c83a5";
-- reverse: modify "incident_debrief_question_incident_severities" table
ALTER TABLE "incident_debrief_question_incident_severities" DROP CONSTRAINT "incident_debrief_question_inci_31f81d29bd5de0dbc30bab30ee6a1ce2", DROP CONSTRAINT "incident_debrief_question_inci_97954e08b0b3c4887d54c7b57902b851";
-- reverse: modify "incident_debrief_question_incident_roles" table
ALTER TABLE "incident_debrief_question_incident_roles" DROP CONSTRAINT "incident_debrief_question_incident_roles_incident_role_id", DROP CONSTRAINT "incident_debrief_question_inci_88623030b1280506f5687158ce17d47b";
-- reverse: modify "incident_debrief_question_incident_fields" table
ALTER TABLE "incident_debrief_question_incident_fields" DROP CONSTRAINT "incident_debrief_question_incident_fields_incident_field_id", DROP CONSTRAINT "incident_debrief_question_inci_44abe8f51887ab1da22a39603e050506";
-- reverse: modify "incident_review_sessions" table
ALTER TABLE "incident_review_sessions" DROP CONSTRAINT "incident_review_sessions_meeting_session_id", DROP CONSTRAINT "incident_review_sessions_incident_id";
-- reverse: modify "incident_tag_assignments" table
ALTER TABLE "incident_tag_assignments" DROP CONSTRAINT "incident_tag_assignments_incident_tag_id", DROP CONSTRAINT "incident_tag_assignments_incident_id";
-- reverse: modify "incident_field_selections" table
ALTER TABLE "incident_field_selections" DROP CONSTRAINT "incident_field_selections_incident_field_option_id", DROP CONSTRAINT "incident_field_selections_incident_id";
-- reverse: modify "video_conferences" table
ALTER TABLE "video_conferences" DROP CONSTRAINT "video_conferences_tenants_tenant", DROP CONSTRAINT "video_conferences_meeting_sessions_video_conference", DROP CONSTRAINT "video_conferences_incidents_video_conferences";
-- reverse: modify "users" table
ALTER TABLE "users" DROP CONSTRAINT "users_tenants_tenant";
-- reverse: modify "tickets" table
ALTER TABLE "tickets" DROP CONSTRAINT "tickets_tenants_tenant";
-- reverse: modify "team_memberships" table
ALTER TABLE "team_memberships" DROP CONSTRAINT "team_memberships_users_user", DROP CONSTRAINT "team_memberships_teams_team", DROP CONSTRAINT "team_memberships_tenants_tenant";
-- reverse: modify "teams" table
ALTER TABLE "teams" DROP CONSTRAINT "teams_tenants_tenant";
-- reverse: modify "tasks" table
ALTER TABLE "tasks" DROP CONSTRAINT "tasks_users_created_tasks", DROP CONSTRAINT "tasks_users_assigned_tasks", DROP CONSTRAINT "tasks_tenants_tenant", DROP CONSTRAINT "tasks_incidents_tasks";
-- reverse: modify "system_topology_snapshot_relationships" table
ALTER TABLE "system_topology_snapshot_relationships" DROP CONSTRAINT "system_topology_snapshot_relat_49bd6a99a61ed2571218b82bfd309a9d", DROP CONSTRAINT "system_topology_snapshot_relat_020fdabf27aca33de3b8244fcbb40d4d", DROP CONSTRAINT "system_topology_snapshot_relat_8618d8e705fc800d4b45f7ba21b7151f", DROP CONSTRAINT "system_topology_snapshot_relat_5f2c55c7b6ad86bc36bd1003d1e2579f", DROP CONSTRAINT "system_topology_snapshot_relationships_tenants_tenant";
-- reverse: modify "system_topology_snapshot_entities" table
ALTER TABLE "system_topology_snapshot_entities" DROP CONSTRAINT "system_topology_snapshot_entit_6acf350be9828a2b8ba4f3166002a8fd", DROP CONSTRAINT "system_topology_snapshot_entit_624ed73670c571ebba5995e15ccf51fd", DROP CONSTRAINT "system_topology_snapshot_entities_tenants_tenant";
-- reverse: modify "system_topology_snapshots" table
ALTER TABLE "system_topology_snapshots" DROP CONSTRAINT "system_topology_snapshots_tenants_tenant";
-- reverse: modify "system_analysis_topology_nodes" table
ALTER TABLE "system_analysis_topology_nodes" DROP CONSTRAINT "system_analysis_topology_nodes_1dea46edc4c4b6fa57943e6a9c4a3f02", DROP CONSTRAINT "system_analysis_topology_nodes_system_analyses_analysis", DROP CONSTRAINT "system_analysis_topology_nodes_tenants_tenant";
-- reverse: modify "system_analysis_topology_edges" table
ALTER TABLE "system_analysis_topology_edges" DROP CONSTRAINT "system_analysis_topology_edges_c4b40b33a79f2054fdf3e995173eda08", DROP CONSTRAINT "system_analysis_topology_edges_system_analyses_analysis", DROP CONSTRAINT "system_analysis_topology_edges_tenants_tenant";
-- reverse: modify "system_analyses" table
ALTER TABLE "system_analyses" DROP CONSTRAINT "system_analyses_system_topology_snapshots_topology_snapshot", DROP CONSTRAINT "system_analyses_tenants_tenant";
-- reverse: modify "retrospective_reviews" table
ALTER TABLE "retrospective_reviews" DROP CONSTRAINT "retrospective_reviews_retrospective_comments_comment", DROP CONSTRAINT "retrospective_reviews_users_reviewer", DROP CONSTRAINT "retrospective_reviews_users_requester", DROP CONSTRAINT "retrospective_reviews_retrospectives_retrospective", DROP CONSTRAINT "retrospective_reviews_tenants_tenant";
-- reverse: modify "retrospective_comments" table
ALTER TABLE "retrospective_comments" DROP CONSTRAINT "retrospective_comments_retrospective_reviews_review", DROP CONSTRAINT "retrospective_comments_users_user", DROP CONSTRAINT "retrospective_comments_retrospectives_retrospective", DROP CONSTRAINT "retrospective_comments_tenants_tenant";
-- reverse: modify "retrospectives" table
ALTER TABLE "retrospectives" DROP CONSTRAINT "retrospectives_system_analyses_retrospective", DROP CONSTRAINT "retrospectives_tenants_tenant", DROP CONSTRAINT "retrospectives_incidents_retrospective", DROP CONSTRAINT "retrospectives_documents_retrospective";
-- reverse: modify "provider_event_sync_runs" table
ALTER TABLE "provider_event_sync_runs" DROP CONSTRAINT "provider_event_sync_runs_tenants_tenant";
-- reverse: modify "provider_event_sync_cursors" table
ALTER TABLE "provider_event_sync_cursors" DROP CONSTRAINT "provider_event_sync_cursors_tenants_tenant";
-- reverse: modify "playbooks" table
ALTER TABLE "playbooks" DROP CONSTRAINT "playbooks_tenants_tenant";
-- reverse: modify "organization_roles" table
ALTER TABLE "organization_roles" DROP CONSTRAINT "organization_roles_users_organization_role", DROP CONSTRAINT "organization_roles_organizations_organization", DROP CONSTRAINT "organization_roles_tenants_tenant";
-- reverse: modify "organizations" table
ALTER TABLE "organizations" DROP CONSTRAINT "organizations_tenants_tenant";
-- reverse: modify "oncall_shift_metrics" table
ALTER TABLE "oncall_shift_metrics" DROP CONSTRAINT "oncall_shift_metrics_tenants_tenant", DROP CONSTRAINT "oncall_shift_metrics_oncall_shifts_metrics";
-- reverse: modify "oncall_shift_handovers" table
ALTER TABLE "oncall_shift_handovers" DROP CONSTRAINT "oncall_shift_handovers_tenants_tenant", DROP CONSTRAINT "oncall_shift_handovers_oncall_shifts_handover";
-- reverse: modify "oncall_shifts" table
ALTER TABLE "oncall_shifts" DROP CONSTRAINT "oncall_shifts_oncall_rosters_roster", DROP CONSTRAINT "oncall_shifts_users_user", DROP CONSTRAINT "oncall_shifts_tenants_tenant";
-- reverse: modify "oncall_schedule_participants" table
ALTER TABLE "oncall_schedule_participants" DROP CONSTRAINT "oncall_schedule_participants_users_user", DROP CONSTRAINT "oncall_schedule_participants_tenants_tenant", DROP CONSTRAINT "oncall_schedule_participants_oncall_schedules_participants";
-- reverse: modify "oncall_schedules" table
ALTER TABLE "oncall_schedules" DROP CONSTRAINT "oncall_schedules_tenants_tenant", DROP CONSTRAINT "oncall_schedules_oncall_rosters_schedules";
-- reverse: modify "oncall_roster_metrics" table
ALTER TABLE "oncall_roster_metrics" DROP CONSTRAINT "oncall_roster_metrics_oncall_rosters_roster", DROP CONSTRAINT "oncall_roster_metrics_tenants_tenant";
-- reverse: modify "oncall_rosters" table
ALTER TABLE "oncall_rosters" DROP CONSTRAINT "oncall_rosters_tenants_tenant", DROP CONSTRAINT "oncall_rosters_oncall_handover_templates_roster";
-- reverse: modify "oncall_handover_templates" table
ALTER TABLE "oncall_handover_templates" DROP CONSTRAINT "oncall_handover_templates_tenants_tenant";
-- reverse: modify "normalized_event_projection_status" table
ALTER TABLE "normalized_event_projection_status" DROP CONSTRAINT "normalized_event_projection_st_57b31f9b9ba804f03db1c8815e863e31", DROP CONSTRAINT "normalized_event_projection_status_tenants_tenant";
-- reverse: modify "normalized_events" table
ALTER TABLE "normalized_events" DROP CONSTRAINT "normalized_events_tenants_tenant";
-- reverse: modify "meeting_sessions" table
ALTER TABLE "meeting_sessions" DROP CONSTRAINT "meeting_sessions_meeting_schedules_schedule", DROP CONSTRAINT "meeting_sessions_tenants_tenant";
-- reverse: modify "meeting_schedules" table
ALTER TABLE "meeting_schedules" DROP CONSTRAINT "meeting_schedules_tenants_tenant";
-- reverse: modify "knowledge_relationships" table
ALTER TABLE "knowledge_relationships" DROP CONSTRAINT "knowledge_relationships_knowledge_entities_target_entity", DROP CONSTRAINT "knowledge_relationships_knowledge_entities_source_entity", DROP CONSTRAINT "knowledge_relationships_tenants_tenant";
-- reverse: modify "knowledge_evidences" table
ALTER TABLE "knowledge_evidences" DROP CONSTRAINT "knowledge_evidences_normalized_events_normalized_event", DROP CONSTRAINT "knowledge_evidences_knowledge_entity_alias_alias", DROP CONSTRAINT "knowledge_evidences_knowledge_relationships_relationship", DROP CONSTRAINT "knowledge_evidences_knowledge_entities_entity", DROP CONSTRAINT "knowledge_evidences_tenants_tenant";
-- reverse: modify "knowledge_entity_alias" table
ALTER TABLE "knowledge_entity_alias" DROP CONSTRAINT "knowledge_entity_alias_knowledge_entities_entity", DROP CONSTRAINT "knowledge_entity_alias_tenants_tenant";
-- reverse: modify "knowledge_entities" table
ALTER TABLE "knowledge_entities" DROP CONSTRAINT "knowledge_entities_tenants_tenant";
-- reverse: modify "integration_oauth_states" table
ALTER TABLE "integration_oauth_states" DROP CONSTRAINT "integration_oauth_states_users_user", DROP CONSTRAINT "integration_oauth_states_tenants_tenant";
-- reverse: modify "integrations" table
ALTER TABLE "integrations" DROP CONSTRAINT "integrations_tenants_tenant";
-- reverse: modify "incident_types" table
ALTER TABLE "incident_types" DROP CONSTRAINT "incident_types_tenants_tenant";
-- reverse: modify "incident_timeline_event_topology_contexts" table
ALTER TABLE "incident_timeline_event_topology_contexts" DROP CONSTRAINT "incident_timeline_event_topolo_c9aecaf9d13cd231fe8f3c66ff0e4671", DROP CONSTRAINT "incident_timeline_event_topolo_71fa443670cce4d56bfb9e7041730983", DROP CONSTRAINT "incident_timeline_event_topolo_9d336da69e411af955f7f9ddff677001", DROP CONSTRAINT "incident_timeline_event_topology_contexts_tenants_tenant";
-- reverse: modify "incident_timeline_event_evidences" table
ALTER TABLE "incident_timeline_event_evidences" DROP CONSTRAINT "incident_timeline_event_evidences_tenants_tenant", DROP CONSTRAINT "incident_timeline_event_eviden_37786b98ea2184b38a27c223bdf28160";
-- reverse: modify "incident_timeline_event_contributing_factors" table
ALTER TABLE "incident_timeline_event_contributing_factors" DROP CONSTRAINT "incident_timeline_event_contributing_factors_tenants_tenant", DROP CONSTRAINT "incident_timeline_event_contri_0aecb2f20121e2d628c1402580fe5d71";
-- reverse: modify "incident_timeline_event_contexts" table
ALTER TABLE "incident_timeline_event_contexts" DROP CONSTRAINT "incident_timeline_event_contexts_tenants_tenant", DROP CONSTRAINT "incident_timeline_event_contex_5ac24bfc474fb61b9351fb5cba7c87cb";
-- reverse: modify "incident_timeline_events" table
ALTER TABLE "incident_timeline_events" DROP CONSTRAINT "incident_timeline_events_normalized_events_event", DROP CONSTRAINT "incident_timeline_events_tenants_tenant", DROP CONSTRAINT "incident_timeline_events_incidents_timeline_events";
-- reverse: modify "incident_tags" table
ALTER TABLE "incident_tags" DROP CONSTRAINT "incident_tags_tenants_tenant";
-- reverse: modify "incident_severities" table
ALTER TABLE "incident_severities" DROP CONSTRAINT "incident_severities_normalized_events_projected_from", DROP CONSTRAINT "incident_severities_tenants_tenant";
-- reverse: modify "incident_role_assignments" table
ALTER TABLE "incident_role_assignments" DROP CONSTRAINT "incident_role_assignments_incident_roles_role", DROP CONSTRAINT "incident_role_assignments_users_user", DROP CONSTRAINT "incident_role_assignments_incidents_incident", DROP CONSTRAINT "incident_role_assignments_tenants_tenant";
-- reverse: modify "incident_roles" table
ALTER TABLE "incident_roles" DROP CONSTRAINT "incident_roles_tenants_tenant";
-- reverse: modify "incident_milestones" table
ALTER TABLE "incident_milestones" DROP CONSTRAINT "incident_milestones_users_incident_milestones", DROP CONSTRAINT "incident_milestones_tenants_tenant", DROP CONSTRAINT "incident_milestones_incidents_milestones";
-- reverse: modify "incident_links" table
ALTER TABLE "incident_links" DROP CONSTRAINT "incident_links_incidents_linked_incident", DROP CONSTRAINT "incident_links_incidents_incident", DROP CONSTRAINT "incident_links_tenants_tenant";
-- reverse: modify "incident_field_options" table
ALTER TABLE "incident_field_options" DROP CONSTRAINT "incident_field_options_tenants_tenant", DROP CONSTRAINT "incident_field_options_incident_fields_options";
-- reverse: modify "incident_fields" table
ALTER TABLE "incident_fields" DROP CONSTRAINT "incident_fields_tenants_tenant";
-- reverse: modify "incident_debrief_suggestions" table
ALTER TABLE "incident_debrief_suggestions" DROP CONSTRAINT "incident_debrief_suggestions_tenants_tenant", DROP CONSTRAINT "incident_debrief_suggestions_incident_debriefs_suggestions";
-- reverse: modify "incident_debrief_questions" table
ALTER TABLE "incident_debrief_questions" DROP CONSTRAINT "incident_debrief_questions_tenants_tenant";
-- reverse: modify "incident_debrief_messages" table
ALTER TABLE "incident_debrief_messages" DROP CONSTRAINT "incident_debrief_messages_inci_0d1f0b105ef851edb04b442b34a0d17f", DROP CONSTRAINT "incident_debrief_messages_tenants_tenant", DROP CONSTRAINT "incident_debrief_messages_incident_debriefs_messages";
-- reverse: modify "incident_debriefs" table
ALTER TABLE "incident_debriefs" DROP CONSTRAINT "incident_debriefs_users_incident_debriefs", DROP CONSTRAINT "incident_debriefs_tenants_tenant", DROP CONSTRAINT "incident_debriefs_incidents_debriefs";
-- reverse: modify "incidents" table
ALTER TABLE "incidents" DROP CONSTRAINT "incidents_incident_types_type", DROP CONSTRAINT "incidents_incident_severities_severity", DROP CONSTRAINT "incidents_normalized_events_projected_from", DROP CONSTRAINT "incidents_tenants_tenant";
-- reverse: modify "event_annotations" table
ALTER TABLE "event_annotations" DROP CONSTRAINT "event_annotations_users_creator", DROP CONSTRAINT "event_annotations_normalized_events_event", DROP CONSTRAINT "event_annotations_tenants_tenant";
-- reverse: modify "document_accesses" table
ALTER TABLE "document_accesses" DROP CONSTRAINT "document_accesses_teams_team", DROP CONSTRAINT "document_accesses_users_user", DROP CONSTRAINT "document_accesses_documents_document", DROP CONSTRAINT "document_accesses_tenants_tenant";
-- reverse: modify "documents" table
ALTER TABLE "documents" DROP CONSTRAINT "documents_tenants_tenant";
-- reverse: modify "alert_feedbacks" table
ALTER TABLE "alert_feedbacks" DROP CONSTRAINT "alert_feedbacks_normalized_events_alert_feedback", DROP CONSTRAINT "alert_feedbacks_normalized_events_alert_instance", DROP CONSTRAINT "alert_feedbacks_alerts_alert", DROP CONSTRAINT "alert_feedbacks_tenants_tenant";
-- reverse: modify "alerts" table
ALTER TABLE "alerts" DROP CONSTRAINT "alerts_oncall_rosters_alerts", DROP CONSTRAINT "alerts_normalized_events_projected_from", DROP CONSTRAINT "alerts_tenants_tenant";
-- reverse: create "user_watched_oncall_rosters" table
DROP TABLE "user_watched_oncall_rosters";
-- reverse: create "team_oncall_rosters" table
DROP TABLE "team_oncall_rosters";
-- reverse: create "task_tickets" table
DROP TABLE "task_tickets";
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
-- reverse: create index "videoconference_meeting_session_id_status" to table: "video_conferences"
DROP INDEX "videoconference_meeting_session_id_status";
-- reverse: create index "videoconference_incident_id_status" to table: "video_conferences"
DROP INDEX "videoconference_incident_id_status";
-- reverse: create index "videoconference_tenant_id" to table: "video_conferences"
DROP INDEX "videoconference_tenant_id";
-- reverse: create index "video_conferences_meeting_session_id_key" to table: "video_conferences"
DROP INDEX "video_conferences_meeting_session_id_key";
-- reverse: create "video_conferences" table
DROP TABLE "video_conferences";
-- reverse: create index "user_tenant_id_email" to table: "users"
DROP INDEX "user_tenant_id_email";
-- reverse: create index "user_auth_provider_id" to table: "users"
DROP INDEX "user_auth_provider_id";
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
-- reverse: create index "teammembership_team_id_user_id" to table: "team_memberships"
DROP INDEX "teammembership_team_id_user_id";
-- reverse: create index "teammembership_tenant_id" to table: "team_memberships"
DROP INDEX "teammembership_tenant_id";
-- reverse: create "team_memberships" table
DROP TABLE "team_memberships";
-- reverse: create index "team_tenant_id" to table: "teams"
DROP INDEX "team_tenant_id";
-- reverse: create index "teams_slug_key" to table: "teams"
DROP INDEX "teams_slug_key";
-- reverse: create "teams" table
DROP TABLE "teams";
-- reverse: create index "task_tenant_id" to table: "tasks"
DROP INDEX "task_tenant_id";
-- reverse: create "tasks" table
DROP TABLE "tasks";
-- reverse: create index "systemtopologysnapshotrelation_96315e4d75ceda9b93720e1ad418db13" to table: "system_topology_snapshot_relationships"
DROP INDEX "systemtopologysnapshotrelation_96315e4d75ceda9b93720e1ad418db13";
-- reverse: create index "systemtopologysnapshotrelation_8bfb42b7b88bb99b0dab90f97e7c8e86" to table: "system_topology_snapshot_relationships"
DROP INDEX "systemtopologysnapshotrelation_8bfb42b7b88bb99b0dab90f97e7c8e86";
-- reverse: create index "systemtopologysnapshotrelation_ce56930e9b9fd6d129909352010bc447" to table: "system_topology_snapshot_relationships"
DROP INDEX "systemtopologysnapshotrelation_ce56930e9b9fd6d129909352010bc447";
-- reverse: create index "systemtopologysnapshotrelation_a06f34ed83db45fd13d05244cf667720" to table: "system_topology_snapshot_relationships"
DROP INDEX "systemtopologysnapshotrelation_a06f34ed83db45fd13d05244cf667720";
-- reverse: create index "systemtopologysnapshotrelationship_tenant_id_snapshot_id" to table: "system_topology_snapshot_relationships"
DROP INDEX "systemtopologysnapshotrelationship_tenant_id_snapshot_id";
-- reverse: create index "systemtopologysnapshotrelationship_tenant_id" to table: "system_topology_snapshot_relationships"
DROP INDEX "systemtopologysnapshotrelationship_tenant_id";
-- reverse: create "system_topology_snapshot_relationships" table
DROP TABLE "system_topology_snapshot_relationships";
-- reverse: create index "systemtopologysnapshotentity_t_f5c4f8cfa84671bf6a92d9f49c2f214d" to table: "system_topology_snapshot_entities"
DROP INDEX "systemtopologysnapshotentity_t_f5c4f8cfa84671bf6a92d9f49c2f214d";
-- reverse: create index "systemtopologysnapshotentity_tenant_id_knowledge_entity_id" to table: "system_topology_snapshot_entities"
DROP INDEX "systemtopologysnapshotentity_tenant_id_knowledge_entity_id";
-- reverse: create index "systemtopologysnapshotentity_tenant_id_snapshot_id" to table: "system_topology_snapshot_entities"
DROP INDEX "systemtopologysnapshotentity_tenant_id_snapshot_id";
-- reverse: create index "systemtopologysnapshotentity_tenant_id" to table: "system_topology_snapshot_entities"
DROP INDEX "systemtopologysnapshotentity_tenant_id";
-- reverse: create "system_topology_snapshot_entities" table
DROP TABLE "system_topology_snapshot_entities";
-- reverse: create index "systemtopologysnapshot_tenant_id_created_at" to table: "system_topology_snapshots"
DROP INDEX "systemtopologysnapshot_tenant_id_created_at";
-- reverse: create index "systemtopologysnapshot_tenant_id_as_of" to table: "system_topology_snapshots"
DROP INDEX "systemtopologysnapshot_tenant_id_as_of";
-- reverse: create index "systemtopologysnapshot_tenant_id" to table: "system_topology_snapshots"
DROP INDEX "systemtopologysnapshot_tenant_id";
-- reverse: create "system_topology_snapshots" table
DROP TABLE "system_topology_snapshots";
-- reverse: create index "systemanalysistopologynode_tenant_id" to table: "system_analysis_topology_nodes"
DROP INDEX "systemanalysistopologynode_tenant_id";
-- reverse: create "system_analysis_topology_nodes" table
DROP TABLE "system_analysis_topology_nodes";
-- reverse: create index "systemanalysistopologyedge_tenant_id" to table: "system_analysis_topology_edges"
DROP INDEX "systemanalysistopologyedge_tenant_id";
-- reverse: create "system_analysis_topology_edges" table
DROP TABLE "system_analysis_topology_edges";
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
-- reverse: create index "retrospective_tenant_id" to table: "retrospectives"
DROP INDEX "retrospective_tenant_id";
-- reverse: create index "retrospectives_system_analysis_id_key" to table: "retrospectives"
DROP INDEX "retrospectives_system_analysis_id_key";
-- reverse: create index "retrospectives_incident_id_key" to table: "retrospectives"
DROP INDEX "retrospectives_incident_id_key";
-- reverse: create index "retrospectives_document_id_key" to table: "retrospectives"
DROP INDEX "retrospectives_document_id_key";
-- reverse: create "retrospectives" table
DROP TABLE "retrospectives";
-- reverse: create index "providereventsyncrun_tenant_id_status_started_at" to table: "provider_event_sync_runs"
DROP INDEX "providereventsyncrun_tenant_id_status_started_at";
-- reverse: create index "providereventsyncrun_tenant_id_provider_started_at" to table: "provider_event_sync_runs"
DROP INDEX "providereventsyncrun_tenant_id_provider_started_at";
-- reverse: create index "providereventsyncrun_tenant_id" to table: "provider_event_sync_runs"
DROP INDEX "providereventsyncrun_tenant_id";
-- reverse: create "provider_event_sync_runs" table
DROP TABLE "provider_event_sync_runs";
-- reverse: create index "providereventsynccursor_tenant_id_provider_provider_source" to table: "provider_event_sync_cursors"
DROP INDEX "providereventsynccursor_tenant_id_provider_provider_source";
-- reverse: create index "providereventsynccursor_tenant_id" to table: "provider_event_sync_cursors"
DROP INDEX "providereventsynccursor_tenant_id";
-- reverse: create "provider_event_sync_cursors" table
DROP TABLE "provider_event_sync_cursors";
-- reverse: create index "playbook_tenant_id" to table: "playbooks"
DROP INDEX "playbook_tenant_id";
-- reverse: create "playbooks" table
DROP TABLE "playbooks";
-- reverse: create index "organizationrole_org_id_user_id" to table: "organization_roles"
DROP INDEX "organizationrole_org_id_user_id";
-- reverse: create index "organizationrole_tenant_id" to table: "organization_roles"
DROP INDEX "organizationrole_tenant_id";
-- reverse: create index "organization_roles_user_id_key" to table: "organization_roles"
DROP INDEX "organization_roles_user_id_key";
-- reverse: create "organization_roles" table
DROP TABLE "organization_roles";
-- reverse: create index "organization_auth_provider_id" to table: "organizations"
DROP INDEX "organization_auth_provider_id";
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
-- reverse: create index "normalizedeventprojectionstatus_tenant_id_status_updated_at" to table: "normalized_event_projection_status"
DROP INDEX "normalizedeventprojectionstatus_tenant_id_status_updated_at";
-- reverse: create index "normalizedeventprojectionstatu_26223016baedaca5556963f76f513be8" to table: "normalized_event_projection_status"
DROP INDEX "normalizedeventprojectionstatu_26223016baedaca5556963f76f513be8";
-- reverse: create index "normalizedeventprojectionstatus_tenant_id" to table: "normalized_event_projection_status"
DROP INDEX "normalizedeventprojectionstatus_tenant_id";
-- reverse: create "normalized_event_projection_status" table
DROP TABLE "normalized_event_projection_status";
-- reverse: create index "normalizedevent_tenant_id_kind_occurred_at" to table: "normalized_events"
DROP INDEX "normalizedevent_tenant_id_kind_occurred_at";
-- reverse: create index "normalizedevent_tenant_id_prov_089950886f426b5eaeeba9e4f3d2645c" to table: "normalized_events"
DROP INDEX "normalizedevent_tenant_id_prov_089950886f426b5eaeeba9e4f3d2645c";
-- reverse: create index "normalizedevent_tenant_id" to table: "normalized_events"
DROP INDEX "normalizedevent_tenant_id";
-- reverse: create "normalized_events" table
DROP TABLE "normalized_events";
-- reverse: create index "meetingsession_tenant_id" to table: "meeting_sessions"
DROP INDEX "meetingsession_tenant_id";
-- reverse: create "meeting_sessions" table
DROP TABLE "meeting_sessions";
-- reverse: create index "meetingschedule_tenant_id" to table: "meeting_schedules"
DROP INDEX "meetingschedule_tenant_id";
-- reverse: create "meeting_schedules" table
DROP TABLE "meeting_schedules";
-- reverse: create index "knowledgerelationship_tenant_id_kind_deleted_at" to table: "knowledge_relationships"
DROP INDEX "knowledgerelationship_tenant_id_kind_deleted_at";
-- reverse: create index "knowledgerelationship_tenant_id_kind_last_observed_at" to table: "knowledge_relationships"
DROP INDEX "knowledgerelationship_tenant_id_kind_last_observed_at";
-- reverse: create index "knowledgerelationship_tenant_id_updated_at" to table: "knowledge_relationships"
DROP INDEX "knowledgerelationship_tenant_id_updated_at";
-- reverse: create index "knowledgerelationship_tenant_id_target_entity_id" to table: "knowledge_relationships"
DROP INDEX "knowledgerelationship_tenant_id_target_entity_id";
-- reverse: create index "knowledgerelationship_tenant_id_source_entity_id" to table: "knowledge_relationships"
DROP INDEX "knowledgerelationship_tenant_id_source_entity_id";
-- reverse: create index "knowledgerelationship_tenant_id_kind" to table: "knowledge_relationships"
DROP INDEX "knowledgerelationship_tenant_id_kind";
-- reverse: create index "knowledgerelationship_tenant_i_c2e180b6bf727a089ab234a0504ce8ba" to table: "knowledge_relationships"
DROP INDEX "knowledgerelationship_tenant_i_c2e180b6bf727a089ab234a0504ce8ba";
-- reverse: create index "knowledgerelationship_tenant_id" to table: "knowledge_relationships"
DROP INDEX "knowledgerelationship_tenant_id";
-- reverse: create "knowledge_relationships" table
DROP TABLE "knowledge_relationships";
-- reverse: create index "knowledgeevidence_tenant_id_no_87f9cb1d0d06fba01667644def3b8e9c" to table: "knowledge_evidences"
DROP INDEX "knowledgeevidence_tenant_id_no_87f9cb1d0d06fba01667644def3b8e9c";
-- reverse: create index "knowledgeevidence_tenant_id_no_be29c0c818fd60bfb834f7c44768c470" to table: "knowledge_evidences"
DROP INDEX "knowledgeevidence_tenant_id_no_be29c0c818fd60bfb834f7c44768c470";
-- reverse: create index "knowledgeevidence_tenant_id_as_facfb4a424d2bd618ffae3bf7d544693" to table: "knowledge_evidences"
DROP INDEX "knowledgeevidence_tenant_id_as_facfb4a424d2bd618ffae3bf7d544693";
-- reverse: create index "knowledgeevidence_tenant_id_normalized_event_id" to table: "knowledge_evidences"
DROP INDEX "knowledgeevidence_tenant_id_normalized_event_id";
-- reverse: create index "knowledgeevidence_tenant_id_alias_id" to table: "knowledge_evidences"
DROP INDEX "knowledgeevidence_tenant_id_alias_id";
-- reverse: create index "knowledgeevidence_tenant_id_relationship_id" to table: "knowledge_evidences"
DROP INDEX "knowledgeevidence_tenant_id_relationship_id";
-- reverse: create index "knowledgeevidence_tenant_id_entity_id" to table: "knowledge_evidences"
DROP INDEX "knowledgeevidence_tenant_id_entity_id";
-- reverse: create index "knowledgeevidence_tenant_id" to table: "knowledge_evidences"
DROP INDEX "knowledgeevidence_tenant_id";
-- reverse: create "knowledge_evidences" table
DROP TABLE "knowledge_evidences";
-- reverse: create index "knowledgeentityalias_tenant_id_bedde5b02dd708153cce4f6b86ce1c2c" to table: "knowledge_entity_alias"
DROP INDEX "knowledgeentityalias_tenant_id_bedde5b02dd708153cce4f6b86ce1c2c";
-- reverse: create index "knowledgeentityalias_tenant_id_entity_id" to table: "knowledge_entity_alias"
DROP INDEX "knowledgeentityalias_tenant_id_entity_id";
-- reverse: create index "knowledgeentityalias_tenant_id" to table: "knowledge_entity_alias"
DROP INDEX "knowledgeentityalias_tenant_id";
-- reverse: create "knowledge_entity_alias" table
DROP TABLE "knowledge_entity_alias";
-- reverse: create index "knowledgeentity_tenant_id_kind_deleted_at" to table: "knowledge_entities"
DROP INDEX "knowledgeentity_tenant_id_kind_deleted_at";
-- reverse: create index "knowledgeentity_tenant_id_kind_last_observed_at" to table: "knowledge_entities"
DROP INDEX "knowledgeentity_tenant_id_kind_last_observed_at";
-- reverse: create index "knowledgeentity_tenant_id_updated_at" to table: "knowledge_entities"
DROP INDEX "knowledgeentity_tenant_id_updated_at";
-- reverse: create index "knowledgeentity_tenant_id_kind" to table: "knowledge_entities"
DROP INDEX "knowledgeentity_tenant_id_kind";
-- reverse: create index "knowledgeentity_tenant_id" to table: "knowledge_entities"
DROP INDEX "knowledgeentity_tenant_id";
-- reverse: create "knowledge_entities" table
DROP TABLE "knowledge_entities";
-- reverse: create index "integrationoauthstate_tenant_id" to table: "integration_oauth_states"
DROP INDEX "integrationoauthstate_tenant_id";
-- reverse: create "integration_oauth_states" table
DROP TABLE "integration_oauth_states";
-- reverse: create index "integration_tenant_id_provider_external_ref" to table: "integrations"
DROP INDEX "integration_tenant_id_provider_external_ref";
-- reverse: create index "integration_tenant_id_provider" to table: "integrations"
DROP INDEX "integration_tenant_id_provider";
-- reverse: create index "integration_tenant_id" to table: "integrations"
DROP INDEX "integration_tenant_id";
-- reverse: create "integrations" table
DROP TABLE "integrations";
-- reverse: create index "incidenttype_tenant_id" to table: "incident_types"
DROP INDEX "incidenttype_tenant_id";
-- reverse: create "incident_types" table
DROP TABLE "incident_types";
-- reverse: create index "incidenttimelineeventtopologycontext_tenant_id" to table: "incident_timeline_event_topology_contexts"
DROP INDEX "incidenttimelineeventtopologycontext_tenant_id";
-- reverse: create "incident_timeline_event_topology_contexts" table
DROP TABLE "incident_timeline_event_topology_contexts";
-- reverse: create index "incidenttimelineeventevidence_tenant_id" to table: "incident_timeline_event_evidences"
DROP INDEX "incidenttimelineeventevidence_tenant_id";
-- reverse: create "incident_timeline_event_evidences" table
DROP TABLE "incident_timeline_event_evidences";
-- reverse: create index "incidenttimelineeventcontributingfactor_tenant_id" to table: "incident_timeline_event_contributing_factors"
DROP INDEX "incidenttimelineeventcontributingfactor_tenant_id";
-- reverse: create "incident_timeline_event_contributing_factors" table
DROP TABLE "incident_timeline_event_contributing_factors";
-- reverse: create index "incidenttimelineeventcontext_tenant_id" to table: "incident_timeline_event_contexts"
DROP INDEX "incidenttimelineeventcontext_tenant_id";
-- reverse: create index "incident_timeline_event_contexts_incident_timeline_event_context_key" to table: "incident_timeline_event_contexts"
DROP INDEX "incident_timeline_event_contexts_incident_timeline_event_context_key";
-- reverse: create "incident_timeline_event_contexts" table
DROP TABLE "incident_timeline_event_contexts";
-- reverse: create index "incidenttimelineevent_kind" to table: "incident_timeline_events"
DROP INDEX "incidenttimelineevent_kind";
-- reverse: create index "incidenttimelineevent_tenant_id" to table: "incident_timeline_events"
DROP INDEX "incidenttimelineevent_tenant_id";
-- reverse: create "incident_timeline_events" table
DROP TABLE "incident_timeline_events";
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
-- reverse: create index "incidentlink_incident_id_linked_incident_id" to table: "incident_links"
DROP INDEX "incidentlink_incident_id_linked_incident_id";
-- reverse: create index "incidentlink_tenant_id" to table: "incident_links"
DROP INDEX "incidentlink_tenant_id";
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
-- reverse: create index "incident_tenant_id" to table: "incidents"
DROP INDEX "incident_tenant_id";
-- reverse: create index "incidents_slug_key" to table: "incidents"
DROP INDEX "incidents_slug_key";
-- reverse: create "incidents" table
DROP TABLE "incidents";
-- reverse: create index "eventannotation_tenant_id" to table: "event_annotations"
DROP INDEX "eventannotation_tenant_id";
-- reverse: create "event_annotations" table
DROP TABLE "event_annotations";
-- reverse: create index "documentaccess_tenant_id" to table: "document_accesses"
DROP INDEX "documentaccess_tenant_id";
-- reverse: create "document_accesses" table
DROP TABLE "document_accesses";
-- reverse: create index "document_tenant_id" to table: "documents"
DROP INDEX "document_tenant_id";
-- reverse: create "documents" table
DROP TABLE "documents";
-- reverse: create index "alertfeedback_tenant_id" to table: "alert_feedbacks"
DROP INDEX "alertfeedback_tenant_id";
-- reverse: create "alert_feedbacks" table
DROP TABLE "alert_feedbacks";
-- reverse: create index "alert_tenant_id" to table: "alerts"
DROP INDEX "alert_tenant_id";
-- reverse: create "alerts" table
DROP TABLE "alerts";
