-- Unique visitors per day. visitor is a daily-rotating HMAC (never an IP);
-- rows older than yesterday are deleted on the next visit.
CREATE TABLE visits (
    day     DATE NOT NULL,
    visitor TEXT NOT NULL,
    PRIMARY KEY (day, visitor)
);

CREATE TABLE counters (
    name TEXT PRIMARY KEY,
    n    BIGINT NOT NULL CHECK (n >= 0)
);
INSERT INTO counters (name, n) VALUES ('visitors', 0);

-- visitor is kept only long enough for rate limits, then set to NULL.
CREATE TABLE guestbook (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT NOT NULL CHECK (char_length(name) BETWEEN 1 AND 40),
    homepage   TEXT NOT NULL DEFAULT '' CHECK (char_length(homepage) <= 200),
    message    TEXT NOT NULL CHECK (char_length(message) BETWEEN 1 AND 500),
    visitor    TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX guestbook_recent ON guestbook (created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX guestbook_visitor ON guestbook (visitor, created_at) WHERE visitor IS NOT NULL;
