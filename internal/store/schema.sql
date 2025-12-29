CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE,
    credentials BLOB -- WebAuthn credentials
);

CREATE TABLE IF NOT EXISTS photos (
    day TEXT PRIMARY KEY,
    id TEXT,
    filepath TEXT,
    thumbnail_path TEXT,
    lat REAL,
    lon REAL,
    notes TEXT,
    exif_data TEXT,
    created_at DATETIME
);

CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY,
    user_id TEXT,
    expires_at DATETIME
);
