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
	"log"
	"net/http"

	db "github.com/croll/arkeogis-server/db"
	routes "github.com/croll/arkeogis-server/webserver/routes"
)

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:        "/api/stats",
			Description: "Main stats",
			Func:        StatsGet,
			Method:      "GET",
		},
	}
	routes.RegisterMultiple(Routes)
}

type StatsResult struct {
	DbCount   int `json:"dbcount"`
	SiteCount int `json:"sitecount"`
}

// StatsGet return some stats
func StatsGet(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	res := StatsResult{}

	err := db.DB.Get(&res.DbCount, `SELECT count(*) from "database" WHERE "published" = true`)
	if err != nil {
		fmt.Println("get dbcount failed", err)
	}
	err = db.DB.Get(&res.SiteCount, `SELECT count(*) from "site" LEFT JOIN "database" ON "site".database_id = "database".id WHERE "database"."published" = true`)
	if err != nil {
		fmt.Println("get sitecount failed", err)
	}

	j, err := json.Marshal(res)
	if err != nil {
		log.Println("marshal failed: ", err)
	}
	//log.Println("result: ", string(j))
	w.Write(j)
}
