-- name: CreateCustomer :one
INSERT INTO customers (
    id, email, name, phone, description, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetCustomer :one
SELECT * FROM customers
WHERE id = $1 LIMIT 1;

-- name: GetCustomerByEmail :one
SELECT * FROM customers
WHERE email = $1 LIMIT 1;

-- name: UpdateCustomer :one
UPDATE customers
SET email = $2, name = $3, phone = $4, description = $5, metadata = $6, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteCustomer :exec
DELETE FROM customers
WHERE id = $1;

-- name: ListCustomers :many
SELECT * FROM customers
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CreatePaymentMethod :one
INSERT INTO payment_methods (
    id, type, customer_id, card_last4, card_brand, card_exp_month, card_exp_year, card_fingerprint, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetPaymentMethod :one
SELECT * FROM payment_methods
WHERE id = $1 LIMIT 1;

-- name: ListPaymentMethods :many
SELECT * FROM payment_methods
WHERE customer_id = $1
ORDER BY created_at DESC;

-- name: DeletePaymentMethod :exec
DELETE FROM payment_methods
WHERE id = $1 AND customer_id = $2;

-- name: CreateCharge :one
INSERT INTO charges (
    id, amount, currency, status, customer_id, payment_method_id, description, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetCharge :one
SELECT * FROM charges
WHERE id = $1 LIMIT 1;

-- name: ListCharges :many
SELECT * FROM charges
WHERE customer_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAllCharges :many
SELECT * FROM charges
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateChargeStatus :one
UPDATE charges
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetCustomerStats :one
SELECT 
    COUNT(*) as total_customers,
    COUNT(CASE WHEN created_at >= NOW() - INTERVAL '30 days' THEN 1 END) as new_customers_30d,
    COUNT(CASE WHEN created_at >= NOW() - INTERVAL '7 days' THEN 1 END) as new_customers_7d
FROM customers;

-- name: GetChargeStats :one
SELECT 
    COUNT(*) as total_charges,
    SUM(amount) as total_amount,
    COUNT(CASE WHEN status = 'succeeded' THEN 1 END) as successful_charges,
    SUM(CASE WHEN status = 'succeeded' THEN amount ELSE 0 END) as successful_amount
FROM charges;

-- name: CreateRefund :one
INSERT INTO refunds (
    id, charge_id, amount, currency, status, reason, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetRefund :one
SELECT * FROM refunds
WHERE id = $1 LIMIT 1;

-- name: ListRefunds :many
SELECT * FROM refunds
WHERE charge_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAllRefunds :many
SELECT * FROM refunds
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateRefundStatus :one
UPDATE refunds
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetRefundStats :one
SELECT 
    COUNT(*) as total_refunds,
    SUM(amount) as total_amount,
    COUNT(CASE WHEN status = 'succeeded' THEN 1 END) as successful_refunds,
    SUM(CASE WHEN status = 'succeeded' THEN amount ELSE 0 END) as successful_amount
FROM refunds;
