-- Очищаем таблицу users, если она существует
drop table if exists users;

-- Создаем таблицу users
create table if not exists users(
                                    id serial primary key,
                                    uuid UUID unique not null,
                                    login text not null,
                                    name text,
                                    surname text,
                                    sex int,
                                    birth_date date,
                                    psw_hash text not null,
                                    salt text not null
);

-- Очищаем таблицу posts, если она существует
drop table if exists posts cascade;

-- Создаем таблицу posts
CREATE table if not exists posts (
                                     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                     creator_id UUID NOT NULL,
                                     description TEXT,
                                     created_at TIMESTAMP DEFAULT now(),
                                     like_count INT DEFAULT 0,
                                     repost_count INT DEFAULT 0,
                                     comment_count INT DEFAULT 0,
                                     FOREIGN KEY (creator_id) REFERENCES users(uuid) ON DELETE CASCADE
);

drop table if exists post_photos;
-- Создаем таблицу post_photos
CREATE TABLE if not exists post_photos (
                                           id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                           post_id UUID NOT NULL,
                                           photo_path TEXT NOT NULL,
                                           FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);
