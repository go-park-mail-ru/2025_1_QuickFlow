create table if not exists "user" (
                                      id uuid primary key,
                                      username text not null unique,
                                      psw_hash text not null,
                                      salt text not null
);

CREATE TABLE IF NOT EXISTS university (
                                          id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
                                          name TEXT NOT NULL,
                                          city TEXT NOT NULL,
                                          UNIQUE(name, city)
);

CREATE TABLE IF NOT EXISTS faculty (
                                       id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
                                       university_id INT REFERENCES university(id) ON DELETE CASCADE,
                                       name TEXT NOT NULL,
                                       UNIQUE (university_id, name)
);


create table if not exists school(
                                     id int primary key generated always as identity,
                                     city text not null,
                                     name text not null
);

create table if not exists contact_info(
                                           id int primary key generated always as identity,
                                           city text,
                                           phone_number text,
                                           email text
);

create table if not exists profile(
                                      id uuid primary key,
                                      bio text,
                                      profile_avatar text,
                                      profile_background text,
                                      firstname text not null,
                                      lastname text not null,
                                      sex int default 0 check (sex >= 0),
                                      birth_date date,
                                      contact_info_id int references contact_info(id) on delete set null,
                                      school_id int references school(id) on delete set null,

                                      last_seen timestamptz not null default now(),
                                      foreign key (id) references "user"(id) on delete cascade
);

CREATE TABLE IF NOT EXISTS education (
                                         id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
                                         profile_id UUID unique REFERENCES profile(id) ON DELETE CASCADE,
                                         faculty_id INT REFERENCES faculty(id) ON DELETE SET NULL,
                                         graduation_year INT CHECK (graduation_year >= 0),
                                         UNIQUE(profile_id, faculty_id, graduation_year)
);


create table if not exists post(
                                   id uuid primary key,
                                   creator_id uuid references "user"(id) on delete cascade,
                                   text text,
                                   created_at timestamptz not null default now(),
                                   updated_at timestamptz not null default now(),
                                   like_count int default 0 check (like_count >= 0),
                                   repost_count int default 0 check(repost_count >= 0),
                                   comment_count int default 0 check(comment_count >= 0),
                                   is_repost bool default false
);

create table if not exists comment(
                                      id uuid primary key,
                                      post_id uuid references post(id) on delete cascade,
                                      user_id uuid  references "user"(id) on delete cascade,
                                      created_at timestamptz not null default now(),
                                      like_count int default 0 check (like_count >= 0),
                                      text text not null
);

create table if not exists post_file(
                                        id int generated always as identity primary key,
                                        post_id uuid references post(id) on delete cascade,
                                        file_url text not null,
                                        added_at timestamptz not null default now()
);

create table if not exists repost(
                                     repost_id uuid primary key,
                                     original_id uuid references post(id) on delete cascade,

                                     foreign key (repost_id) references post(id) on delete cascade
);

create table if not exists like_post(
                                        id int generated always as identity primary key,
                                        user_id uuid references "user"(id) on delete cascade,
                                        post_id uuid references post(id) on delete cascade,
                                        unique (user_id, post_id)
);

create table if not exists like_comment(
                                           id int generated always as identity primary key,
                                           user_id uuid references "user"(id) on delete cascade,
                                           comment_id uuid references comment(id) on delete cascade,
                                           unique (user_id, comment_id)
);

create table if not exists friendship(
                                         id int generated always as identity primary key,
                                         user1_id uuid references "user"(id) on delete cascade,
                                         user2_id uuid references "user"(id) on delete cascade,
                                         status text not null default 'following',
                                         unique (user1_id, user2_id),
                                         check (user1_id < user2_id)
);

create table if not exists chat(
                                   id uuid primary key,
                                   type int default 0,
                                   name text check (length(name) > 0),
                                   avatar_url text check (length(name) > 0),
                                   created_at timestamptz not null default now(),
                                   updated_at timestamptz not null default now()
);

create table if not exists chat_user(
                                        id int generated always as identity primary key,
                                        chat_id uuid references chat(id) on delete cascade,
                                        user_id uuid references "user"(id) on delete cascade,
                                        last_read timestamptz,
                                        unique(chat_id, user_id)
);

create table if not exists message(
                                      id uuid primary key,
                                      text text check (length(text) > 0),
                                      sender_id uuid references "user"(id) on delete cascade,
                                      chat_id uuid references chat(id) on delete cascade,
                                      created_at timestamptz not null default now(),
                                      updated_at timestamptz not null default now()
);

create table if not exists message_file(
                                           id int generated always as identity primary key,
                                           message_id uuid references message(id) on delete cascade,
                                           file_url text not null
);

create table if not exists community(
                                        id uuid primary key,
                                        owner_id uuid references "user"(id) on delete cascade,
                                        name text not null unique,
                                        description text,
                                        created_at timestamptz not null default now()
);

create table if not exists community_user(
                                             id int generated always as identity primary key,
                                             community_id uuid references community(id) on delete cascade,
                                             user_id uuid references "user"(id) on delete cascade,
                                             role int not null default 0,
                                             joined_at timestamptz not null default now(),
                                             unique (community_id, user_id)
);

create table if not exists user_follow(
                                          id int generated always as identity primary key,
                                          following_id uuid references "user"(id) on delete cascade,
                                          followed_id uuid references "user"(id) on delete cascade,
                                          unique (following_id, followed_id),
                                          check (following_id != followed_id)
);

create table if not exists feedback(
                                       id uuid primary key default gen_random_uuid(),
                                       rating int not null,
                                       respondent_id uuid references "user"(id) on delete set null,
                                       text text,
                                       type text not null,
                                       created_at timestamptz not null default now(),
                                       check (rating >= 0 and rating <= 10)
);

create extension if not exists pg_trgm;
SET pg_trgm.similarity_threshold = 0.3; -- for fuzzy search
