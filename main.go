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
	"tinyint":   "int",
	"smallint":  "int",
	"mediumint": "int",
	"int":       "int",
	"integer":   "int",
	"bigint":    "uint",
	"float":     "float64",
	"double":    "float64",

	"decimal":   "float64",
	"datetime":  "restgo.DateTime",
	"date":      "restgo.Date",
	"timestamp": "uint",
	"char":      "string",
	"varchar":   "string",
	"bit":       "bool",
	"numeric":   "float64",
	"text":      "string",
}

var datatypemapjava map[string]string = map[string]string{
	"tinyint":   "Integer",
	"int":       "Integer",
	"smallint":  "Integer",
	"mediumint": "Integer",
	"integer":   "Integer",
	"bigint":    "Long",
	"float":     "BigDecimal",
	"double":    "BigDecimal",
	"datetime":  "Timestamp",
	"date":      "Timestamp",
	"timestamp": "Timestamp",
	"varchar":   "String",
	"char":      "String",
	"bit":       "Boolean",
	"decimal":   "BigDecimal",
	"numeric":   "BigDecimal",
	"text":      "String",
}

//Col int
func datatype(col Column, lang string) string {
	//tinyint(1) ????????????
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

//??????tag
func buildtag(col Column, useGorm bool, lang string) template.HTML {
	uname := transfer(col.ColumnName)
	lname := lcfirst(uname)
	if col.ColumnName == "id" {
		return `restgo.BaseModel`
	}
	//?????????????????????????????????????????????
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

//????????????
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

//??????????????????
var tpldir = flag.String("tpldir", "", "templete for code ")

var showversion = flag.Bool("v", false, "show restctl version")

//?????????????????????????????????
var reverse = flag.Bool("reverse", false, "generate code from all table in curent database")
var exclude = flag.String("exclude", "", "available when use reverse, generate code from all table in curent database exclude those ,use `,` to exclude more than one ,it ")
var trimprefix = flag.String("trimprefix", "", "trim the prefix of tablename used for model, use `,` to trim more than one")

var lang = flag.String("lang", "go", "language eg:go/java")

var model = ""
var config *Config = new(Config)

const version = `
 ____  _____ ____  _____  ____  _____  _    
/  __\/  __// ___\/__ __\/   _\/__ __\/ \   
|  \/||  \  |    \  / \  |  /    / \  | |   
|    /|  /_ \___ |  | |  |  \_   | |  | |_/\
\_/\_\\____\\____/  \_/  \____/  \_/  \____/ restctl@0.0.9,

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
	//???????????????????????????
	if exist, err := PathExists(*cfgpath); err != nil || !exist {
		if err != nil {
			fmt.Println(err.Error())
		} else {
			if !exist {
				f, _ := os.OpenFile(*cfgpath, os.O_WRONLY|os.O_CREATE, 0666) //????????????
				f.Close()
			}
		}
		//????????????(?????????)
	}
	//????????????reverse?????????

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

	//????????????
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
	//??????????????????-pkg?????? ??? ??????package
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
	// Open?????????????????????:  ?????????:??????@??????(ip:??????)/?????????
	dnsstr := config.Dns
	//fmt.Println(dnsstr)
	MtsqlDb, err := sql.Open("mysql", dnsstr)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer MtsqlDb.Close()

	//???????????????????????????
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
			//????????????
			if !strings.Contains(*exclude, tablename) {
				tables = append(tables, tablename)
			}

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
		//????????????????????????
		if *reverse {
			trimprefixs := strings.Split(*trimprefix, ",")
			model = tablename
			for i, _ := range trimprefixs {
				model = strings.TrimPrefix(model, trimprefixs[i])
			}

		} else {
			//??????
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
			//?????????abC?????????
			col.ColumnJsonName = lcfirst(transfer(col.ColumnName))
			col.ModelTag = buildtag(col, true, config.Lang)
			col.ArgTag = buildtag(col, false, config.Lang)
			col.DataTypeGo = datatype(col, "go")
			col.DataTypeJava = datatype(col, "java")
			columns = append(columns, col)
		}
		//???????????????

		//????????????
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
			//????????????html?????????
			if strings.HasSuffix(tplName, ".html") {
				continue
			}
			//???
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
			//????????????????????????
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
		fmt.Println("generate code " + tablename + "->" + model + " ???")
	}

}
