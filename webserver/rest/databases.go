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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"time"

	db "github.com/croll/arkeogis-server/db"
	"github.com/croll/arkeogis-server/model"
	routes "github.com/croll/arkeogis-server/webserver/routes"
)

// DatabaseInfosParams are params received by REST query
type DatabaseInfosParams struct {
	Id       int `min:"0" error:"Database Id is mandatory"`
	ImportID int
}

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:        "/api/database",
			Description: "Get list of all databases in arkeogis",
			Func:        DatabaseList,
			Method:      "GET",
			Permissions: []string{},
			Params:      reflect.TypeOf(DatabaseListParams{}),
		},
		&routes.Route{
			Path:        "/api/database/{id:[0-9]+}",
			Description: "Get infos on an arkeogis database",
			Func:        DatabaseInfos,
			Method:      "GET",
			Permissions: []string{},
			Params:      reflect.TypeOf(DatabaseInfosParams{}),
		},
		&routes.Route{
			Path:        "/api/database/{id:[0-9]+}/export",
			Description: "Export database as csv",
			Func:        DatabaseExportCSV,
			Method:      "GET",
			Permissions: []string{},
			Params:      reflect.TypeOf(DatabaseInfosParams{}),
		},
		&routes.Route{
			Path:        "/api/database/{id:[0-9]+}/csv/{importid:[0-9]{0,}}",
			Description: "Get the csv used at import",
			Func:        DatabaseGetImportedCSV,
			Method:      "GET",
			Permissions: []string{},
			Params:      reflect.TypeOf(DatabaseInfosParams{}),
		},
		&routes.Route{
			Path:        "/api/database/delete",
			Description: "Delete database",
			Func:        DatabaseDelete,
			Method:      "POST",
			Permissions: []string{},
			Json:        reflect.TypeOf(DatabaseInfosParams{}),
		},
		&routes.Route{
			Path:        "/api/licences",
			Description: "Get list of licenses",
			Func:        LicensesList,
			Method:      "GET",
			Permissions: []string{
			//"AdminUsers",
			},
		},
	}
	routes.RegisterMultiple(Routes)
}

type DatabaseListParams struct {
	Bounding_box string
	Start_date   int  `json:"start_date"`
	End_date     int  `json:"end_date"`
	Check_dates  bool `json:"check_dates"`
}

// DatabaseList returns the list of databases
func DatabaseList(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Params.(*DatabaseListParams)

	tx, err := db.DB.Beginx()
	if err != nil {
		userSqlError(w, err)
		return
	}

	_user, _ := proute.Session.Get("user")
	user := _user.(model.User)

	q := "SELECT d.*, ST_AsGeoJSON(d.geographical_extent_geom) as geographical_extent_geom, l.name as license, (SELECT count(*) from site WHERE database_id = d.id) AS number_of_sites, (SELECT number_of_lines FROM import WHERE database_id = d.id ORDER BY id DESC LIMIT 1) AS number_of_lines, u.firstname || ' ' || u.lastname as author FROM \"database\" d LEFT JOIN \"user\" u ON d.owner = u.id LEFT JOIN license l ON l.id = d.license_id WHERE 1 = 1"

	type dbInfos struct {
		model.Database
		Author              string            `json:"author"`
		Description         map[string]string `json:"description"`
		Geographical_limit  map[string]string `json:"geographical_limit"`
		Bibliography        map[string]string `json:"bibliography"`
		Context_description map[string]string `json:"context_description"`
		Source_description  map[string]string `json:"source_description"`
		Source_relation     map[string]string `json:"source_relation"`
		Copyright           map[string]string `json:"copyright"`
		Subject             map[string]string `json:"subject"`
		Number_of_lines     int               `json:"number_of_lines"`
		Number_of_sites     int               `json:"number_of_sites"`
		License             string            `json:"license"`
		Contexts            []string          `json:"context"`
		Countries           []struct {
			Id   int
			Name string
		} `json:"countries"`
		Continents []struct {
			Id   int
			Name string
		} `json:"continents"`
		Authors []struct {
			Id       int
			Fullname string
		} `json:"authors"`
	}

	if params.Bounding_box != "" {
		q += " AND ST_Contains(ST_GeomFromGeoJSON(:bounding_box), geographical_extent_geom::::geometry)"
	}

	if params.Check_dates {
		q += " AND d.start_date > :start_date AND d.end_date < :end_date"
	}

	databases := []dbInfos{}

	nstmt, err := tx.PrepareNamed(q)
	if err != nil {
		err = errors.New("rest.databases::DatabaseList : (infos) " + err.Error())
		log.Println(err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}
	err = nstmt.Select(&databases, params)

	for _, database := range databases {

		// Authors
		astmt, err := tx.PrepareNamed("SELECT id, firstname || ' ' || lastname  as fullname FROM \"user\" u LEFT JOIN database__authors da ON u.id = da.user_id WHERE da.database_id = :id")
		if err != nil {
			err = errors.New("rest.databases::DatabaseList (authors) : " + err.Error())
			log.Println(err)
			userSqlError(w, err)
			_ = tx.Rollback()
			return
		}
		err = astmt.Select(&database.Authors, database)

		// Contexts
		cstmt, err := tx.PrepareNamed("SELECT context FROM database_context WHERE database_id = :id")
		if err != nil {
			err = errors.New("rest.databases::DatabaseList (contexts) : " + err.Error())
			log.Println(err)
			userSqlError(w, err)
			_ = tx.Rollback()
			return
		}
		err = cstmt.Select(&database.Authors, database)

		// Countries
		if database.Geographical_extent == "country" {
			coustmt, err := tx.Preparex("SELECT ctr.name FROM country_tr ctr LEFT JOIN country c ON ctr.country_geonameid = c.geonameid LEFT JOIN database__country dc ON c.geonameid = dc.country_geonameid WHERE dc.database_id = $1 AND ctr.lang_isocode = $2")
			if err != nil {
				err = errors.New("rest.databases::DatabaseList (countries) : " + err.Error())
				log.Println(err)
				userSqlError(w, err)
				_ = tx.Rollback()
				return
			}
			err = coustmt.Select(&database.Countries, database.Id, user.First_lang_isocode)
		}

		// Continents
		if database.Geographical_extent == "continent" {
			constmt, err := tx.Preparex("SELECT ctr.name FROM continent_tr ctr LEFT JOIN continent c ON ctr.continent_geonameid = c.geonameid LEFT JOIN database__continent dc ON c.geonameid = dc.continent_geonameid WHERE dc.database_id = $1 AND ctr.lang_isocode = $2")
			if err != nil {
				err = errors.New("rest.databases::DatabaseList : (continents) " + err.Error())
				log.Println(err)
				userSqlError(w, err)
				_ = tx.Rollback()
				return
			}
			err = constmt.Select(&database.Continents, database.Id, user.First_lang_isocode)
		}
	}

	if err != nil {
		log.Println(err)
		_ = tx.Rollback()
		userSqlError(w, err)
		return
	}

	for _, database := range databases {
		tr := []model.Database_tr{}
		err = tx.Select(&tr, "SELECT * FROM database_tr WHERE database_id = "+strconv.Itoa(database.Id))
		if err != nil {
			log.Println(err)
			_ = tx.Rollback()
			userSqlError(w, err)
			return
		}
		database.Description = model.MapSqlTranslations(tr, "Lang_isocode", "Description")
		database.Geographical_limit = model.MapSqlTranslations(tr, "Lang_isocode", "Geographical_limit")
		database.Bibliography = model.MapSqlTranslations(tr, "Lang_isocode", "Bibliography")
		database.Context_description = model.MapSqlTranslations(tr, "Lang_isocode", "Context_description")
		database.Source_description = model.MapSqlTranslations(tr, "Lang_isocode", "Source_description")
		database.Source_relation = model.MapSqlTranslations(tr, "Lang_isocode", "Source_relation")
		database.Copyright = model.MapSqlTranslations(tr, "Lang_isocode", "Copyright")
		database.Subject = model.MapSqlTranslations(tr, "Lang_isocode", "Subject")
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		userSqlError(w, err)
		return
	}

	l, _ := json.Marshal(databases)
	w.Header().Set("Content-Type", "application/json")
	w.Write(l)
}

// LicensesList returns the list of licenses which can be assigned to databases
func LicensesList(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	databases := []model.License{}
	err := db.DB.Select(&databases, "SELECT * FROM \"license\"")
	if err != nil {
		routes.ServerError(w, 500, "INTERNAL ERROR")
		return
	}
	l, _ := json.Marshal(databases)
	w.Header().Set("Content-Type", "application/json")
	w.Write(l)
}

// DatabaseEnumList returns the list of enums fields
// We have to link them with a translation manually clientside
func DatabaseEnumList(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	enums := struct {
		ScaleResolution    []string
		GeographicalExtent []string
		Type               []string
		State              []string
		Context            []string
		Occupation         []string
		KnowledgeType      []string
	}{}
	db.DB.Select(&enums.ScaleResolution, "SELECT unnest(enum_range(NULL::database_scale_resolution))")
	db.DB.Select(&enums.GeographicalExtent, "SELECT unnest(enum_range(NULL::database_geographical_extent))")
	db.DB.Select(&enums.Type, "SELECT unnest(enum_range(NULL::database_type))")
	db.DB.Select(&enums.State, "SELECT unnest(enum_range(NULL::database_state))")
	db.DB.Select(&enums.Context, "SELECT unnest(enum_range(NULL::database_context))")
}

// DatabaseInfos return detailed infos on an database
func DatabaseInfos(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	params := proute.Params.(*DatabaseInfosParams)
	tx, err := db.DB.Beginx()
	if err != nil {
		log.Println("can't start transaction")
		userSqlError(w, err)
		return
	}
	d := model.DatabaseFullInfos{}
	d.Id = params.Id

	// dbInfos, err := d.GetFullInfosAsJSON(tx, proute.Lang1.Isocode)
	dbInfos, err := d.GetFullInfos(tx, proute.Lang1.Isocode)

	if err != nil {
		log.Println("Error getting database infos", err)
	}

	w.Header().Set("Content-Type", "application/json")
	// w.Write([]byte(dbInfos))
	l, _ := json.Marshal(dbInfos)
	w.Write(l)
}

func DatabaseExportCSV(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	params := proute.Params.(*DatabaseInfosParams)
	tx, err := db.DB.Beginx()
	if err != nil {
		log.Println("can't start transaction")
		userSqlError(w, err)
		return
	}
	d := model.DatabaseFullInfos{}
	d.Id = params.Id

	csvContent, err := d.ExportCSV(tx)
	if err != nil {
		log.Println("Unable to export database")
		userSqlError(w, err)
		return
	}
	t := time.Now()
	filename := d.Name + "-" + fmt.Sprintf("%d-%d-%d %d:%d:%d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second()) + ".csv"
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Write([]byte(csvContent))
}

func DatabaseDelete(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	params := proute.Json.(*DatabaseInfosParams)
	tx, err := db.DB.Beginx()
	if err != nil {
		log.Println("can't start transaction")
		userSqlError(w, err)
		return
	}
	d := model.Database{}
	d.Id = params.Id

	if params.Id == 0 {
		log.Println("Unable to delete database. Id is not provided")
		userSqlError(w, err)
		tx.Rollback()
		return
	}

	err = d.Delete(tx)
	if err != nil {
		log.Println("Unable to delete database")
		userSqlError(w, err)
		tx.Rollback()
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Unable to delete database")
		userSqlError(w, err)
		tx.Rollback()
		return
	}

}

func DatabaseGetImportedCSV(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	params := proute.Params.(*DatabaseInfosParams)

	var err error

	var infos struct {
		Md5sum   string
		Filename string
	}

	if params.ImportID > 0 {
		err = db.DB.Get(&infos, "SELECT md5sum, filename FROM import WHERE database_id = $1 AND id = $2", params.Id, params.ImportID)
	} else {
		err = db.DB.Get(&infos, "SELECT md5sum, filename FROM import WHERE database_id = $1 ORDER BY id DESC LIMIT 1", params.Id)
	}

	if infos.Md5sum == "" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("No file found. Please reimport database."))
		return
	}

	filename := infos.Md5sum + "_" + infos.Filename

	if err != nil {
		log.Println("Unable to get imported csv database")
		userSqlError(w, err)
		return
	}

	content, err := ioutil.ReadFile("./uploaded/databases/" + filename)

	if err != nil {
		log.Println("Unable to read the csv file")
		userSqlError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename="+infos.Filename)
	w.Write([]byte(content))
}
