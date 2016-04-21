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
	"net/http"
	"reflect"

	db "github.com/croll/arkeogis-server/db"
	routes "github.com/croll/arkeogis-server/webserver/routes"
)

type ContinentsListParams struct {
}

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:        "/api/continents",
			Description: "Search for continents available on our world, using a search string",
			Func:        ContinentsList,
			Params:      reflect.TypeOf(ContinentsListParams{}),
			Method:      "GET",
		},
	}
	routes.RegisterMultiple(Routes)
}

func ContinentsList(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	//params := proute.Params.(*ContinentsListParams)

	continents := []struct {
		Geonameid uint32 `json:"geonameid"`
		Name      string `json:"name"`
	}{}

	err := db.DB.Select(&continents, "SELECT geonameid, name FROM continent LEFT JOIN continent_tr ON continent.geonameid = continent_tr.continent_geonameid LEFT JOIN lang ON continent_tr.lang_id = lang.id WHERE active = true AND continent.iso_code != 'U' AND (lang.Id = $1 OR lang.iso_code = 'D')", proute.Lang1.Id)
	if err != nil {
		fmt.Println("err: ", err)
		return
	}

	l, _ := json.Marshal(continents)

	w.Header().Set("Content-Type", "application/json")
	w.Write(l)
}
