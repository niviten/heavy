CREATE TABLE IF NOT EXISTS `muscle_groups` (
  `muscle_group_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `muscle_group_name` varchar(100) NOT NULL,
  `muscle_group_display_name` varchar(100) NOT NULL,
  PRIMARY KEY (`muscle_group_id`),
  UNIQUE KEY `muscle_group_unique` (`muscle_group_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_ai_ci;

CREATE TABLE IF NOT EXISTS `exercises` (
  `exercise_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `exercise_name` varchar(100) NOT NULL,
  `muscle_group_id` bigint(20) unsigned NOT NULL,
  `is_deleted` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`exercise_id`),
  UNIQUE KEY `exercise_unique` (`exercise_name`),
  KEY `exercise_muscle_group_FK` (`muscle_group_id`),
  CONSTRAINT `exercise_muscle_group_FK`
    FOREIGN KEY (`muscle_group_id`)
    REFERENCES `muscle_groups` (`muscle_group_id`)
    ON DELETE CASCADE
    ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_ai_ci;

CREATE TABLE IF NOT EXISTS `workout_routines` (
  `workout_routine_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `workout_routine_name` varchar(100) NOT NULL,
  `workout_routine_description` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`workout_routine_id`),
  UNIQUE KEY `workout_routine_unique` (`workout_routine_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_ai_ci;

CREATE TABLE IF NOT EXISTS `workout_routine_exercises` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `workout_routine_id` bigint(20) unsigned NOT NULL,
  `exercise_id` bigint(20) unsigned NOT NULL,
  `exercise_order` decimal(10,3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `workout_routine_exercises_unique` (`workout_routine_id`,`exercise_id`),
  KEY `workout_routine_exercises_exercises_FK` (`exercise_id`),
  KEY `workout_routine_exercises_workout_routine_id_IDX` (`workout_routine_id`,`exercise_order`) USING BTREE,
  CONSTRAINT `workout_routine_exercises_exercises_FK`
    FOREIGN KEY (`exercise_id`)
    REFERENCES `exercises` (`exercise_id`)
    ON DELETE CASCADE
    ON UPDATE CASCADE,
  CONSTRAINT `workout_routine_exercises_workout_routines_FK`
    FOREIGN KEY (`workout_routine_id`)
    REFERENCES `workout_routines` (`workout_routine_id`)
    ON DELETE CASCADE
    ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_ai_ci;
