CREATE ROLE app_feedback LOGIN PASSWORD 'feedback_password';

GRANT SELECT, INSERT, UPDATE, DELETE ON feedback TO app_feedback;
