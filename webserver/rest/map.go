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

package rest

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	db "github.com/croll/arkeogis-server/db"
	"github.com/croll/arkeogis-server/model"
	routes "github.com/croll/arkeogis-server/webserver/routes"
	"github.com/jmoiron/sqlx"
)

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:        "/api/map/search",
			Description: "Main map search function",
			Func:        MapSearch,
			Method:      "POST",
			Json:        reflect.TypeOf(MapSearchParams{}),
		},
	}
	routes.RegisterMultiple(Routes)
}

type MapSqlQuery struct {
	Tables        map[string]bool
	GroupedWheres map[string][]string
}

func (sql *MapSqlQuery) Init() {
	sql.Tables = map[string]bool{}
	sql.GroupedWheres = make(map[string][]string)
}

func (sql *MapSqlQuery) AddTable(name string) {
	sql.Tables[name] = true
}

func (sql *MapSqlQuery) AddFilter(filtergroup string, where string) {
	if _, ok := sql.GroupedWheres[filtergroup]; ok {
		sql.GroupedWheres[filtergroup] = append(sql.GroupedWheres[filtergroup], where)
	} else {
		sql.GroupedWheres[filtergroup] = []string{
			where,
		}
	}
}

func (sql *MapSqlQuery) BuildQuery() string {
	q := "SELECT site.id FROM site"

	// dependences
	if join, ok := sql.Tables["site_range__charac"]; ok && join {
		sql.AddTable("site_range")
	}

	if join, ok := sql.Tables["database"]; ok && join {
		q += ` LEFT JOIN "database" ON "site".database_id = "database".id`
	}
	if join, ok := sql.Tables["site_range"]; ok && join {
		q += ` LEFT JOIN "site_range" ON "site_range".site_id = "site".id`
	}
	if join, ok := sql.Tables["site_range__charac"]; ok && join {
		q += ` LEFT JOIN "site_range__charac" ON "site_range__charac".site_range_id = "site_range".id`
	}

	q += " WHERE 1=1"
	for _, groupfilters := range sql.GroupedWheres {
		q += " AND ( 1=0 "
		for _, filter := range groupfilters {
			q += " OR " + filter
		}
		q += ")"
	}

	q += " GROUP BY site.id"

	return q
}

// MapSearchParams is the query filter for searching sites
type MapSearchParams struct {
	Centroid   map[string]bool            `json:"centroid"`
	Knowledge  map[string]bool            `json:"knowledge"`
	Occupation map[string]bool            `json:"occupation"`
	Database   map[string]map[string]bool `json:"database"`
	Chronology map[string]map[string]bool `json:"chronology"`
	Charac     map[string]map[string]bool `json:"charac"`
}

// MapSearch search for sites using many filters
func MapSearch(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	// for measuring execution time
	start := time.Now()

	params := proute.Json.(*MapSearchParams)

	//fmt.Println("params: ", params)

	tx, err := db.DB.Beginx()
	if err != nil {
		log.Println("can't start transaction")
		userSqlError(w, err)
		return
	}

	//q := `SELECT site.id FROM site`

	filters := MapSqlQuery{}
	filters.Init()

	// custom hard coded / mandatory filters
	filters.AddTable("database")
	filters.AddFilter("database_published", `"database".published = true`)

	// add database filter
	for iddb, subfilters := range params.Database {
		if strings.HasPrefix(iddb, "type:") {
			var dbtype = strings.Replace(iddb, "type:", "", 1)
			if dbtype == "inventory" || dbtype == "research" || dbtype == "literary-work" || dbtype == "undefined" {
				for subfilter, yesno := range subfilters {
					if subfilter == "inclorexcl" {
						var compare string
						if yesno {
							compare = "="
						} else {
							compare = "!="
						}
						filters.AddTable("database")
						filters.AddFilter("database", `"database".type" `+compare+` '`+dbtype+`'`)
					}
				}
			}
		} else {
			id, err := strconv.Atoi(iddb)
			if err == nil {
				for subfilter, yesno := range subfilters {
					if subfilter == "inclorexcl" {
						var compare string
						if yesno {
							compare = "="
						} else {
							compare = "!="
						}
						filters.AddFilter("database", `"site".database_id `+compare+` `+strconv.Itoa(id))
					}
				}
			}
		}
	}

	// add centroid filter
	for inclorexcl, yesno := range params.Centroid {
		var compare string
		if yesno {
			compare = "="
		} else {
			compare = "!="
		}

		if inclorexcl == "centroid-include" {
			filters.AddFilter("centroid", `"site".centroid `+compare+` true`)
		} else if inclorexcl == "centroid-exclude" {
			filters.AddFilter("centroid", `"site".centroid `+compare+` false`)
		}
	}

	// add knowledge filter
	for knowledge, yesno := range params.Knowledge {
		var compare string
		if yesno {
			compare = "="
		} else {
			compare = "!="
		}

		switch knowledge {
		case "literature", "surveyed", "dig", "not_documented", "prospected_aerial", "prospected_pedestrian":
			filters.AddTable(`site_range__charac`)
			filters.AddFilter("knowledge", `"site_range__charac".knowledge_type `+compare+` '`+knowledge+`'`)
		}
	}

	// add occupation filter
	for occupation, yesno := range params.Occupation {
		var compare string
		if yesno {
			compare = "="
		} else {
			compare = "!="
		}

		switch occupation {
		case "not_documented", "single", "continuous", "multiple":
			filters.AddFilter("occupation", `"site".occupation `+compare+` '`+occupation+`'`)
		}
	}

	// characs filters
	for characidstr, subfilters := range params.Charac {
		characid, _ := strconv.Atoi(characidstr)
		q := "( 1=1 "
		if yesno, ok := subfilters["inclorexcl"]; ok {
			var compare string
			if yesno {
				compare = "="
			} else {
				compare = "!="
			}
			q += ` AND "site_range__charac".charac_id ` + compare + ` ` + strconv.Itoa(characid)
		}
		if yesno, ok := subfilters["exceptional"]; ok {
			var compare string
			if yesno {
				compare = "="
			} else {
				compare = "!="
			}
			q += ` AND "site_range__charac".exceptional ` + compare + ` true`
		}

		q += ")"

		filters.AddTable("site_range__charac")
		filters.AddFilter("charac", q)
	}

	// chronologies filters
	/*
		for chronocode, subfilters := range params.Chronology {
			chronocode1 := strings.Split(chronocode, "#")
			chronocode2 := strings.Split(chronocode1[1], ":")

			//chronoid, _ := strconv.Atoi(chronocode1[0])
			date_start, _ := strconv.Atoi(chronocode2[0])
			date_end, _ := strconv.Atoi(chronocode2[1])

			if yesno, ok := subfilters["inclorexcl"]; ok {
				q := `"site_range".start_date1 >= ` + strconv.Itoa(date_start) + ` AND "site_range".start_date2 <= ` + strconv.Itoa(date_start)
				q += ` AND "site_range".end_date1 >= ` + strconv.Itoa(date_end) + ` AND "site_range".end_date2 <= ` + strconv.Itoa(date_end)
				if yesno {
				} else {
					q = `NOT (` + q + `)`
				}
				filters.AddTable("site_range")
				filters.AddFilter("chronology", q)
			}
		}
	*/

	// chronologies filters
	for chronocode, subfilters := range params.Chronology {
		chronocode1 := strings.Split(chronocode, "#")
		chronocode2 := strings.Split(chronocode1[1], ":")

		//chronoid, _ := strconv.Atoi(chronocode1[0])
		date_start, _ := strconv.Atoi(chronocode2[0])
		date_end, _ := strconv.Atoi(chronocode2[1])

		if yesno, ok := subfilters["inclorexcl"]; ok {
			var q string
			if yesno {
				q = `"site_range".start_date1 >= ` + strconv.Itoa(date_start) + ` AND "site_range".end_date2 <= ` + strconv.Itoa(date_end)
			} else {
				q = `"site_range".start_date2 < ` + strconv.Itoa(date_start) + ` AND "site_range".end_date1 > ` + strconv.Itoa(date_end)
			}
			filters.AddTable("site_range")
			filters.AddFilter("chronology", q)
		}
	}

	q := filters.BuildQuery()
	fmt.Println("q: ", q)

	site_ids := []int{}
	err = tx.Select(&site_ids, q)
	//fmt.Println("site_ids : ", site_ids)

	elapsed := time.Since(start)
	fmt.Printf("Search took %s", elapsed)

	jsonString := mapGetSitesAsJson(site_ids, tx)
	mapDebug(site_ids, tx)

	err = tx.Commit()
	if err != nil {
		log.Println("can't commit")
		userSqlError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonString))
	return
}

func mapDebug(sites []int, tx *sqlx.Tx) {
	type row struct {
		Id          int
		Start_date1 int
		Start_date2 int
		End_date1   int
		End_date2   int
	}
	rows := []row{}
	err := tx.Select(&rows, "SELECT site.id, sr.start_date1, sr.start_date2, sr.end_date1, sr.end_date2 FROM site LEFT JOIN site_range sr ON sr.site_id = site.id WHERE site.id IN("+model.IntJoin(sites, true)+")")
	if err != nil {
		fmt.Println("err: ", err)
	} else {
		for _, r := range rows {
			fmt.Println("r: ", r)
		}
	}
}

func mapGetSitesAsJson(sites []int, tx *sqlx.Tx) string {

	// for measuring execution time
	start := time.Now()

	var jsonResult []string

	q := `SELECT '{"type": "Feature", ' ||`
	q += `'"geometry": {"type": "Point", "coordinates": [' || (`
	q += `	SELECT ST_X(geom::geometry) || ', ' || ST_Y(geom::geometry) AS coordinates FROM site WHERE id = s.id`
	q += `) || ']}, ' ||`
	q += `'"properties": {"infos": ' || (`
	q += `	SELECT row_to_json(site_infos) || ',' || `
	q += `	'"site_ranges": ' || (`
	q += `		SELECT  array_to_json(array_agg(row_to_json(q_src))) FROM (`
	q += `			SELECT start_date1 as start_date, end_date2 as end_date, `
	q += `			(`
	q += `				SELECT array_to_json(array_agg(row_to_json(q_src2))) FROM (`
	q += `					SELECT src.* FROM site_range__charac src WHERE src.site_range_id IN (SELECT site_range_id FROM site_range__charac WHERE site_range_id = sr.id)`
	q += `				) q_src2`
	q += `			) characs`
	q += `	   	FROM site_range sr WHERE sr.site_id = s.id) q_src`
	q += `	)`
	q += `	 FROM (SELECT si.id, si.code, si.name, si.centroid, si.occupation, d.id AS database_id, d.name as database_name FROM site si LEFT JOIN database d ON si.database_id = d.id WHERE si.id = s.id) site_infos`
	q += `)`
	q += `|| '}}'`
	q += ` FROM site s WHERE s.id IN (` + model.IntJoin(sites, true) + `)`

	err := tx.Select(&jsonResult, q)

	elapsed := time.Since(start)
	fmt.Printf("mapGetSitesAsJson took %s", elapsed)

	if err != nil {
		fmt.Println(err.Error())
	}

	jsonString := `{"type": "FeatureCollection", "features": [` + strings.Join(jsonResult, ",") + `]}`

	return jsonString
}
