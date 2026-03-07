CREATE TABLE `roles`
(
    `id`          CHAR(36) PRIMARY KEY,
    `name`        VARCHAR(255) NOT NULL,
    `description` TEXT NOT NULL default '',
    `created_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE = InnoDB
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

CREATE INDEX `idx_roles_name` ON `roles` (`name`);
