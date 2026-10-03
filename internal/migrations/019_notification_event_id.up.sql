ALTER TABLE notifications_types ADD COLUMN event_id INTEGER REFERENCES group_events(id) ON DELETE CASCADE;
