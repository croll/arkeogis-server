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
	"reflect"

	"net/http"

	db "github.com/croll/arkeogis-server/db"
	"github.com/croll/arkeogis-server/model"

	routes "github.com/croll/arkeogis-server/webserver/routes"
)

type CityListParams struct {
	Id_country int    `default:"0" min:"0"`
	Search     string `default:"" regexp:"^[^%]*$"`
}

type CityGetParams struct {
	Id_city int `default:"0" min:"0"`
}

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:   "/api/cities",
			Func:   CityCreate,
			Method: "POST",
		},
		&routes.Route{
			Path:        "/api/cities",
			Func:        CityList,
			Description: "Search for cities available in a country, using a search string",
			Params:      reflect.TypeOf(CityListParams{}),
			Method:      "GET",
		},
		&routes.Route{
			Path:        "/api/cities/{id_city:[0-9]+}",
			Func:        CityGet,
			Description: "Get a city, using a city id",
			Params:      reflect.TypeOf(CityGetParams{}),
			Method:      "GET",
		},
		&routes.Route{
			Path:   "/api/cities",
			Func:   CityUpdate,
			Method: "PUT",
		},
		&routes.Route{
			Path:   "/api/cities",
			Func:   CityDelete,
			Method: "DELETE",
		},
	}
	routes.RegisterMultiple(Routes)
}

func CityList(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Params.(*CityListParams)

	type row struct {
		model.City
		model.City_tr
	}

	cities := []row{}

	err := db.DB.Select(&cities, "SELECT geonameid, country_geonameid, geom, geom_centroid, lang_isocode, name, name_ascii FROM city JOIN city_tr ON city_tr.city_geonameid = city.geonameid LEFT JOIN lang ON city_tr.lang_isocode = lang.isocode WHERE (name_ascii LIKE lower(f_unaccent($1)) OR lower(f_unaccent(name)) LIKE lower(f_unaccent($1))) AND country_geonameid = $2 AND (lang.iso_code = $3 OR lang.iso_code = 'D')", params.Search+"%", params.Id_country, proute.Lang1.Id)
	if err != nil {
		fmt.Println("err: ", err)
		return
	}

	j, err := json.Marshal(cities)
	w.Write(j)
}

func CityGet(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Params.(*CityGetParams)

	res := model.CityAndCountry_wtr{}

	tx, err := db.DB.Beginx()
	if err != nil {
		log.Println("err on begin", err)
		routes.ServerError(w, 500, "INTERNAL ERROR")
		return
	}

	err = res.Get(tx, params.Id_city, proute.Lang1.Id)
	if err != nil {
		log.Println("err while getting city and country: ", err)
		routes.ServerError(w, 500, "INTERNAL ERROR")
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println("err on commit", err)
		routes.ServerError(w, 500, "INTERNAL ERROR")
		return
	}

	j, err := json.Marshal(res)
	w.Write(j)
}

func CityCreate(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	fmt.Println("request :", r)
}

func CityUpdate(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	//params := mux.Vars(r)
	//uid := params["id"]
	//email := r.FormValue("email")
}

func CityDelete(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
}

func CityInfos(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	w.Header().Set("Allow", "DELETE,GET,HEAD,OPTIONS,POST,PUT")
}
