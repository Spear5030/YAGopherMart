-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- DROP TABLE order_status cascade;
CREATE TABLE IF NOT EXISTS order_status
(   id integer primary key,
    status varchar(10)
);
INSERT INTO order_status (id, status) VALUES (1, 'NEW');
INSERT INTO order_status (id, status) VALUES (2, 'PROCESSING');
INSERT INTO order_status (id, status) VALUES (3, 'INVALID');
INSERT INTO order_status (id, status) VALUES (4, 'PROCESSED');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd