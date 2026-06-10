CREATE TABLE IF NOT EXISTS sessions (
    id                      VARCHAR(16)  PRIMARY KEY,
    picks                   JSONB        NOT NULL DEFAULT '[]',
    constructor_skips_left  INT          NOT NULL DEFAULT 1,
    era_skips_left          INT          NOT NULL DEFAULT 1,
    pending_spin            JSONB,
    wins                    INT,
    tier                    VARCHAR(80),
    completed               BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS leaderboard (
    id          SERIAL       PRIMARY KEY,
    session_id  VARCHAR(16)  REFERENCES sessions(id) ON DELETE CASCADE,
    name        VARCHAR(50)  NOT NULL,
    wins        INT          NOT NULL,
    tier        VARCHAR(80)  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (session_id)
);

CREATE INDEX IF NOT EXISTS leaderboard_wins_idx ON leaderboard (wins DESC, created_at ASC);
