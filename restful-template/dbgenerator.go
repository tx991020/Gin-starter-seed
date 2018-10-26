package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"unicode"

	_ "github.com/go-sql-driver/mysql"
)

var config Configuration

type Configuration struct {
	DbUser     string `json:"db_user"`
	DbPassword string `json:"db_password"`
	DbAddress  string `json:"db_address"`
	DbName     string `json:"db_name"`
	// DaoPkgName gives name of the package using the stucts
	DaoPkgPath   string `json:"dao_pkg_path"`
	DaoPkgName   string `json:"dao_pkg_name"`
	CachePkgName string `json:"cache_pkg_name"`
	// TagLabel produces tags commonly used to match database field names with Go struct members
	TagLabel string `json:"tag_label"`
}

type ColumnSchema struct {
	TableName              string
	ColumnName             string
	IsNullable             string
	DataType               string
	CharacterMaximumLength sql.NullInt64
	NumericPrecision       sql.NullInt64
	NumericScale           sql.NullInt64
	ColumnType             string
	ColumnKey              string
}
//生成目录
func initDirs() {
	daoPath := strings.ToLower(config.DaoPkgName)
	cachePath := strings.ToLower(config.CachePkgName)
	os.Mkdir("generated/", 0755)
	os.Mkdir("generated/"+daoPath, 0755)
	os.Mkdir("generated/"+cachePath, 0755)
	os.Mkdir("generated/"+"handler", 0755)
	os.Mkdir("generated/"+"service", 0755)
}

func writeStructs(tables map[string][]*ColumnSchema) error {
	var cacheParam = ""
	var handlerParam =""
	// To store the keys in slice in sorted order
	var keys []string
	for k := range tables {
		keys = append(keys, k)
	}
	//按表名排序
	sort.Strings(keys)
	fmt.Println(keys)
	for _, tableName := range keys {
		var columns = tables[tableName]
		//按表生成结构体
		generateModel(tableName, columns)
		//生成dao CRUD
		generateCacheCRUD(tableName)
		//生成Handler
		generateHandler(tableName)
		//生成Service
		generateService(tableName)
		ftn := formatName(tableName)
		ctn := getVarPrefix(ftn)
		cacheParam += "\n\"" + ctn + "\":" + ctn + "CacheInit,"
		handlerParam += "\n" + ctn + "Init(r)"


	}
	//生成CaChe Init
	generateCacheInit(cacheParam)
	//生成Handler Init
	generateHandlerInit(handlerParam)

	return nil
}

//反射表结构
func getSchema() map[string][]*ColumnSchema {
	conn, err := sql.Open("mysql", config.DbUser+":"+config.DbPassword+"@tcp("+config.DbAddress+")/information_schema")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	q := "SELECT TABLE_NAME, COLUMN_NAME, IS_NULLABLE, DATA_TYPE, " +
		"CHARACTER_MAXIMUM_LENGTH, NUMERIC_PRECISION, NUMERIC_SCALE, COLUMN_TYPE, " +
		"COLUMN_KEY FROM COLUMNS WHERE TABLE_SCHEMA = ? ORDER BY TABLE_NAME, ORDINAL_POSITION"
	rows, err := conn.Query(q, config.DbName)
	if err != nil {
		log.Fatal(err)
	}
	tables := make(map[string][]*ColumnSchema)
	for rows.Next() {
		cs := ColumnSchema{}
		err := rows.Scan(&cs.TableName, &cs.ColumnName, &cs.IsNullable, &cs.DataType,
			&cs.CharacterMaximumLength, &cs.NumericPrecision, &cs.NumericScale,
			&cs.ColumnType, &cs.ColumnKey)
		if err != nil {
			log.Fatal(err)
		}

		if _, ok := tables[cs.TableName]; !ok {
			tables[cs.TableName] = make([]*ColumnSchema, 0)
		}
		tables[cs.TableName] = append(tables[cs.TableName], &cs)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return tables
}

func formatName(name string) string {
	parts := strings.Split(name, "_")
	newName := ""
	for _, p := range parts {
		if len(p) < 1 {
			continue
		}
		newName = newName + strings.Replace(p, string(p[0]), strings.ToUpper(string(p[0])), 1)
	}
	return newName
}

func getVarPrefix(tableName string) string {
	r := []rune(tableName)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func getDeclarePrefix(tableName string) string {
	r1 := []rune(tableName)
	r1[0] = unicode.ToUpper(r1[0])
	return string(r1)
}
//生成cache Init 列表
func generateCacheInit(cacheParam string) {
	b, err := ioutil.ReadFile("./template/cache.go")
	if err != nil {
		fmt.Println(err)
	}

	b0 := strings.Replace(string(b), "{namelist}", cacheParam, -1)
	b1 := strings.Replace(string(b0), "{package}", strings.ToLower(config.CachePkgName), -1)
	os.Mkdir("generated/"+config.CachePkgName+"/", os.FileMode(0755))
	ioutil.WriteFile("generated/"+config.CachePkgName+"/cache.go", []byte(b1), os.ModeAppend|os.FileMode(0664))
}

//生成handler Init 列表
func generateHandlerInit(handlerParam string) {
	b, err := ioutil.ReadFile("./template/handler.go")
	if err != nil {
		fmt.Println(err)
	}

	b0 := strings.Replace(string(b), "{namelist}", handlerParam, -1)
	b1 := strings.Replace(string(b0), "{package}", strings.ToLower(config.CachePkgName), -1)
	os.Mkdir("generated/"+config.CachePkgName+"/", os.FileMode(0755))
	ioutil.WriteFile("generated/"+"handler"+"/handler.go", []byte(b1), os.ModeAppend|os.FileMode(0664))
}

func generateCacheCRUD(tableName string) {
	b, err := ioutil.ReadFile("./template/cache_template.go")
	if err != nil {
		fmt.Println(err)
	}

	ftn := formatName(tableName)
	ctn := getVarPrefix(ftn)

	b0 := strings.Replace(string(b), "{package}", strings.ToLower(config.CachePkgName), -1)
	b1 := strings.Replace(b0, "{table}", ctn, -1)
	b2 := strings.Replace(b1, "{Table}", ftn, -1)
	b3 := strings.Replace(b2, "{DaoPkgPath}", config.DaoPkgPath, -1)
	os.Mkdir("generated/"+config.CachePkgName+"/", os.FileMode(0755))
	ioutil.WriteFile("generated/"+config.CachePkgName+"/"+tableName+".go", []byte(b3), os.ModeAppend|os.FileMode(0664))
}


func generateHandler(tableName string) {
	b, err := ioutil.ReadFile("./template/handler_template.go")
	if err != nil {
		fmt.Println(err)
	}

	ftn := formatName(tableName)
	ctn := getVarPrefix(ftn)

	b0 := strings.Replace(string(b), "{package}", strings.ToLower(config.CachePkgName), -1)
	b1 := strings.Replace(b0, "{table}", ctn, -1)
	b2 := strings.Replace(b1, "{Table}", ftn, -1)
	b3 := strings.Replace(b2, "{DaoPkgPath}", config.DaoPkgPath, -1)
	os.Mkdir("generated/"+"handler/", os.FileMode(0755))
	ioutil.WriteFile("generated/"+"handler"+"/"+tableName+".go", []byte(b3), os.ModeAppend|os.FileMode(0664))
}

func generateService(tableName string) {
	b, err := ioutil.ReadFile("./template/service_template.go")
	if err != nil {
		fmt.Println(err)
	}

	ftn := formatName(tableName)
	ctn := getVarPrefix(ftn)

	b0 := strings.Replace(string(b), "{package}", strings.ToLower(config.CachePkgName), -1)
	b1 := strings.Replace(b0, "{table}", ctn, -1)
	b2 := strings.Replace(b1, "{Table}", ftn, -1)
	b3 := strings.Replace(b2, "{DaoPkgPath}", config.DaoPkgPath, -1)
	os.Mkdir("generated/"+"service/", os.FileMode(0755))
	ioutil.WriteFile("generated/"+"service"+"/"+tableName+".go", []byte(b3), os.ModeAppend|os.FileMode(0664))
}

func generateDatabase() {
	path := strings.ToLower(config.DaoPkgName)
	b, err := ioutil.ReadFile("./template/database.go")
	if err != nil {
		log.Fatal(err)
	}
	b0 := strings.Replace(string(b), "{package}", strings.ToLower(config.DaoPkgName), -1)
	ioutil.WriteFile("generated/"+path+"/database.go", []byte(b0), os.FileMode(0644))
}

func generateModel(tableName string, columns []*ColumnSchema) {
	var out bytes.Buffer
	// generate model
	ftn := formatName(tableName)
	ctn := getVarPrefix(ftn)

	out.WriteString("type ")
	out.WriteString(ftn)
	out.WriteString(" struct{\n")

	for _, column := range columns {

		//下划线转驼峰
		fcn := formatName(column.ColumnName)
		ccn := getVarPrefix(fcn)

		goType, _, err := goType(column)

		if err != nil {
			log.Fatal(err)
		}
		out.WriteString("\t")
		out.WriteString(fcn)
		out.WriteString(" ")
		out.WriteString(goType)
		if len(config.TagLabel) > 0 || goType == "*htutil.JSONTime" {
			out.WriteString("\t`")
		}
		if len(config.TagLabel) > 0 {
			if goType == "*htutil.JSONTime" {
				out.WriteString(config.TagLabel)
				out.WriteString(":\"-\" ")
			} else {
				out.WriteString(config.TagLabel)
				out.WriteString(":\"")
				out.WriteString(ccn)
				out.WriteString("\"")
			}
		}
		if goType == "*htutil.JSONTime" {
			out.WriteString("sql:\"-\"")
		}
		out.WriteString("`\n")
	}

	out.WriteString("}")

	b, err := ioutil.ReadFile("./template/model_template.go")
	if err != nil {
		fmt.Println(err)
	}
	bb := strings.Replace(string(b), "{model}", out.String(), -1)

	b0 := strings.Replace(bb, "{package}", strings.ToLower(config.DaoPkgName), -1)
	b1 := strings.Replace(b0, "{table}", ctn, -1)
	b2 := strings.Replace(b1, "{Table}", ftn, -1)
	b3 := strings.Replace(b2, "{snake}", tableName, -1)
	os.Mkdir("generated/"+config.DaoPkgName+"/", os.FileMode(0755))
	ioutil.WriteFile("generated/"+config.DaoPkgName+"/"+tableName+".go", []byte(b3), os.ModeAppend|os.FileMode(0664))
}

//数据库字段类型转go类型
func goType(col *ColumnSchema) (string, string, error) {
	requiredImport := ""
	//   if col.IsNullable == "YES" {
	//     requiredImport = "database/sql"
	//   }
	var gt string = ""
	switch col.DataType {
	case "char", "varchar", "enum", "text", "longtext", "mediumtext", "tinytext":
		//     if col.IsNullable == "YES" {
		//       gt = "sql.NullString"
		//     } else {
		gt = "string"
		//     }
	case "blob", "mediumblob", "longblob", "varbinary", "binary":
		gt = "[]byte"
	case "date", "time", "datetime", "timestamp":
		//     gt, requiredImport = "time.Time", "time"
		gt, requiredImport = "*htutil.JSONTime", "git.corp.hetao101.com/backend/htutil"
	case "smallint", "int", "mediumint", "bigint":
		//     if col.IsNullable == "YES" {
		//       gt = "sql.NullInt64"
		//     } else {
		gt = "int64"
		//     }
	case "float", "decimal", "double":
		//     if col.IsNullable == "YES" {
		//       gt = "sql.NullFloat64"
		//     } else {
		gt = "float64"
		//     }
	case "tinyint":
		gt = "bool"
	}
	if gt == "" {
		n := col.TableName + "." + col.ColumnName
		return "", "", errors.New("No compatible datatype (" + col.DataType + ") for " + n + " found")
	}
	return gt, requiredImport, nil
}

func main() {

	//dbUser := flag.String("dbuser", "root", "database user name")
	//dbPassword := flag.String("dbpassword", "password", "password for user name")
	//dbAddress := flag.String("dbaddress", "127.0.0.1:3306", "database address")
	//dbName := flag.String("dbname", "dbname", "database name")
	//daoPkgPath := flag.String("daopkgpath", "dao package path", "dao package path")
	//daoPkgName := flag.String("daopkgname", "dao package name", "dao package name")
	//cachePkgName := flag.String("cachepkgname", "cache package name", "cache package name")
	//tagLabel := flag.String("taglabel", "json", "json or xml")
	//flag.Parse()
	//
	//config.DbUser = *dbUser
	//config.DbPassword = *dbPassword
	//config.DbAddress = *dbAddress
	//config.DbName = *dbName
	//config.DaoPkgPath = *daoPkgPath
	//config.DaoPkgName = *daoPkgName
	//config.CachePkgName = *cachePkgName
	//config.TagLabel = *tagLabel

	config.DbUser = "root"
	config.DbPassword = "123456"
	config.DbAddress = "127.0.0.1:3306"
	config.DbName = "logic"
	config.DaoPkgPath = ""
	config.DaoPkgName = "dao"
	config.CachePkgName = "cache"
	config.TagLabel = "json"
	//反射表结构
	tables := getSchema()

	//生成目录
	initDirs()
	//生成databse配置
	generateDatabase()
	//结构体
	err := writeStructs(tables)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Printf("Done\n")
}
