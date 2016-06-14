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

// LangGetParams are params received by REST query
type LangGetParams struct {
	Active int
}

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:        "/api/langs",
			Description: "Get languages list that are available",
			Func:        LangList,
			Method:      "GET",
			Params:      reflect.TypeOf(LangGetParams{}),
			Permissions: []string{},
		},
	}
	routes.RegisterMultiple(Routes)
}

// LangList return the list of langs in arkeogis
func LangList(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Params.(*LangGetParams)

	langs := []struct {
		Id      uint32 `json:"id"`
		IsoCode string `json:"iso_code"`
		Name    string `json:"name"`
	}{}

	fmt.Println(params)

	q := "SELECT id, iso_code as isocode, name FROM lang LEFT JOIN lang_tr ON lang_tr.lang_isocode = lang.isocode WHERE iso_code != 'D' AND lang_tr.lang_isocode_tr = $1"

	if params.Active == 1 {
		q += " AND active = true"
	}

	err := db.DB.Select(&langs, q, proute.Lang1.Id)
	if err != nil {
		fmt.Println("err: ", err)
		return
	}

	l, _ := json.Marshal(langs)

	w.Header().Set("Content-Type", "application/json")
	w.Write(l)
}
