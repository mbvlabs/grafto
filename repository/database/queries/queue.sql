-- name: InsertJob :exec
insert into queue
    (id, created_at, updated_at, scheduled_for, failed_attempts, state, message, processor)
values
    ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: QueryJobs :many
update queue
    set state = $1, updated_at = $2
    where id in (
        select id
        from queue as inner_queue
        where inner_queue.state = sqlc.arg(inner_state)::int 
        and inner_queue.scheduled_for::time <= sqlc.arg(inner_scheduled_for)::time 
        and inner_queue.failed_attempts < sqlc.arg(inner_failed_attempts)::int
        order by inner_queue.scheduled_for
        for update skip locked
        limit $3
    )
returning *;

-- name: DeleteJob :exec
delete from queue where id = $1;

-- name: FailJob :exec
update queue
    SET state = $1, updated_at = $2, scheduled_for = $3, failed_attempts = failed_attempts + 1
WHERE id = $4;

-- name: ClearQueue :exec
delete from queue;
