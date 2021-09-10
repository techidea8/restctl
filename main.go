package main

//select COLUMN_NAME,DATA_TYPE,COLUMN_TYPE  from information_schema.COLUMNS where  table_schema = 'dbiot' and  table_name = 'mqtt_acl';
//COLUMN_NAME,DATA_TYPE,COLUMN_TYPE
import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

type Column struct {
	ColumnName      string        `json:"colname"`
	ColumnJsonName  string        `json:"coljsonname"`
	DataType        string        `json:"datatype"`
	CharMaxLen      int           `json:"maxlen"`
	ColumnType      string        `json:"coltype"`
	Nump            int           `json:"nump"`
	Nums            int           `json:"nums"`
	Comment         string        `json:"comment"`
	DataTypeJava    string        `json:"datatypejava"`
	DataTypeGo      string        `json:"datatypego"`
	ColumnKey       string        `json:"columnkey"`
	Extra           string        `json:"extra"`
	OrdinalPosition string        `json:"position"`
	ModelTag        template.HTML `json:"modeltag"`
	ArgTag          template.HTML `json:"argtag"`
}

func (col *Column) IsKey() bool {
	return col.ColumnKey == "PRI"
}

func (col *Column) AutoIncrement() bool {
	return strings.Index(col.Extra, "auto_increment") > -1
}

func (col *Column) Build() string {
	return col.ColumnName + " " + col.DataType
}

type DstData struct {
	Package    string      `json:"package'`
	Model      string      `json:"model'`
	TableName  string      `json:"tablename"`
	ModelL     string      `json:"modell"`
	ModelApi   template.JS `json:"modelapi"`
	DefaultObj template.JS `json:"defaultobj"`
	Columns    []Column    `json:"columns"`
	Comment    string      `json:"comment"`
	Now        string      `json:"now"`
}

func ucfirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}
func lcfirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

//abc_def_ghi=> AbcDefGhi
func transfer(in string) string {
	dstdata := make([]string, 0)
	inarr := strings.Split(in, "_")
	for _, v := range inarr {
		dstdata = append(dstdata, ucfirst(v))
	}
	return strings.Join(dstdata, "")
}

var datatypemapgo map[string]string = map[string]string{
	"int":      "int",
	"bigint":   "uint",
	"datetime": "restgo.DateTime",
	"date":     "restgo.Date",
	"varchar":  "string",
	"bit":      "bool",
	"decimal":  "float64",
	"numeric":  "float64",
	"float":    "float64",
	"text":     "string",
	"tinyint":  "int",
}

var datatypemapjava map[string]string = map[string]string{
	"int":      "Integer",
	"bigint":   "Long",
	"datetime": "Timestamp",
	"date":     "Timestamp",
	"varchar":  "String",
	"bit":      "Boolean",
	"decimal":  "BigDecimal",
	"numeric":  "BigDecimal",
	"float":    "BigDecimal",
	"text":     "String",
	"tinyint":  "Integer",
}

//Col int
func datatype(col Column, lang string) string {
	//tinyint(1) 特殊处理
	if lang == "go" {
		if col.ColumnType == "tinyint(1)" {
			return "bool"
		}
		t := col.DataType
		r, ok := datatypemapgo[t]
		if ok {
			return r
		} else {
			return t
		}
	} else if lang == "java" {
		if col.ColumnType == "tinyint(1)" {
			return "Boolean"
		}
		t := col.DataType
		r, ok := datatypemapjava[t]
		if ok {
			return r
		} else {
			return t
		}
	} else {
		return col.DataType
	}

}

var baseModel []string = []string{
	"create_at", "update_by", "create_by", "delete_at", "update_at", "deleted",
}

func contains(arr []string, str string) bool {
	ret := false
	for _, v := range arr {
		if v == str {
			ret = true
		}
	}
	return ret
}

//构造tag
func buildtag(col Column, useGorm bool, lang string) template.HTML {
	uname := transfer(col.ColumnName)
	lname := lcfirst(uname)
	if col.ColumnName == "id" {
		return `restgo.BaseModel`
	}
	//如果是一些关键数值那么直接处理
	if contains(baseModel, col.ColumnName) {
		return ""
	}
	ret := uname + " " + datatype(col, lang) + " " + " `" + "json:\"" + lname + "\" form:\"" + lname + "\""
	if col.DataType == "date" || col.DataType == "datetime" {
		ret = ret + ` time_format:"2006-01-02 15:04:05" time_utc:"1"`
	}
	if useGorm {

		ret = ret + ` gorm:"comment:` + col.Comment
		if col.DataType == "varchar" {
			if col.CharMaxLen == 0 {
				col.CharMaxLen = 250
			}
			ret = ret + `;type:varchar(` + strconv.Itoa(col.CharMaxLen) + `)`
		}
		ret = ret + "\"` "
	} else {
		ret = ret + "`"
	}

	return template.HTML(ret)
}

//配置文件
type Config struct {
	Table   string `mapstructure:"table" json:"table"`
	Dns     string `mapstructure:"dns" json:"dns"`
	Model   string `mapstructure:"model" json:"model"`
	Package string `mapstructure:"package" json:"package"`
	Dstdir  string `mapstructure:"dstdir" json:"dstdir"`
	Lang    string `mapstructure:"lang" json:"lang"`
	Tpldir  string `mapstructure:"tpldir" json:"tpldir"`
}

var table = flag.String("t", "test", "table name")
var modelin = flag.String("m", "", "out model")

const dnsStr = "root:root@(127.0.0.1:3306)/test"

var dns = flag.String("dns", dnsStr, "dns link to mysql")

//#
var pkg = flag.String("pkg", "turinapp", "application package")
var cfgpath = flag.String("c", "./restgo.yaml", "config file path")

//代码模板路径
var tpldir = flag.String("tpldir", "", "templete for code ")

var showversion = flag.Bool("v", false, "show restctl version")

//根据数据库生成全部代码
var reverse = flag.Bool("reverse", false, "generate code from all table in curent database")
var trimprefix = flag.String("trimprefix", "", "trim the prefix of tablename used for model, use `,` to trim more than one")

var lang = flag.String("lang", "go", "language eg:go/java")

var model = ""
var config *Config = new(Config)

const version = `
 ____  _____ ____  _____  ____  _____  _    
/  __\/  __// ___\/__ __\/   _\/__ __\/ \   
|  \/||  \  |    \  / \  |  /    / \  | |   
|    /|  /_ \___ |  | |  |  \_   | |  | |_/\
\_/\_\\____\\____/  \_/  \____/  \_/  \____/ restctl@0.0.8,

email=271151388@qq.com,author=winlion,all rights reserved!

`

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
func main() {
	if len(os.Args) == 1 {
		fmt.Println(version)
		flag.CommandLine.Parse([]string{"-h"})
	} else {
		flag.Parse()
	}

	fmt.Println(version)
	//如果需要展示版本号
	if exist, err := PathExists(*cfgpath); err != nil || !exist {
		if err != nil {
			fmt.Println(err.Error())
		} else {
			if !exist {
				f, _ := os.OpenFile(*cfgpath, os.O_WRONLY|os.O_CREATE, 0666) //打开文件
				f.Close()
			}
		}
		//写入文件(字符串)
	}
	//如果需要reverse数据库

	v := viper.New()

	v.SetConfigFile(*cfgpath)
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	v.Unmarshal(config)

	if config.Model == "" {
		v.SetDefault("model", "test")
	}
	if config.Table == "" {
		v.SetDefault("table", "test")
	}

	if config.Lang != *lang {
		v.SetDefault("table", *lang)
	}

	//设置模板
	if *tpldir != "" {
		v.SetDefault("tpldir", *tpldir)
		config.Tpldir = *tpldir
	}

	if *table != "test" {
		config.Table = *table
		v.SetDefault("table", *table)
	}
	if *modelin != "" {
		config.Model = *modelin
		v.SetDefault("model", *modelin)
	}

	if *dns != dnsStr {
		v.SetDefault("dns", *dns)
		config.Dns = *dns
	}
	//如果指定默认-pkg参数 则 默认package
	if *pkg != "turinapp" {
		v.SetDefault("package", *pkg)
		config.Package = *pkg
	}

	v.WriteConfig()
	if *showversion {
		return
	}

	model = config.Model
	if model == "" {
		model = config.Table
	}
	model = strings.ToLower(model)
	// Open方法第二个参数:  用户名:密码@协议(ip:端口)/数据库
	dnsstr := config.Dns
	//fmt.Println(dnsstr)
	MtsqlDb, err := sql.Open("mysql", dnsstr)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer MtsqlDb.Close()

	//解析得到数据库名称
	dbname := "test"
	arr := strings.Split(config.Dns, "/")
	arr2 := strings.Split(arr[1], "?")
	dbname = arr2[0]
	tables := make([]string, 0)
	if !*reverse {
		tables = append(tables, config.Table)
	} else {
		rows, err := MtsqlDb.Query(`select table_name from information_schema.tables where table_schema=?`, dbname)
		if err != nil {
			fmt.Println(err)
			return
		}
		for rows.Next() {
			tablename := ""
			err := rows.Scan(&tablename)
			if err != nil {
				fmt.Println(err)
				return
			}
			tables = append(tables, tablename)
		}
	}
	tmpls, err := template.ParseGlob(config.Tpldir + "/*")
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Println("tables->"+strings.Join(tables,","))
	for _, tablename := range tables {
		columns := make([]Column, 0)
		//是否从数据库生成
		if *reverse {
			trimprefixs := strings.Split(*trimprefix, ",")
			model = tablename
			for i, _ := range trimprefixs {
				model = strings.TrimPrefix(model, trimprefixs[i])
			}

		} else {
			//不是
			if *modelin == "" {
				model = tablename
			} else {
				model = *modelin
			}

		}

		rows, err := MtsqlDb.Query(`select COLUMN_NAME ,DATA_TYPE,IFNULL(CHARACTER_MAXIMUM_LENGTH,0),COLUMN_TYPE,IFNULL(NUMERIC_PRECISION,0),IFNULL(NUMERIC_SCALE,0),COLUMN_COMMENT,column_key,extra,ORDINAL_POSITION  from information_schema.COLUMNS where  table_schema = ? and  table_name = ?`, dbname, tablename)
		if err != nil {
			fmt.Println(err)
			return
		}
		//
		for rows.Next() {

			col := Column{}
			err := rows.Scan(&col.ColumnName, &col.DataType, &col.CharMaxLen, &col.ColumnType, &col.Nump, &col.Nums, &col.Comment, &col.ColumnKey, &col.Extra, &col.OrdinalPosition)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			//转换成abC的形式
			col.ColumnJsonName = lcfirst(transfer(col.ColumnName))
			col.ModelTag = buildtag(col, true, config.Lang)
			col.ArgTag = buildtag(col, false, config.Lang)
			col.DataTypeGo = datatype(col, "go")
			col.DataTypeJava = datatype(col, "java")
			columns = append(columns, col)
		}
		//输出表注释

		//输出模板
		dstdata := new(DstData)
		dstdata.Package = config.Package
		dstdata.Model = ucfirst(transfer(model))
		dstdata.ModelL = lcfirst(transfer(model))
		dstdata.Columns = columns
		dstdata.ModelApi = template.JS(lcfirst(transfer(model)) + "Api")
		dstdata.TableName = tablename
		dstdata.Now = time.Now().Format("2006-01-02 15:04:05")
		//
		comments, err := MtsqlDb.Query(`select table_comment from information_schema.tables where table_schema=? and table_name = ?`, dbname, tablename)

		if err != nil {
			fmt.Println(err)
			return
		}

		for comments.Next() {
			tmp := ""
			comments.Scan(&tmp)
			dstdata.Comment = tmp
		}

		for _, tpl := range tmpls.Templates() {
			tplName := tpl.Name()
			//过滤掉以html结尾的
			if strings.HasSuffix(tplName, ".html") {
				continue
			}
			//将
			dstFile := strings.ReplaceAll(tplName, "[model]", dstdata.ModelL)
			dstFile = strings.ReplaceAll(dstFile, "[Model]", dstdata.Model)
			pkgpath := strings.ReplaceAll(dstdata.Package, ".", "/")
			dstFile = strings.ReplaceAll(dstFile, "[pkgpath]", pkgpath)

			dstFile = strings.TrimSuffix(dstFile, ".tpl")
			os.MkdirAll(filepath.Dir(dstFile), fs.FileMode(os.O_CREATE))

			f, err := os.OpenFile(dstFile, os.O_WRONLY|os.O_CREATE, 0766)

			if err != nil {
				log.Fatalln(err.Error())
				return
			}
			//文件需要再次清空
			err = f.Truncate(0)
			if err != nil {
				log.Fatalln(err.Error())
				return
			}

			tpl.ExecuteTemplate(f, tplName, *dstdata)

			f.Close()
			buf, _ := ioutil.ReadFile(dstFile)
			content := string(buf)
			content = strings.ReplaceAll(content, "&lt;", "<")
			ioutil.WriteFile(dstFile, []byte(content), 0766)
		}
		fmt.Println("generate code " + tablename + "->" + model + " √")
	}

}
