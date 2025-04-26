-- +migrate Up
alter table post
    add column if not exists updated_at timestamptz not null default now();

update post
set updated_at = created_at;

-- +migrate Down
alter table post
    drop column if exists updated_at;