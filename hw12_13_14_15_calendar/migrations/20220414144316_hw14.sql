-- +goose Up
-- +goose StatementBegin
ALTER TABLE events ADD notification_sent BOOLEAN DEFAULT FALSE;
CREATE INDEX events_time_end_index ON events (time_end);
CREATE INDEX events_notify_at_notification_sent_index ON events (notify_at, notification_sent);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX events_time_end_index;
DROP INDEX events_notify_at_notification_sent_index;
ALTER TABLE events DROP COLUMN notification_sent;
-- +goose StatementEnd
