CREATE TABLE `resources`
(
    `id`            CHAR(36) PRIMARY KEY,
    `parent_id`     CHAR(36)     NOT NULL DEFAULT '',
    `name`          VARCHAR(255) NOT NULL,
    `identifier`    VARCHAR(255) NOT NULL UNIQUE,
    `type`          VARCHAR(50)  NOT NULL,
    `path`          VARCHAR(255),
    `component`     VARCHAR(255),
    `display_order` INT,
    `description`   TEXT         NOT NULL default '',
    `metadata`      JSON,
    `status`        VARCHAR(50)  NOT NULL,
    `created_at`    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE = InnoDB
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

CREATE INDEX `idx_resources_parent_id` ON `resources` (`parent_id`);
CREATE INDEX `idx_resources_name` ON `resources` (`name`);
CREATE INDEX `idx_resources_type` ON `resources` (`type`);
CREATE INDEX `idx_resources_display_order` ON `resources` (`display_order`);
