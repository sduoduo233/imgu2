CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    email_verified BOOLEAN,
    role INTEGER,
    space INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS social_logins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL,
    user INTEGER REFERENCES users(id),
    account_id TEXT NOT NULL,
    UNIQUE (type, account_id),
    UNIQUE (type, user)
);

CREATE TABLE IF NOT EXISTS storages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL,
    config TEXT NOT NULL,
    enabled BOOLEAN,
    allow_upload BOOLEAN
);

CREATE TABLE IF NOT EXISTS images (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    storage INTEGER REFERENCES storages(id),
    uploader INTEGER,
    file_name TEXT NOT NULL UNIQUE,
    uploader_ip TEXT NOT NULL,
    time INTEGER,
    expire_time INTEGER
);

CREATE TABLE IF NOT EXISTS settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key TEXT NOT NULL UNIQUE,
    value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    token TEXT NOT NULL UNIQUE,
    user INTEGER REFERENCES users(id),
    expire_at INTEGER
);

-- default values
INSERT OR IGNORE INTO settings(key, value) VALUES('CAPTCHA', 'none');
INSERT OR IGNORE INTO settings(key, value) VALUES('RECAPTCHA_CLIENT', '');
INSERT OR IGNORE INTO settings(key, value) VALUES('RECAPTCHA_SERVER', '');

INSERT OR IGNORE INTO settings(key, value) VALUES('SITE_URL', 'http://127.0.0.1:3000');
INSERT OR IGNORE INTO settings(key, value) VALUES('SITE_NAME', 'imgu2 dev');

INSERT OR IGNORE INTO settings(key, value) VALUES('GOOGLE_SIGNIN', 'false');
INSERT OR IGNORE INTO settings(key, value) VALUES('GOOGLE_CLIENT_ID', '');
INSERT OR IGNORE INTO settings(key, value) VALUES('GOOGLE_SECRET', '');

INSERT OR IGNORE INTO settings(key, value) VALUES('GITHUB_SIGNIN', 'false');
INSERT OR IGNORE INTO settings(key, value) VALUES('GITHUB_CLIENT_ID', '');
INSERT OR IGNORE INTO settings(key, value) VALUES('GITHUB_SECRET', '');

INSERT OR IGNORE INTO settings(key, value) VALUES('MAX_IMAGE_SIZE', '10485760'); -- 10 MiB
