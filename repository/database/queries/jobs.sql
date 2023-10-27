-- name: InsertJob :exec
insert into jobs
    (id, created_at, updated_at, failed_attempts, state, instructions, repeatable_job_id, scheduled_for, executor)
values
    ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: QueryJobs :many
update jobs
    set state = $1, updated_at = $2
    where id in (
        select id
        from jobs as inner_jobs
        where inner_jobs.state = sqlc.arg(inner_state)::int 
        and inner_jobs.scheduled_for::timestamptz <= sqlc.arg(inner_scheduled_for)::timestamptz 
        and inner_jobs.failed_attempts < sqlc.arg(inner_failed_attempts)::int
        order by inner_jobs.scheduled_for
        for update skip locked
        limit $3
    )
returning *;

-- name: DeleteJob :exec
delete from jobs where id = $1;

-- name: FailJob :exec
update jobs
    SET state = $1, updated_at = $2, scheduled_for = $3, failed_attempts = failed_attempts + 1
WHERE id = $4;

-- name: ClearJobs :exec
delete from jobs;

-- name: CheckIfRepeatableJobExists :one
select exists(select 1 from jobs where repeatable_job_id = $1);

-- name: RescheduleRepeatableJob :exec
update jobs
    set state = $1, updated_at = $2, scheduled_for  = $3, failed_attempts = 0
    where id = $4;
