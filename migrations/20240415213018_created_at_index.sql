-- +goose Up
-- +goose NO TRANSACTION
-- +goose StatementBegin
CREATE UNIQUE INDEX CONCURRENTLY messages_created_at on messages(created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS messages_created_at;
-- +goose StatementEnd
