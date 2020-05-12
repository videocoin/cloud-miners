-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE miners ADD `reward` DECIMAL(10,4) DEFAULT 0;


-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE miners DROP `reward`;