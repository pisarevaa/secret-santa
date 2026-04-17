CREATE TABLE users (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    email       TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE groups (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    invite_code  TEXT NOT NULL UNIQUE,
    title        TEXT NOT NULL,
    organizer_id INTEGER NOT NULL REFERENCES users(id),
    status       TEXT NOT NULL DEFAULT 'open',
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    drawn_at     DATETIME
);

CREATE TABLE memberships (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id      INTEGER NOT NULL REFERENCES groups(id),
    user_id       INTEGER NOT NULL REFERENCES users(id),
    wishlist      TEXT NOT NULL DEFAULT '',
    recipient_id  INTEGER REFERENCES users(id),
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(group_id, user_id)
);

CREATE TABLE magic_links (
    token       TEXT PRIMARY KEY,
    email       TEXT NOT NULL,
    expires_at  DATETIME NOT NULL,
    used_at     DATETIME
);

CREATE TABLE sessions (
    token       TEXT PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id),
    expires_at  DATETIME NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE messages (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id       INTEGER NOT NULL REFERENCES groups(id),
    sender_id      INTEGER NOT NULL REFERENCES users(id),
    recipient_id   INTEGER NOT NULL REFERENCES users(id),
    direction      TEXT NOT NULL,
    body           TEXT NOT NULL,
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_messages_pair ON messages(group_id, sender_id, recipient_id, created_at);
CREATE INDEX idx_memberships_group ON memberships(group_id);
