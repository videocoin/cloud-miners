-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE miners ADD `org_name` VARCHAR(255) DEFAULT NULL;
ALTER TABLE miners ADD `org_email` VARCHAR(255) DEFAULT NULL;
ALTER TABLE miners ADD `org_desc` TEXT DEFAULT NULL;
ALTER TABLE miners ADD `allow_thirdparty_delegates` TINYINT(1) DEFAULT 0;
ALTER TABLE miners ADD `delegate_policy` TEXT DEFAULT NULL;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE miners DROP `org_name`;
ALTER TABLE miners DROP `org_email`;
ALTER TABLE miners DROP `org_desc`;
ALTER TABLE miners DROP `allow_thirdparty_delegates`;
ALTER TABLE miners DROP `delegate_policy`;