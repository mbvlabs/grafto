-- name: InsertToken :exec
insert into tokens
    (id, created_at, hash, expires_at, meta_information) values ($1, $2, $3, $4, $5) 
returning *;

-- name: QueryTokenByHash :one
select * from tokens where hash=$1;

-- name: DeleteTokenByHash :exec
delete from tokens
where hash = $1;
