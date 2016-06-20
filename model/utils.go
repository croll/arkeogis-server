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

package model

import (
	"fmt"
	"reflect"
	"strings"

	db "github.com/croll/arkeogis-server/db"
)

// MapSqlTranslations return a mapped isocode:translation of a [{ isocode: translation }] array from sql
func MapSqlTranslations(src interface{}, isofieldname string, valuefieldname string) map[string]string {
	s := reflect.ValueOf(src)
	if s.Kind() != reflect.Slice {
		panic("MapSqlTranslations() given a non-slice type for src")
	}

	var res = make(map[string]string, s.Len())
	for i := 0; i < s.Len(); i++ {
		e := s.Index(i)
		viso := e.FieldByName(isofieldname)
		siso := viso.String()

		vf := e.FieldByName(valuefieldname)
		tr := vf.String()

		res[siso] = tr
	}
	return res
}

// GetQueryTranslationsAsJSON load translations from database
func GetQueryTranslationsAsJSON(tableName, where, wrapTo string, fields ...string) string {
	var f = "*"
	if len(fields) > 0 {
		f = strings.Join(fields, ", tbl.")
	}
	return db.AsJSON("SELECT tbl."+f+", la.isocode FROM "+tableName+" tbl LEFT JOIN lang la ON tbl.lang_isocode = la.isocode WHERE "+where, true, wrapTo, true)
}

// GetQueryTranslationsAsJSONObject load translations from database
func GetQueryTranslationsAsJSONObject(tableName, where string, wrapTo string, noBrace bool, fields ...string) string {

	jsonQuery := "SELECT '"

	if noBrace == false {
		jsonQuery += "{"
	}
	if wrapTo != "" {
		jsonQuery += "\"" + wrapTo + "\": {' || "
	} else {
		jsonQuery += "' || "

	}
	numFields := len(fields)
	if numFields == 0 {
		fmt.Println("ERROR : GetQueryTranslationsAsJSONObject: You have to provide at least one field")
		return ""
	}
	for k, f := range fields {
		jsonQuery += "'\"" + f + "\": ' || json_object_agg(la.isocode, tbl." + f + ")"
		if k < numFields-1 {
			jsonQuery += " || ',' || "
		}
	}
	if wrapTo != "" {
		jsonQuery += " || '}'"
	}
	if noBrace == false {
		jsonQuery += " || '}'"
	}
	jsonQuery += " FROM " + tableName + " tbl LEFT JOIN lang la ON tbl.lang_isocode = la.isocode WHERE " + where
	fmt.Println(jsonQuery)
	return jsonQuery
}
