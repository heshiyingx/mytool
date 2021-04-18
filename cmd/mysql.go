/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/jmoiron/sqlx"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
)
var username string
var prefix string
var password string
var tableName string
var host string
var port string
var dataBase string
var dbr int8

// mysqlCmd represents the mysql command
var mysqlCmd = &cobra.Command{
	Use:   "mysql ",
	Short: "mysql table to struct ",
	Long: `mysql table to struct
eg:mytool mysql -H=127.0.0.1 -d=dbName -u=admin -p=pwd -P=3306 -t=prefix_order -D=1 --prefix=prefix_
`,
	Run: func(cmd *cobra.Command, args []string) {
		info := DBInfo{
			Host:     host,
			UserName: username,
			Password: password,
			Port:     port,
		}
		dbModel := &DBModel{
			DB:     nil,
			DBInfo: info,
		}
		err := dbModel.ConnectMysql()
		if err != nil {
			panic(err)
		}

		_, err = dbModel.GetColumns(dataBase, tableName)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(mysqlCmd)
	mysqlCmd.PersistentFlags().StringVarP(&username, "username", "u", "", "数据库用户名")
	mysqlCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "数据库密码")
	mysqlCmd.PersistentFlags().StringVarP(&tableName, "tableName", "t", "", "表名")
	mysqlCmd.PersistentFlags().StringVarP(&prefix, "prefix", "", "", "表名前缀")
	mysqlCmd.PersistentFlags().StringVarP(&host, "host", "H", "", "数据库地址")
	mysqlCmd.PersistentFlags().StringVarP(&port, "port", "P", "3306", "数据库端口")
	mysqlCmd.PersistentFlags().StringVarP(&dataBase, "dataBase", "d", "", "数据库名")
	mysqlCmd.PersistentFlags().Int8VarP(&dbr,"dbr","D",0,"是否使用dbr(default 0),0:不使用,1:使用")

}

func (d *DBModel)ConnectMysql()error  {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/information_schema",d.DBInfo.UserName,d.DBInfo.Password,d.DBInfo.Host,d.DBInfo.Port)

	timeout, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelFunc()

	connectDB, err := sqlx.ConnectContext(timeout, "mysql", dsn)
	connectDB.SetMaxIdleConns(2)
	connectDB.SetMaxOpenConns(2)
	if err != nil {
		return err
	}
	d.DB = connectDB
	return nil
}

func (d *DBModel)GetColumns(dbName,tableName string)([]TableInfo,error)  {
	var columns []TableInfo
	err := d.DB.Select(&columns, "select COLUMN_NAME,IS_NULLABLE,DATA_TYPE,COLUMN_COMMENT from COLUMNS where TABLE_SCHEMA=? and TABLE_NAME=? order by ORDINAL_POSITION asc ", dbName, tableName)

	for i,column := range columns{
		columns[i].ColumnComment = strings.Replace(column.ColumnComment,"\n"," ",-1)
		if dbr==1 && column.IsNullable==StringBoolYes{
			columns[i].StructType = DBTypeToStructTypeDbrInfo[column.DataType]
		}else {
			columns[i].StructType = DBTypeToStructTypeInfo[column.DataType]
		}

	}

	if !strings.HasPrefix(tableName,prefix){
		prefix = ""
	}
	tpl := StructTpl{
		TableName: tableName,
		TagAround: "`",
		Columns:   columns,
		Class: strings.Replace(tableName,prefix,"",1),
	}
	tpl.Parse()


	return columns,err
}

var DBTypeToStructTypeInfo = map[string]string{
	"int":"int64",
	"tinyint":"int8",
	"smallint":"int16",
	"bigint":"int64",
	"varchar":"string",
	"text":"string",
	"mediumtext":"string",
	"json":"string",
	"longtext":"string",
	"char":"string",
	"timestamp":"time.Time",
	"datetime":"time.Time",
	"date":"time.Time",
	"time":"time.Time",
	"double":"float64",
	"decimal":"float64",
	"float":"float64",
}
var DBTypeToStructTypeDbrInfo = map[string]string{
	"int":"dbr.NullInt64",
	"tinyint":"dbr.NullInt64",
	"smallint":"dbr.NullInt64",
	"bigint":"dbr.NullInt64",
	"varchar":"dbr.NullString",
	"text":"dbr.NullString",
	"mediumtext":"dbr.NullString",
	"json":"dbr.NullString",
	"longtext":"dbr.NullString",
	"char":"dbr.NullString",
	"timestamp":"dbr.NullTime",
	"datetime":"dbr.NullTime",
	"date":"dbr.NullTime",
	"time":"dbr.NullTime",
	"double":"dbr.NullFloat64",
	"decimal":"dbr.NullFloat64",
	"float":"dbr.NullFloat64",
}

type DBModel struct {
	DB *sqlx.DB
	DBInfo DBInfo

}
type DBInfo struct {
	Host string
	UserName string
	Password string
	DataBase string
	Port string
}
type StringBool string

const (
	StringBoolYes StringBool = "YES"
	StringBoolNO StringBool = "NO"
)
type TableInfo struct {
	ColumnName string `db:"COLUMN_NAME" json:"column_name"`
	IsNullable StringBool `db:"IS_NULLABLE" json:"is_nullable"`
	DataType string `db:"DATA_TYPE" json:"data_type"`
	ColumnComment string `db:"COLUMN_COMMENT" json:"column_comment"`
	StructType string `db:"-" json:"-"`
	TagAround string `db:"-" json:"-"`
}
type StructTpl struct {
	TableName string
	TagAround string
	Class string
	Columns []TableInfo
}

const tplText =`
package model
// {{.Class | ToCamelCase}} ...
const {{.Class | ToCamelCase}}Table = "{{.TableName}}"
// {{.Class | ToCamelCase}} ...
type {{.Class | ToCamelCase}} struct {
{{- $tag := .TagAround}}
{{- range .Columns}}
	{{.ColumnName | ToCamelCase}}	{{.StructType}}		{{$tag}}db:"{{.ColumnName}}"{{$tag}} //{{.ColumnComment}}
{{- end}}
}`

func (s *StructTpl)Parse()  {
	tpl := template.Must(template.New("table2struct").Funcs(template.FuncMap{
		"ToCamelCase": ToCamel,
	}).Parse(string(tplText)))
	dir, _ := filepath.Abs("./model")
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		os.Mkdir(dir,0755)
	}
	fileName := filepath.Join(dir, s.Class+".go")
	file,err := os.OpenFile(fileName,os.O_CREATE|os.O_RDWR,0755)
	if err != nil {
		panic(err)
	}
	err = tpl.Execute(file, s)
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Millisecond*500)
	file.Close()
	time.Sleep(time.Millisecond*500)

	exec.Command("go","fmt",fileName).Run()

}

func ToCamel(s string) string {
	if s == "id"{
		return "ID"
	}
	if strings.HasSuffix(s,"id"){
		s = strings.Replace(s,"id","ID",1)
	}

	return strcase.ToCamel( s)
}