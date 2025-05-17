CREATE ROLE app_messenger LOGIN PASSWORD 'messenger_password';

GRANT ALL ON TABLE chat TO app_messenger;
GRANT ALL ON TABLE chat_user TO app_messenger;
GRANT ALL ON TABLE message TO app_messenger;
GRANT ALL ON TABLE message_file TO app_messenger;
