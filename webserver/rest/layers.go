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
	"crypto/md5"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	db "github.com/croll/arkeogis-server/db"
	"github.com/croll/arkeogis-server/translate"
	"github.com/jmoiron/sqlx"

	"github.com/croll/arkeogis-server/model"
	routes "github.com/croll/arkeogis-server/webserver/routes"
)

type LayersParams struct {
	Ids          []int `json:"ids"`
	Type         string
	Published    bool
	Author       int
	Iso_code     string
	Bounding_box string
	Start_date   int  `json:"start_date"`
	End_date     int  `json:"end_date"`
	Check_dates  bool `json:"check_dates"`
}

type LayersExportParams struct {
	LayersParams
	Lang string `json:"lang" min:"0" max:"2"`
}

type LayerParams struct {
	Id      int    `json:"id" min:"1"`
	Type    string `json:"type" min:"3" max:"3"`
	GeoJson bool   `json:"geojson"`
}

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:        "/api/wmLayer",
			Description: "Save wm(t)s layer",
			Func:        SaveWmLayer,
			Permissions: []string{"manage all wms/wmts"},
			Json:        reflect.TypeOf(SaveWmLayerParams{}),
			Method:      "POST",
		},
		&routes.Route{
			Path:        "/api/shpLayer",
			Description: "Save shapefile layer",
			Func:        SaveShpLayer,
			Permissions: []string{"manage all wms/wmts"},
			Json:        reflect.TypeOf(SaveShpParams{}),
			Method:      "POST",
		},
		&routes.Route{
			Path:        "/api/layers",
			Description: "Get wms, wmts and shapefiles",
			Func:        GetLayers,
			Permissions: []string{
				"request map",
			},
			Params: reflect.TypeOf(LayersParams{}),
			Method: "GET",
		},
		&routes.Route{
			Path:        "/api/layers/export",
			Description: "Get wms, wmts and shapefiles that are published, as csv",
			Func:        GetExportLayers,
			Permissions: []string{},
			Params:      reflect.TypeOf(LayersExportParams{}),
			Method:      "GET",
		},
		&routes.Route{
			Path:        "/api/layer",
			Description: "Get layer informations",
			Func:        GetLayer,
			Permissions: []string{
				"request map",
			},
			Params: reflect.TypeOf(LayerParams{}),
			Method: "GET",
		},
		&routes.Route{
			Path:        "/api/layer/delete",
			Description: "Delete layer",
			Func:        DeleteLayer,
			Permissions: []string{"manage all wms/wmts"},
			Json:        reflect.TypeOf(LayerParams{}),
			Method:      "POST",
		},
		&routes.Route{
			Path:        "/api/layer/{id:[0-9]+}/geojson",
			Description: "Get SHP geojson",
			Func:        GetShpGeojson,
			Permissions: []string{
				"request map",
			},
			Params: reflect.TypeOf(struct{ Id int }{}),
			Method: "GET",
		},
	}
	routes.RegisterMultiple(Routes)
}

type SaveShpParams struct {
	Id                       int
	Author_Id                int
	Authors                  []int
	Filename                 string
	Geojson                  string
	Geojson_with_data        string
	Start_date               int
	End_date                 int
	Geographical_extent_geom string
	Published                bool
	File                     *routes.File
	License                  string
	License_id               int
	Declared_creation_date   time.Time
	Attribution              string
	Copyright                string
	Name                     []struct {
		Lang_Isocode string
		Text         string
	}
	Description []struct {
		Lang_Isocode string
		Text         string
	}
}

// SaveShpLayer saves shp file on filesystem and datas into database
func SaveShpLayer(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Json.(*SaveShpParams)

	// get the user
	_user, ok := proute.Session.Get("user")
	if !ok {
		log.Println("SaveShpLayer: can't get user in session.", _user)
		return
	}
	user, ok := _user.(model.User)
	if !ok {
		log.Println("Save: can't cast user.", _user)
		return
	}

	//user.First_lang_isocode

	tx, err := db.DB.Beginx()
	if err != nil {
		http.Error(w, "Error saving shapefile informations: "+err.Error(), http.StatusBadRequest)
		return
	}

	var filehash string

	if params.File != nil {

		filehash = fmt.Sprintf("%x", md5.Sum([]byte(params.File.Name)))
		filepath := "./uploaded/shp/" + filehash + "_" + params.File.Name

		outfile, err := os.Create(filepath)
		if err != nil {
			_ = tx.Rollback()
			http.Error(w, "Error saving file: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Save the file on filesystem
		_, err = io.WriteString(outfile, string(params.File.Content))
		if err != nil {
			_ = tx.Rollback()
			http.Error(w, "Error saving file: "+err.Error(), http.StatusBadRequest)
			return
		}

	}

	var layer = &model.Shapefile{
		Creator_user_id:          params.Authors[0],
		Filename:                 params.Filename,
		Geojson:                  params.Geojson,
		Geojson_with_data:        params.Geojson_with_data,
		Start_date:               params.Start_date,
		End_date:                 params.End_date,
		Geographical_extent_geom: params.Geographical_extent_geom,
		Published:                params.Published,
		License:                  params.License,
		License_id:               params.License_id,
		Declared_creation_date:   params.Declared_creation_date,
	}

	if params.File != nil {
		layer.Md5sum = filehash
	}

	if params.Id > 0 {
		layer.Id = params.Id
		err = layer.Update(tx)
		if err != nil {
			log.Println(err)
			_ = tx.Rollback()
			userSqlError(w, err)
			return
		}
		err = layer.DeleteAuthors(tx)
		if err != nil {
			log.Println(err)
			_ = tx.Rollback()
			userSqlError(w, err)
			return
		}
	} else {
		err = layer.Create(tx)
		if err != nil {
			log.Println(err)
			_ = tx.Rollback()
			userSqlError(w, err)
			return
		}
	}

	err = layer.SetAuthors(tx, params.Authors)
	if err != nil {
		log.Println("Error setting database authors: ", err)
		_ = tx.Rollback()
		userSqlError(w, err)
		return
	}

	// For now attribution is not translatable but store it in database_tr anyway
	var attribution = []struct {
		Lang_Isocode string
		Text         string
	}{
		{user.First_lang_isocode, params.Attribution},
	}
	err = layer.SetTranslations(tx, "attribution", attribution)
	if err != nil {
		log.Println("Error setting attribution: ", err)
		_ = tx.Rollback()
		userSqlError(w, err)
		return
	}

	// For now attribution is not translatable but store it in database_tr anyway
	var copyright = []struct {
		Lang_Isocode string
		Text         string
	}{
		{user.First_lang_isocode, params.Copyright},
	}
	err = layer.SetTranslations(tx, "copyright", copyright)
	if err != nil {
		log.Println("Error setting copyright: ", err)
		_ = tx.Rollback()
		userSqlError(w, err)
		return
	}

	err = layer.SetTranslations(tx, "name", params.Name)
	if err != nil {
		log.Println("Error setting name: ", err)
		_ = tx.Rollback()
		userSqlError(w, err)
		return
	}

	err = layer.SetTranslations(tx, "description", params.Description)
	if err != nil {
		log.Println("Error setting description: ", err)
		_ = tx.Rollback()
		userSqlError(w, err)
		return
	}

	if err != nil {
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Error commiting changes: ", err)
		_ = tx.Rollback()
		userSqlError(w, err)
		return
	}

}

type SaveWmLayerParams struct {
	Id                       int
	Authors                  []int
	Type                     string
	Url                      string
	Identifier               string
	Min_scale                int
	Max_scale                int
	Start_date               int
	End_date                 int
	Image_format             string
	Geographical_extent_geom string
	Published                bool
	License                  string
	License_id               int
	Attribution              string
	Copyright                string
	Tile_matrix_set          string
	Tile_matrix_string       string
	Use_proxy                bool
	Max_usage_date           time.Time
	Name                     []struct {
		Lang_Isocode string
		Text         string
	}
	Description []struct {
		Lang_Isocode string
		Text         string
	}
}

// SaveWmLayer save wm(t)s layer informations into database
func SaveWmLayer(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Json.(*SaveWmLayerParams)

	// get the user
	_user, ok := proute.Session.Get("user")
	if !ok {
		log.Println("SaveShpLayer: can't get user in session.", _user)
		return
	}
	user, ok := _user.(model.User)
	if !ok {
		log.Println("Save: can't cast user.", _user)
		return
	}

	tx, err := db.DB.Beginx()
	if err != nil {
		http.Error(w, "Error saving wm(t)s informations: "+err.Error(), http.StatusBadRequest)
		return
	}

	var layer = &model.Map_layer{
		Creator_user_id:          params.Authors[0],
		Type:                     params.Type,
		Identifier:               params.Identifier,
		Url:                      params.Url,
		Min_scale:                params.Min_scale,
		Max_scale:                params.Max_scale,
		Start_date:               params.Start_date,
		End_date:                 params.End_date,
		Image_format:             params.Image_format,
		Geographical_extent_geom: params.Geographical_extent_geom,
		Published:                params.Published,
		License:                  params.License,
		License_id:               params.License_id,
		Max_usage_date:           params.Max_usage_date,
		Tile_matrix_set:          params.Tile_matrix_set,
		Tile_matrix_string:       params.Tile_matrix_string,
		Use_proxy:                params.Use_proxy,
	}

	if params.Id > 0 {
		layer.Id = params.Id
		err = layer.Update(tx)
		if err != nil {
			log.Println(err)
			_ = tx.Rollback()
			userSqlError(w, err)
			return
		}
		err = layer.DeleteAuthors(tx)
		if err != nil {
			log.Println(err)
			_ = tx.Rollback()
			userSqlError(w, err)
			return
		}
	} else {
		err = layer.Create(tx)
		if err != nil {
			log.Println(err)
			_ = tx.Rollback()
			userSqlError(w, err)
			return
		}
	}

	err = layer.SetAuthors(tx, params.Authors)
	if err != nil {
		log.Println("Error setting wm(t)s layer authors: ", err)
		_ = tx.Rollback()
		userSqlError(w, err)
		return
	}

	// For now attribution is not translatable but store it in database_tr anyway
	var attribution = []struct {
		Lang_Isocode string
		Text         string
	}{
		{user.First_lang_isocode, params.Attribution},
	}
	err = layer.SetTranslations(tx, "attribution", attribution)
	if err != nil {
		log.Println("Error setting attribution: ", err)
		_ = tx.Rollback()
		userSqlError(w, err)
		return
	}

	// For now attribution is not translatable but store it in database_tr anyway
	var copyright = []struct {
		Lang_Isocode string
		Text         string
	}{
		{user.First_lang_isocode, params.Copyright},
	}
	err = layer.SetTranslations(tx, "copyright", copyright)
	if err != nil {
		log.Println("Error setting copyright: ", err)
		_ = tx.Rollback()
		userSqlError(w, err)
		return
	}

	err = layer.SetTranslations(tx, "name", params.Name)
	if err != nil {
		log.Println("Error setting name: ", err)
		_ = tx.Rollback()
		userSqlError(w, err)
		return
	}

	err = layer.SetTranslations(tx, "description", params.Description)
	if err != nil {
		log.Println("Error setting description: ", err)
		_ = tx.Rollback()
		userSqlError(w, err)
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		userSqlError(w, err)
		return
	}
}

func getLayers(w http.ResponseWriter, r *http.Request, proute routes.Proute, params *LayersParams) []*model.LayerFullInfos {
	var err error

	// get the user
	_user, ok := proute.Session.Get("user")
	if !ok {
		log.Println("GetLayers: can't get user in session.", _user)
		userSqlError(w, err)
		return nil
	}
	user, ok := _user.(model.User)
	if !ok {
		log.Println("GetLayers: can't cast user.", _user)
		userSqlError(w, err)
		return nil
	}

	params.Iso_code = user.First_lang_isocode

	// transaction begin...
	tx, err := db.DB.Beginx()
	if err != nil {
		userSqlError(w, err)
		return nil
	}

	viewUnpublished, err := user.HavePermissions(tx, "manage all databases")
	if err != nil {
		userSqlError(w, err)
		tx.Rollback()
		return nil
	}

	var result = []*model.LayerFullInfos{}

	if params.Type == "" || params.Type == "shp" {
		infos := []*model.LayerFullInfos{}
		infos, err = getShpLayers(params, viewUnpublished, tx)
		if err != nil {
			log.Println(err)
			userSqlError(w, err)
			return nil
		}
		result = append(result, infos...)
	}

	if params.Type == "" || params.Type != "shp" {
		infos := []*model.LayerFullInfos{}
		infos, err = getWmLayers(params, viewUnpublished, tx)
		if err != nil {
			log.Println(err)
			userSqlError(w, err)
			return nil
		}
		result = append(result, infos...)
	}

	// commit...
	err = tx.Commit()
	if err != nil {
		log.Println("commit failed")
		_ = tx.Rollback()
		return nil
	}

	return result
}

// GetLayers returns layers list
func GetLayers(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Params.(*LayersParams)

	var result = getLayers(w, r, proute, params)

	l, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json")
	w.Write(l)
}

// GetLayers returns layers list
func GetExportLayers(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	paramsExport := proute.Params.(*LayersExportParams)
	params := paramsExport.LayersParams

	params.Published = true // force published

	var result = getLayers(w, r, proute, &params)

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	var csvW = csv.NewWriter(w)

	csvW.Write([]string{
		translate.TWeb(paramsExport.Lang, "MAP.EXPORT_NAME.T_HEADER"),
		translate.TWeb(paramsExport.Lang, "MAP.EXPORT_LICENSE.T_HEADER"),
		translate.TWeb(paramsExport.Lang, "MAP.EXPORT_START_DATE.T_HEADER"),
		translate.TWeb(paramsExport.Lang, "MAP.EXPORT_END_DATE.T_HEADER"),
		translate.TWeb(paramsExport.Lang, "MAP.EXPORT_SCALE.T_HEADER"),
		translate.TWeb(paramsExport.Lang, "MAP.EXPORT_TYPE.T_HEADER"),
		translate.TWeb(paramsExport.Lang, "MAP.EXPORT_DESCRIPTION.T_HEADER"),
	})

	for _, line := range result {
		csvW.Write([]string{
			translate.GetTranslated(line.Name, paramsExport.Lang),
			translate.GetTranslated(line.Copyright, paramsExport.Lang),
			dateToDate(paramsExport.Lang, line.Start_date),
			dateToDate(paramsExport.Lang, line.End_date),
			strconv.Itoa(line.Min_scale) + "/" + strconv.Itoa(line.Max_scale),
			translate.TWeb(paramsExport.Lang, "DATABASE"+"."+"TYPE_"+strings.ToUpper(strings.Replace(line.Type, "-", "", 1))+"."+"T"+"_TITLE"),
			translate.GetTranslated(line.Description, paramsExport.Lang),
		})
	}

	csvW.Flush()

	fmt.Println("result : ", result)

}

func getShpLayers(params *LayersParams, viewUnpublished bool, tx *sqlx.Tx) (layers []*model.LayerFullInfos, err error) {

	layers = []*model.LayerFullInfos{}

	q := "SELECT m.id, m.start_date, m.end_date, ST_AsGeoJSON(m.geographical_extent_geom) as geographical_extent_geom, m.published, m.created_at, m.creator_user_id, u.firstname || ' ' || u.lastname as author, 'shp' AS type FROM shapefile m LEFT JOIN \"user\" u ON m.creator_user_id = u.id WHERE m.id > 0"

	if params.Author > 0 {
		q += " AND u.id = :author"
	}

	if params.Published || !viewUnpublished {
		q += " AND m.published = 't'"
	}

	if params.Bounding_box != "" {
		q += " AND (ST_Contains(ST_GeomFromGeoJSON(:bounding_box), m.geographical_extent_geom::::geometry) OR ST_Contains(m.geographical_extent_geom::::geometry, ST_GeomFromGeoJSON(:bounding_box)) OR ST_Overlaps(ST_GeomFromGeoJSON(:bounding_box), m.geographical_extent_geom::::geometry))"
	}

	if params.Check_dates {
		q += " AND m.start_date >= :start_date AND m.end_date <= :end_date"
	}

	in := model.IntJoin(params.Ids, false)

	if in != "" {
		q += " AND m.id IN (" + in + ")"
	}

	nstmt, err := tx.PrepareNamed(q)
	if err != nil {
		_ = tx.Rollback()
		log.Println(err)
		return
	}
	err = nstmt.Select(&layers, params)
	if err != nil {
		_ = tx.Rollback()
		log.Println(err)
		return
	}

	for _, layer := range layers {
		tr := []model.Shapefile_tr{}
		err = tx.Select(&tr, "SELECT * FROM shapefile_tr WHERE shapefile_id = "+strconv.Itoa(layer.Id))
		if err != nil {
			_ = tx.Rollback()
			return
		}
		layer.Uniq_code = layer.Type + strconv.Itoa(layer.Id)
		layer.Name = model.MapSqlTranslations(tr, "Lang_isocode", "Name")
		layer.Attribution = model.MapSqlTranslations(tr, "Lang_isocode", "Attribution")
		layer.Copyright = model.MapSqlTranslations(tr, "Lang_isocode", "Copyright")
		layer.Description = model.MapSqlTranslations(tr, "Lang_isocode", "Description")
	}
	return
}

func getWmLayers(params *LayersParams, viewUnpublished bool, tx *sqlx.Tx) (layers []*model.LayerFullInfos, err error) {

	layers = []*model.LayerFullInfos{}

	q := "SELECT m.id, m.type, m.start_date, m.end_date, m.min_scale, m.max_scale, m.tile_matrix_set, m.tile_matrix_string, m.use_proxy, ST_AsGeoJSON(m.geographical_extent_geom) as geographical_extent_geom, m.published, m.created_at, m.creator_user_id, u.firstname || ' ' || u.lastname as author FROM map_layer m LEFT JOIN \"user\" u ON m.creator_user_id = u.id WHERE m.id > 0"

	if params.Author > 0 {
		q += " AND u.id = :author"
	}

	if params.Published || !viewUnpublished {
		q += " AND m.published = 't'"
	}

	if params.Type != "" {
		q += " AND m.type= :type"
	}

	if params.Bounding_box != "" {
		q += " AND (ST_Contains(ST_GeomFromGeoJSON(:bounding_box), m.geographical_extent_geom::::geometry) OR ST_Contains(m.geographical_extent_geom::::geometry, ST_GeomFromGeoJSON(:bounding_box)) OR ST_Overlaps(ST_GeomFromGeoJSON(:bounding_box), m.geographical_extent_geom::::geometry))"
	}

	if params.Check_dates {
		q += " AND m.start_date >= :start_date AND m.end_date <= :end_date"
	}

	in := model.IntJoin(params.Ids, false)

	if in != "" {
		q += " AND m.id IN (" + in + ")"
	}

	nstmt, err := tx.PrepareNamed(q)
	if err != nil {
		log.Println(err)
		_ = tx.Rollback()
		return
	}
	err = nstmt.Select(&layers, params)

	for _, layer := range layers {

		tr := []model.Map_layer_tr{}
		err = tx.Select(&tr, "SELECT * FROM map_layer_tr WHERE map_layer_id = "+strconv.Itoa(layer.Id))
		if err != nil {
			log.Println(err)
			_ = tx.Rollback()
			return
		}
		layer.Uniq_code = layer.Type + strconv.Itoa(layer.Id)
		layer.Name = model.MapSqlTranslations(tr, "Lang_isocode", "Name")
		layer.Attribution = model.MapSqlTranslations(tr, "Lang_isocode", "Attribution")
		layer.Copyright = model.MapSqlTranslations(tr, "Lang_isocode", "Copyright")
		layer.Description = model.MapSqlTranslations(tr, "Lang_isocode", "Description")
	}

	return
}

// GetLayer returns all infos about a layer
func GetLayer(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Params.(*LayerParams)

	var q []string
	var query string
	var jsonString string

	if params.Type == "shp" {
		q = make([]string, 3)
		if params.GeoJson {
			q[0] = db.AsJSON("SELECT s.id, s.creator_user_id, 'shp' AS type, s.filename, s.md5sum,  s.geojson, s.start_date, s.end_date, ST_AsGeoJSON(s.geographical_extent_geom) as geographical_extent_geom, s.published, s.license_id, s.license, s.declared_creation_date, s.created_at, s.updated_at, u.firstname || ' ' || u.lastname as author FROM shapefile s LEFT JOIN \"user\" u ON s.creator_user_id = u.id WHERE s.id = sl.id", false, "infos", true)
		} else {
			q[0] = db.AsJSON("SELECT s.id, s.creator_user_id, 'shp' AS type, s.filename, s.md5sum, s.start_date, s.end_date, ST_AsGeoJSON(s.geographical_extent_geom) as geographical_extent_geom, s.published, s.license_id, s.license, s.declared_creation_date, s.created_at, s.updated_at, u.firstname || ' ' || u.lastname as author FROM shapefile s LEFT JOIN \"user\" u ON s.creator_user_id = u.id WHERE s.id = sl.id", false, "infos", true)
		}
		q[1] = db.AsJSON("SELECT u.id, u.firstname, u.lastname FROM \"user\" u LEFT JOIN shapefile__authors sa ON u.id = sa.user_id WHERE sa.shapefile_id = sl.id", true, "authors", true)
		q[2] = model.GetQueryTranslationsAsJSONObject("shapefile_tr", "shapefile_id = sl.id", "translations", true, "name", "attribution", "copyright", "description")
		query = db.JSONQueryBuilder(q, "shapefile sl", "sl.id = "+strconv.Itoa(params.Id))
	} else {
		q = make([]string, 3)
		q[0] = db.AsJSON("SELECT m.id, m.creator_user_id, m.type, m.url, m.identifier, m.min_scale, m.max_scale, m.start_date, m.end_date, m.image_format, ST_AsGeoJSON(m.geographical_extent_geom) as geographical_extent_geom, m.published, m.license_id, m.license, m.tile_matrix_set, m.tile_matrix_string, m.use_proxy, m.max_usage_date, m.created_at, m.updated_at, u.firstname || ' ' || u.lastname as author FROM map_layer m LEFT JOIN \"user\" u ON m.creator_user_id = u.id WHERE m.id = ml.id", false, "infos", true)
		q[1] = db.AsJSON("SELECT u.id, u.firstname, u.lastname FROM \"user\" u LEFT JOIN map_layer__authors ma ON u.id = ma.user_id WHERE ma.map_layer_id = ml.id", true, "authors", true)
		q[2] = model.GetQueryTranslationsAsJSONObject("map_layer_tr", "map_layer_id = ml.id", "translations", true, "name", "attribution", "copyright", "description")
		query = db.JSONQueryBuilder(q, "map_layer ml", "ml.id = "+strconv.Itoa(params.Id))
	}

	err := db.DB.Get(&jsonString, query)
	if err != nil {
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonString))
	return

}

func DeleteLayer(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Json.(*LayerParams)

	tx, err := db.DB.Beginx()
	if err != nil {
		log.Println("can't start transaction")
		userSqlError(w, err)
		return
	}

	fmt.Println("TYPE DE LAYER A DEL", params.Type)
	if params.Type == "shp" {
		_, err = tx.Exec("DELETE FROM project__shapefile WHERE shapefile_id = $1", params.Id)
		if err != nil {
			log.Println("Unable to delete project map layer link:", err)
		}
		_, err = tx.Exec("DELETE FROM shapefile_tr WHERE shapefile_id = $1", params.Id)
		if err != nil {
			log.Println("Unable to delete layer translation:", err)
		}
		_, err = tx.Exec("DELETE FROM shapefile__authors WHERE shapefile_id = $1", params.Id)
		if err != nil {
			log.Println("Unable to delete layer author:", err)
		}
		_, err = tx.Exec("DELETE FROM shapefile WHERE id = $1", params.Id)
	} else {
		_, err = tx.Exec("DELETE FROM project__map_layer WHERE map_layer_id = $1", params.Id)
		if err != nil {
			log.Println("Unable to delete project map layer link:", err)
		}
		_, err = tx.Exec("DELETE FROM map_layer_tr WHERE map_layer_id = $1", params.Id)
		if err != nil {
			log.Println("Unable to delete layer translation:", err)
		}
		_, err = tx.Exec("DELETE FROM map_layer__authors WHERE map_layer_id = $1", params.Id)
		if err != nil {
			log.Println("Unable to delete layer author:", err)
		}
		_, err = tx.Exec("DELETE FROM map_layer WHERE id = $1", params.Id)
	}

	if err != nil {
		log.Println("Unable to delete layer:", err)
		_ = tx.Rollback()
		userSqlError(w, err)
		return
	}

	err = tx.Commit()

	if err != nil {
		log.Println("Unable to delete layer:", err)
		userSqlError(w, err)
		return
	}
}

func GetShpGeojson(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Params.(*struct{ Id int })

	var geoJSON string

	tx, err := db.DB.Beginx()
	if err != nil {
		log.Println("can't start transaction")
		userSqlError(w, err)
		return
	}

	err = tx.Get(&geoJSON, "SELECT geojson FROM shapefile WHERE id = $1", params.Id)
	if err != nil {
		log.Println("can't get geojson")
		_ = tx.Rollback()
		userSqlError(w, err)
		return
	}

	err = tx.Commit()

	if err != nil {
		log.Println("GetShpGeojson commit failed:", err)
		userSqlError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(geoJSON))
	return

}
