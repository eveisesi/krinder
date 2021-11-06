CREATE TABLE `groups` (
    `category_id` INT(11) NOT NULL,
    `group_id` INT(11) NOT NULL,
    `name` VARCHAR(255) NOT NULL COLLATE 'utf8mb4_general_ci',
    `published` TINYINT(1) NOT NULL,
    `etag` VARCHAR(255) NOT NULL COLLATE 'utf8mb4_general_ci',
    `expires` DATETIME NOT NULL,
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    PRIMARY KEY (`group_id`) USING BTREE
) COLLATE = 'utf8mb4_general_ci' ENGINE = InnoDB;