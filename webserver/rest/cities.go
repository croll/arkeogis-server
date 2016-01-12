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
	"fmt"

	db "github.com/croll/arkeogis-server/db"
	"github.com/emicklei/go-restful"

	//model "github.com/croll/arkeogis-server/model"
)

func init() {
	register(Cities{})
}

type ResCities struct {
	Geonameid int    `json:"value"`
	Name      string `json:"display"`
}

type Cities struct {
}

func (me Cities) Name() string {
	return "Cities"
}

func (me Cities) Register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path("/api/cities").
		Doc("Manage cities").
		Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML) // you can specify this per route as well

	ws.Route(ws.GET("/").To(me.list).
		// docs
		Doc("get the cities list").
		Operation("list").
		Returns(200, "OK", []ResCities{}))

	container.Add(ws)
}

func (me Cities) list(request *restful.Request, response *restful.Response) {

	r := request.Request
	err := r.ParseForm()
	if err != nil {
		fmt.Println("ParseForm err: ", err)
		return
	}

	cities := []ResCities{}

	err = db.DB.Select(&cities, "SELECT geonameid,name FROM city JOIN city_translation ON city_translation.city_geonameid = city.geonameid LEFT JOIN lang ON city_translation.lang_id = lang.id WHERE (name_ascii ILIKE $1 OR name ILIKE $1) AND country_geonameid = $2 AND (lang.iso_code = $3 OR lang.iso_code = 'D')", r.FormValue("search")+"%", r.FormValue("id_country"), r.FormValue("lang"))
	if err != nil {
		fmt.Println("err: ", err)
		return
	}

	response.WriteEntity(cities)
}

func (me Cities) create(request *restful.Request, response *restful.Response) {
	fmt.Println("request :", request.Request)
}

func (me Cities) update(request *restful.Request, response *restful.Response) {
	//params := mux.Vars(r)
	//uid := params["id"]
	//email := r.FormValue("email")
}

func (me Cities) delete(request *restful.Request, response *restful.Response) {
}
