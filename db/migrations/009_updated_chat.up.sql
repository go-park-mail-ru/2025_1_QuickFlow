CREATE OR REPLACE FUNCTION update_chat_updated_at()
    RETURNS TRIGGER AS $$
BEGIN
    -- Обновляем поле updated_at в таблице chat
    UPDATE chat
    SET updated_at = NOW()
    WHERE id = NEW.chat_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER trg_update_chat_updated_at
    AFTER INSERT ON message
    FOR EACH ROW
EXECUTE FUNCTION update_chat_updated_at();
