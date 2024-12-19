-- name: QueryUserByID :one
select * from users where id=$1;

-- name: QueryUserByEmail :one
select * from users where email=$1;

-- name: QueryUsers :many
select * from users;

-- name: InsertUser :one
insert into
    users (id, created_at, updated_at, name, email, password)
values
    ($1, $2, $3, $4, $5, $6)
returning *;

-- name: UpdateUser :one
update users
    set updated_at=$2, name=$3, email=$4, password=$5
where id = $1
returning *;

-- name: DeleteUser :exec
delete from users where id=$1;

-- name: ChangeUserPassword :exec
update users set updated_at=$2, password=$3 where id=$1;

-- name: VerifyUserEmail :exec
update users set updated_at=$2, email_verified_at=$3 where email=$1;
