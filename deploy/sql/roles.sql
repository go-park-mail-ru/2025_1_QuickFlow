-- community --
CREATE ROLE app_community LOGIN PASSWORD 'community_password';

GRANT SELECT, INSERT, UPDATE, DELETE ON
    community,
    community_user,
    contact_info
    TO app_community;

-- feedback --
CREATE ROLE app_feedback LOGIN PASSWORD 'feedback_password';
GRANT SELECT, INSERT, UPDATE, DELETE ON feedback TO app_feedback;

-- friends --
CREATE ROLE app_friends LOGIN PASSWORD 'friends_password';
GRANT SELECT, INSERT, UPDATE, DELETE ON friendship TO app_friends;
GRANT SELECT ON "user" TO app_friends;
GRANT SELECT ON profile TO app_friends;
GRANT SELECT ON education TO app_friends;
GRANT SELECT ON faculty TO app_friends;
GRANT SELECT ON university TO app_friends;

-- messenger --
CREATE ROLE app_messenger LOGIN PASSWORD 'messenger_password';
GRANT ALL ON TABLE chat TO app_messenger;
GRANT ALL ON TABLE chat_user TO app_messenger;
GRANT ALL ON TABLE message TO app_messenger;
GRANT ALL ON TABLE message_file TO app_messenger;

-- post --
CREATE ROLE app_post LOGIN PASSWORD 'post_password';
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE post TO app_post;
GRANT SELECT, INSERT, DELETE ON TABLE post_file TO app_post;
GRANT SELECT, INSERT, DELETE ON TABLE like_post TO app_post;
GRANT SELECT ON TABLE friendship TO app_post;
GRANT SELECT ON TABLE community_user TO app_post;

-- user --
CREATE ROLE app_user LOGIN PASSWORD 'user_password';
GRANT ALL ON TABLE profile TO app_user;
GRANT ALL ON TABLE school TO app_user;
GRANT ALL ON TABLE contact_info TO app_user;
GRANT ALL ON TABLE university TO app_user;
GRANT ALL ON TABLE faculty TO app_user;
GRANT ALL ON TABLE education TO app_user;
GRANT ALL ON TABLE "user" TO app_user;
