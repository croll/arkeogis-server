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
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"log"
	"net/http"
	"os"
	"reflect"
	"time"

	"unicode/utf8"

	"github.com/croll/arkeogis-server/databaseimport"
	db "github.com/croll/arkeogis-server/db"
	"github.com/croll/arkeogis-server/model"
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
			Permissions: []string{
				"import",
			},
		},
		&routes.Route{
			Path:        "/api/import/update-step1",
			Description: "Four step of ArkeoGIS import procedure",
			Func:        ImportStep1Update,
			Method:      "POST",
			Json:        reflect.TypeOf(ImportStep1UpdateT{}),
			Permissions: []string{
				"import",
			},
		},
		&routes.Route{
			Path:        "/api/import/step3",
			Description: "Third step of ArkeoGIS import procedure",
			Func:        ImportStep3,
			Method:      "POST",
			Json:        reflect.TypeOf(ImportStep3T{}),
			Permissions: []string{
				"import",
			},
		},
		&routes.Route{
			Path:        "/api/import/step4",
			Description: "Four step of ArkeoGIS import procedure",
			Func:        ImportStep4,
			Method:      "POST",
			Json:        reflect.TypeOf(ImportStep4T{}),
			Permissions: []string{
				"import",
			},
		},
	}
	routes.RegisterMultiple(Routes)
}

// ImportStep1T struct holds information provided by user
type ImportStep1T struct {
	Name                string `min:"3" max:"50" error:"Wrong length for database name"`
	Geographical_extent string `enum:"undefined,country,continent,international_waters,world" error:"Wrong geographical extent"`
	Default_language    string `min:"2" max:"2" error:"Wrong lang"`
	Continents          []model.Continent
	Countries           []model.Country
	UseGeonames         bool
	Separator           string `min:"1" max:"1" error:"Wrong separator"`
	EchapCharacter      string `min:"1" max:"1" error:"Wrong echap characted"`
	File                *routes.File
}

// ImportStep1 is called by rest
func ImportStep1(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Json.(*ImportStep1T)

	if params.File == nil {
		userSqlError(w, errors.New("No file provided"))
		return
	}

	_user, _ := proute.Session.Get("user")
	user := _user.(model.User)

	filehash := fmt.Sprintf("%x", md5.Sum([]byte(params.File.Name)))
	filepath := "./uploaded/databases/" + filehash + "_" + params.File.Name

	var dbImport *databaseimport.DatabaseImport

	outfile, err := os.Create(filepath)
	if err != nil {
		http.Error(w, "Error saving file: "+err.Error(), http.StatusBadRequest)
		log.Println(err)
		return
	}

	// Save the file on filesystem
	_, err = io.WriteString(outfile, string(params.File.Content))
	if err != nil {
		http.Error(w, "Error saving file: "+err.Error(), http.StatusBadRequest)
		log.Println(err)
		return
	}

	// Parse the file
	parser, err := databaseimport.NewParser(filepath, params.Default_language, user.First_lang_isocode)
	if err != nil {
		parser.AddError("IMPORT.CSV_FILE.T_ERROR_PARSING_FAILED")
	}

	// utf8 validation
	if !utf8.ValidString(string(params.File.Content)) {
		parser.AddError("IMPORT.CSV_FILE.T_ERROR_NOT_UTF8_ENCODING")
	}

	if parser.HasError() {
		sendError(w, parser.Errors)
		return
	}

	// Set parser preferences
	parser.SetUserChoices("UseGeonames", params.UseGeonames)

	// Init import
	dbImport = new(databaseimport.DatabaseImport)
	err = dbImport.New(parser, user.Id, params.Name, params.Default_language, filehash)
	if err != nil {
		parser.AddError(err.Error())
		sendError(w, parser.Errors)
		dbImport.Tx.Rollback()
		return
	}

	// Analyze csv headers
	if err = parser.CheckHeader(); err != nil {
		if err != nil {
			sendError(w, parser.Errors)
			dbImport.Tx.Rollback()
			return
		}
	}

	// Record database essentials infos
	var continentsID = make([]int, 0)
	for _, c := range params.Continents {
		continentsID = append(continentsID, c.Geonameid)
	}
	var countriesID = make([]int, 0)
	for _, c := range params.Countries {
		countriesID = append(countriesID, c.Geonameid)
	}
	err = dbImport.ProcessEssentialDatabaseInfos(params.Name, params.Geographical_extent, continentsID, countriesID)
	if err != nil {
		parser.AddError("Import: error processing essential infos " + err.Error())
		sendError(w, parser.Errors)
		dbImport.Tx.Rollback()
		return
	}
	ticker := time.NewTicker(time.Second * 10)
	w.Header().Set("Content-Type", "application/json")
	go func() {
		for range ticker.C {
			w.Write([]byte(" "))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}()
	parser.Parse(dbImport.ProcessRecord)
	ticker.Stop()
	/*
		if err != nil {
			for siteCode, e := range dbImport.Errors {
				fmt.Println("Error line", e.Line, "-- Site Code:", siteCode, "-- Column", strings.Join(e.Columns, ","), ":", e.ErrMsg)
			}
		}
	*/

	import_id, err := dbImport.Save(params.File.Name)
	if err != nil {
		parser.AddError("Error saving import " + err.Error())
	}

	if len(dbImport.SitesWithError) < dbImport.NumberOfSites-1 {
		// Cache geom
		err = dbImport.Database.CacheGeom(dbImport.Tx)
		if err != nil {
			parser.AddError("Error caching geom " + err.Error())
		}

		// Cache dates
		err = dbImport.Database.CacheDates(dbImport.Tx)
		if err != nil {
			parser.AddError("Error caching dates" + err.Error())
		}

		err = dbImport.Tx.Commit()
		if err != nil {
			parser.AddError("Error when inserting import into database: " + err.Error())
			sendError(w, parser.Errors)
			return
		}
	} else {
		err = dbImport.Tx.Rollback()
	}

	// Prepare response

	var sitesWithError []string

	for id := range dbImport.SitesWithError {
		sitesWithError = append(sitesWithError, id)
	}

	response := struct {
		DatabaseId     int                           `json:"database_id"`
		ImportId       int                           `json:"import_id"`
		NumberOfSites  int                           `json:"nbSites"`
		SitesWithError []string                      `json:"sitesWithError"`
		Errors         []*databaseimport.ImportError `json:"errors"`
		Lines          int                           `json:"nbLines"`
	}{
		DatabaseId:     dbImport.Database.Id,
		ImportId:       import_id,
		NumberOfSites:  dbImport.NumberOfSites,
		SitesWithError: sitesWithError,
		Errors:         dbImport.Errors,
		Lines:          dbImport.Parser.Line - 1, // Remove first line
	}
	lok, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(lok)
}

// ImportStep1UpdateT struct holds information provided by user
type ImportStep1UpdateT struct {
	Id                  int
	Geographical_extent string
	Continents          []model.Continent
	Countries           []model.Country
}

// ImportStep1Update is called by rest
func ImportStep1Update(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Json.(*ImportStep1UpdateT)

	tx, err := db.DB.Beginx()
	if err != nil {
		http.Error(w, "Error updating step1 informations: "+err.Error(), http.StatusBadRequest)
		return
	}

	var d = &model.Database{}
	d.Id = params.Id

	// Update datatabase name and geographical extent
	err = d.UpdateFields(tx, params, "geographical_extent")
	if err != nil {
		log.Println("Error updating database fields: ", err)
		tx.Rollback()
		userSqlError(w, err)
		return
	}

	// Delete linked continents
	err = d.DeleteContinents(tx)
	if err != nil {
		log.Println("Error updating database : unabled to delete continents", err)
		tx.Rollback()
		return
	}

	// Delete linked countries
	err = d.DeleteCountries(tx)
	if err != nil {
		log.Println("Error updating database : unabled to delete countries", err)
		tx.Rollback()
		return
	}

	if params.Geographical_extent == "country" {
		// Insert countries
		var countriesID = make([]int, 0)
		for _, c := range params.Countries {
			countriesID = append(countriesID, c.Geonameid)
		}
		err = d.AddCountries(tx, countriesID)
		if err != nil {
			log.Println("Error updating database : unabled to add countries", err)
			tx.Rollback()
			return
		}
	} else if params.Geographical_extent == "continent" {
		// Insert continents
		var continentsID = make([]int, 0)
		for _, c := range params.Continents {
			continentsID = append(continentsID, c.Geonameid)
		}
		err = d.AddContinents(tx, continentsID)
		if err != nil {
			log.Println("Error updating database : unabled to add continents", err)
			tx.Rollback()
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Error updating database infos: ", err)
		userSqlError(w, err)
		return
	}

}

func sendError(w http.ResponseWriter, errors []*databaseimport.ParserError) {
	// Prepare response
	response := struct {
		Errors []*databaseimport.ParserError `json:"errors"`
	}{
		errors,
	}
	l, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(l)
	return
}

/*
func writeResponse(w http.ResponseWriter, numberOfSites int, sitesWithError []string, errors []*databaseimport.ParserError, lines int) {
	response := struct {
		NumberOfSites  int                           `json:"nbSites"`
		SitesWithError []string                      `json:"sitesWithError"`
		Errors         []*databaseimport.ImportError `json:"errors"`
		Lines          int                           `json:"nbLines"`
	}{
		NumberOfSites:  numberOfSites,
		SitesWithError: sitesWithError,
		Errors:         errors,
		Lines:          lines, // Remove first line
	}

	w.Header().Set("Content-Type", "application/json")
	lok, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(lok)
}
*/

type ImportStep3T struct {
	Id                     int    `min:"1" error:"Wrong database id"`
	Authors                []int  `min:"1" error:"Wrong num of authors"`
	Type                   string `enum:"undefined,inventory,research,literary-work" error:"Wrong database type"`
	Declared_creation_date time.Time
	Contexts               []string `min:"1" error:"Wrong context"`
	License_ID             int      `min:"1" error:"Wrong license id"`
	Scale_resolution       string   `min:"1" error:"Scale resolution is mandatory"`
	Subject                string   `min:"1" error:"Subject is mandatory"`
	State                  string   `min:"1" error:"State is mandatory"`
	Project_ID             int      `min:"1" error:"Project id is mandatory"`
	Published              bool
	Description            []struct {
		Lang_Isocode string
		Text         string
	} `min:"1" error:"Description is mandatory"`
}

func ImportStep3(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Json.(*ImportStep3T)

	tx, err := db.DB.Beginx()
	if err != nil {
		http.Error(w, "Error step3 informations: "+err.Error(), http.StatusBadRequest)
		return
	}

	d := &model.Database{Id: params.Id}
	fmt.Println("DATABASE", d)
	d.Get(tx)

	err = d.UpdateFields(tx, params, "type", "declared_creation_date", "license_id", "scale_resolution", "state", "published")

	if err != nil {
		log.Println("Error updating database fields: ", err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}
	err = d.DeleteAuthors(tx)
	if err != nil {
		log.Println("Error deleting database authors: ", err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}
	err = d.SetAuthors(tx, params.Authors)
	if err != nil {
		log.Println("Error setting database authors: ", err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}
	err = d.DeleteContexts(tx)
	if err != nil {
		log.Println("Error deleting database contexts: ", err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}
	err = d.SetContexts(tx, params.Contexts)
	if err != nil {
		log.Println("Error setting database contexts: ", err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	// For now subject is not translatable but store it in database_tr anyway
	fmt.Println("DATABASE DEFAULT LANGUAGE " + d.Default_language)
	var subject = []struct {
		Lang_Isocode string
		Text         string
	}{
		{d.Default_language, params.Subject},
	}
	err = d.SetTranslations(tx, "subject", subject)
	if err != nil {
		log.Println("Error setting subject: ", err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	err = d.SetTranslations(tx, "description", params.Description)
	if err != nil {
		log.Println("Error setting database description: ", err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	// If published link database to user project
	if params.Published {
		linked, _ := d.IsLinkedToProject(tx, params.Project_ID)
		if err != nil {
			log.Println("Error checking if database is linked to project: ", err)
			userSqlError(w, err)
			_ = tx.Rollback()
			return
		}
		if !linked {
			err = d.LinkToUserProject(tx, params.Project_ID)
		}
		if err != nil {
			log.Println("Error auto linking database to user project: ", err)
			userSqlError(w, err)
			_ = tx.Rollback()
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Error commiting step 3 infos: ", err)
		userSqlError(w, err)
		return
	}

}

type ImportStep4T struct {
	Id                            int
	Import_ID                     int
	Editor                        string
	Contributor                   string
	Source_description            string
	Context_description           string
	Source_url                    string
	Source_declared_creation_date time.Time
	Source_relation               string
	Source_identifier             string
	Geographical_Limit            []struct {
		Lang_Isocode string
		Text         string
	}
	Bibliography []struct {
		Lang_Isocode string
		Text         string
	}
}

func ImportStep4(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Json.(*ImportStep4T)

	tx, err := db.DB.Beginx()
	if err != nil {
		http.Error(w, "Error saving step3 informations: "+err.Error(), http.StatusBadRequest)
		return
	}

	d := &model.Database{Id: params.Id}
	d.Get(tx)

	err = d.UpdateFields(tx, params, "editor", "contributor")
	if err != nil {
		log.Println("Error updating database fields: ", err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	// For now source description is not translatable but store it in database_tr anyway
	var source_desc = []struct {
		Lang_Isocode string
		Text         string
	}{
		{d.Default_language, params.Source_description},
	}
	err = d.SetTranslations(tx, "source_description", source_desc)
	if err != nil {
		log.Println("Error setting source description: ", err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	// For now source relation is not translatable but store it in database_tr anyway
	var source_relation = []struct {
		Lang_Isocode string
		Text         string
	}{
		{d.Default_language, params.Source_relation},
	}
	err = d.SetTranslations(tx, "source_relation", source_relation)
	if err != nil {
		log.Println("Error setting source relation: ", err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	// For now context description is not translatable but store it in database_tr anyway
	var context_desc = []struct {
		Lang_Isocode string
		Text         string
	}{
		{d.Default_language, params.Context_description},
	}
	err = d.SetTranslations(tx, "context_description", context_desc)
	if err != nil {
		log.Println("Error setting context description: ", err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	err = d.SetTranslations(tx, "geographical_limit", params.Geographical_Limit)
	if err != nil {
		log.Println("Error setting geographical limit: ", err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	err = d.SetTranslations(tx, "bibliography", params.Bibliography)
	if err != nil {
		log.Println("Error setting bibliography: ", err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	// Database handle

	currentHandle, err := d.GetLastHandle(tx)

	if err != nil && err != sql.ErrNoRows {
		log.Println("Error getting last handle: ", err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	handle := &model.Database_handle{
		Database_id: params.Id,
		Import_id:   params.Import_ID,
		Identifier:  params.Source_identifier,
		Url:         params.Source_url,
		Declared_creation_date: params.Source_declared_creation_date,
		Created_at:             time.Now(),
	}

	if currentHandle.Identifier == params.Source_identifier {
		handle.Id = currentHandle.Id
		err = d.UpdateHandle(tx, handle)
	} else {
		_, err = d.AddHandle(tx, handle)
	}

	if err != nil {
		log.Println("Error setting handle informations: ", err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Error commiting step 4 infos: ", err)
		userSqlError(w, err)
		return
	}
}
