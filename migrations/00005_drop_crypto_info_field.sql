-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE miners DROP `crypto_info`;


-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE miners ADD `crypto_info` JSON DEFAULT NULL;