-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS users
(   id serial primary key,
    login varchar unique not null ,
    hash  varchar not null ,
    balance numeric(12,2) default 0
);


-- +goose StatementEnd

-- +goose Down
-- DROP TABLE if exists users cascade;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd