package main

//select COLUMN_NAME,DATA_TYPE,COLUMN_TYPE  from information_schema.COLUMNS where  table_schema = 'dbiot' and  table_name = 'mqtt_acl';
//COLUMN_NAME,DATA_TYPE,COLUMN_TYPE
import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

type Column struct {
	ColumnName      string        `json:"colname"`
	ColumnJsonName      string        `json:"coljsonname"`
	DataType        string        `json:"datatype"`
	CharMaxLen          int         `json:"maxlen"`
	ColumnType      string        `json:"coltype"`
	Nump            int           `json:"nump"`
	Nums            int           `json:"nums"`
	Comment         string        `json:"comment"`
	ColumnKey       string        `json:"columnkey"`
	Extra           string        `json:"extra"`
	OrdinalPosition string        `json:"position"`
	ModelTag          template.HTML `json:"modeltag"`
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
	Package string   `json:"package'`
	Model   string   `json:"model'`
	ModelL  string   `json:"modell"`
	ModelApi  template.JS   `json:"modelapi"`
	DefaultObj template.JS `json:"defaultobj"`
	Columns []Column `json:"columns"`
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

var datatypemap map[string]string = map[string]string{
	"int":      "int",
	"bigint":   "uint",
	"datetime": "restgo.DateTime",
	"date":     "restgo.Date",
	"varchar":  "string",
	"bit":      "int",
	"decimal":  "float64",
	"numeric":  "float64",
}

//Col int
func datatype(col Column) string {
	t := col.DataType
	r, ok := datatypemap[t]
	if ok {
		return r
	} else {
		return t
	}
}

//构造tag
func buildtag(col Column,useGorm bool) template.HTML {
	uname := transfer(col.ColumnName)
	lname := lcfirst(uname)
	if col.ColumnName == "id" {
		return `restgo.BaseModel`
	}
	ret := uname + " " + datatype(col) + " " + " `" + "json:\"" + lname + "\" form:\"" + lname + "\""
	if col.DataType == "date" || col.DataType == "datetime" {
		ret = ret + ` time_format:"2006-01-02 15:04:05" time_utc:"1"`
	}
	if useGorm{
		ret = ret + ` gorm:"comment:`+col.Comment
		if col.DataType == "varchar" {
			if(col.CharMaxLen==0){
				col.CharMaxLen = 250
			}
			ret = ret + `;type:varchar(`+ strconv.Itoa(col.CharMaxLen)+`)`
		}
		ret = ret + "\"` "
	}else{
		ret = ret + "`"
	}


	return template.HTML(ret)
}

//配置文件
type Config struct {

	Table    string `mapstructure:"table" json:"table"`
	Dns     string `mapstructure:"dns" json:"dns"`
	Model    string `mapstructure:"model" json:"model"`
	Package  string `mapstructure:"package" json:"package"`
	Dstdir   string `mapstructure:"dstdir" json:"dstdir"`
	Lang     string `mapstructure:"lang" json:"lang"`
	Tpldir  string `mapstructure:"tpldir" json:"tpldir"`
	ServerDir string `mapstructure:"serverdir" json:"serverdir"`
	FrontDir string `mapstructure:"frontdir" json:"frontdir"`
}

var table = flag.String("t", "test", "table name")
var modelin = flag.String("m", "", "out model")

const  dnsStr =  "root:root@(127.0.0.1:3306)/test"
var dns = flag.String("dns", dnsStr, "dns link to mysql")

var lang = flag.String("l", "go", "code language,eg:go || java || php ")
//#
var pkg = flag.String("pkg", "turinapp", "application package")
var cfgpath = flag.String("c", "./restgo.yaml", "config file path")

//代码模板路径
var tpldir = flag.String("tpldir", "./tmpl-go", "templete for code ")

//前端代码路径
var serverdir = flag.String("serverdir", "./server", "dir for server code")

//前端代码路径
var frontdir = flag.String("frontdir", "./console/src", "dir for front code")



var showversion = flag.Bool("v", false, "show restctl version")

var model = ""
var config *Config = new(Config)

const version = `
 ____  _____ ____  _____  ____  _____  _    
/  __\/  __// ___\/__ __\/   _\/__ __\/ \   
|  \/||  \  |    \  / \  |  /    / \  | |   
|    /|  /_ \___ |  | |  |  \_   | |  | |_/\
\_/\_\\____\\____/  \_/  \____/  \_/  \____/ restcrl@0.0.2,
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
		flag.CommandLine.Parse([]string{"-h"})
	} else {
		flag.Parse()
	}

	fmt.Println(version)
	//如果需要展示版本号
	if exist, err := PathExists(*cfgpath); err != nil || !exist {
		if err!=nil{
			fmt.Println(err.Error())
		}else{
			if !exist{
				f, _ := os.OpenFile(*cfgpath, os.O_WRONLY|os.O_CREATE, 0666) //打开文件
				f.Close()
			}
		}
		                                                  //写入文件(字符串)
	}
	//如果需要reversion数据库

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

	if config.Lang == "" {
		v.SetDefault("lang", "go")
	}

	//设置模板
	if *tpldir!="./tmpl-go"{
		v.SetDefault("tpldir", *tpldir)
		config.Tpldir = *tpldir
	}

	//设置模板
	if *serverdir!="./server"{
		v.SetDefault("serverdir", *serverdir)
		config.ServerDir = *serverdir
	}

	//设置模板
	if *frontdir!="./console/src"{
		v.SetDefault("frontdir", *frontdir)
		config.ServerDir = *frontdir
	}

	//设置模板
	if *tpldir!="./tmpl-go"{
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

	//如果指定默认-pkg参数 则 默认package
	if *lang != "go" {
		v.SetDefault("lang", *lang)
		config.Lang = *lang
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
	columns := make([]Column, 0)

	//解析得到数据库名称
	dbname := "test"
	arr := strings.Split(config.Dns,"/")
	arr2 := strings.Split(arr[1],"?")
	dbname =  arr2[0]

	rows, err := MtsqlDb.Query(`select COLUMN_NAME ,DATA_TYPE,IFNULL(CHARACTER_MAXIMUM_LENGTH,0),COLUMN_TYPE,IFNULL(NUMERIC_PRECISION,0),IFNULL(NUMERIC_SCALE,0),COLUMN_COMMENT,column_key,extra,ORDINAL_POSITION  from information_schema.COLUMNS where  table_schema = ? and  table_name = ?`, dbname,config.Table)
	if err != nil {
		fmt.Println(err)
		return
	}
	//
	for rows.Next() {
		col :=Column{}
		err := rows.Scan(&col.ColumnName, &col.DataType,&col.CharMaxLen, &col.ColumnType, &col.Nump, &col.Nums, &col.Comment, &col.ColumnKey, &col.Extra, &col.OrdinalPosition)
		if err!=nil{
			fmt.Println(err.Error())
			return
		}
		//转换成abC的形式
		col.ColumnJsonName = lcfirst(transfer(col.ColumnName))
		col.ModelTag = buildtag(col,true)
		col.ArgTag = buildtag(col,false)
		columns = append(columns, col)
	}

	tmpl, err :=template.ParseGlob("tmpl-"+config.Lang+"/*")
	if err != nil {
		fmt.Println(err)
		return
	}

	tpls := []string{
		"server/args", "server/model", "server/ctrl", "server/service",
	}

	for _, tpl := range tpls {
		os.MkdirAll(config.ServerDir+"/"+tpl, fs.FileMode(os.O_CREATE))
		f, err := os.OpenFile(config.ServerDir+"/"+tpl+"/"+model+"."+config.Lang, os.O_WRONLY|os.O_CREATE, 0766)
		if err != nil {
			log.Fatalln(err.Error())
				return
		}

		tmpl.ExecuteTemplate(f, tpl, DstData{
			Package: config.Package,
			Model:   ucfirst(transfer(model)),
			ModelL:  lcfirst(transfer(model)),
			Columns: columns,
		})
	}

	os.MkdirAll(config.FrontDir+"/views/"+model, fs.FileMode(os.O_CREATE))
	f, err := os.OpenFile(config.FrontDir+"/views/"+model+"/list.vue", os.O_WRONLY|os.O_CREATE, 0766)
	if err != nil {
		fmt.Println(err)
		return
	}
	tmpl.ExecuteTemplate(f, "view/list", DstData{
		Package: config.Package,
		Model:   ucfirst(transfer(model)),
		ModelL:  lcfirst(transfer(model)),
		ModelApi: template.JS(lcfirst(transfer(model))+"Api"),
		Columns: columns,
	})
	//并不需要创建目录
	os.MkdirAll(config.FrontDir+"/api/", fs.FileMode(os.O_CREATE))
	f, err = os.OpenFile(config.FrontDir+"/api/"+model+".js", os.O_WRONLY|os.O_CREATE, 0766)
	if err != nil {
		fmt.Println(err)
		return
	}
	tmpl.ExecuteTemplate(f, "view/api", DstData{
		Package: config.Package,
		Model:   ucfirst(transfer(model)),
		ModelL:  lcfirst(transfer(model)),
		ModelApi: template.JS(lcfirst(transfer(model))+"Api"),
		Columns: columns,
	})

}
