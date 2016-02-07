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

	db "github.com/croll/arkeogis-server/db"
	routes "github.com/croll/arkeogis-server/webserver/routes"
)

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:        "/api/langs",
			Description: "Get languages list that are available",
			Func:        LangList,
			Method:      "GET",
		},
	}
	routes.RegisterMultiple(Routes)
}

func LangList(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	langs := []struct {
		Id      uint32 `json:"id"`
		IsoCode string `json:"iso_code"`
	}{}

	err := db.DB.Select(&langs, "SELECT id, iso_code as isocode FROM lang WHERE active = true AND iso_code != 'D'")
	if err != nil {
		fmt.Println("err: ", err)
		return
	}

	l, _ := json.Marshal(langs)

	w.Header().Set("Content-Type", "application/json")
	w.Write(l)
}
