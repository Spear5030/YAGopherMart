-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS withdrawals
(
    id serial primary key,
    order_id  varchar references orders not null ,
    sum numeric(12,2) not null ,
    proccesed_at timestamp with time zone default now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
