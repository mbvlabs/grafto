-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
create schema if not exists queue;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
drop schema if exists queue cascade;
-- +goose StatementEnd
