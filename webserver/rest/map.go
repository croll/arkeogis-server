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
	"log"
	"net/http"
	"reflect"
	"strings"

	db "github.com/croll/arkeogis-server/db"
	routes "github.com/croll/arkeogis-server/webserver/routes"
)

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:        "/api/search/sites/{id:[0-9]+}",
			Description: "Get list of all databases in arkeogis",
			Func:        SearchSites,
			Method:      "GET",
			Params:      reflect.TypeOf(SearchSitesParams{}),
		},
	}
	routes.RegisterMultiple(Routes)
}

// DatabaseGetParams are params received by REST query
type SearchSitesParams struct {
	Id int
}

// GetSitesAsJSON returns all infos about a database
func SearchSites(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Params.(*SearchSitesParams)

	tx, err := db.DB.Beginx()
	if err != nil {
		log.Println("can't start transaction")
		userSqlError(w, err)
		return
	}

	var jsonResult []string

	q := "SELECT '{\"type\": \"Feature\", ' ||"
	q += "'\"geometry\": {\"type\": \"Point\", \"coordinates\": [' || ("
	q += "	SELECT ST_X(geom::geometry) || ', ' || ST_Y(geom::geometry) AS coordinates FROM site WHERE id = s.id"
	q += ") || ']}, ' ||"
	q += "'\"properties\": {\"infos\": ' || ("
	q += "	SELECT row_to_json(site_infos) || ',' || "
	q += "	'\"site_ranges\": ' || ("
	q += "		SELECT  array_to_json(array_agg(row_to_json(q_src))) FROM ("
	q += "			SELECT *,"
	q += "			("
	q += "				SELECT array_to_json(array_agg(row_to_json(q_src2))) FROM ("
	q += "					SELECT src.*, srctr.comment, srctr.bibliography FROM site_range__charac src LEFT JOIN site_range__charac_tr srctr ON src.id = srctr.site_range__charac_id WHERE src.site_range_id IN (SELECT site_range_id FROM site_range__charac WHERE site_range_id = sr.id)"
	q += "				) q_src2"
	q += "			) characs"
	q += "	   	FROM site_range sr WHERE sr.site_id = s.id) q_src"
	q += "	)"
	q += "	 FROM (SELECT code, name, city_name, city_geonameid, centroid, occupation, created_at, updated_at FROM site WHERE id = s.id) site_infos"
	q += ")"
	q += "|| '}}'"
	q += " FROM site s WHERE database_id = $1"

	err = tx.Select(&jsonResult, q, params.Id)

	if err != nil {
		fmt.Println(err.Error())
	}

	jsonString := "{\"type\": \"FeatureCollection\", \"features\": [" + strings.Join(jsonResult, ",") + "]}"

	if jsonString == "" {
		jsonString = "null"
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonString))

	return

}
