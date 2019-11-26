## yml格式说明
----
```yml
format: json
itemFormat:
project:
  name: gitlab-config-server
  description: gitlab-config-server
  id: 123
configs:
  projecta:
    - configa
    - configb
    - configc
  projectb:
    - configb
    - configc
    - configd
branch: test

```
### `format`
表示输出文件格式，值如下：

`js`(es6)

`json`

`ini`

`yml`

------
### `project`
表示工程信息，主要是相关配置更新后用于重启服务用的

> `id` 项目在gitlab中的ID

> `name` 项目名称

> `description` 项目描述信息

------


### `branch`
表示要输出的分支配置

------
### `itemFormat`
表示生成配置项的方式，取值如下

`ignore` 忽略项目名

`prefix` 项目名+配置名

`dot` 项目名+ `.` +连接配置

`tree`  保持层级结构

`project_without_prefix` 当前项目不加前缀，其它项目同 `prefix`

`project_without_dot`    当前项目不用`.`连接，其它项目同 `dot`

`project_without_tree`  当前项目不保持层级结构，其它项目同 `tree`

### `configs`
项目配置信息，具体参考后台数据


### `CI配置`
所有配置信息更新后，会触发依赖项目的`CI`并携带参数 `readboy_trigger=config`，如果某些阶段在自动触发的CI中不要执行，可以在`CI`文件中加上条件
```
only:
  variables
    - $config_trigger == 'gitlab_config_server'
```
或
```
except:
  variables
    - $config_trigger == 'gitlab_config_server'
```

