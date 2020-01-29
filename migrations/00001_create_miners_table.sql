-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE IF NOT EXISTS `miners` (
  `id` varchar(255) NOT NULL,
  `user_id` varchar(255) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `status` varchar(100) DEFAULT NULL,
  `last_ping_at` timestamp NULL DEFAULT NULL,
  `current_task_id` varchar(255) DEFAULT NULL,
  `address` varchar(255) DEFAULT NULL,
  `tags` JSON DEFAULT NULL,
  `system_info` JSON DEFAULT NULL,
  `crypto_info` JSON DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE miners;
