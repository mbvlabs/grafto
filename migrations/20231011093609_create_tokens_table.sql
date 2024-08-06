-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
create table if not exists tokens (
    id uuid not null,
    primary key (id),
    created_at timestamp with time zone not null,
    hash text not null,
    expires_at timestamp with time zone not null,
	meta_information jsonb not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
drop table if exists tokens;
-- +goose StatementEnd
