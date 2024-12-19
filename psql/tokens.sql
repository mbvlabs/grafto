-- name: InsertToken :one
INSERT INTO tokens (
    id,
    created_at,
    hash,
    expires_at,
    meta_information
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
) RETURNING *;

-- name: QueryTokenByID :one
SELECT * FROM tokens
WHERE id = $1;

-- name: QueryTokenByHash :one
SELECT * FROM tokens
WHERE hash = $1;

-- name: QueryValidTokens :many
SELECT * FROM tokens
WHERE expires_at > NOW();

-- name: QueryExpiredTokens :many
SELECT * FROM tokens
WHERE expires_at <= NOW();

-- name: UpdateTokenExpiresAt :one
UPDATE tokens
SET expires_at = $2
WHERE id = $1
RETURNING *;

-- name: UpdateTokenMetaInformation :one
UPDATE tokens
SET meta_information = $2
WHERE id = $1
RETURNING *;

-- name: DeleteToken :exec
DELETE FROM tokens
WHERE id = $1;

-- name: DeleteExpiredTokens :exec
DELETE FROM tokens
WHERE expires_at <= NOW();

