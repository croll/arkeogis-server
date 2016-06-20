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
	"fmt"
	"io"
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
	}
	routes.RegisterMultiple(Routes)
}

type SaveWmLayerParams struct {
	Id                       int
	Type                     string
	Url                      string
	Identifier               string
	Min_scale                int
	Max_scale                int
	Start_date               time.Time
	End_date                 time.Time
	Image_format             string
	Geographical_extent_geom string
	Published                bool
	License                  string
	License_id               int
	Max_usage_date           time.Time
	Translations             []struct {
		Name []struct {
			Lang_Isocode string
			Text         string
		}
		Attribution []struct {
			Lang_Isocode string
			Text         string
		}
		Copyright []struct {
			Lang_Isocode string
			Text         string
		}
		Description []struct {
			Lang_Isocode string
			Text         string
		}
	}
}

type SaveShpParams struct {
	Id                       int
	Authors                  []int
	Filename                 string
	Geojson                  string
	Identifier               string
	Start_date               int
	End_date                 int
	Geographical_extent_geom string
	Published                bool
	File                     *routes.File
	License                  string
	License_id               int
	Declared_creation_date   time.Time
	Name                     []struct {
		Lang_Isocode string
		Text         string
	}
	Attribution []struct {
		Lang_Isocode string
		Text         string
	}
	Copyright []struct {
		Lang_Isocode string
		Text         string
	}
	Description []struct {
		Lang_Isocode string
		Text         string
	}
}

// ShpToGeoJSON get shapefile and convert it to geojson
func SaveShpLayer(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Json.(*SaveShpParams)

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
		Start_date:               params.Start_date,
		End_date:                 params.End_date,
		Geographical_extent_geom: params.Geographical_extent_geom,
		Published:                params.Published,
		License:                  params.License,
		License_id:               params.License_id,
		Declared_creation_date:   params.Declared_creation_date,
	}

	fmt.Println(layer)

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

}

// SaveWmLayer save shapefile layer informations into database
func SaveWmLayer(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	params := proute.Json.(*SaveWmLayerParams)
	fmt.Println(params)
}
