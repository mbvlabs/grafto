-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
create table if not exists queue (
  id uuid primary key,
  created_at timestamp with time zone not null,
  updated_at timestamp with time zone not null,
  scheduled_for timestamp with time zone not null,
  failed_attempts int not null,
  state int not null,
  message jsonb not null,
  processor varchar(255) not null
);

create index index_queue_on_scheduled_for on queue (scheduled_for);
create index index_queue_on_state on queue (state);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
drop table if exists queue;
-- +goose StatementEnd
