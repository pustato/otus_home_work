-- +goose Up
-- +goose StatementBegin
ALTER TABLE events
    ALTER COLUMN title TYPE text USING title::text;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events
    ALTER COLUMN title TYPE varchar(100) USING title::varchar(100);
-- +goose StatementEnd
