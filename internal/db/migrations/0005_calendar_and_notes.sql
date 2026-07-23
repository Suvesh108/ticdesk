-- Up migration: schedule_events and user_notes for Outlook Calendar & Notes module

CREATE TABLE schedule_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title       TEXT NOT NULL,
    description TEXT,
    event_type  TEXT NOT NULL DEFAULT 'maintenance',
    start_time  TIMESTAMPTZ NOT NULL,
    end_time    TIMESTAMPTZ NOT NULL,
    created_by  UUID NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE user_notes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    content     TEXT NOT NULL,
    color       TEXT NOT NULL DEFAULT 'blue',
    is_pinned   BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_schedule_events_start ON schedule_events(start_time);
CREATE INDEX idx_user_notes_user_id ON user_notes(user_id);
