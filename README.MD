# GitLab配置管理API

## 介绍
配置服务的由来，请看这篇文章，[一个简单的基于gitlab的配置服务
](https://juejin.im/post/5dc0036cf265da4d4704246a)。

[前端项目](https://github.com/iroben/gitlab-config-web)

[API接口](https://github.com/iroben/gitlab-config-server)

## 安装
> 1. `conf/dev`下添加`app.conf`配置文件，配置项说明看后面
> 2. `go mod vendor` 安装依赖
> 3. `go run migrate.go` 初始化数据库信息
> 4. `go run main.go` 启动服务
> 5. 打开 `gitlab-config-web` 界面

## 注意
> 1. 第一次启动`API`或者`Web`的时候，需要手动编辑`app.conf`,`config.js`，等服务启动成功后，就可以利用配置服务自举。
> 2. 修改`.config.sh`的`${GITLAB_CONFIG_SERVER}`为`${API}/v1/generate`，或者设置`CI`的`Variables`，然后运行`./.config.sh > conf/dev/app.conf`。
> 3. 配置编辑、删除权限是基于`gitlab`的用户在这个项目权限来的，只有`MaintainerAccess`用户有编辑、删除权限，其它用户只可以查看
> 4. 项目信息会在每个用户登录后同步一次，服务器每`5`分钟同步一次，可以修改`GitLabSyncTTL`配置来设置同步时间，单位是`分钟`
> 5. `.gitlib-ci.yml`的`deploy_test`根据自己实际情况修改，直接在`gitlab`上跑应该是跑不通过的。

## .config.sh 说明

[.config.sh 说明](https://github.com/iroben/gitlab-config-server/blob/master/config.MD)


## app.conf说明
    domain = "http://localhost:9090"   #API域名
    appname = "gitlab-config-server" #随意
    httpport = 9090  #API监听的端口
    CopyRequestBody = true #不用改
    runmode = "dev"  #生产环境用prod
    Graceful = false  #生产环境用true
    servername = "nginx/1.8"  #随意

    GitLabDomain = "https://gitlab.example.cn"  #gitlab域名
    GitLabToken = ""  #用管理员账号创建一个GitLabToken，拥有API权限
    GitLabClientId = ""  #admin/applications下创建一个 application, Redirect URI配置：${API}/gitlab/callback
    GitLabClientSecret = "" #admin/applications下创建一个 application, Redirect URI配置：${API}/gitlab/callback

    dbhost = "localhost:3306"  #数据库地址
    dbuser = "root"  #数据库用户
    dbpasswd = ""   #数据库密码
    dbname = "gitlab-config" #数据库名

    redisHost = "localhost:6379"  #redis地址
    redisPasswd =  #redis密码
    redisDB = 1   #redis DB
    gitlab-config-web.domain = "http://localhost:8080/"  #gitlab web域名，用于登录成功后跳转用


## 使用规范

该配置服务是基于分支或者tag来拉取配置的，所以项目要统一分支信息要统一才能更好的利用起来，暂时不能跨分支引用，也就是说`master`分支只能引用`master`分支的，不可以引用`test`的，所有`tag`只有一份配置，如果配置修改了，先找跟该配置所在项目的最新`tag`的同名`tag`，更新依赖项目的最新`tag`，如果该模式不能满足，大家可以再扩展一下，我们公司的`.config.sh`加了个`envBranch`字段，手动配置环境和分支的映射关系，`branch`换成`env`
