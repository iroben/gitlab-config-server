CREATE TABLE `gitlab_project` (
  `id`          INT(10) UNSIGNED NOT NULL,
  `name`        VARCHAR(128)     NOT NULL,
  `description` TEXT             NOT NULL,
  `branches`    TEXT             NOT NULL,
  `tags`        TEXT             NOT NULL,
  PRIMARY KEY (`id`)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8