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
-- +goose StatementBegin
DROP TABLE withdrawals;
SELECT 'down SQL query';
-- +goose StatementEnd