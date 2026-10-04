INSERT INTO `muscle_groups` (
  `muscle_group_id`,
  `muscle_group_name`,
  `muscle_group_display_name`
) VALUES
  (1, 'chest', 'Chest'),
  (2, 'back', 'Back'),
  (3, 'shoulders', 'Shoulders'),
  (4, 'biceps', 'Biceps'),
  (5, 'triceps', 'Triceps'),
  (6, 'forearms', 'Forearms'),
  (7, 'quads', 'Quadriceps'),
  (8, 'hams', 'Hamstrings'),
  (9, 'glutes', 'Glutes'),
  (10, 'calves', 'Calves'),
  (11, 'core_or_abs', 'Core / Abs')
ON DUPLICATE KEY UPDATE
  `muscle_group_name` = VALUES(`muscle_group_name`),
  `muscle_group_display_name` = VALUES(`muscle_group_display_name`);
