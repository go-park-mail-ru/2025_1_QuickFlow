
alter table community add column "avatar_url" text;
alter table post add column "creator_type" text;
update post set creator_type = 'user' where creator_id is not null;
alter table post add constraint creator_type_check check (creator_type in ('user', 'community') and creator_id is not null);



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
$$ LANGUAGE plpgsql;

CREATE TRIGGER validate_post_owner
    BEFORE INSERT OR UPDATE ON post
    FOR EACH ROW
EXECUTE FUNCTION check_owner_exists();

