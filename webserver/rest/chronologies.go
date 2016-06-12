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
	"github.com/croll/arkeogis-server/translate"

	routes "github.com/croll/arkeogis-server/webserver/routes"
	sqlx_types "github.com/jmoiron/sqlx/types"
)

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:        "/api/chronologies",
			Func:        ChronologiesAll,
			Description: "Get all chronologies in all languages",
			Method:      "GET",
		},
	}
	routes.RegisterMultiple(Routes)
}

func ChronologiesAll(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	type row struct {
		Parent_id int                 `db:"parent_id" json:"parent_id"`
		Id        int                 `db:"id" json:"id"`
		Tr        sqlx_types.JSONText `db:"tr" json:"tr"`
	}

	chronologies := []row{}

	//err := db.DB.Select(&chronologies, "select parent_id, id, to_json((select array_agg(chronology_tr.*) from chronology_tr where chronology_tr.chronology_id = chronology.id)) as tr FROM chronology order by parent_id, \"order\", id")
	transquery, err := translate.GetQueryTranslationsAsJSONObject("chronology_tr", "tbl.chronology_id = chronology.id", "", false, "name")
	q := "select parent_id, id, (" + transquery + ") as tr FROM chronology order by parent_id, \"order\", id"
	fmt.Println("q: ", q)
	err = db.DB.Select(&chronologies, q)
	fmt.Println("chronologies: ", chronologies)
	if err != nil {
		fmt.Println("err: ", err)
		return
	}

	j, err := json.Marshal(chronologies)
	w.Write(j)
}
