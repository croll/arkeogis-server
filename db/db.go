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
	"strconv"
	"strings"

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

func AsJSON(query string, wrapTo string, noBrace bool) (q string) {
	q = "SELECT array_to_json(array_agg(row_to_json(t))) FROM (" + query + ") t"
	if wrapTo != "" {
		if noBrace {
			q = "SELECT ('\"" + wrapTo + "\": ' || (" + q + "))"
		} else {
			q = "SELECT ('{\"" + wrapTo + "\": ' || (" + q + ") || '}')"
		}
	}
	return
}

func JSONQueryBuilder(subQueries []string, databaseName, where string) string {
	//outp := "SELECT (" + strings.Join(subQueries, "), (") + ") FROM " + databaseName + " WHERE " + where
	outp := "SELECT('{' || (SELECT "
	for k, sq := range subQueries {
		outp += "COALESCE((" + sq + "), '\"q" + strconv.Itoa(k) + "\": null')"
		if k < len(subQueries)-1 {
			outp += " || ',' || "
		}
		fmt.Println(outp)
	}
	outp += " FROM " + databaseName + " WHERE " + where
	outp += ") || '}')"
	return outp
}
