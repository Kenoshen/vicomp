CREATE TABLE IF NOT EXISTS counties (
    id           BIGSERIAL PRIMARY KEY,
    name         TEXT NOT NULL,
    contact_name TEXT NOT NULL DEFAULT '',
    address      TEXT NOT NULL DEFAULT '',
    city         TEXT NOT NULL DEFAULT '',
    state        TEXT NOT NULL DEFAULT '',
    zip          TEXT NOT NULL DEFAULT '',
    phone        TEXT NOT NULL DEFAULT '',
    email        TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS clinicians (
    id          BIGSERIAL PRIMARY KEY,
    first_name  TEXT NOT NULL,
    last_name   TEXT NOT NULL,
    credentials TEXT NOT NULL DEFAULT '',
    npi         TEXT NOT NULL DEFAULT '',
    tax_id      TEXT NOT NULL DEFAULT '',
    address     TEXT NOT NULL DEFAULT '',
    city        TEXT NOT NULL DEFAULT '',
    state       TEXT NOT NULL DEFAULT '',
    zip         TEXT NOT NULL DEFAULT '',
    phone       TEXT NOT NULL DEFAULT '',
    email       TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS clients (
    id            BIGSERIAL PRIMARY KEY,
    first_name    TEXT NOT NULL,
    last_name     TEXT NOT NULL,
    date_of_birth DATE,
    county_id     BIGINT REFERENCES counties(id) ON DELETE SET NULL,
    clinician_id  BIGINT REFERENCES clinicians(id) ON DELETE SET NULL,
    claim_number  TEXT NOT NULL DEFAULT '',
    notes         TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS invoices (
    id             BIGSERIAL PRIMARY KEY,
    client_id      BIGINT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    invoice_number TEXT NOT NULL UNIQUE,
    status         TEXT NOT NULL DEFAULT 'draft',
    total          NUMERIC(10,2) NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sessions (
    id               BIGSERIAL PRIMARY KEY,
    client_id        BIGINT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    session_date     DATE NOT NULL,
    duration_minutes INT NOT NULL DEFAULT 0,
    amount           NUMERIC(10,2) NOT NULL DEFAULT 0,
    notes            TEXT NOT NULL DEFAULT '',
    invoice_id       BIGINT REFERENCES invoices(id) ON DELETE SET NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_clients_county_id ON clients(county_id);
CREATE INDEX IF NOT EXISTS idx_clients_clinician_id ON clients(clinician_id);
CREATE INDEX IF NOT EXISTS idx_sessions_client_id ON sessions(client_id);
CREATE INDEX IF NOT EXISTS idx_sessions_invoice_id ON sessions(invoice_id);
CREATE INDEX IF NOT EXISTS idx_invoices_client_id ON invoices(client_id);
