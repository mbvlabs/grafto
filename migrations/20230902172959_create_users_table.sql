-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
create table if not exists users (
    id uuid not null,
    primary key (id),
    created_at timestamp with time zone not null,
    updated_at timestamp with time zone not null,
    name varchar(255) not null,
    mail varchar(255) unique not null,
    mail_verified_at timestamp with time zone,
    password text not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
drop table if exists users;
-- +goose StatementEnd
