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

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/croll/arkeogis-server/databaseimport"
)

func main() {

	parser, err := databaseimport.NewParser("essai-import_Bernard2003_revue_extrait-qqusaccent-demoins.csv", 48)
	if err != nil {
		log.Fatalln(err)
	}
	if err := parser.CheckHeader(); err != nil {
		if err != nil {
			for _, e := range parser.Errors {
				fmt.Println("column", strings.Join(e.Columns, ","), ":", e.ErrMsg)
			}
		}
		log.Fatalln(err)
	}
	dbImport := new(databaseimport.DatabaseImport)
	dbImport.New(parser, 1, "My test database", 48)
	parser.SetUserChoices("UseGeonames", true)
	err = parser.Parse(dbImport.ProcessRecord)
	if err != nil {
		for _, e := range dbImport.Errors {
			fmt.Println("Error line", e.Line, "-- Site Code:", e.SiteCode, "-- Column", strings.Join(e.Columns, ","), ":", e.ErrMsg)
		}
	}
	// Commit or Rollback if we are in simulation mode
	err = dbImport.Tx.Rollback()
	if err != nil {
		log.Fatalln(err)
	}
}
