create table if not exists comment_file(
                                           id int generated always as identity primary key,
                                           comment_id uuid references comment(id) on delete cascade,
                                           file_url text not null,
                                           added_at timestamptz not null default now(),
                                           file_type text not null default 'image'
);

alter table comment add column updated_at timestamptz not null default now();

