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
	"log"

	"github.com/jmoiron/sqlx"
	sqlx_types "github.com/jmoiron/sqlx/types"
)

type ProjectLayerInfos struct {
	Id                       int                 `json:"id"`
	Type                     string              `json:"type"`
	Uniq_code                string              `json:"uniq_code"`
	Geographical_extent_geom string              `json:"geom"`
	Min_scale                int                 `json:"min_scale"`
	Max_scale                int                 `json:"max_scale"`
	Translations             sqlx_types.JSONText `db:"translations" json:"translations"`
	Url                      string              `json:"url"`
	Identifier               string              `json:"identifier"`
}

type ProjectFullInfos struct {
	Project
	Chronologies []struct {
		Root_chronology_id int `json:"id"`
	} `json:"chronologies"`
	Layers  []ProjectLayerInfos `json:"layers"`
	Characs []struct {
		Root_charac_id int `json:"id"`
	} `json:"characs"`
	Databases []struct {
		Database_id int `json:"id"`
	} `json:"databases"`
}

func (pfi *ProjectFullInfos) Get(tx *sqlx.Tx) (err error) {

	// Infos
	err = tx.Get(pfi, "SELECT *,ST_AsGeoJSON(geom) as geom from project WHERE id = $1", pfi.Id)
	if err != nil {
		log.Println(err)
		return
	}

	// Chronologies
	err = tx.Select(&pfi.Chronologies, "SELECT root_chronology_id from project__chronology WHERE project_id = $1", pfi.Id)
	if err != nil {
		log.Println(err)
		return
	}

	// Characs
	err = tx.Select(&pfi.Characs, "SELECT project__charac.root_charac_id from project__charac LEFT JOIN charac ON charac.id = project__charac.root_charac_id WHERE project_id = $1 ORDER BY charac.order", pfi.Id)
	if err != nil {
		log.Println(err)
		return
	}

	// Databases
	err = tx.Select(&pfi.Databases, "SELECT database_id from project__database WHERE project_id = $1", pfi.Id)
	if err != nil {
		log.Println(err)
		return
	}

	// Layers WMS
	transquery := GetQueryTranslationsAsJSONObject("map_layer_tr", "tbl.map_layer_id = ml.id", "", false, "name", "attribution", "copyright")
	err = tx.Select(&pfi.Layers, "SELECT ml.id, ST_AsGeojson(ml.geographical_extent_geom) as geographical_extent_geom, url, identifier, ("+transquery+") as translations, ml.min_scale, ml.max_scale, ml.type, 'wms' || ml.id AS uniq_code FROM project__map_layer pml LEFT JOIN map_layer ml ON pml.map_layer_id = ml.id WHERE pml.project_id = $1", pfi.Id)
	if err != nil {
		log.Println(err)
		return
	}

	// Layers Shapefile
	transquery = GetQueryTranslationsAsJSONObject("shapefile_tr", "tbl.shapefile_id = s.id", "", false, "name", "attribution", "copyright")
	err = tx.Select(&pfi.Layers, "SELECT s.id, ST_AsGeojson(s.geographical_extent_geom) as geographical_extent_geom, ("+transquery+") as translations, 'shp' as type, 'shp' || s.id AS uniq_code from project__shapefile ps LEFT JOIN shapefile s ON ps.shapefile_id = s.id WHERE ps.project_id = $1", pfi.Id)
	if err != nil {
		log.Println(err)
		return
	}

	return
}
