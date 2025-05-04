

alter table community drop constraint if exists nickname_check;
alter table community drop constraint if exists unique_nickname;
alter table community drop column if exists nickname;

alter table community drop column if exists contact_info;
alter table community drop column if exists cover_url;
