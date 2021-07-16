#安装
```
go install github.com/techidea8/restgo/restctl@latest
```

#配置文件说明
```yaml
#数据库地址
addr: 127.0.0.1:3306
#数据库名称
database: test
#数据库用户名
username: root
#数据库密码
password: test
#要对哪个表进行处理
table: test
#模型名称
model: test
#应用包名称
package: turinapp
#代码输出目录
dstdir: ./
#默认语言输出go/java
lang: go
```


#使用方法
```bash
Usage of restctl:
  -addr string
        mysql database host (default "127.0.0.1:3306")
  -c string
        config file path (default "./restgo.yaml")
  -db string
        database name (default "test")
  -init
        init restgo project
  -m string
        out model
  -o string
        dist dir (default "./")
  -p string
        password
  -pkg string
        application package (default "turinapp")
  -t string
        table name (default "test")
  -u string
        user name (default "root")
  -v    show restctl version
```

# 效果
系统将自动生成如下文件
```bash
restctl -addr 127.0.0.1:3306 -c ./restgo.yaml -db test -t biz_order -m order  -u root -p root  -o ./code -pkg turingapp
```
```bash
├─server
│  ├─args
│  │      order.go
│  │
│  ├─ctrl
│  │      order.go
│  │
│  ├─model
│  │      order.go
│  │
│  └─service
│          order.go
│
└─ui
    ├─api
    │      order.js
    │
    └─view
        └─order
                list.vue

```