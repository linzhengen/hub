CREATE TABLE `permissions`
(
    `id`          CHAR(36) PRIMARY KEY,
    `verb`        VARCHAR(255) NOT NULL,
    `resource_id` CHAR(36)     NOT NULL,
    `description` TEXT NOT NULL default '',
    `created_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (`resource_id`) REFERENCES `resources` (`id`) ON DELETE CASCADE
) ENGINE = InnoDB
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

CREATE INDEX `idx_permissions_verb` ON `permissions` (`verb`);
CREATE INDEX `idx_permissions_resource_id` ON `permissions` (`resource_id`);
