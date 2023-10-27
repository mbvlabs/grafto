-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
create table if not exists jobs (
  id uuid primary key,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  scheduled_for timestamptz not null,
  failed_attempts int not null,
  state int not null,
  instructions jsonb not null,
  executor varchar(255) not null,
  repeatable_job_id text unique null
);

create index index_jobs_on_scheduled_for on jobs (scheduled_for);
create index index_jobs_on_state on jobs (state);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
drop table if exists jobs;
-- +goose StatementEnd
