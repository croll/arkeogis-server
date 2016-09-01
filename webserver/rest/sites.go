/* ArkeoGIS - The Geographic Information System for Archaeologists
 * Copyright (C) 2015-2016 CROLL SAS
 *
 * Authors :
 *  Christophe Beveraggi <beve@croll.fr>
 *  Nicolas Dimitrijevic <nicolas@croll.fr> *
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
	"time"

	db "github.com/croll/arkeogis-server/db"

	routes "github.com/croll/arkeogis-server/webserver/routes"
)

type GetSiteParams struct {
	ID int `min:"1"`
}

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:        "/api/site/{id:[0-9]+}",
			Description: "Get site infos",
			Func:        GetSite,
			Permissions: []string{},
			Params:      reflect.TypeOf(GetSiteParams{}),
			Method:      "GET",
		},
	}
	routes.RegisterMultiple(Routes)
}

// GetSite returns all informations about a site
func GetSite(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Params.(*GetSiteParams)

	tx, err := db.DB.Beginx()
	if err != nil {
		log.Println("can't start transaction")
		userSqlError(w, err)
		return
	}

	// for measuring execution time
	start := time.Now()

	var jsonResult []string

	q := `SELECT '{"type": "Feature", ' ||`
	q += `'"geometry": {"type": "Point", "coordinates": [' || (`
	q += `	SELECT ST_X(geom::geometry) || ', ' || ST_Y(geom::geometry) AS coordinates FROM site WHERE id = s.id`
	q += `) || ']}, ' ||`
	q += `'"properties": {"infos": ' || (`
	q += `	SELECT row_to_json(site_infos) || ',' || `
	q += `	'"site_ranges": ' || (`
	q += `		SELECT  array_to_json(array_agg(row_to_json(q_src))) FROM (`
	q += `			SELECT *,`
	q += `			(`
	q += `				SELECT array_to_json(array_agg(row_to_json(q_src2))) FROM (`
	q += `					SELECT src.*, srctr.comment, srctr.bibliography FROM site_range__charac src LEFT JOIN site_range__charac_tr srctr ON src.id = srctr.site_range__charac_id WHERE src.site_range_id IN (SELECT site_range_id FROM site_range__charac WHERE site_range_id = sr.id)`
	q += `				) q_src2`
	q += `			) characs`
	q += `	   	FROM site_range sr WHERE sr.site_id = s.id) q_src`
	q += `	)`
	q += `	 FROM (SELECT si.code, si.name, si.city_name, si.city_geonameid, si.centroid, si.occupation, si.altitude, si.created_at, si.updated_at, d.name as database_name, (SELECT array_to_json(array_agg(row_to_json(d_src))) FROM (SELECT firstname || ' ' || lastname as fullname FROM "user" u LEFT JOIN database__authors da ON u.id = da.user_id WHERE da.database_id = 43) d_src) as authors FROM site si LEFT JOIN database d ON si.database_id = d.id WHERE si.id = s.id) site_infos`
	q += `)`
	q += `|| '}}'`
	q += ` FROM site s WHERE s.id = ($1)`

	err = tx.Select(&jsonResult, q, params.ID)

	elapsed := time.Since(start)
	fmt.Printf("mapGetSitesAsJson took %s", elapsed)

	if err != nil {
		fmt.Println(err.Error())
		tx.Rollback()
	}

	tx.Commit()

	jsonString := `{"type": "FeatureCollection", "features": [` + strings.Join(jsonResult, ",") + `]}`
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonString))

}
