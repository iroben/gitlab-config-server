CREATE TABLE `config` (
  `id`          INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `key`         VARCHAR(64)      NOT NULL,
  `val`         TEXT             NOT NULL,
  `description` VARCHAR(255)     NOT NULL,
  `project_id`  INT(11)          NOT NULL,
  `dependent`   TEXT,
  PRIMARY KEY (`id`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8