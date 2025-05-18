-- triggers for updating like_count in post table
CREATE OR REPLACE FUNCTION update_post_like_count()
    RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE post SET like_count = like_count + 1 WHERE id = NEW.post_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE post SET like_count = like_count - 1 WHERE id = OLD.post_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_post_like_count
    AFTER INSERT OR DELETE ON like_post
    FOR EACH ROW
EXECUTE FUNCTION update_post_like_count();

CREATE OR REPLACE FUNCTION check_owner_exists()
    RETURNS TRIGGER AS $$
BEGIN
    IF NEW.creator_type = 'user' THEN
        IF NOT EXISTS (SELECT 1 FROM "user" WHERE id = NEW.creator_id) THEN
            RAISE EXCEPTION 'User with id % does not exist', NEW.creator_id;
        END IF;
    ELSIF NEW.creator_type = 'community' THEN
        IF NOT EXISTS (SELECT 1 FROM community WHERE id = NEW.creator_id) THEN
            RAISE EXCEPTION 'Community with id % does not exist', NEW.creator_id;
        END IF;
    ELSE
        RAISE EXCEPTION 'Invalid owner_type: %', NEW.creator_type;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

CREATE TRIGGER validate_post_owner
    BEFORE INSERT OR UPDATE ON post
    FOR EACH ROW
EXECUTE FUNCTION check_owner_exists();

CREATE OR REPLACE FUNCTION delete_posts_by_owner()
    RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM post
    WHERE creator_id = OLD.id AND (
        (TG_TABLE_NAME = 'user' AND creator_type = 'user') OR
        (TG_TABLE_NAME = 'community' AND creator_type = 'community')
        );
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_delete_user_posts
    AFTER DELETE ON "user"
    FOR EACH ROW
EXECUTE FUNCTION delete_posts_by_owner();