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
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"

	db "github.com/croll/arkeogis-server/db"

	"github.com/croll/arkeogis-server/model"
	routes "github.com/croll/arkeogis-server/webserver/routes"
)

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:        "/api/wmLayer",
			Description: "Save wm(t)s layer",
			Func:        SaveWmLayer,
			Permissions: []string{},
			Json:        reflect.TypeOf(SaveWmLayerParams{}),
			Method:      "POST",
		},
		&routes.Route{
			Path:        "/api/shpLayer",
			Description: "Save shapefile layer",
			Func:        SaveShpLayer,
			Permissions: []string{},
			Json:        reflect.TypeOf(SaveShpParams{}),
			Method:      "POST",
		},
		&routes.Route{
			Path:        "/api/layers",
			Description: "Get wms, wmts and shapefiles",
			Func:        GetLayers,
			Permissions: []string{},
			Params:      reflect.TypeOf(GetLayersParams{}),
			Method:      "GET",
		},
	}
	routes.RegisterMultiple(Routes)
}

type SaveShpParams struct {
	Id                       int
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

	filehash := fmt.Sprintf("%x", md5.Sum([]byte(params.File.Name)))
	filename := params.File.Name
	filepath := "./uploaded/shp/" + filehash + "_" + filename

	outfile, err := os.Create(filepath)
	if err != nil {
		http.Error(w, "Error saving file: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Save the file on filesystem
	_, err = io.WriteString(outfile, string(params.File.Content))
	if err != nil {
		http.Error(w, "Error saving file: "+err.Error(), http.StatusBadRequest)
		return
	}

	var layer = &model.Shapefile{
		Creator_user_id:          params.Authors[0],
		Filename:                 params.Filename,
		Md5sum:                   filehash,
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

	if params.Id > 0 {
		layer.Id = params.Id
		err = layer.Update(tx)
		if err != nil {
			userSqlError(w, err)
			return
		}
		err = layer.DeleteAuthors(tx)
		if err != nil {
			userSqlError(w, err)
			return
		}
	} else {
		err = layer.Create(tx)
		if err != nil {
			userSqlError(w, err)
			return
		}
	}

	err = layer.SetAuthors(tx, params.Authors)
	if err != nil {
		log.Println("Error setting database authors: ", err)
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
		userSqlError(w, err)
		return
	}

	err = layer.SetTranslations(tx, "name", params.Name)
	if err != nil {
		log.Println("Error setting name: ", err)
		userSqlError(w, err)
		return
	}

	err = layer.SetTranslations(tx, "description", params.Description)
	if err != nil {
		log.Println("Error setting description: ", err)
		userSqlError(w, err)
		return
	}

	if err != nil {
		userSqlError(w, err)
		return
	}

	tx.Commit()

}

type SaveWmLayerParams struct {
	Id                       int
	Author_Id                int
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
		Creator_user_id:          params.Author_Id,
		Type:                     params.Type,
		Identifier:               params.Identifier,
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
	}

	if params.Id > 0 {
		layer.Id = params.Id
		err = layer.Update(tx)
	} else {
		err = layer.Create(tx)
	}

	if err != nil {
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
		userSqlError(w, err)
		return
	}

	err = layer.SetTranslations(tx, "name", params.Name)
	if err != nil {
		log.Println("Error setting name: ", err)
		userSqlError(w, err)
		return
	}

	err = layer.SetTranslations(tx, "description", params.Description)
	if err != nil {
		log.Println("Error setting description: ", err)
		userSqlError(w, err)
		return
	}

	if err != nil {
		userSqlError(w, err)
		return
	}

	tx.Commit()
}

type GetLayersParams struct {
	Type      string
	Published bool
	Author    int
	Iso_code  string
}

type LayerInfos struct {
	Id                       int       `json:"id"`
	Geographical_extent_geom string    `json:"geographical_extent_geom"`
	Creator_user_id          int       `json:"creator_user_id"`
	Published                bool      `json:"published"`
	Description              string    `json:"description"`
	Description_en           string    `json:"description_en"`
	Created_at               time.Time `json:"created_at"`
	Author                   string    `json:"author"`
	Type                     string    `json:"type"`
	Start_date               int       `json:"start_date"`
	End_date                 int       `json:"end_date"`
	Min_scale                int       `json:"min_scale"`
	Max_scale                int       `json:"max_scale"`
}

// GetLayers returns layers list
func GetLayers(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Params.(*GetLayersParams)

	var err error

	// get the user
	_user, ok := proute.Session.Get("user")
	if !ok {
		log.Println("GetLayers: can't get user in session.", _user)
		return
	}
	user, ok := _user.(model.User)
	if !ok {
		log.Println("GetLayers: can't cast user.", _user)
		return
	}

	params.Iso_code = user.First_lang_isocode

	// user.First_lang_isocode

	var result = []LayerInfos{}

	if params.Type == "" || params.Type == "shp" {
		infos := &[]LayerInfos{}
		infos, err = getShpLayers(params)
		if err != nil {
			http.Error(w, "Error getting shp layers list: "+err.Error(), http.StatusBadRequest)
			return
		}
		result = append(result, *infos...)
	}

	if params.Type == "" || params.Type != "shp" {
		infos := &[]LayerInfos{}
		infos, err = getWmLayers(params)
		if err != nil {
			http.Error(w, "Error getting shp layers list: "+err.Error(), http.StatusBadRequest)
			return
		}
		result = append(result, *infos...)
	}

	l, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json")
	w.Write(l)
	return

}

func getShpLayers(params *GetLayersParams) (result *[]LayerInfos, err error) {
	fmt.Println(params)
	q := "SELECT m.id, m.start_date, m.end_date, ST_AsGeoJSON(m.geographical_extent_geom) as geographical_extent_geom, m.published, m.created_at, m.creator_user_id, u.firstname || ' ' || u.lastname as author, 'shp' AS type, t.description, (SELECT description FROM shapefile_tr WHERE shapefile_id = m.id AND lang_isocode = :iso_code) AS description_en FROM shapefile m LEFT JOIN \"user\" u ON m.creator_user_id = u.id LEFT JOIN shapefile_tr t ON m.id = t.shapefile_id WHERE t.lang_isocode = :iso_code"

	result = &[]LayerInfos{}

	if params.Author > 0 {
		q += " AND u.id = :author"
	}

	if params.Published {
		q += " AND u.published = :published"
	}

	nstmt, err := db.DB.PrepareNamed(q)
	if err != nil {
		return
	}
	err = nstmt.Select(result, params)
	return
}

func getWmLayers(params *GetLayersParams) (result *[]LayerInfos, err error) {

	result = &[]LayerInfos{}

	q := "SELECT m.id, m.type, m.start_date, m.end_date, m.min_scale, m.max_scale, ST_AsGeoJSON(m.geographical_extent_geom) as geographical_extent_geom, m.published, m.created_at, m.creator_user_id, u.firstname || ' ' || u.lastname as author, (SELECT description FROM map_layer_tr WHERE map_layer_id = m.id AND lang_isocode = :iso_code) AS description_en  FROM map_layer m LEFT JOIN \"user\" u ON m.creator_user_id = u.id LEFT JOIN map_layer_tr t ON m.id = t.map_layer_id WHERE t.lang_isocode = :iso_code"

	if params.Author > 0 {
		q += " AND u.id = :author"
	}

	if params.Published {
		q += " AND u.published = :published"
	}

	if params.Type != "" {
		q += " AND u.type= :type"
	}

	nstmt, err := db.DB.PrepareNamed(q)
	if err != nil {
		return
	}
	err = nstmt.Select(result, params)
	return
}
