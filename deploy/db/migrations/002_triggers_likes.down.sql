
-- +migrate Down
DROP TRIGGER IF EXISTS trg_update_post_like_count ON like_post;
DROP FUNCTION IF EXISTS update_post_like_count();
