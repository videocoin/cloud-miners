-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE miners ADD `secret` TEXT DEFAULT NULL;


-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE miners DROP `secret`;