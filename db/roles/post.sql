CREATE ROLE app_post LOGIN PASSWORD 'post_password';

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE post TO app_post_service;
GRANT SELECT, INSERT, DELETE ON TABLE post_file TO app_post_service;
GRANT SELECT, INSERT, DELETE ON TABLE like_post TO app_post_service;
GRANT SELECT ON TABLE friendship TO app_post_service;
GRANT SELECT ON TABLE community_user TO app_post_service;

GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_post_service;
