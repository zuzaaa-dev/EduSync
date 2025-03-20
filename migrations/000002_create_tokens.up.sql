CREATE TABLE tokens
(
    id            SERIAL PRIMARY KEY,
    user_id       INT       NOT NULL,
    access_token  TEXT      NOT NULL,
    refresh_token TEXT      NOT NULL,
    user_agent    VARCHAR(255),
    ip_address    VARCHAR(45),
    expires_at    TIMESTAMP NOT NULL,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
