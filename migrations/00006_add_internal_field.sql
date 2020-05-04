-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE miners ADD `is_internal` TINYINT(1) DEFAULT 0;


-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE miners DROP `is_internal`;