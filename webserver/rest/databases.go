/* ArkeoGIS - The Geographic Information System for Archaeologists
 * Copyright (C) 2015-2016 CROLL SAS
 *
 * Authors :
 *  Christophe Beveraggi <beve@croll.fr>
 *  Nicolas Dimitrijevic <nicolas@croll.fr>
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
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"

	db "github.com/croll/arkeogis-server/db"
	"github.com/croll/arkeogis-server/model"
	routes "github.com/croll/arkeogis-server/webserver/routes"
)

// DatabaseInfosParams are params received by REST query
type DatabaseInfosParams struct {
	Id int `min:"0" error:"Database Id is mandatory"`
}

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:        "/api/database",
			Description: "Get list of all databases in arkeogis",
			Func:        DatabasesList,
			Method:      "GET",
			Permissions: []string{},
			Params:      reflect.TypeOf(DatabaseListParams{}),
		},
		&routes.Route{
			Path:        "/api/database/{id:[0-9]+}",
			Description: "Get infos on an arkeogis database",
			Func:        DatabaseInfos,
			Method:      "GET",
			Permissions: []string{},
			Params:      reflect.TypeOf(DatabaseInfosParams{}),
		},
		&routes.Route{
			Path:        "/api/licences",
			Description: "Get list of licenses",
			Func:        LicensesList,
			Method:      "GET",
			Permissions: []string{
			//"AdminUsers",
			},
		},
	}
	routes.RegisterMultiple(Routes)
}

type DatabaseListParams struct {
	Bounding_box string
	Start_date   int  `json:"start_date"`
	End_date     int  `json:"end_date"`
	Check_dates  bool `json:"check_dates"`
}

// DatabasesList returns the list of databases
func DatabasesList(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Params.(*DatabaseListParams)

	tx, err := db.DB.Beginx()
	if err != nil {
		userSqlError(w, err)
		return
	}

	q := "SELECT d.*, ST_AsGeoJSON(d.geographical_extent_geom) as geographical_extent_geom, u.firstname || ' ' || u.lastname as author FROM \"database\" d LEFT JOIN \"user\" u ON d.owner = u.id WHERE d.id > 0"

	type dbInfos struct {
		model.Database
		Description         map[string]string
		Geographical_limit  map[string]string
		Bibliography        map[string]string
		Context_description map[string]string
		Source_description  map[string]string
		Source_relation     map[string]string
		Copyright           map[string]string
		Subject             map[string]string
		Author              string
	}

	if params.Bounding_box != "" {
		q += " AND ST_Contains(ST_GeomFromGeoJSON(:bounding_box), geographical_extent_geom::::geometry)"
	}

	if params.Check_dates {
		q += " AND d.start_date > :start_date AND d.end_date < :end_date"
	}

	databases := []dbInfos{}

	nstmt, err := db.DB.PrepareNamed(q)
	if err != nil {
		fmt.Println(err)
		userSqlError(w, err)
		return
	}
	err = nstmt.Select(&databases, params)

	if err != nil {
		fmt.Println(err)
		userSqlError(w, err)
		return
	}

	for _, database := range databases {
		tr := []model.Database_tr{}
		err = tx.Select(&tr, "SELECT * FROM database_tr WHERE database_id = "+strconv.Itoa(database.Id))
		if err != nil {
			fmt.Println(err)
			userSqlError(w, err)
			_ = tx.Rollback()
			return
		}
		database.Description = model.MapSqlTranslations(tr, "Lang_isocode", "Description")
		database.Geographical_limit = model.MapSqlTranslations(tr, "Lang_isocode", "Geographical_limit")
		database.Bibliography = model.MapSqlTranslations(tr, "Lang_isocode", "Bibliography")
		database.Context_description = model.MapSqlTranslations(tr, "Lang_isocode", "Context_description")
		database.Source_description = model.MapSqlTranslations(tr, "Lang_isocode", "Source_description")
		database.Source_relation = model.MapSqlTranslations(tr, "Lang_isocode", "Source_relation")
		database.Copyright = model.MapSqlTranslations(tr, "Lang_isocode", "Copyright")
		database.Subject = model.MapSqlTranslations(tr, "Lang_isocode", "Subject")
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println(err)
		userSqlError(w, err)
		return
	}

	l, _ := json.Marshal(databases)
	w.Header().Set("Content-Type", "application/json")
	w.Write(l)
}

// LicensesList returns the list of licenses which can be assigned to databases
func LicensesList(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	databases := []model.License{}
	err := db.DB.Select(&databases, "SELECT * FROM \"license\"")
	if err != nil {
		routes.ServerError(w, 500, "INTERNAL ERROR")
		return
	}
	l, _ := json.Marshal(databases)
	w.Header().Set("Content-Type", "application/json")
	w.Write(l)
}

// DatabaseEnumList returns the list of enums fields
// We have to link them with a translation manually clientside
func DatabaseEnumList(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	enums := struct {
		ScaleResolution    []string
		GeographicalExtent []string
		Type               []string
		State              []string
		Context            []string
		Occupation         []string
		KnowledgeType      []string
	}{}
	db.DB.Select(&enums.ScaleResolution, "SELECT unnest(enum_range(NULL::database_scale_resolution))")
	db.DB.Select(&enums.GeographicalExtent, "SELECT unnest(enum_range(NULL::database_geographical_extent))")
	db.DB.Select(&enums.Type, "SELECT unnest(enum_range(NULL::database_type))")
	db.DB.Select(&enums.State, "SELECT unnest(enum_range(NULL::database_state))")
	db.DB.Select(&enums.Context, "SELECT unnest(enum_range(NULL::database_context))")
}

// DatabaseInfos return detailed infos on an database
func DatabaseInfos(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	params := proute.Params.(*DatabaseInfosParams)
	tx, err := db.DB.Beginx()
	if err != nil {
		log.Println("can't start transaction")
		userSqlError(w, err)
		return
	}
	d := model.DatabaseFullInfos{}
	d.Id = params.Id

	dbInfos, err := d.GetFullInfosAsJSON(tx, proute.Lang1.Isocode)

	if err != nil {
		log.Println("Error getting database infos", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(dbInfos))
}
