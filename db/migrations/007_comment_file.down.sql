drop table if exists comment_file cascade;

alter table comment drop column if exists updated_at;