package store

// Schema is the index schema, executed as one string.
const Schema = `
CREATE TABLE IF NOT EXISTS chunks(
 id INTEGER PRIMARY KEY, session_id TEXT NOT NULL, source_store TEXT NOT NULL, agent TEXT NOT NULL,
 project TEXT NOT NULL, source_path TEXT NOT NULL, source_ref TEXT NOT NULL, seq INTEGER NOT NULL,
 window_idx INTEGER NOT NULL, ts INTEGER NOT NULL, role TEXT NOT NULL, text TEXT NOT NULL,
 content_hash TEXT NOT NULL, is_error INTEGER NOT NULL DEFAULT 0,
 tool_id TEXT NOT NULL DEFAULT '', tool_name TEXT NOT NULL DEFAULT '',
 UNIQUE(session_id, source_ref, window_idx));
CREATE INDEX IF NOT EXISTS chunks_ts ON chunks(ts);
CREATE INDEX IF NOT EXISTS chunks_hash ON chunks(content_hash);
CREATE TABLE IF NOT EXISTS cursors(store TEXT NOT NULL, key TEXT NOT NULL, cursor TEXT NOT NULL, skipped INTEGER NOT NULL DEFAULT 0, PRIMARY KEY(store,key));
`
