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
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
	"net/url"

	db "github.com/croll/arkeogis-server/db"
	export "github.com/croll/arkeogis-server/export"
	"github.com/croll/arkeogis-server/model"
	translate "github.com/croll/arkeogis-server/translate"
	routes "github.com/croll/arkeogis-server/webserver/routes"
)

// DatabaseInfosParams are params received by REST query
type DatabaseInfosParams struct {
	Id       int `min:"0" error:"Database Id is mandatory"`
	ImportID int
}

type DatabaseExportInfosParams struct {
	Id       int `min:"0" error:"Database Id is mandatory"`
	ImportID int
	IncludeSiteId bool
	IncludeInterop bool
}

type DatabaseExportOmekaParams struct {
	Id           int `min:"0" error:"Database Id is mandatory"`
	ChronologyId int `min:"0" error:"chronology is mandatory"`
}

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:        "/api/database",
			Description: "Get list of all databases in arkeogis",
			Func:        DatabaseList,
			Method:      "GET",
			Permissions: []string{
				"request map",
			},
			Params: reflect.TypeOf(DatabaseListParams{}),
		},
		&routes.Route{
			Path:        "/api/database/export",
			Description: "Get list of all databases export",
			Func:        DatabaseExportList,
			Method:      "GET",
			Permissions: []string{},
			Params:      reflect.TypeOf(DatabaseListExportParams{}),
		},
		&routes.Route{
			Path:        "/api/database/{id:[0-9]+}",
			Description: "Get infos on an arkeogis database",
			Func:        DatabaseInfos,
			Method:      "GET",
			Permissions: []string{
				"request map",
			},
			Params: reflect.TypeOf(DatabaseInfosParams{}),
		},
		&routes.Route{
			Path:        "/api/database/{id:[0-9]+}/export",
			Description: "Export database as csv",
			Func:        DatabaseExportCSVArkeogis,
			Method:      "GET",
			Permissions: []string{
				"request map",
			},
			Params: reflect.TypeOf(DatabaseExportInfosParams{}),
		},
		&routes.Route{
			Path:        "/api/database/{id:[0-9]+}/exportOmeka",
			Description: "Export database as csvs in a zip",
			Func:        DatabaseExportZIPOmeka,
			Method:      "GET",
			Permissions: []string{
				"request map",
			},
			Params: reflect.TypeOf(DatabaseExportOmekaParams{}),
		},
		&routes.Route{
			Path:        "/api/database/{id:[0-9]+}/exportxml",
			Description: "Export database informations as XML",
			Func:        DatabaseExportXML,
			Method:      "GET",
			Permissions: []string{
				"request map",
			},
			Params: reflect.TypeOf(DatabaseInfosParams{}),
		},
		&routes.Route{
			Path:        "/api/database/{id:[0-9]+}/csv/{importid:[0-9]{0,}}",
			Description: "Get the csv used at import",
			Func:        DatabaseGetImportedCSV,
			Method:      "GET",
			Permissions: []string{
				"import",
			},
			Params: reflect.TypeOf(DatabaseInfosParams{}),
		},
		&routes.Route{
			Path:        "/api/database/delete",
			Description: "Delete database",
			Func:        DatabaseDelete,
			Method:      "POST",
			Permissions: []string{
				"import",
			},
			Json: reflect.TypeOf(DatabaseInfosParams{}),
		},
		&routes.Route{
			Path:        "/api/licences",
			Description: "Get list of licenses",
			Func:        LicensesList,
			Method:      "GET",
			Permissions: []string{
				"request map",
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
		Re_use        		map[string]string `json:"re_use"`
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
			Id   int    `json:"id"`
			Name string `json:"name"`
		} `json:"countries"`
		Continents []struct {
			Id   int    `json:"id"`
			Name string `json:"name"`
		} `json:"continents"`
		Authors []struct {
			Id       int    `json:"id"`
			Fullname string `json:"fullname"`
		} `json:"authors"`
	}

	if params.Bounding_box != "" {
		q += " AND (ST_Contains(ST_GeomFromGeoJSON(:bounding_box), geographical_extent_geom::::geometry) OR ST_Contains(geographical_extent_geom::::geometry, ST_GeomFromGeoJSON(:bounding_box)) OR ST_Overlaps(ST_GeomFromGeoJSON(:bounding_box), geographical_extent_geom::::geometry))"
	}

	if params.Check_dates {
		q += " AND ((d.start_date = " + strconv.Itoa(math.MinInt32) + " OR d.start_date >= :start_date) AND (d.end_date = " + strconv.Itoa(math.MaxInt32) + " OR d.end_date <= :end_date))"
	}

	viewUnpublished, err := user.HavePermissions(tx, "manage all databases")
	if err != nil {
		userSqlError(w, err)
		return
	}

	if !viewUnpublished {
		q += " AND published = 't' OR d.owner = " + strconv.Itoa(user.Id)
	}

	q += " ORDER BY d.Id DESC"

	// fmt.Println(q)

	databases := []dbInfos{}

	nstmt, err := tx.PrepareNamed(q)
	if err != nil {
		err = errors.New("rest.databases::DatabaseList : (infos) " + err.Error())
		log.Println(err)
		userSqlError(w, err)
		tx.Rollback()
		return
	}
	err = nstmt.Select(&databases, params)
	if err != nil {
		err = errors.New("rest.databases::DatabaseList : (infos) " + err.Error())
		log.Println(err)
		userSqlError(w, err)
		tx.Rollback()
		return
	}

	returnedDatabases := []dbInfos{}

	for _, database := range databases {

		// Authors
		astmt, err2 := tx.PrepareNamed("SELECT id, firstname || ' ' || lastname  as fullname FROM \"user\" u LEFT JOIN database__authors da ON u.id = da.user_id WHERE da.database_id = :id")
		if err2 != nil {
			err = errors.New("rest.databases::DatabaseList (authors) : " + err2.Error())
			log.Println(err)
			userSqlError(w, err)
			tx.Rollback()
			return
		}
		err = astmt.Select(&database.Authors, database)
		if err != nil {
			err = errors.New("rest.databases::DatabaseList (authors) : " + err.Error())
			log.Println(err)
			userSqlError(w, err)
			tx.Rollback()
			return
		}

		// Contexts
		cstmt, err3 := tx.PrepareNamed("SELECT context FROM database_context WHERE database_id = :id")
		if err3 != nil {
			err = errors.New("rest.databases::DatabaseList (contexts) : " + err3.Error())
			log.Println(err)
			userSqlError(w, err)
			tx.Rollback()
			return
		}
		err = cstmt.Select(&database.Contexts, database)

		// Countries
		if database.Geographical_extent == "country" {
			coustmt, err4 := tx.Preparex("SELECT ctr.name FROM country_tr ctr LEFT JOIN country c ON ctr.country_geonameid = c.geonameid LEFT JOIN database__country dc ON c.geonameid = dc.country_geonameid WHERE dc.database_id = $1 AND ctr.lang_isocode = $2")
			if err4 != nil {
				err = errors.New("rest.databases::DatabaseList (countries) : " + err4.Error())
				log.Println(err)
				userSqlError(w, err)
				tx.Rollback()
				return
			}
			err = coustmt.Select(&database.Countries, database.Id, user.First_lang_isocode)
		}

		// Continents
		if database.Geographical_extent == "continent" {
			constmt, err5 := tx.Preparex("SELECT ctr.name FROM continent_tr ctr LEFT JOIN continent c ON ctr.continent_geonameid = c.geonameid LEFT JOIN database__continent dc ON c.geonameid = dc.continent_geonameid WHERE dc.database_id = $1 AND ctr.lang_isocode = $2")
			if err5 != nil {
				err = errors.New("rest.databases::DatabaseList : (continents) " + err5.Error())
				log.Println(err)
				userSqlError(w, err)
				tx.Rollback()
				return
			}
			err = constmt.Select(&database.Continents, database.Id, user.First_lang_isocode)
		}

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
		database.Re_use = model.MapSqlTranslations(tr, "Lang_isocode", "Re_use")
		database.Context_description = model.MapSqlTranslations(tr, "Lang_isocode", "Context_description")
		database.Source_description = model.MapSqlTranslations(tr, "Lang_isocode", "Source_description")
		database.Source_relation = model.MapSqlTranslations(tr, "Lang_isocode", "Source_relation")
		database.Copyright = model.MapSqlTranslations(tr, "Lang_isocode", "Copyright")
		database.Subject = model.MapSqlTranslations(tr, "Lang_isocode", "Subject")
		returnedDatabases = append(returnedDatabases, database)
	}

	if err != nil {
		log.Println(err)
		tx.Rollback()
		userSqlError(w, err)
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		userSqlError(w, err)
		return
	}

	l, _ := json.Marshal(returnedDatabases)
	w.Header().Set("Content-Type", "application/json")
	w.Write(l)
}

/**
 *
 * Export Database List CSV
 *
**/

type DatabaseListExportParams struct {
	Lang string
}

// DatabaseExportList returns the list of databases as csv
func DatabaseExportList(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Params.(*DatabaseListExportParams)

	tx, err := db.DB.Beginx()
	if err != nil {
		userSqlError(w, err)
		return
	}

	_user, _ := proute.Session.Get("user")
	user := _user.(model.User)

	if len(params.Lang) == 0 {
		params.Lang = user.First_lang_isocode
	}

	q := `SELECT d.*,
			ST_AsGeoJSON(d.geographical_extent_geom) as geographical_extent_geom,
			l.name as license,
			(SELECT count(*) from site WHERE database_id = d.id) AS number_of_sites,
			(SELECT number_of_lines FROM import WHERE database_id = d.id ORDER BY id DESC LIMIT 1) AS number_of_lines,
			u.firstname || ' ' || u.lastname as author
		  FROM "database" d
		  LEFT JOIN "user" u ON d.owner = u.id
		  LEFT JOIN license l ON l.id = d.license_id
		  WHERE 1 = 1`

	type dbInfos struct {
		model.Database
		Author              string            `json:"author"`
		Description         map[string]string `json:"description"`
		Geographical_limit  map[string]string `json:"geographical_limit"`
		Bibliography        map[string]string `json:"bibliography"`
		Re_use		        map[string]string `json:"re_use"`
		Context_description map[string]string `json:"context_description"`
		Source_description  map[string]string `json:"source_description"`
		Source_relation     map[string]string `json:"source_relation"`
		Copyright           map[string]string `json:"copyright"`
		Subject             map[string]string `json:"subject"`
		Number_of_lines     int               `json:"number_of_lines"`
		Number_of_sites     int               `json:"number_of_sites"`
		License             string            `json:"license"`
		//Contexts            []string          `json:"context"`
		Countries []struct {
			Id   int    `json:"id"`
			Name string `json:"name"`
		} `json:"countries"`
		Continents []struct {
			Id   int    `json:"id"`
			Name string `json:"name"`
		} `json:"continents"`
		Authors []struct {
			Id       int    `json:"id"`
			Fullname string `json:"fullname"`
		} `json:"authors"`
	}

	// do not show unpublished
	q += " AND published = 't'"

	q += " ORDER BY d.updated_at DESC"

	// fmt.Println(q)

	databases := []dbInfos{}

	nstmt, err := tx.PrepareNamed(q)
	if err != nil {
		err = errors.New("rest.databases::DatabaseList : (infos) " + err.Error())
		log.Println(err)
		userSqlError(w, err)
		tx.Rollback()
		return
	}
	err = nstmt.Select(&databases, params)
	if err != nil {
		err = errors.New("rest.databases::DatabaseList : (infos) " + err.Error())
		log.Println(err)
		userSqlError(w, err)
		tx.Rollback()
		return
	}

	returnedDatabases := []dbInfos{}

	for _, database := range databases {

		// Authors
		astmt, err2 := tx.PrepareNamed("SELECT id, firstname || ' ' || lastname  as fullname FROM \"user\" u LEFT JOIN database__authors da ON u.id = da.user_id WHERE da.database_id = :id")
		if err2 != nil {
			err = errors.New("rest.databases::DatabaseList (authors) : " + err2.Error())
			log.Println(err)
			userSqlError(w, err)
			tx.Rollback()
			return
		}
		err = astmt.Select(&database.Authors, database)
		if err != nil {
			err = errors.New("rest.databases::DatabaseList (authors) : " + err.Error())
			log.Println(err)
			userSqlError(w, err)
			tx.Rollback()
			return
		}

		// Contexts
		/*
			cstmt, err3 := tx.PrepareNamed("SELECT context FROM database_context WHERE database_id = :id")
			if err3 != nil {
				err = errors.New("rest.databases::DatabaseList (contexts) : " + err3.Error())
				log.Println(err)
				userSqlError(w, err)
				tx.Rollback()
				return
			}
			err = cstmt.Select(&database.Contexts, database)
		*/

		// Countries
		if database.Geographical_extent == "country" {
			coustmt, err4 := tx.Preparex("SELECT ctr.name FROM country_tr ctr LEFT JOIN country c ON ctr.country_geonameid = c.geonameid LEFT JOIN database__country dc ON c.geonameid = dc.country_geonameid WHERE dc.database_id = $1 AND ctr.lang_isocode = $2")
			if err4 != nil {
				err = errors.New("rest.databases::DatabaseList (countries) : " + err4.Error())
				log.Println(err)
				userSqlError(w, err)
				tx.Rollback()
				return
			}
			err = coustmt.Select(&database.Countries, database.Id, params.Lang)
		}

		// Continents
		if database.Geographical_extent == "continent" {
			constmt, err5 := tx.Preparex("SELECT ctr.name FROM continent_tr ctr LEFT JOIN continent c ON ctr.continent_geonameid = c.geonameid LEFT JOIN database__continent dc ON c.geonameid = dc.continent_geonameid WHERE dc.database_id = $1 AND ctr.lang_isocode = $2")
			if err5 != nil {
				err = errors.New("rest.databases::DatabaseList : (continents) " + err5.Error())
				log.Println(err)
				userSqlError(w, err)
				tx.Rollback()
				return
			}
			err = constmt.Select(&database.Continents, database.Id, params.Lang)
		}

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
		database.Re_use = model.MapSqlTranslations(tr, "Lang_isocode", "Re_use")
		database.Context_description = model.MapSqlTranslations(tr, "Lang_isocode", "Context_description")
		database.Source_description = model.MapSqlTranslations(tr, "Lang_isocode", "Source_description")
		database.Source_relation = model.MapSqlTranslations(tr, "Lang_isocode", "Source_relation")
		database.Copyright = model.MapSqlTranslations(tr, "Lang_isocode", "Copyright")
		database.Subject = model.MapSqlTranslations(tr, "Lang_isocode", "Subject")
		returnedDatabases = append(returnedDatabases, database)
	}

	if err != nil {
		log.Println(err)
		tx.Rollback()
		userSqlError(w, err)
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		userSqlError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	//w.Header().Set("Content-Disposition", "attachment; filename=\"databases.csv\"")
	var csvW = csv.NewWriter(w)

	// LANG;NAME;AUTHORS;SUBJET;TYPE;LINES;SITES;SCALE;START_DATE;END_DATE;STATE;GEOGRAPHICAL_EXTENT;LICENSE;DESCRIPTION

	csvW.Write([]string{
		translate.TWeb(params.Lang, "DATABASE.EXPORT_NAME.T_HEADER"),
		translate.TWeb(params.Lang, "DATABASE.EXPORT_AUTHORS.T_HEADER"),
		translate.TWeb(params.Lang, "DATABASE.EXPORT_SUBJECT.T_HEADER"),
		translate.T(params.Lang, "DATABASE.EXPORT_DATE_UPDATED.T_HEADER"),
		translate.T(params.Lang, "DATABASE.EXPORT_DATE_CREATED.T_HEADER"),
		translate.TWeb(params.Lang, "DATABASE.EXPORT_LICENSE.T_HEADER"),
		translate.TWeb(params.Lang, "DATABASE.EXPORT_START_DATE.T_HEADER"),
		translate.TWeb(params.Lang, "DATABASE.EXPORT_END_DATE.T_HEADER"),
		translate.TWeb(params.Lang, "DATABASE.EXPORT_LINES.T_HEADER"),
		translate.TWeb(params.Lang, "DATABASE.EXPORT_SITES.T_HEADER"),
		translate.TWeb(params.Lang, "DATABASE.EXPORT_TYPE.T_HEADER"),
		translate.TWeb(params.Lang, "DATABASE.EXPORT_STATE.T_HEADER"),
		translate.TWeb(params.Lang, "DATABASE.EXPORT_GEOGRAPHICAL_EXTENT.T_HEADER"),
		translate.T(params.Lang, "DATABASE.EXPORT_GEOGRAPHICAL_EXTENT_GEOM.T_HEADER"),
		translate.TWeb(params.Lang, "DATABASE.EXPORT_SCALE.T_HEADER"),
		translate.TWeb(params.Lang, "DATABASE.EXPORT_LANG.T_HEADER"),
		translate.TWeb(params.Lang, "DATABASE.EXPORT_DESCRIPTION.T_HEADER"),
	})

	for _, line := range returnedDatabases {
		// element is the element from someSlice for where we are
		var authors []string
		for _, author := range line.Authors {
			authors = append(authors, author.Fullname)
		}

		csvW.Write([]string{
			line.Name,
			strings.Join(authors, " - "),
			translate.GetTranslated(line.Subject, params.Lang),
			line.Updated_at.Local().Format("2006-01-02 15:04"),
			line.Created_at.Local().Format("2006-01-02 15:04"),
			line.License,
			dateToDate(params.Lang, line.Start_date),
			dateToDate(params.Lang, line.End_date),
			strconv.Itoa(line.Number_of_lines),
			strconv.Itoa(line.Number_of_sites),
			translate.TWeb(params.Lang, "DATABASE"+"."+"TYPE_"+strings.ToUpper(strings.Replace(line.Type, "-", "", 1))+"."+"T"+"_TITLE"),
			translate.TWeb(params.Lang, "DATABASE"+"."+"STATE_"+strings.ToUpper(strings.Replace(line.State, "-", "", 1))+"."+"T"+"_TITLE"),
			translate.TWeb(params.Lang, "DATABASE"+"."+"GEOGRAPHICAL_EXTENT_"+strings.ToUpper(strings.Replace(line.Geographical_extent, "-", "", 1))+"."+"T"+"_TITLE"),
			line.Geographical_extent_geom,
			translate.TWeb(params.Lang, "DATABASE"+"."+"SCALE_RESOLUTION_"+strings.ToUpper(strings.Replace(line.Scale_resolution, "-", "", 1))+"."+"T"+"_TITLE"),
			line.Default_language,
			translate.GetTranslated(line.Description, params.Lang),
		})
	}

	csvW.Flush()
}

func dateToDate(lang string, date int) string {
	if date == -2147483648 {
		return translate.TWeb(lang, "MAIN.LABEL.T_UNDETERMINED")
	} else if date == 2147483647 {
		return translate.TWeb(lang, "MAIN.LABEL.T_UNDETERMINED")
	} else if date <= 0 {
		return strconv.Itoa(date - 1)
	} else {
		return strconv.Itoa(date)
	}
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

	dbInfos, err := d.GetFullInfos(tx, proute.Lang1.Isocode)

	if err != nil {
		log.Println("Error getting database infos", err)
		tx.Rollback()
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Error getting database infos", err)
		userSqlError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// w.Write([]byte(dbInfos))
	l, _ := json.Marshal(dbInfos)
	w.Write(l)
}

func DatabaseExportXML(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	params := proute.Params.(*DatabaseInfosParams)
	tx, err := db.DB.Beginx()
	if err != nil {
		log.Println("can't start transaction")
		userSqlError(w, err)
		return
	}

	buf := bytes.NewBufferString("")
	dbInfos, err := export.InteroperableExportXml(tx, buf, params.Id, proute.Lang1.Isocode)
	if err != nil {
		log.Println("Error creating Interoperable Export XML", err)
		userSqlError(w, err)
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Error getting database infos", err)
		userSqlError(w, err)
		return
	}

	t := time.Now()
	filename := fmt.Sprintf("ArkeoGIS-export-%d-%d-%d-%s-%s.xml",
							t.Year(), t.Month(), t.Day(),
							dbInfos.Name,
							dbInfos.GetAuthorsString())
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Disposition", "attachment; filename*=utf-8''"+url.PathEscape(filename))
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))

	buf.WriteTo(w)
}

func DatabaseExportCSVArkeogis(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	params := proute.Params.(*DatabaseExportInfosParams)
	tx, err := db.DB.Beginx()
	if err != nil {
		log.Println("can't start transaction")
		userSqlError(w, err)
		return
	}

	// Datatabase isocode

	_user, _ := proute.Session.Get("user")
	user := _user.(model.User)

	d := model.Database{}
	d.Id = params.Id
	dbInfos, err := d.GetFullInfos(tx, proute.Lang1.Isocode)

	//err = tx.Get(&dbName, "SELECT name FROM \"database\" WHERE id = $1", params.Id)

	if err != nil {
		log.Println("Unable to export database")
		userSqlError(w, err)
		tx.Rollback()
		return
	}

	var sites []int

	err = tx.Select(&sites, "SELECT id FROM site where database_id = $1", params.Id)
	if err != nil {
		log.Println("Unable to export database")
		userSqlError(w, err)
		tx.Rollback()
		return
	}

	csvContent, err := export.SitesAsCSV(&dbInfos, sites, user.First_lang_isocode, false, params.IncludeSiteId, params.IncludeInterop, tx)

	if err != nil {
		log.Println("Unable to export database")
		userSqlError(w, err)
		tx.Rollback()
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Unable to export database")
		userSqlError(w, err)
		return
	}
	t := time.Now()
	/*
	filename := dbName + "-" + fmt.Sprintf("%d-%d-%d %d:%d:%d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second()) + ".csv"
	*/	
	filename := fmt.Sprintf("ArkeoGIS-export-%d-%d-%d-%s-%s.csv",
							t.Year(), t.Month(), t.Day(),
							dbInfos.Name,
							dbInfos.GetAuthorsString())
							
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename*=utf-8''"+url.PathEscape(filename))
	w.Write([]byte(csvContent))
}

func DatabaseExportZIPOmeka(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	params := proute.Params.(*DatabaseExportOmekaParams)
	tx, err := db.DB.Beginx()
	if err != nil {
		log.Println("can't start transaction")
		userSqlError(w, err)
		return
	}

	// Datatabase isocode

	_user, _ := proute.Session.Get("user")
	user := _user.(model.User)

	dbName := ""

	err = tx.Get(&dbName, "SELECT name FROM \"database\" WHERE id = $1", params.Id)

	if err != nil {
		log.Println("Unable to export database", err)
		userSqlError(w, err)
		tx.Rollback()
		return
	}

	csvSitesContent, csvCaracsContent, err := export.SitesAsOmeka(params.Id, params.ChronologyId, user.First_lang_isocode, tx)

	if err != nil {
		log.Println("Unable to export database", err)
		userSqlError(w, err)
		tx.Rollback()
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Unable to export database", err)
		userSqlError(w, err)
		return
	}
	t := time.Now()
	filename := dbName + "-" + fmt.Sprintf("%.4d-%.2d-%.2d_%.2d-%.2d-%.2d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	filename = strings.ReplaceAll(filename, "\"", "_")

	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	// Create a new zip archive.
	wZip := zip.NewWriter(buf)

	// Add some files to the archive.
	var files = []struct {
		Name, Body string
	}{
		{"akg2omk-sites_" + filename + ".csv", csvSitesContent},
		{"akg2omk-caracterisations_" + filename + ".csv", csvCaracsContent},
	}

	for _, file := range files {
		f, err := wZip.Create(file.Name)
		if err != nil {
			log.Fatal(err)
		}
		_, err = f.Write([]byte(file.Body))
		if err != nil {
			log.Fatal(err)
		}
	}

	err = wZip.Flush()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("buf size : ", buf.Len())

	// Make sure to check the error on Close.
	err = wZip.Close()
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+".zip\"")
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))

	/*
		hackb := make([]byte, buf.Len())
		var readed int
		readed, err = buf.Read(hackb)
		log.Println("readed: ", readed, err)
		w.Write(hackb)
		err = ioutil.WriteFile("/tmp/dat1.zip", hackb, 0644)
	*/
	buf.WriteTo(w)
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
