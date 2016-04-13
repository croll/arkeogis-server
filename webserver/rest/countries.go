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
	"github.com/croll/arkeogis-server/model"
	//model "github.com/croll/arkeogis-server/model"
	routes "github.com/croll/arkeogis-server/webserver/routes"
	//"github.com/gorilla/mux"
	"net/http"
)

type CountryListParams struct {
	Search string
}

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:   "/api/countries",
			Func:   CountryCreate,
			Method: "POST",
		},
		&routes.Route{
			Path:        "/api/countries",
			Description: "Search for countries available on our world, using a search string",
			Func:        CountryList,
			Params:      reflect.TypeOf(CountryListParams{}),
			Method:      "GET",
		},
		&routes.Route{
			Path:   "/api/countries",
			Func:   CountryUpdate,
			Method: "PUT",
		},
		&routes.Route{
			Path:   "/api/countries",
			Func:   CountryDelete,
			Method: "DELETE",
		},
	}
	routes.RegisterMultiple(Routes)
}

func CountryList(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Params.(*CountryListParams)

	type row struct {
		model.Country
		model.Country_tr
	}

	countries := []row{}

	err := db.DB.Select(&countries, "SELECT country.*, country_tr.* FROM \"country\" JOIN country_tr ON country_tr.country_geonameid = country.geonameid LEFT JOIN lang ON country_tr.lang_id = lang.id WHERE (lang.iso_code = $1 OR lang.iso_code = 'D') AND (name_ascii ILIKE $2 OR name ILIKE $2)", proute.Lang1.Id, params.Search+"%")
	if err != nil {
		fmt.Println("err: ", err)
		return
	}

	j, err := json.Marshal(countries)
	w.Write(j)
}

func CountryCreate(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	fmt.Println("request :", r)
}

func CountryUpdate(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	//params := mux.Vars(r)
	//uid := params["id"]
	//email := r.FormValue("email")
}

func CountryDelete(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
}

func CountryInfos(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	w.Header().Set("Allow", "DELETE,GET,HEAD,OPTIONS,POST,PUT")
}
