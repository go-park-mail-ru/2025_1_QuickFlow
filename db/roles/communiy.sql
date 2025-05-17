
CREATE ROLE app_community LOGIN PASSWORD 'community_password';

GRANT SELECT, INSERT, UPDATE, DELETE ON
    community,
    community_user,
    contact_info
TO app_community;
