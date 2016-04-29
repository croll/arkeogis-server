/* ArkeoGIS - The Geographic Information System for Archaeologists
 * Copyright (C) 2015-2016 CROLL SAS
 *
 * Authors :
 *  Nicolas Dimitrijevic <nicolas@croll.fr>
 *  Christophe Beveraggi <beve@croll.fr>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

// This tool convert the xml file generated by SQL Designer to a PostgreSQL importable file
// You have to name the input file "in.xml", output will be print in stdout

package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"regexp"

	config "github.com/croll/arkeogis-server/config"
)

// Theses struct are a modelisation of the XML file
type Relation struct {
	XMLName xml.Name `xml:"relation"`
	Table   string   `xml:"table,attr"`
	Row     string   `xml:"row,attr"`
}

type Row struct {
	XMLName       xml.Name   `xml:"row"`
	Name          string     `xml:"name,attr"`
	Null          int        `xml:"null,attr"`
	Autoincrement int        `xml:"autoincrement,attr"`
	Datatype      string     `xml:"datatype"`
	Default       string     `xml:"default"`
	Comment       string     `xml:"comment"`
	Relations     []Relation `xml:"relation"`
	PsqlType      string
}

type Key struct {
	XMLName xml.Name `xml:"key"`
	Name    string   `xml:"name,attr"`
	Type    string   `xml:"type,attr"`
	Parts   []string `xml:"part"`
}

type Table struct {
	XMLName xml.Name `xml:"table"`
	Name    string   `xml:"name,attr"`
	Rows    []Row    `xml:"row"`
	Keys    []Key    `xml:"key"`
}

type Sql struct {
	XMLName xml.Name `xml:"sql"`
	Tables  []Table  `xml:"table"`
}

// func mysqlToPsqlType will return a string representing the postgresql type of the original type in SQL Designer
func mysqlToPsqlType(row Row) string {
	if row.Autoincrement > 0 {
		return "serial"
	}

	if row.Datatype == "DATETIME" {
		return "timestamp"
	}

	if row.Datatype == "MEDIUMTEXT" {
		return "text"
	}

	if row.Datatype == "bit" {
		return "boolean"
	}

	if row.Datatype == "VARCHAR(POINT)" {
		return "geography(POINT,4326)"
	}

	if row.Datatype == "VARCHAR(POLYGON)" {
		return "geography(POLYGON,4326)"
	}

	if row.Datatype == "VARCHAR(MULTIPOLYGON)" {
		return "geography(MULTIPOLYGON,4326)"
	}

	return row.Datatype
}

// func printPsql will print all the postgresql code from the Sql structure
func printPsql(sql Sql) {
	types := ""
	creates := ""
	geoms := ""
	constraints := ""
	indexes := ""

	var constraintRegexp = regexp.MustCompile(`^xmltopsql:"([a-z]+)\:{1}([a-z]+)"`)

	for i1 := range sql.Tables {
		table := &sql.Tables[i1]
		creates += fmt.Sprintf("CREATE TABLE \"%s\" (\n", table.Name)
		for i2 := range table.Rows {
			row := &table.Rows[i2]
			nullstr := ""
			if row.Null == 0 {
				nullstr = " NOT NULL"
			}
			row.PsqlType = mysqlToPsqlType(*row)

			// use comment to define constraint params
			constraintFromComment := ""
			if len(row.Comment) > 0 {
				tmp := constraintRegexp.FindStringSubmatch(row.Comment)
				if len(tmp) == 3 {
						switch(tmp[1]) {
						case "ondelete":
							constraintFromComment = "ON DELETE "+strings.ToUpper(tmp[2])
						case "onupdate":
							constraintFromComment = "ON UPDATE "+strings.ToUpper(tmp[2])
						}
				}
			}

			// special case for enums in postgres, we have to create a type
			if strings.Index(row.Datatype, "ENUM(") == 0 {
				types += fmt.Sprintf("CREATE TYPE %s_%s AS %s;\n", table.Name, row.Name, row.Datatype)
				row.PsqlType = fmt.Sprintf("%s_%s", table.Name, row.Name)
			}
			if strings.Index(row.Datatype, "BIT(") == 0 {
				types += fmt.Sprintf("CREATE TYPE %s_%s AS %s;\n", table.Name, row.Name, strings.Replace(row.Datatype, "BIT", "ENUM", -1))
				row.PsqlType = fmt.Sprintf("%s_%s", table.Name, row.Name)
			}

			creates += fmt.Sprintf("  \"%s\" %s%s,\n", row.Name, row.PsqlType, nullstr)

			// create all foreign keys
			for _, relation := range row.Relations {
				constraints += fmt.Sprintf("ALTER TABLE \"%s\" ADD CONSTRAINT \"c_%s.%s\" FOREIGN KEY (\"%s\") REFERENCES \"%s\" (\"%s\") %s DEFERRABLE INITIALLY DEFERRED;\n", table.Name, table.Name, row.Name, row.Name, relation.Table, relation.Row, constraintFromComment)

				// also create an index for the foreign key.
				indexes += fmt.Sprintf("CREATE INDEX \"i_%s.%s\" ON \"%s\" (\"%s\");\n",
					table.Name, row.Name, table.Name, row.Name)
			}
		}

		// search all indexs for this table
		for _, key := range table.Keys {
			if key.Type == "PRIMARY" {
				creates += fmt.Sprintf("  PRIMARY KEY (\"%s\")\n", strings.Join(key.Parts, "\", \""))
			}

			if key.Type == "INDEX" {

				// search if the index concern a geographic column, so we use the correct index type to it
				isgeo := false
				for _, keyname := range key.Parts {
					for _, row := range table.Rows {
						if row.Name == keyname {
							if strings.Index(row.PsqlType, "geography(") == 0 {
								isgeo = true
							}
						}
					}
				}

				if isgeo {
					indexes += fmt.Sprintf("CREATE INDEX \"i_%s.%s\" ON \"%s\" USING GIST ( %s );\n",
						table.Name, strings.Join(key.Parts, ","), table.Name, strings.Join(key.Parts, "\", \""))
				} else {
					indexes += fmt.Sprintf("CREATE INDEX \"i_%s.%s\" ON \"%s\" ( \"%s\" );\n",
						table.Name, strings.Join(key.Parts, ","), table.Name, strings.Join(key.Parts, "\", \""))
				}

			} else if key.Type == "UNIQUE" {
				indexes += fmt.Sprintf("CREATE UNIQUE INDEX \"i_%s.%s\" ON \"%s\" ( \"%s\" );\n",
					table.Name, strings.Join(key.Parts, ","), table.Name, strings.Join(key.Parts, "\", \""))
			}
		}
		creates += fmt.Sprintf(");\n\n")
	}
	fmt.Println(types)
	fmt.Println(creates)
	fmt.Println(geoms)
	fmt.Println(constraints)
	fmt.Println(indexes)
}

func main() {
	filepath := config.DevDistPath + "/src/github.com/croll/arkeogis-server/db-schema.xml"
	//fmt.Println("Opening: " + filepath)
	in, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Println("Unable to open input file : ", err)
		os.Exit(1)
	}

	x := Sql{}

	err2 := xml.Unmarshal(in, &x)
	if err2 != nil {
		log.Printf("error: %v", err2)
		return
	}
	//fmt.Println("result: ", x)
	//fmt.Println("result len: ", len(x.Tables))

	printPsql(x)
}
