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

	db "github.com/croll/arkeogis-server/db"
	//model "github.com/croll/arkeogis-server/model"
	"net/http"

	routes "github.com/croll/arkeogis-server/webserver/routes"
	"github.com/croll/arkeogis-server/webserver/session"
)

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:   "/api/cities",
			Func:   CityCreate,
			Method: "POST",
		},
		&routes.Route{
			Path:   "/api/cities",
			Func:   CityList,
			Method: "GET",
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

func CityList(w http.ResponseWriter, r *http.Request, o interface{}, s *session.Session) {

	err := r.ParseForm()
	if err != nil {
		fmt.Println("ParseForm err: ", err)
		return
	}

	type res struct {
		Geonameid int    `json:"value"`
		Name      string `json:"display"`
	}

	city := []res{}

	err = db.DB.Select(&city, "SELECT geonameid,name FROM city JOIN city_translation ON city_translation.city_geonameid = city.geonameid LEFT JOIN lang ON city_translation.lang_id = lang.id WHERE (name_ascii ILIKE $1 OR name ILIKE $1) AND country_geonameid = $2 AND (lang.iso_code = $3 OR lang.iso_code = 'D')", r.FormValue("search")+"%", r.FormValue("id_country"), r.FormValue("lang"))
	if err != nil {
		fmt.Println("err: ", err)
		return
	}

	//fmt.Println("c: ", city)
	j, err := json.Marshal(city)
	w.Write(j)
}

func CityCreate(w http.ResponseWriter, r *http.Request, o interface{}, s *session.Session) {
	fmt.Println("request :", r)
}

func CityUpdate(w http.ResponseWriter, r *http.Request, o interface{}, s *session.Session) {
	//params := mux.Vars(r)
	//uid := params["id"]
	//email := r.FormValue("email")
}

func CityDelete(w http.ResponseWriter, r *http.Request, o interface{}, s *session.Session) {
}

func CityInfos(w http.ResponseWriter, r *http.Request, o interface{}, s *session.Session) {
	w.Header().Set("Allow", "DELETE,GET,HEAD,OPTIONS,POST,PUT")
}
