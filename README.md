#安装
```
go install github.com/techidea8/restctl
```

#配置文件说明
```yaml
#数据库连接串,目前只支持mysql
dns: root:root@(127.0.0.1:3306)/test?charset=utf8mb4&loc=Local
#编程语言
lang: go
#输出模型
model: test
#输出包名称
package: turingapp
#代码保存位置
serverdir: ./
#当前默认哪个表
table: test
#当前使用哪个模板
tpldir: ./tmpl-go
```


#使用方法
```bash

 ____  _____ ____  _____  ____  _____  _
/  __\/  __// ___\/__ __\/   _\/__ __\/ \
|  \/||  \  |    \  / \  |  /    / \  | |
|    /|  /_ \___ |  | |  |  \_   | |  | |_/\
\_/\_\\____\\____/  \_/  \____/  \_/  \____/ restcrl@0.0.5,
email=271151388@qq.com,author=winlion,all rights reserved!

Usage of restctl.exe:
  -c string
        config file path (default "./restgo.yaml")
  -dns string
        dns link to mysql (default "root:root@(127.0.0.1:3306)/test")
  -m string
        out model
  -pkg string
        application package (default "turinapp")
  -reverse
        generate code from all table in curent database
  -t string
        table name (default "test")
  -tpldir string
        templete for code  (default "./tmpl-go")
  -trimprefix string
        trim the prefix of tablename used for model
  -v    show restctl version
```

# 使用示例
## 根据数据表生成代码
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

## 根据数据库生成代码
假设数据库有如下表
`trkf_account,trkf_attach,trkf_charge`
则运行如下指令后系统系统将自动生成`account,attach,charge`三个模块
```bash
restctl -c ./restgo.yaml  -reverse -trimprefix trkf_ 
```
