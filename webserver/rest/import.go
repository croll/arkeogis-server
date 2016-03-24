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
	//	"github.com/croll/arkeogis-server/csvimport"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	//	"strings"

	"github.com/croll/arkeogis-server/databaseimport"
	routes "github.com/croll/arkeogis-server/webserver/routes"
)

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:        "/api/import/step1",
			Description: "Fist step of CSV importation of sites in arkeogis",
			Func:        ImportStep1,
			Method:      "POST",
			Json:        reflect.TypeOf(ImportStep1T{}),
		},
	}
	routes.RegisterMultiple(Routes)
}

// ImportStep1T struct holds information provided by user
type ImportStep1T struct {
	DatabaseLang       int
	DatabaseName       string
	SelectedContinent  int
	SelectedCountries  []int
	UseGeonames        bool
	GeographicalExtent string
	Separator          string
	EchapCharacter     string
	File               *routes.File
}

// ImportStep1 is called by rest
func ImportStep1(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Json.(*ImportStep1T)

	fmt.Println(params)
	filepath := "./uploaded/" + params.File.Name
	outfile, err := os.Create(filepath)
	if err != nil {
		http.Error(w, "Error saving file: "+err.Error(), http.StatusBadRequest)
		return
	}

	_, err = io.WriteString(outfile, params.File.Content)
	if err != nil {
		http.Error(w, "Error saving file: "+err.Error(), http.StatusBadRequest)
		return
	}

	parser, err := databaseimport.NewParser(filepath, params.DatabaseLang)
	if err != nil {
		http.Error(w, "Error parsing file: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Set parser preferences
	parser.SetUserChoices("UseGeonames", params.UseGeonames)

	// Init import
	dbImport := new(databaseimport.DatabaseImport)
	dbImport.New(parser, 1, params.DatabaseName, params.DatabaseLang, true)

	// Analyze csv headers
	if err := parser.CheckHeader(); err != nil {
		if err != nil {

			// Prepare response
			response := struct {
				Errors []*databaseimport.ParserError `json:"errors"`
			}{
				parser.Errors,
			}
			l, _ := json.Marshal(response)
			w.Header().Set("Content-Type", "application/json")
			w.Write(l)
			return
		}
	}

	err = parser.Parse(dbImport.ProcessRecord)
	/*
		if err != nil {
			for siteCode, e := range dbImport.Errors {
				fmt.Println("Error line", e.Line, "-- Site Code:", siteCode, "-- Column", strings.Join(e.Columns, ","), ":", e.ErrMsg)
			}
		}
	*/

	// If error ...
	if err != nil {
		dbImport.Tx.Rollback()
	} else {
		// Commit or Rollback if we are in simulation mode
		switch dbImport.Simulate {
		case true:
			err = dbImport.Tx.Rollback()
		case false:
			err = dbImport.Tx.Commit()
		}
	}

	// Prepare response

	var sitesWithError []string

	for id := range dbImport.SitesWithError {
		sitesWithError = append(sitesWithError, id)
	}

	response := struct {
		NumberOfSites  int                           `json:"nbSites"`
		SitesWithError []string                      `json:"sitesWithError"`
		Errors         []*databaseimport.ImportError `json:"errors"`
		Lines          int                           `json:"nbLines"`
	}{
		NumberOfSites:  dbImport.NumberOfSites,
		SitesWithError: sitesWithError,
		Errors:         dbImport.Errors,
		Lines:          dbImport.Parser.Line - 1, // Remove first line
	}

	w.Header().Set("Content-Type", "application/json")
	lok, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(lok)
}
