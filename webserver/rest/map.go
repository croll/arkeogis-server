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
			Permissions: []string{
			//"request map",
			},
		},
	}
	routes.RegisterMultiple(Routes)
}

type MapSqlJoin struct {
	JoinLeftTable  string
	JoinLeftKey    string
	JoinRightTable string
	JoinRightKey   string
}

type MapSqlTableDef struct {
	TableName string
	Joins     []MapSqlJoin
}

var MapSqlDefSite = MapSqlTableDef{
	TableName: "site",
	Joins:     []MapSqlJoin{},
}

var MapSqlDefXSite = MapSqlTableDef{
	TableName: "site",
	Joins: []MapSqlJoin{
		{
			JoinLeftTable:  "site",
			JoinLeftKey:    "id",
			JoinRightTable: "site",
			JoinRightKey:   "id",
		},
	},
}

var MapSqlDefDatabase = MapSqlTableDef{
	TableName: "database",
	Joins: []MapSqlJoin{
		{
			JoinLeftTable:  "site",
			JoinLeftKey:    "database_id",
			JoinRightTable: "database",
			JoinRightKey:   "id",
		},
	},
}

var MapSqlDefSiteRange = MapSqlTableDef{
	TableName: "site_range",
	Joins: []MapSqlJoin{
		{
			JoinLeftTable:  "site",
			JoinLeftKey:    "id",
			JoinRightTable: "site_range",
			JoinRightKey:   "site_id",
		},
	},
}

var MapSqlDefSiteRangeCharac = MapSqlTableDef{
	TableName: "site_range__charac",
	Joins: []MapSqlJoin{
		{
			JoinLeftTable:  "site_range",
			JoinLeftKey:    "id",
			JoinRightTable: "site_range__charac",
			JoinRightKey:   "site_range_id",
		},
	},
}

var MapSqlDefSiteRangeCharacTr = MapSqlTableDef{
	TableName: "site_range__charac_tr",
	Joins: []MapSqlJoin{
		{
			JoinLeftTable:  "site_range__charac",
			JoinLeftKey:    "id",
			JoinRightTable: "site_range__charac_tr",
			JoinRightKey:   "site_range__charac_id",
		},
		{
			JoinLeftTable:  "database",
			JoinLeftKey:    "default_language",
			JoinRightTable: "site_range__charac_tr",
			JoinRightKey:   "lang_isocode",
		},
	},
}

type MapSqlQueryTable struct {
	TableDef       *MapSqlTableDef
	As             string
	UsedForExclude bool
}

type MapSqlQueryWhere struct {
	Table *MapSqlQueryTable
	Where string
	Args  []interface{}
}

type MapSqlQuery struct {
	Tables []*MapSqlQueryTable
	Wheres []*MapSqlQueryWhere
}

func (sql *MapSqlQuery) Init() {
	sql.Tables = make([]*MapSqlQueryTable, 0)
	sql.Wheres = make([]*MapSqlQueryWhere, 0)
}

func (sql *MapSqlQuery) AddTable(tabledef *MapSqlTableDef, as string, usedforexclude bool) {
	for _, t := range sql.Tables {
		if t.TableDef == tabledef && t.As == as && t.UsedForExclude == usedforexclude {
			return // don't add any table, we already have one
		}
	}

	t := MapSqlQueryTable{
		TableDef:       tabledef,
		As:             as,
		UsedForExclude: usedforexclude,
	}

	sql.Tables = append(sql.Tables, &t)
}

func (sql *MapSqlQuery) FindTable(tableas string, trymebefore *MapSqlQueryTable) (table *MapSqlQueryTable, ok bool) {
	if trymebefore != nil && (trymebefore.As == tableas || trymebefore.TableDef.TableName == tableas) {
		return trymebefore, true
	}
	for _, t := range sql.Tables {
		if t.As == tableas {
			return t, true
		}
	}
	return nil, false
}

func (sql *MapSqlQuery) AddFilter(tableas string, where string, args ...interface{}) {
	if table, ok := sql.FindTable(tableas, nil); ok {
		sql.Wheres = append(sql.Wheres, &MapSqlQueryWhere{
			Table: table,
			Where: where,
			Args:  args,
		})
	} else {
		fmt.Println("BAD: add filter on an unknow tableas : ", tableas, where)
	}
}

func (sql *MapSqlQuery) BuildQuery() (string, []interface{}) {
	q := ""
	joins_str := ""
	joins_args := []interface{}{}
	where_str := "1=1"
	where_args := []interface{}{}

	for _, table := range sql.Tables {
		if len(table.TableDef.Joins) > 0 {
			joins_str += ` LEFT JOIN "` + table.TableDef.TableName + `" AS "` + table.As + `"`
		} else { // no join mean first table
			q += `SELECT "` + table.As + `"."id" FROM "` + table.TableDef.TableName + `" AS "` + table.As + `" `
		}
		for i, join := range table.TableDef.Joins {
			lefttable, _ := sql.FindTable(join.JoinLeftTable, nil)
			righttable, _ := sql.FindTable(join.JoinRightTable, table)
			if righttable == nil {
				fmt.Println("BAD: right table not found : ", join.JoinRightTable)
			}
			if lefttable != nil {
				if i == 0 {
					joins_str += ` ON "`
				} else {
					joins_str += ` AND "`
				}
				joins_str += lefttable.As + `"."` + join.JoinLeftKey + `" = "` + righttable.As + `"."` + join.JoinRightKey + `"`
			} else {
				fmt.Println("BAD: left table not found : ", join.JoinLeftTable)
			}
		}

		if table.UsedForExclude == false {
			for _, where := range sql.Wheres {
				if where.Table == table {
					where_str += ` AND (` + where.Where + `)`
					where_args = append(where_args, where.Args...)
				}
			}
		} else {
			for _, where := range sql.Wheres {
				if where.Table == table {
					joins_str += ` AND (` + where.Where + `)`
					joins_args = append(joins_args, where.Args...)
				}
			}
			where_str += ` AND "` + table.As + `".id IS NULL`
		}
	}

	q += " " + joins_str + " WHERE " + where_str

	// query end
	q += " GROUP BY site.id"

	// replace $$
	q_copy := ""
	for i := 1; i < 999; i++ {
		q_copy = strings.Replace(q, "$$", "$"+strconv.Itoa(i), 1)
		if q_copy == q {
			break
		}
		q = q_copy
	}

	return q, append(joins_args, where_args...)
}

type MapSearchParamsOthers struct {
	Centroid      string   `json:"centroid"`
	CharacsLinked string   `json:"characs_linked"`
	Knowledges    []string `json:"knowledges"`
	Occupation    []string `json:"occupation"`
	TextSearch    string   `json:"text_search"`
	TextSearchIn  []string `json:"text_search_in"`
}

type MapSearchParamsAreaGeometry struct {
	Geometry sqlx_types.JSONText `json:"geometry"`
}

type MapSearchParamsArea struct {
	Type    string                      `json:"type"`
	Lat     float32                     `json:"lat"`
	Lng     float32                     `json:"lng"`
	Radius  float32                     `json:"radius"`
	Geojson MapSearchParamsAreaGeometry `json:"geojson"`
}

type MapSearchParamsCharac struct {
	Include     bool `json:"include"`
	Exceptional bool `json:"exceptional"`
	RootId      int  `json:"root_id"`
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
	Knowledge    map[string]bool               `json:"knowledge"`
	Occupation   map[string]bool               `json:"occupation"`
	Database     []int                         `json:"database"`
	Chronologies []MapSearchParamsChronology   `json:"chronologies"`
	Characs      map[int]MapSearchParamsCharac `json:"characs"`
	Others       MapSearchParamsOthers         `json:"others"`
	Area         MapSearchParamsArea           `json:"area"`
}

// MapSearch search for sites using many filters
func MapSearch(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
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
	filters.AddTable(&MapSqlDefSite, "site", false)
	filters.AddTable(&MapSqlDefDatabase, "database", false)
	filters.AddFilter("database", `"database".published = true`)

	// add database filter
	filters.AddFilter("database", `"site".database_id IN (`+model.IntJoin(params.Database, true)+`)`)

	// Area filter
	fmt.Println(params.Area.Type, params.Area.Lat, params.Area.Lng, params.Area.Radius)
	if params.Area.Type == "disc" || params.Area.Type == "custom" {
		filters.AddFilter("site", `ST_DWithin("site".geom, Geography(ST_MakePoint($$, $$)), $$)`,
			params.Area.Lng, params.Area.Lat, params.Area.Radius)
	} else {
		filters.AddFilter("site", `ST_Within("site".geom::geometry, ST_SetSRID(ST_GeomFromGeoJSON($$),4326))`,
			params.Area.Geojson.Geometry)
	}

	// add centroid filter
	switch params.Others.Centroid {
	case "with":
		filters.AddFilter("site", `"site".centroid = true`)
	case "without":
		filters.AddFilter("site", `"site".centroid = false`)
	case "":
		// do not filter
	}

	// add knowledge filter
	str := ""
	for _, knowledge := range params.Others.Knowledges {
		switch knowledge {
		case "literature", "surveyed", "dig", "not_documented", "prospected_aerial", "prospected_pedestrian":
			if str == "" {
				str += "'" + knowledge + "'"
			} else {
				str += ",'" + knowledge + "'"
			}
		}
	}
	if str != "" {
		filters.AddTable(&MapSqlDefSiteRange, `site_range`, false)
		filters.AddTable(&MapSqlDefSiteRangeCharac, `site_range__charac`, false)
		filters.AddFilter("site_range__charac", `"site_range__charac".knowledge_type IN (`+str+`)`)
	}

	// add occupation filter
	str = ""
	for _, occupation := range params.Others.Knowledges {
		switch occupation {
		case "not_documented", "single", "continuous", "multiple":
			if str == "" {
				str += "'" + occupation + "'"
			} else {
				str += ",'" + occupation + "'"
			}
		}
	}
	if str != "" {
		filters.AddFilter("site", `"site".occupation IN (`+str+`)`)
	}

	// text filter
	if params.Others.TextSearch != "" {
		str = "1=0"
		args := []interface{}{}
		for _, textSearchIn := range params.Others.TextSearchIn {
			switch textSearchIn {
			case "site_name":
				args = append(args, "%"+params.Others.TextSearch+"%")
				str += ` OR "site".name ILIKE $$`
			case "city_name":
				args = append(args, "%"+params.Others.TextSearch+"%")
				str += ` OR "site".city_name ILIKE $$`
			case "bibliography":
				args = append(args, "%"+params.Others.TextSearch+"%")
				filters.AddTable(&MapSqlDefSiteRange, `site_range`, false)
				filters.AddTable(&MapSqlDefSiteRangeCharac, `site_range__charac`, false)
				filters.AddTable(&MapSqlDefSiteRangeCharacTr, `site_range__charac_tr`, false)
				str += ` OR "site_range__charac_tr".bibliography ILIKE $$`
			case "comment":
				args = append(args, "%"+params.Others.TextSearch+"%")
				filters.AddTable(&MapSqlDefSiteRange, `site_range`, false)
				filters.AddTable(&MapSqlDefSiteRangeCharac, `site_range__charac`, false)
				filters.AddTable(&MapSqlDefSiteRangeCharacTr, `site_range__charac_tr`, false)
				str += ` OR "site_range__charac_tr".comment ILIKE $$`
			}
		}
		if str != "1=0" {
			filters.AddFilter("site", str, args...)
		}
	}

	// characs filters
	includes := make(map[int][]int, 0)
	excludes := make(map[int][]int, 0)
	exceptionals := make(map[int][]int, 0)

	for characid, sel := range params.Characs {
		if sel.Include && !sel.Exceptional {
			if _, ok := includes[sel.RootId]; !ok {
				includes[sel.RootId] = make([]int, 0)
			}
			includes[sel.RootId] = append(includes[sel.RootId], characid)
		} else if sel.Include && sel.Exceptional {
			if _, ok := exceptionals[sel.RootId]; !ok {
				exceptionals[sel.RootId] = make([]int, 0)
			}
			exceptionals[sel.RootId] = append(exceptionals[sel.RootId], characid)
		} else if !sel.Include {
			if _, ok := excludes[sel.RootId]; !ok {
				excludes[sel.RootId] = make([]int, 0)
			}
			excludes[sel.RootId] = append(excludes[sel.RootId], characid)
		}
	}

	if params.Others.CharacsLinked == "all" {
		for rootid, characids := range includes {
			tableas := "site_range__charac_" + strconv.Itoa(rootid)
			filters.AddTable(&MapSqlDefSiteRange, "site_range", false)
			filters.AddTable(&MapSqlDefSiteRangeCharac, tableas, false)
			filters.AddFilter(tableas, tableas+`.charac_id IN (`+model.IntJoin(characids, true)+`)`)
		}

		for rootid, characids := range exceptionals {
			tableas := "site_range__charac_" + strconv.Itoa(rootid)
			filters.AddTable(&MapSqlDefSiteRange, "site_range", false)
			filters.AddTable(&MapSqlDefSiteRangeCharac, tableas, false)

			q := "1=0"
			for _, characid := range characids {
				q += " OR " + tableas + ".charac_id = " + strconv.Itoa(characid) + " AND " + tableas + ".exceptional = true"
			}

			filters.AddFilter(tableas, q)
		}

		for rootid, characids := range excludes {
			tableas := "x_site_range__charac_" + strconv.Itoa(rootid)
			filters.AddTable(&MapSqlDefSiteRange, "site_range", false)
			filters.AddTable(&MapSqlDefSiteRangeCharac, tableas, true)
			filters.AddFilter(tableas, tableas+".charac_id IN ("+model.IntJoin(characids, true)+")")
		}

	} else if params.Others.CharacsLinked == "at-least-one" { // default
		s_includes := []int{}
		s_excludes := []int{}
		s_exceptionals := []int{}

		for _, characids := range includes {
			s_includes = append(s_includes, characids...)
		}
		for _, characids := range excludes {
			s_excludes = append(s_excludes, characids...)
		}
		for _, characids := range exceptionals {
			s_exceptionals = append(s_exceptionals, characids...)
		}

		if len(s_includes) > 0 {
			filters.AddTable(&MapSqlDefSiteRange, "site_range", false)
			filters.AddTable(&MapSqlDefSiteRangeCharac, "site_range__charac", false)
			filters.AddFilter("site_range__charac", `site_range__charac.charac_id IN (`+model.IntJoin(s_includes, true)+`)`)
		}

		if len(s_excludes) > 0 {
			filters.AddTable(&MapSqlDefSiteRange, "site_range", false)
			filters.AddTable(&MapSqlDefSiteRangeCharac, "x_site_range__charac", true)
			filters.AddFilter("x_site_range__charac", `x_site_range__charac.charac_id IN (`+model.IntJoin(s_includes, true)+`)`)
		}

		if len(s_exceptionals) > 0 {
			filters.AddTable(&MapSqlDefSiteRange, "site_range", false)
			filters.AddTable(&MapSqlDefSiteRangeCharac, "site_range__charac", false)
			q := "1=0"
			for _, characid := range s_exceptionals {
				q += " OR site_range__charac.charac_id = " + strconv.Itoa(characid) + " AND site_range__charac.exceptional = true"
			}
			filters.AddFilter("site_range__charac", q)
		}
	}

	/*
		if len(excludes) > 0 {
			filters.AddExclude("site_range__charac", "x_site_range__charac.charac_id IN ("+model.IntJoin(excludes, true)+")")
		}
	*/
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
				filters.AddFilter("site", q)
			} else if chronology.ExistenceInsideInclude == "-" {
				filters.AddTable(&MapSqlDefXSite, "x_site", true)
				filters.AddFilter("x_site", q)
			}
		}
	}

	q, q_args := filters.BuildQuery()
	fmt.Println("q: ", q, q_args)

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
	q += `			SELECT start_date1, start_date2, end_date1, end_date2, `
	q += `			(`
	q += `				SELECT array_to_json(array_agg(row_to_json(q_src2))) FROM (`
	q += `					SELECT src.* FROM site_range__charac src WHERE src.site_range_id IN (SELECT site_range_id FROM site_range__charac WHERE site_range_id = sr.id)`
	q += `				) q_src2`
	q += `			) characs`
	q += `	   	FROM site_range sr WHERE sr.site_id = s.id) q_src`
	q += `	)`
	q += `	 FROM (SELECT si.id, si.code, si.name, si.centroid, si.occupation, si.start_date1, si.start_date2, si.end_date1, si.end_date2, d.id AS database_id, d.name as database_name FROM site si LEFT JOIN database d ON si.database_id = d.id WHERE si.id = s.id) site_infos`
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
