CREATE TABLE chains_states(
    chain_id INTEGER PRIMARY KEY NOT NULL,
    last_block_number TEXT NULL,
    last_block_hash TEXT NULL,
    
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE TRIGGER update_chains_states_modified_time 
    AFTER UPDATE ON chains_states
    FOR EACH ROW
    WHEN NEW.updated_at = OLD.updated_at OR NEW.updated_at IS NULL
BEGIN
    UPDATE chains_states 
    SET updated_at = unixepoch()
    WHERE chain_id = NEW.chain_id;
END;