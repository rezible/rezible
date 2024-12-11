ALTER TABLE ladders
    ADD CONSTRAINT ladder_parent_uniq
        UNIQUE (id, parent_id);

ALTER TABLE ladder_ancestries
    ADD CONSTRAINT ladder_intermediary_dist_uniq
        UNIQUE (ladder_id, intermediary_id, distance_to_intermediary);

ALTER TABLE ladder_ancestries
    ADD CONSTRAINT ladder_ancestor_dist_uniq
        UNIQUE (ladder_id, ancestor_id, distance);

ALTER TABLE ladder_ancestries
    ADD CONSTRAINT ladder_ancestries_intermediary_distance_fk
        FOREIGN KEY (ladder_id, intermediary_id, distance_to_intermediary)
            REFERENCES ladder_ancestries (ladder_id, ancestor_id, distance)
            DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ladder_ancestries
    ADD CONSTRAINT ladder_ancestries_intermediary_fk
        FOREIGN KEY (intermediary_id, ancestor_id)
            REFERENCES ladders (id, parent_id)
            ON DELETE CASCADE
            DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE ladder_ancestries
    ADD CONSTRAINT ladder_ancestries_ancestor_fk
        FOREIGN KEY (ancestor_id)
            REFERENCES ladders (id)
            ON DELETE CASCADE
            DEFERRABLE INITIALLY DEFERRED;

-- recursively crawl parents and calculate distance to ancestors
CREATE OR REPLACE FUNCTION materialize_ladders_ancestors(VARIADIC uuid[]) RETURNS void LANGUAGE plpgsql AS $$BEGIN
    WITH RECURSIVE closure AS (
        SELECT
            ladders.id AS ladder_id,
            ladders.id AS ancestor_id,
            0 AS distance,
            null::uuid AS intermediary_id,
            null::integer AS distance_to_intermediary
        FROM ladders
            WHERE ladders.id = any ($1)
        UNION
        SELECT
            closure.ladder_id AS ladder_id,
            ladders.parent_id AS ancestor_id,
            closure.distance + 1 AS distance,
            ladders.id AS intermediary_id,
            closure.distance AS distance_to_intermediary
        FROM ladders
            JOIN closure ON (closure.ancestor_id = ladders.id)
        WHERE ladders.parent_id IS NOT NULL
    )

    INSERT INTO ladder_ancestries (ladder_id, ancestor_id, distance, intermediary_id, distance_to_intermediary)
    SELECT * FROM closure
    ON CONFLICT (ladder_id, ancestor_id) DO NOTHING;
END$$;

CREATE OR REPLACE FUNCTION ladders_rematerialize_closure(uuid) RETURNS void LANGUAGE plpgsql AS $$BEGIN
    PERFORM materialize_ladders_ancestors($1);
    PERFORM ladders_rematerialize_closure(ladder_id) FROM (
        SELECT id AS ladder_id FROM ladders WHERE ladders.parent_id = $1
    ) children;
END$$;

CREATE FUNCTION trigger_materialize_ancestry_closure() RETURNS trigger LANGUAGE plpgsql AS $$BEGIN
    PERFORM ladders_rematerialize_closure(NEW.id);
    RETURN NEW;
END$$;

CREATE TRIGGER materialize_ancestry_closure
    AFTER INSERT ON ladders
    FOR EACH ROW EXECUTE FUNCTION trigger_materialize_ancestry_closure();
------

------
CREATE FUNCTION trigger_block_ancestry_cycle() RETURNS trigger LANGUAGE plpgsql AS $$BEGIN
    IF EXISTS (SELECT FROM ladder_ancestries WHERE ancestor_id = NEW.id AND ladder_ancestries.ladder_id = NEW.parent_id) THEN
        RAISE EXCEPTION 'Cannot add descendant as parent, would create cycle';
    ELSE
        RETURN NEW;
    END IF;
END$$;

CREATE TRIGGER prevent_ancestry_cycle
    BEFORE INSERT ON ladders
    FOR EACH ROW EXECUTE FUNCTION trigger_block_ancestry_cycle();
------

------
CREATE FUNCTION ladders_cleanup_ancestry_before_parent_updated() RETURNS trigger LANGUAGE plpgsql AS $$BEGIN
    -- update children, set their parent to what was previously grandparent
    UPDATE ladders
        SET parent_id=OLD.parent_id
        WHERE parent_id=OLD.id;
    -- remove old ancestries for this node or that used this node as an intermediary
    DELETE FROM ladder_ancestries
        WHERE intermediary_id IN (OLD.id, OLD.parent_id, NEW.parent_id) OR
              ancestor_id IN (OLD.id, OLD.parent_id, NEW.parent_id) OR
              ladder_id = OLD.id;
    RETURN NEW;
END$$;

CREATE TRIGGER cleanup_ladder_ancestry_on_parent_updated
    BEFORE UPDATE OF parent_id ON ladders FOR EACH ROW
    WHEN (pg_trigger_depth() < 1)
EXECUTE FUNCTION ladders_cleanup_ancestry_before_parent_updated();

CREATE FUNCTION ladders_after_parent_updated() RETURNS trigger LANGUAGE plpgsql AS $$BEGIN
    PERFORM ladders_rematerialize_closure(OLD.parent_id);
    PERFORM ladders_rematerialize_closure(NEW.parent_id);
    RETURN NEW;
END$$;

CREATE TRIGGER rematerialize_ladder_ancestry_after_parent_updated
    AFTER UPDATE OF parent_id ON ladders FOR EACH ROW
    WHEN (pg_trigger_depth() < 1)
    EXECUTE FUNCTION ladders_after_parent_updated();
------

CREATE FUNCTION ladders_cleanup_ancestry_after_delete() RETURNS trigger LANGUAGE plpgsql AS $$BEGIN
    -- remove old ancestries for this node
    DELETE FROM ladder_ancestries
        WHERE ladder_id = OLD.id;
    RETURN NEW;
END$$;

CREATE TRIGGER cleanup_ladder_ancestry_after_delete
    AFTER DELETE ON ladders FOR EACH ROW
    EXECUTE FUNCTION ladders_cleanup_ancestry_after_delete();