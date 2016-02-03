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
	"encoding/json"
	"fmt"
	"reflect"

	db "github.com/croll/arkeogis-server/db"
	//model "github.com/croll/arkeogis-server/model"
	"net/http"

	routes "github.com/croll/arkeogis-server/webserver/routes"
)

type CompanyListParams struct {
	Search string
}

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:   "/api/companies",
			Func:   CompanyCreate,
			Method: "POST",
		},
		&routes.Route{
			Path:   "/api/companies",
			Func:   CompanyList,
			Params: reflect.TypeOf(CompanyListParams{}),
			Method: "GET",
		},
		&routes.Route{
			Path:   "/api/companies",
			Func:   CompanyUpdate,
			Method: "PUT",
		},
		&routes.Route{
			Path:   "/api/companies",
			Func:   CompanyDelete,
			Method: "DELETE",
		},
	}
	routes.RegisterMultiple(Routes)
}

func CompanyList(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	companies := []Company{}

	params := proute.Params.(*CompanyListParams)

	err := db.DB.Select(&companies, "SELECT * FROM company WHERE name ILIKE $1", params.Search+"%")
	if err != nil {
		fmt.Println("err: ", err)
		return
	}

	j, err := json.Marshal(companies)
	w.Write(j)
}

func CompanyCreate(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	fmt.Println("request :", r)
}

func CompanyUpdate(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	//params := mux.Vars(r)
	//uid := params["id"]
	//email := r.FormValue("email")
}

func CompanyDelete(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
}

func CompanyInfos(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	w.Header().Set("Allow", "DELETE,GET,HEAD,OPTIONS,POST,PUT")
}
