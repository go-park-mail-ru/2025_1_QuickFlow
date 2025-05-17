CREATE ROLE app_friends LOGIN PASSWORD 'friends_password';

GRANT SELECT, INSERT, UPDATE, DELETE ON friendship TO app_friends;

GRANT SELECT ON "user" TO app_friends;
GRANT SELECT ON profile TO app_friends;
GRANT SELECT ON education TO app_friends;
GRANT SELECT ON faculty TO app_friends;
GRANT SELECT ON university TO app_friends;

