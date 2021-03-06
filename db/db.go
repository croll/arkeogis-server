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

package arkeogis

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"

	config "github.com/croll/arkeogis-server/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var DB *sqlx.DB

func init() {
	var err error
	DB, err = sqlx.Open("postgres", formatConnexionString())
	if err != nil {
		log.Fatal(err)
	}
	if err = DB.Ping(); err != nil {
		log.Fatal(err)
	}
	DB.SetConnMaxLifetime(290 * time.Second)
}

func formatConnexionString() string {
	c := ""
	r := reflect.ValueOf(config.Main.Database)
	for i := 0; i < r.NumField(); i++ {
		v := string(r.Field(i).Interface().(string))
		if v != "" {
			c += strings.ToLower(r.Type().Field(i).Name) + "=" + v + " "
		}
	}
	c = strings.Trim(c, " ")
	fmt.Println("*** postgres str :" + c)
	return c
}

func AsJSON(query string, asArray bool, wrapTo string, noBrace bool) (q string) {
	if asArray {
		q = "SELECT array_to_json(array_agg(row_to_json(t))) FROM (" + query + ") t"
	} else {
		q = "SELECT row_to_json(t) FROM (" + query + ") t"
	}
	if wrapTo != "" {
		if noBrace {
			q = "SELECT ('\"" + wrapTo + "\": ' || (" + q + "))"
		} else {
			q = "SELECT ('{\"" + wrapTo + "\": ' || (" + q + ") || '}')"
		}
	}
	return
}

var rgxp = regexp.MustCompile(`^SELECT \('"(\w+)":`)

func JSONQueryBuilder(subQueries []string, tableName, where string) string {
	outp := "SELECT('{' || (SELECT "
	emptyJSON := ""
	for k, sq := range subQueries {
		m := rgxp.FindStringSubmatch(sq)
		if len(m) > 0 {
			emptyJSON = "\"" + m[1] + "\": null"
		}
		outp += "COALESCE((" + sq + "), '" + emptyJSON + "')"
		if k < len(subQueries)-1 {
			outp += " || ',' || "
		}
	}
	outp += " FROM " + tableName + " WHERE " + where
	outp += ") || '}')"
	return outp
}

//'{ "type": "FeatureCollection", "features": [' || ']}'

/*
SELECT '{' ||
	'"site": ' || (
		SELECT row_to_json(site_infos) FROM (SELECT code, name, city_name, city_geonameid, centroid, occupation, created_at, updated_at FROM site WHERE id = s.id) site_infos
	) ||
	'"site_ranges": ' || (
		SELECT  array_to_json(array_agg(row_to_json(q_src))) FROM (
			SELECT *,
			(
				SELECT array_to_json(array_agg(row_to_json(q_src2))) FROM (
					SELECT src.*, srctr.comment, srctr.bibliography FROM site_range__charac src LEFT JOIN site_range__charac_tr srctr ON src.id = srctr.site_range__charac_id WHERE src.site_range_id IN (SELECT site_range_id FROM site_range__charac WHERE srctr.lang_id = 47 AND site_range_id = sr.id)
				) q_src2
			) characs
	   	FROM site_range sr WHERE sr.site_id = s.id) q_src
	)
	|| '}'
FROM site s WHERE id = 1;
*/

/*
SELECT id,
	(
		SELECT (
			SELECT '{"site_ranges": ' || array_to_json(array_agg(row_to_json(q_src))) || ', "characs": ' ||
			(
			SELECT array_to_json(array_agg(row_to_json(q_src2))) FROM (SELECT src.*, srctr.comment, srctr.bibliography FROM site_range__charac src LEFT JOIN site_range__charac_tr srctr ON src.id = srctr.site_range__charac_id WHERE src.site_range_id IN (SELECT id FROM site_range__charac WHERE srctr.lang_id = 47 AND site_range_id IN (SELECT id FROM site_range WHERE site_range.site_id = s.id)) ) q_src2
			) || '}' FROM (SELECT * FROM site_range sr WHERE sr.site_id = s.id) q_src
		)
	) site_ranges_list
FROM site s WHERE id = 13;

*/
//FROM site s WHERE database_id = 13;

/*
SELECT s.id, ST_AsText(geom),
	(
		SELECT (
			SELECT array_to_json(array_agg(row_to_json(q_src))) || ' ' ||
			(
			SELECT array_to_json(array_agg(row_to_json(q_src2))) FROM (SELECT * FROM site_range__charac src WHERE src.site_range_id IS NOT NULL) q_src2
			) characs_list FROM (SELECT id FROM site_range sr WHERE sr.id = s.id) q_src
		)
	) site_ranges_list
FROM site s WHERE s.id = 1;
*/

/*
SELECT s.id, ST_AsText(geom),
	(
		SELECT
			array_to_json(array_agg(row_to_json(q_src))) FROM (SELECT * FROM site_range sr WHERE sr.id = s.id) q_src
	) site_ranges_list
FROM site s WHERE s.database_id = 13;
*/
