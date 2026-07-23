-- Up migration: categories, tickets, ticket_status_history tables and enums

CREATE TYPE ticket_priority AS ENUM ('low', 'medium', 'high');
CREATE TYPE ticket_status   AS ENUM ('open', 'in_progress', 'resolved', 'closed');

CREATE TABLE categories (
    id    SERIAL PRIMARY KEY,
    name  TEXT UNIQUE NOT NULL
);

CREATE TABLE tickets (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_number SERIAL UNIQUE,
    title         TEXT NOT NULL,
    description   TEXT NOT NULL,
    category_id   INT REFERENCES categories(id),
    priority      ticket_priority NOT NULL DEFAULT 'medium',
    status        ticket_status NOT NULL DEFAULT 'open',
    created_by    UUID NOT NULL REFERENCES users(id),
    assigned_to   UUID REFERENCES users(id),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at   TIMESTAMPTZ
);

CREATE TABLE ticket_status_history (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id   UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    changed_by  UUID NOT NULL REFERENCES users(id),
    old_status  ticket_status,
    new_status  ticket_status NOT NULL,
    changed_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Seed initial default categories
INSERT INTO categories (name) VALUES 
    ('Hardware'), 
    ('Software'), 
    ('Network'), 
    ('Access & Auth'), 
    ('General IT') 
ON CONFLICT DO NOTHING;

-- Indices
CREATE INDEX idx_tickets_status ON tickets(status);
CREATE INDEX idx_tickets_priority ON tickets(priority);
CREATE INDEX idx_tickets_assigned_to ON tickets(assigned_to);
CREATE INDEX idx_tickets_created_by ON tickets(created_by);
