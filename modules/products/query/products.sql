-- name: InsertProduct :one
INSERT INTO products (content, metadata, embedding)
VALUES ($1, $2, $3)
RETURNING id, content, metadata, created_at;

-- name: CountProducts :one
SELECT COUNT(*) FROM products;

-- name: DeleteAllProducts :exec
DELETE FROM products;
