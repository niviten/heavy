CREATE TABLE IF NOT EXISTS `muscle_group` (
  `muscle_group_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `muscle_group_name` varchar(100) NOT NULL,
  `muscle_group_display_name` varchar(100) NOT NULL,
  PRIMARY KEY (`muscle_group_id`),
  UNIQUE KEY `muscle_group_unique` (`muscle_group_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_ai_ci;

CREATE TABLE IF NOT EXISTS `excercise` (
  `excercise_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `excercise_name` varchar(100) NOT NULL,
  `muscle_group_id` bigint(20) unsigned NOT NULL,
  `is_deleted` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`excercise_id`),
  UNIQUE KEY `excercise_unique` (`excercise_name`),
  KEY `excercise_muscle_group_FK` (`muscle_group_id`),
  CONSTRAINT `excercise_muscle_group_FK`
    FOREIGN KEY (`muscle_group_id`)
    REFERENCES `muscle_group` (`muscle_group_id`)
    ON DELETE CASCADE
    ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_ai_ci;
