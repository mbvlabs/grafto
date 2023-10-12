-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
create table if not exists tokens (
    id uuid not null,
    primary key (id),
    created_at timestamp with time zone not null,
    hash text not null,
    expires_at timestamp with time zone not null,
    scope varchar(255) not null,
    user_id uuid not null,
    constraint fk_tokens_user_id foreign key (user_id) references users(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
drop table if exists tokens;
-- +goose StatementEnd
