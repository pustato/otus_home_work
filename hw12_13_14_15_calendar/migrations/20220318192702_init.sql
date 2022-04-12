-- +goose Up
-- +goose StatementBegin
CREATE TABLE events
(
    id BIGSERIAL CONSTRAINT events_pk PRIMARY KEY,
    user_id BIGINT NOT NULL,
    description TEXT NOT NULL,
    title VARCHAR (100) NOT NULL,
    time_start TIMESTAMP NOT NULL,
    time_end TIMESTAMP NOT NULL,
    notify_at TIMESTAMP NULL DEFAULT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
CREATE INDEX events_user_id_index ON events (user_id);
CREATE INDEX events_time_start_index ON events (time_start);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS events;
-- +goose StatementEnd
