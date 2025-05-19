
alter table message_file add column file_type text not null default 'image';
alter table post_file add column file_type text not null default 'image';