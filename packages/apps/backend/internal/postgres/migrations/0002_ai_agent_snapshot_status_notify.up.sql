CREATE OR REPLACE FUNCTION rezible.notify_ai_agent_snapshot_status_change()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  IF TG_OP = 'UPDATE' AND OLD.status IS NOT DISTINCT FROM NEW.status THEN
    RETURN NEW;
  END IF;

  PERFORM pg_notify(
    'rezible_ai_agent_snapshot_status',
    json_build_object(
      'tenant_id', NEW.tenant_id,
      'snapshot_id', NEW.id,
      'status', NEW.status,
      'updated_at', NEW.updated_at
    )::text
  );

  RETURN NEW;
END;
$$;

CREATE TRIGGER ai_agent_run_snapshot_status_notify
AFTER INSERT OR UPDATE OF status ON rezible.ai_agent_run_snapshots
FOR EACH ROW
EXECUTE FUNCTION rezible.notify_ai_agent_snapshot_status_change();
