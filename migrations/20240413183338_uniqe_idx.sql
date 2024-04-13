-- +goose Up
-- +goose NO TRANSACTION
CREATE UNIQUE INDEX CONCURRENTLY users_to_chats_idx on users_to_chats(user_id, chat_id);
CREATE UNIQUE INDEX CONCURRENTLY messages_users_to_chats_idx on messages(user_id, chat_id);

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS users_to_chats_idx;
DROP INDEX IF EXISTS messages_users_to_chats_idx;
-- +goose StatementEnd
