
alter table community drop column "avatar_url";
alter table post drop column "creator_type";
alter table post drop constraint creator_type_check;

drop TRIGGER if exists validate_post_owner on post;
drop function if exists check_owner_exists();
