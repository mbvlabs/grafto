-- name: StoreToken :exec
insert into tokens
    (id, created_at, hash, expires_at, scope, user_id) values ($1, $2, $3, $4, $5, $6) 
returning *;

-- name: QueryTokenByHash :one
select * from tokens where hash=$1;

-- name: DeleteToken :exec
delete from tokens where id=$1;
