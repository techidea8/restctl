#安装
```
go install github.com/techidea8/restctl
```

#配置文件说明
```yaml
#数据库连接串,目前只支持mysql
dns: root:root@(127.0.0.1:3306)/test?charset=utf8mb4&loc=Local
#前端代码存放底子好
frontdir: ./console/src
#后端代码语言:go|java
lang: go
#对应表名称
table: test
#生成模型文件
model: test
#对应项目package
package: turingapp
#服务端代码路径
serverdir: ./server
#代码模板
tpldir: ./tmpl-go
```


#使用方法
```bash
Usage of restctl.exe:
  -c string
        config file path (default "./restgo.yaml")
  -dns string
        dns link to mysql (default "root:root@(127.0.0.1:3306)/test")
  -frontdir string
        dir for front code (default "./console/src")
  -l string
        code language,eg:go || java || php  (default "go")
  -m string
        out model
  -pkg string
        application package (default "turinapp")
  -serverdir string
        dir for server code (default "./server")
  -t string
        table name (default "test")
  -tpldir string
        templete for code  (default "./tmpl-go")
  -v    show restctl version
```

# 效果
系统将自动生成如下文件
```bash
restctl -c ./restgo.yaml  -t biz_activity -m activity 
```
```bash
├─console
│  └─src
│      ├─api
│      │      activity.js
│      │
│      └─views
│          └─activity
│                  list.vue
│
├─server
│  └─server
│      ├─args
│      │      activity.go
│      │
│      ├─ctrl
│      │      activity.go
│      │
│      ├─model
│      │      activity.go
│      │
│      └─service
│              activity.go


```