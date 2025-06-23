package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/gertd/go-pluralize"
	"github.com/go-jet/jet/v2/generator/metadata"
	"github.com/go-jet/jet/v2/generator/sqlite"
	"github.com/go-jet/jet/v2/generator/template"
	sqlite2 "github.com/go-jet/jet/v2/sqlite"
	"github.com/iancoleman/strcase"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dsn, exists := os.LookupEnv("SQLITE_DSN")
	if !exists {
		panic("SQLITE_DSN should be set")
	}
	path := "./jet"
	pluralizer := pluralize.NewClient()
	regex := regexp.MustCompile("(?i)(s_|es_|ies_)")
	
	err := sqlite.GenerateDSN(
		dsn, 
		path, 
		template.Default(sqlite2.Dialect).
			UseSchema(func(schemaMetaData metadata.Schema) template.Schema {
				return template.DefaultSchema(schemaMetaData).
					UseModel(
						template.DefaultModel().
							UseTable(func(table metadata.Table) template.TableModel {
								if table.Name == "schema_migrations" {
									return template.TableModel{Skip: true}
								}
								return template.DefaultTableModel(table).
								UseTypeName(strcase.ToCamel(pluralizer.Singular(regex.ReplaceAllString(table.Name, "_")))).
								UseField(func(columnMetaData metadata.Column) template.TableModelField {
										defaultTableModelField := template.DefaultTableModelField(columnMetaData)
										if columnMetaData.DataType.Name == "INTEGER" {
											defaultTableModelField.Type = template.Type{
												Name: "int64",
											}
										}
										// if strings.Contains(defaultTableModelField.Type.Name, "*") {
										// 	defaultTableModelField.Type = template.Type{
										// 		ImportPath: "github.com/LukaGiorgadze/gonull/v2",
										// 		Name: "gonull.Nullable["+strings.TrimPrefix(defaultTableModelField.Type.Name, "*")+"]",
										// 	}
										// }
										
										return defaultTableModelField
								})
							}).
							UsePath("models"),
					).
					UseSQLBuilder(template.DefaultSQLBuilder().
					UseTable(func(table metadata.Table) template.TableSQLBuilder {
						if table.Name == "schema_migrations" {
							return template.TableSQLBuilder{Skip: true}
						}
						return template.DefaultTableSQLBuilder(table).
						UseDefaultAlias(strcase.ToCamel(pluralizer.Singular(regex.ReplaceAllString(table.Name, "_")))).
						UsePath("tables")
					}))
					// UsePath(path + "../../..")
			}),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("Generated models and tables successfully")
}