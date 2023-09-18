-- name: QueryUser :one
select * from users where id=$1;

-- name: QueryUsers :many
select * from users;

-- name: InsertUser :one
insert into
    users (id, created_at, updated_at, name, mail, password)
values
    ($1, $2, $3, $4, $5, $6)
returning *;

-- name: UpdateUser :one
update users
    set updated_at=$2, name=$3, mail=$4, password=$5
where id = $1
returning *;

-- name: DeleteUser :exec
delete from users where id=$1;

-- name: DoesMailExists :one
select exists (select 1 from users where mail = $1) as does_mail_exists;
