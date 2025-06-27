CREATE TABLE addresses(
    chain_id INTEGER NOT NULL,
    address TEXT NOT NULL,

    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE TRIGGER update_addresses_modified_time 
    AFTER UPDATE ON addresses
    FOR EACH ROW
    WHEN NEW.updated_at = OLD.updated_at OR NEW.updated_at IS NULL
BEGIN
    UPDATE addresses 
    SET updated_at = unixepoch()
    WHERE chain_id = NEW.chain_id;
END;