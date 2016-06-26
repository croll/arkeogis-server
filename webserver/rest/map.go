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

	params := proute.Json.(*MapSearchParams)

	fmt.Println("params: ", params)

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
			if err != nil {
				for subfilter, yesno := range subfilters {
					if subfilter == "inclorexcl" {
						var compare string
						if yesno {
							compare = "="
						} else {
							compare = "!="
						}
						filters.AddFilter("database", `"site".id `+compare+` `+strconv.Itoa(id))
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
			filters.AddFilter("centroid", `"site".centroid `+compare+` 't'`)
		} else if inclorexcl == "centroid-exclude" {
			filters.AddFilter("centroid", `"site".centroid `+compare+` 'f'`)
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

	q := filters.BuildQuery()
	fmt.Println("q: ", q)

	site_ids := []int{}
	err = tx.Select(&site_ids, q)
	fmt.Println("site_ids : ", site_ids)

	jsonString := mapGetSitesAsJson(site_ids, tx)

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

func mapGetSitesAsJson(sites []int, tx *sqlx.Tx) string {
	var jsonResult []string

	q := `SELECT '{"type": "Feature", ' ||`
	q += `'"geometry": {"type": "Point", "coordinates": [' || (`
	q += `	SELECT ST_X(geom::geometry) || ', ' || ST_Y(geom::geometry) AS coordinates FROM site WHERE id = s.id`
	q += `) || ']}, ' ||`
	q += `'"properties": {"infos": ' || (`
	q += `	SELECT row_to_json(site_infos) || ',' || `
	q += `	'"site_ranges": ' || (`
	q += `		SELECT  array_to_json(array_agg(row_to_json(q_src))) FROM (`
	q += `			SELECT *,`
	q += `			(`
	q += `				SELECT array_to_json(array_agg(row_to_json(q_src2))) FROM (`
	q += `					SELECT src.*, srctr.comment, srctr.bibliography FROM site_range__charac src LEFT JOIN site_range__charac_tr srctr ON src.id = srctr.site_range__charac_id WHERE src.site_range_id IN (SELECT site_range_id FROM site_range__charac WHERE site_range_id = sr.id)`
	q += `				) q_src2`
	q += `			) characs`
	q += `	   	FROM site_range sr WHERE sr.site_id = s.id) q_src`
	q += `	)`
	q += `	 FROM (SELECT code, name, city_name, city_geonameid, centroid, occupation, created_at, updated_at FROM site WHERE id = s.id) site_infos`
	q += `)`
	q += `|| '}}'`
	q += ` FROM site s WHERE s.id IN (` + model.IntJoin(sites, true) + `)`

	err := tx.Select(&jsonResult, q)

	if err != nil {
		fmt.Println(err.Error())
	}

	jsonString := `{"type": "FeatureCollection", "features": [` + strings.Join(jsonResult, ",") + `]}`

	return jsonString
}
