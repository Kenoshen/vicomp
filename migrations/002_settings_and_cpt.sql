-- Singleton settings row (single-user app): id is pinned to 1.
CREATE TABLE IF NOT EXISTS settings (
    id                      BIGINT PRIMARY KEY DEFAULT 1,
    make_checks_payable_to  TEXT NOT NULL DEFAULT '',
    default_session_rate    NUMERIC(10,2) NOT NULL DEFAULT 0,
    default_session_minutes INT NOT NULL DEFAULT 50,
    default_cpt_code        TEXT NOT NULL DEFAULT '90834',
    CONSTRAINT settings_singleton CHECK (id = 1)
);

INSERT INTO settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

-- Per-session CPT code, defaulting to 90834.
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS cpt_code TEXT NOT NULL DEFAULT '90834';
