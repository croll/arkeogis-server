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

	"net/http"

	db "github.com/croll/arkeogis-server/db"
	"github.com/croll/arkeogis-server/translate"

	routes "github.com/croll/arkeogis-server/webserver/routes"
	sqlx_types "github.com/jmoiron/sqlx/types"
)

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:        "/api/characs",
			Func:        CharacsAll,
			Description: "Get all characs in all languages",
			Method:      "GET",
		},
	}
	routes.RegisterMultiple(Routes)
}

func CharacsAll(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	type row struct {
		Parent_id int                 `db:"parent_id" json:"parent_id"`
		Id        int                 `db:"id" json:"id"`
		Tr        sqlx_types.JSONText `db:"tr" json:"tr"`
	}

	characs := []row{}

	//err := db.DB.Select(&characs, "select parent_id, id, to_json((select array_agg(charac_tr.*) from charac_tr where charac_tr.charac_id = charac.id)) as tr FROM charac order by parent_id, \"order\", id")
	transquery, err := translate.GetQueryTranslationsAsJSONObject("charac_tr", "tbl.charac_id = charac.id", "", false, "name")
	q := "select parent_id, id, (" + transquery + ") as tr FROM charac order by parent_id, \"order\", id"
	fmt.Println("q: ", q)
	err = db.DB.Select(&characs, q)
	fmt.Println("characs: ", characs)
	if err != nil {
		fmt.Println("err: ", err)
		return
	}

	j, err := json.Marshal(characs)
	w.Write(j)
}
