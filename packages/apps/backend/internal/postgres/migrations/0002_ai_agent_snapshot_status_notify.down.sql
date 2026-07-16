DROP TRIGGER IF EXISTS ai_agent_run_snapshot_status_notify ON rezible.ai_agent_run_snapshots;

DROP FUNCTION IF EXISTS rezible.notify_ai_agent_snapshot_status_change();
