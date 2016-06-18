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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"time"

	db "github.com/croll/arkeogis-server/db"
	routes "github.com/croll/arkeogis-server/webserver/routes"
)

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:        "/api/mapLayer",
			Description: "Save wms layer",
			Func:        SaveLayer,
			Permissions: []string{},
			Params:      reflect.TypeOf(SaveLayerParams{}),
			Method:      "POST",
		},
		&routes.Route{
			Path:        "/api/shapefile/togeojson",
			Description: "Get a shapefile and convert it to geojson",
			Func:        ShapefileToGeoJSON,
			Permissions: []string{},
			Params:      reflect.TypeOf(ShapefileToGeoJSONParams{}),
			Method:      "POST",
		},
		&routes.Route{
			Path:        "/api/shapefile",
			Description: "Save shapefile layer",
			Func:        SaveShapefile,
			Permissions: []string{},
			Params:      reflect.TypeOf(SaveShapefileParams{}),
			Method:      "POST",
		},
	}
	routes.RegisterMultiple(Routes)
}

type SaveLayerParams struct {
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

type SaveShapefileParams struct {
	Filename                 string
	Geom                     string
	Identifier               string
	Min_scale                int
	Max_scale                int
	Start_date               time.Time
	End_date                 time.Time
	Geographical_extent_geom string
	Published                bool
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

type ShapefileToGeoJSONParams struct {
	Filename string
	File     *routes.File
}

// SaveLayer saves wm(t)s layer into database
func SaveLayer(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Json.(*SaveLayerParams)

	tx, err := db.DB.Beginx()
	if err != nil {
		http.Error(w, "Error saving layer: "+err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println(tx)

	l, _ := json.Marshal(params)
	w.Header().Set("Content-Type", "application/json")
	w.Write(l)
}

// ShapefileToGeoJSON get shapefile and convert it to geojson
func ShapefileToGeoJSON(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Json.(*ShapefileToGeoJSONParams)

	filepath := "./uploaded/shp" + params.File.Name

	fmt.Println(params.File.Content)

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

}

// SaveShapefileParams save shapefile layer informations into database
func SaveShapefile(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	params := proute.Json.(*SaveShapefileParams)
	fmt.Println(params)
}
