CREATE TABLE crypto_currencies(
    chain_id INTEGER NOT NULL,
    contract_address TEXT NOT NULL,
    symbol TEXT NOT NULL,
    decimals INTEGER NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE TRIGGER update_crypto_currencies_modified_time 
    AFTER UPDATE ON crypto_currencies
    FOR EACH ROW
    WHEN NEW.updated_at = OLD.updated_at OR NEW.updated_at IS NULL
BEGIN
    UPDATE crypto_currencies 
    SET updated_at = unixepoch()
    WHERE chain_id = NEW.chain_id;
END;