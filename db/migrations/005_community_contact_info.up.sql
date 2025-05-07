
alter table community add column nickname text;
update community set nickname = name where nickname is null;
alter table community add constraint nickname_check check (nickname is not null);
alter table community add constraint unique_nickname unique (nickname);

alter table community add column contact_info int references contact_info(id) on delete set null;
alter table community add column cover_url text;
