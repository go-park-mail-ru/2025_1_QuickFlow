
alter table community_user drop column role;
alter table community_user add column role text;
update community_user set role = 'member' where role is null;
alter table community_user add constraint role_check check (role in ('owner', 'admin', 'member') and role is not null);
alter table post drop constraint post_creator_id_fkey;
