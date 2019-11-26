#!/bin/bash
config="
format: ini
project:
  name: ${CI_PROJECT_NAME}
  description: ${CI_PROJECT_TITLE}
  id: ${CI_PROJECT_ID}
itemFormat: project_without_dot
configs:
  gitlab-config-server:
    - CopyRequestBody
    - GitLabClientId
    - GitLabClientSecret
    - GitLabDomain
    - GitLabToken
    - Graceful
    - apiToken
    - appname
    - dbhost
    - dbname
    - dbpasswd
    - dbuser
    - domain
    - httpport
    - redisDB
    - redisHost
    - redisPasswd
    - runmode
    - servername
  gitlab-config-web:
    - domain
    - testkey
branch: ${CI_COMMIT_REF_NAME}
"
curl "${GITLAB_CONFIG_SERVER}" -fd "$config"