CREATE ROLE app_user LOGIN PASSWORD 'user_password';

GRANT ALL ON TABLE profile TO app_user;
GRANT ALL ON TABLE school TO app_user;
GRANT ALL ON TABLE contact_info TO app_user;
GRANT ALL ON TABLE university TO app_user;
GRANT ALL ON TABLE faculty TO app_user;
GRANT ALL ON TABLE education TO app_user;
GRANT ALL ON TABLE "user" TO app_user;
