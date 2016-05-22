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

	db "github.com/croll/arkeogis-server/db"
	"github.com/croll/arkeogis-server/model"
	routes "github.com/croll/arkeogis-server/webserver/routes"
)

// DatabaseGetParams are params received by REST query
type DatabaseGetParams struct {
	Id int `min:"0" error:"Database Id is mandatory"`
}

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:        "/api/database",
			Description: "Get list of all databases in arkeogis",
			Func:        DatabasesList,
			Method:      "GET",
		},
		&routes.Route{
			Path:        "/api/database/{id:[0-9]+}",
			Description: "Get infos on an arkeogis database",
			Func:        DatabaseInfos,
			Method:      "GET",
			Permissions: []string{},
			Params:      reflect.TypeOf(DatabaseGetParams{}),
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

// DatabasesList returns the list of databases
func DatabasesList(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	databases := []model.Database{}
	err := db.DB.Select(&databases, "SELECT * FROM \"database\"")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	db.DB.Select(&enums.Context, "SELECT unnest(enum_range(NULL::database_context))")
	db.DB.Select(&enums.Context, "SELECT unnest(enum_range(NULL::database_context))")
	fmt.Println(enums)
}

// DatabaseInfos return detailed infos on an database
func DatabaseInfos(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	params := proute.Params.(*DatabaseGetParams)
	tx, err := db.DB.Beginx()
	if err != nil {
		log.Println("can't start transaction")
		userSqlError(w, err)
		return
	}
	d := model.DatabaseFullInfos{}
	d.Id = params.Id

	dbInfos, err := d.GetFullInfosAsJSON(tx, proute.Lang1.Id)

	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(dbInfos))
}
