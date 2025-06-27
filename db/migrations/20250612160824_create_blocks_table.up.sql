CREATE TABLE blocks(
    chain_id INTEGER NOT NULL,
    block_number TEXT NOT NULL,
    block_hash TEXT NOT NULL,
    block_parent_hash TEXT NOT NULL,
    block_confirmed INTEGER DEFAULT 0,

    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE TRIGGER update_blocks_modified_time 
    AFTER UPDATE ON blocks
    FOR EACH ROW
    WHEN NEW.updated_at = OLD.updated_at OR NEW.updated_at IS NULL
BEGIN
    UPDATE blocks 
    SET updated_at = unixepoch()
    WHERE chain_id = NEW.chain_id;
END;