-- +migrate Up

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
