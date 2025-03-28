CREATE TABLE "user" (
                        id UUID PRIMARY KEY,
                        username TEXT NOT NULL UNIQUE,
                        psw_hash TEXT NOT NULL,
                        salt TEXT NOT NULL
);

CREATE TABLE profile (
                         id UUID PRIMARY KEY,
                         bio TEXT,
                         profile_avatar TEXT,
                         firstname TEXT NOT NULL,
                         lastname TEXT NOT NULL,
                         sex INT DEFAULT 0,
                         birth_date DATE,
                         FOREIGN KEY (id) REFERENCES "user"(id) ON DELETE CASCADE
);

CREATE TABLE post (
                      id UUID PRIMARY KEY,
                      creator_id UUID REFERENCES "user"(id) ON DELETE CASCADE,
                      text TEXT,
                      created_at timestamp NOT NULL DEFAULT now(),
                      like_count INT DEFAULT 0,
                      repost_count INT DEFAULT 0,
                      comment_count INT DEFAULT 0,
                      is_repost BOOL DEFAULT FALSE
);

CREATE TABLE comment (
                         id UUID PRIMARY KEY,
                         post_id UUID REFERENCES post(id) ON DELETE CASCADE,
                         user_id UUID REFERENCES "user"(id) ON DELETE CASCADE,
                         created_at TIMESTAMP NOT NULL DEFAULT now(),
                         like_count INT DEFAULT 0,
                         text TEXT NOT NULL
);

CREATE TABLE post_file (
                           id SERIAL PRIMARY KEY,
                           post_id UUID REFERENCES post(id) ON DELETE CASCADE,
                           file_url TEXT NOT NULL
);

CREATE TABLE repost (
                        repost_id UUID PRIMARY KEY,
                        original_id UUID REFERENCES post(id) ON DELETE CASCADE,
                        FOREIGN KEY (repost_id) REFERENCES post(id) ON DELETE CASCADE
);

CREATE TABLE like_post (
                           id SERIAL PRIMARY KEY,
                           user_id UUID REFERENCES "user"(id) ON DELETE CASCADE,
                           post_id UUID REFERENCES post(id) ON DELETE CASCADE,
                           UNIQUE (user_id, post_id)
);

CREATE TABLE like_comment (
                              id SERIAL PRIMARY KEY,
                              user_id UUID REFERENCES "user"(id) ON DELETE CASCADE,
                              comment_id UUID REFERENCES comment(id) ON DELETE CASCADE,
                              UNIQUE (user_id, comment_id)
);

CREATE TABLE friendship (
                            id SERIAL PRIMARY KEY,
                            user1_id UUID REFERENCES "user"(id) ON DELETE CASCADE,
                            user2_id UUID REFERENCES "user"(id) ON DELETE CASCADE,
                            status INT NOT NULL DEFAULT 0,
                            UNIQUE (user1_id, user2_id)
);

CREATE TABLE chat (
                      id UUID PRIMARY KEY,
                      type INT DEFAULT 0,
                      created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE chat_user (
                           id SERIAL PRIMARY KEY,
                           chat_id UUID REFERENCES chat(id) ON DELETE CASCADE,
                           user_id UUID REFERENCES "user"(id) ON DELETE CASCADE,
                           UNIQUE (chat_id, user_id)
);

CREATE TABLE message (
                         id UUID PRIMARY KEY,
                         text TEXT CHECK (LENGTH(text) > 0),
                         sender_id UUID REFERENCES "user"(id) ON DELETE CASCADE,
                         chat_id UUID REFERENCES chat(id) ON DELETE CASCADE,
                         created_at TIMESTAMP NOT NULL DEFAULT now(),
                         is_read BOOL NOT NULL DEFAULT FALSE
);

CREATE TABLE community (
                           id UUID PRIMARY KEY,
                           owner_id UUID REFERENCES "user"(id) ON DELETE CASCADE,
                           name TEXT NOT NULL UNIQUE,
                           description TEXT,
                           created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE community_user (
                                id SERIAL PRIMARY KEY,
                                community_id UUID REFERENCES community(id) ON DELETE CASCADE,
                                user_id UUID REFERENCES "user"(id) ON DELETE CASCADE,
                                role INT NOT NULL DEFAULT 0,
                                joined_at TIMESTAMP NOT NULL DEFAULT now(),
                                UNIQUE (community_id, user_id)
);

CREATE TABLE user_follow (
                             id SERIAL PRIMARY KEY,
                             following_id UUID REFERENCES "user"(id) ON DELETE CASCADE,
                             followed_id UUID REFERENCES "user"(id) ON DELETE CASCADE,
                             UNIQUE (following_id, followed_id)
);

---- create above / drop below ----

DROP TABLE IF EXISTS user_follow;
DROP TABLE IF EXISTS community_user;
DROP TABLE IF EXISTS community;
DROP TABLE IF EXISTS message;
DROP TABLE IF EXISTS chat_user;
DROP TABLE IF EXISTS chat;
DROP TABLE IF EXISTS friendship;
DROP TABLE IF EXISTS like_comment;
DROP TABLE IF EXISTS like_post;
DROP TABLE IF EXISTS repost;
DROP TABLE IF EXISTS post_file;
DROP TABLE IF EXISTS comment;
DROP TABLE IF EXISTS post;
DROP TABLE IF EXISTS profile;
DROP TABLE IF EXISTS "user";
