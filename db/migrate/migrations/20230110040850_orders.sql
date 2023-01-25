-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS orders
(
    number varchar primary key,
    user_id  integer references users not null ,
    status  integer references order_status not null,
    accrual numeric(12,2),
    uploaded_at timestamptz default now(),
    updated_at timestamptz
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd