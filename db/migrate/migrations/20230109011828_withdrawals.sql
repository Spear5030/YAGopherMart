-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS withdrawals
(
    number  varchar primary key,
    user_id  integer references users not null ,
    sum numeric(12,2) not null ,
    proccesed_at timestamp with time zone default now()
);
-- +goose StatementEnd

-- +goose Down
-- DROP TABLE if exists withdrawals;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd