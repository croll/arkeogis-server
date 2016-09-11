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
	sqlx_types "github.com/jmoiron/sqlx/types"
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
	Excludes      map[string][]string
}

func (sql *MapSqlQuery) Init() {
	sql.Tables = map[string]bool{}
	sql.GroupedWheres = make(map[string][]string)
	sql.Excludes = make(map[string][]string)
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

func (sql *MapSqlQuery) AddExclude(tablename string, where string) {
	if _, ok := sql.Excludes[tablename]; ok {
		sql.Excludes[tablename] = append(sql.Excludes[tablename], where)
	} else {
		sql.Excludes[tablename] = []string{
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
	if excludes, ok := sql.Excludes["site_range__charac"]; ok && len(excludes) > 0 {
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

	if excludes, ok := sql.Excludes["site_range__charac"]; ok && len(excludes) > 0 {
		q += ` LEFT JOIN "site_range__charac" "x_site_range__charac" ON "x_site_range__charac".site_range_id = "site_range".id AND (1=0`
		for _, exclude := range excludes {
			q += ` OR ` + exclude
		}
		q += ")"
	}
	if excludes, ok := sql.Excludes["site"]; ok && len(excludes) > 0 {
		q += ` LEFT JOIN "site" "x_site" ON "x_site".id = "site".id AND (1=0`
		for _, exclude := range excludes {
			q += ` OR ` + exclude
		}
		q += ")"
	}

	q += " WHERE 1=1"
	for _, groupfilters := range sql.GroupedWheres {
		q += " AND ( 1=0 "
		for _, filter := range groupfilters {
			q += " OR " + filter
		}
		q += ")"
	}

	if excludes, ok := sql.Excludes["site_range__charac"]; ok && len(excludes) > 0 {
		q += " AND x_site_range__charac.id is null"
	}

	if excludes, ok := sql.Excludes["site"]; ok && len(excludes) > 0 {
		q += " AND x_site.id is null"
	}

	q += " GROUP BY site.id"

	return q
}

type MapSearchParamsAreaGeometry struct {
	Geometry sqlx_types.JSONText `json:"geometry"`
}

type MapSearchParamsArea struct {
	Type    string                      `json:"type"`
	Lat     float32                     `json:lat`
	Lng     float32                     `json:'lng'`
	Radius  float32                     `json:'radius'`
	Geojson MapSearchParamsAreaGeometry `json:"geojson"`
}

type MapSearchParamsChronology struct {
	StartDate                int    `json:"start_date"`
	EndDate                  int    `json:"end_date"`
	ExistenceInsideInclude   string `json:"existence_inside_include"`
	ExistenceInsidePart      string `json:"existence_inside_part"`
	ExistenceInsideSureness  string `json:"existence_inside_sureness"`
	ExistenceOutsideInclude  string `json:"existence_outside_include"`
	ExistenceOutsideSureness string `json:"existence_outside_sureness"`
	SelectedChronologyId     int    `json:"selected_chronology_id"`
}

// MapSearchParams is the query filter for searching sites
type MapSearchParams struct {
	Centroid     map[string]bool             `json:"centroid"`
	Knowledge    map[string]bool             `json:"knowledge"`
	Occupation   map[string]bool             `json:"occupation"`
	Database     []int                       `json:"database"`
	Chronologies []MapSearchParamsChronology `json:"chronologies"`
	Characs      map[int]string              `json:"characs"`
	Area         MapSearchParamsArea         `json:"area"`
}

// MapSearch search for sites using many filters
func MapSearch(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	q_args := []interface{}{}

	// for measuring execution time
	start := time.Now()

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
	fmt.Println("database: ", params.Database)
	for _, iddb := range params.Database {
		filters.AddFilter("database", `"site".database_id = `+strconv.Itoa(iddb))
	}

	// if true {
	fmt.Println("geojson.geometry : ", params.Area.Geojson)
	fmt.Println("geojson.geometry : ", params.Area.Geojson.Geometry)
	q_args = append(q_args, params.Area.Geojson.Geometry)
	filters.AddFilter("area", `ST_Contains(ST_SetSRID(ST_GeomFromGeoJSON($`+strconv.Itoa(len(q_args))+`),4326), "site".geom::geometry)`)
	//}

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
	var includes []int
	var excludes []int
	q_exceptional := "1=1"
	for characid, include := range params.Characs {
		if include == "+" {
			includes = append(includes, characid)
		} else if include == "!" {
			q_exceptional += " OR site_range__charac.charac_id=" + strconv.Itoa(characid) + " AND site_range__charac.exceptional=true"
			excludes = append(excludes, characid)
		} else if include == "-" {
			excludes = append(excludes, characid)
		}
	}

	if len(includes) > 0 {
		filters.AddTable("site_range__charac")
		filters.AddFilter("charac", `site_range__charac.charac_id IN (`+model.IntJoin(includes, true)+`)`)
	}

	if q_exceptional != "1=1" {
		filters.AddTable("site_range__charac")
		filters.AddFilter("charac", q_exceptional)
	}

	if len(excludes) > 0 {
		filters.AddExclude("site_range__charac", "x_site_range__charac.charac_id IN ("+model.IntJoin(excludes, true)+")")
	}

	for _, chronology := range params.Chronologies {

		q := "1=1"

		var start_date_str = strconv.Itoa(chronology.StartDate)
		var end_date_str = strconv.Itoa(chronology.EndDate)

		var tblname string
		if chronology.ExistenceInsideInclude == "+" {
			tblname = "site"
		} else if chronology.ExistenceInsideInclude == "-" {
			tblname = "x_site"
		} else {
			log.Println("ExistenceInsideInclude is bad : ", chronology.ExistenceInsideInclude)
			_ = tx.Rollback()
			return
		}

		switch chronology.ExistenceInsideSureness {
		case "potentially":
			q += " AND " + tblname + ".start_date1 <= " + end_date_str + " AND " + tblname + ".end_date2 >= " + start_date_str
			if chronology.ExistenceInsidePart == "full" {
				q += " AND " + tblname + ".start_date1 >= " + start_date_str + " AND " + tblname + ".end_date2 <= " + end_date_str
			}
		case "certainly":
			q += " AND " + tblname + ".start_date2 <= " + end_date_str + " AND " + tblname + ".end_date1 >= " + start_date_str
			if chronology.ExistenceInsidePart == "full" {
				q += " AND " + tblname + ".start_date2 >= " + start_date_str + " AND " + tblname + ".end_date1 <= " + end_date_str
			}
		case "potentially-only":
			q += " AND " + tblname + ".start_date1 <= " + end_date_str + " AND " + tblname + ".end_date2 >= " + start_date_str
			q += " AND " + tblname + ".start_date2 > " + end_date_str + " AND " + tblname + ".end_date1 < " + start_date_str

			if chronology.ExistenceInsidePart == "full" {
				q += " AND " + tblname + ".start_date1 >= " + start_date_str + " AND " + tblname + ".end_date2 <= " + end_date_str
			}
		}

		switch chronology.ExistenceOutsideInclude {
		case "": // it can
			// do nothing
		case "+": // it must
			switch chronology.ExistenceOutsideSureness {
			case "potentially":
				q += " AND (" + tblname + ".start_date2 < " + start_date_str + " OR " + tblname + ".end_date1 >= " + end_date_str + ")"
			case "certainly":
				q += " AND (" + tblname + ".start_date1 < " + start_date_str + " OR " + tblname + ".end_date1 >= " + end_date_str + ")"
			case "potentially-only":
				q += " AND (" + tblname + ".start_date2 < " + start_date_str + " AND " + tblname + ".start_date1 >= " + start_date_str
				q += " OR " + tblname + ".end_date1 > " + end_date_str + " AND " + tblname + ".end_date2 <= " + end_date_str + ")"
			}

		case "-": // it must not
			switch chronology.ExistenceOutsideSureness {
			case "potentially":
				q += " AND NOT (" + tblname + ".start_date2 < " + start_date_str + " OR " + tblname + ".end_date1 >= " + end_date_str + ")"
			case "certainly":
				q += " AND NOT (" + tblname + ".start_date1 < " + start_date_str + " OR " + tblname + ".end_date1 >= " + end_date_str + ")"
			case "potentially-only":
				q += " AND NOT (" + tblname + ".start_date2 < " + start_date_str + " AND " + tblname + ".start_date1 >= " + start_date_str
				q += " OR " + tblname + ".end_date1 > " + end_date_str + " AND " + tblname + ".end_date2 <= " + end_date_str + ")"
			}
		}

		if q != "1=1" {
			if chronology.ExistenceInsideInclude == "+" {
				filters.AddFilter("chronology", q)
			} else if chronology.ExistenceInsideInclude == "-" {
				filters.AddExclude("site", q)
			}
		}
	}

	q := filters.BuildQuery()
	fmt.Println("q: ", q)

	site_ids := []int{}
	err = tx.Select(&site_ids, q, q_args...)
	if err != nil {
		fmt.Println("query failed : ", err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}
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
	} /*else {
		for _, r := range rows {
			fmt.Println("r: ", r)
		}
	}*/
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
