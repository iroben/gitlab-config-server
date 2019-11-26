CREATE TABLE `action_log` (
  `id`     INT(11) NOT NULL                  AUTO_INCREMENT,
  `who`    VARCHAR(32) CHARACTER SET utf8mb4 DEFAULT NULL,
  `action` VARCHAR(16)                       DEFAULT NULL,
  `data`   TEXT COMMENT 'utf8mb4',
  `time`   INT(11)                           DEFAULT NULL,
  `uid`    INT(11)                           DEFAULT NULL,
  `ip`     VARCHAR(32)                       DEFAULT NULL,
  PRIMARY KEY (`id`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

