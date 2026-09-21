CREATE TABLE IF NOT EXISTS group_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    title VARCHAR(100) NOT NULL,
    description VARCHAR(1000) NOT NULL DEFAULT '',
    event_time DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS group_event_responses (
    event_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    response INTEGER NOT NULL CHECK (response IN (1, 0)),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (event_id, user_id),
    FOREIGN KEY (event_id) REFERENCES group_events(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_group_events_group_id
ON group_events(group_id);

CREATE INDEX IF NOT EXISTS idx_group_event_responses_event_id
ON group_event_responses(event_id);

CREATE INDEX IF NOT EXISTS idx_group_event_responses_user_id
ON group_event_responses(user_id);
