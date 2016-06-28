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
)

type ProjectLayerInfos struct {
	Id   int    `json:"id"`
	Type string `json:"type"`
}

type ProjectFullInfos struct {
	Project
	Chronologies []int `json:"chronologies"`
	Layers       []ProjectLayerInfos
	Databases    []int `json:"databases"`
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

	// Databases
	err = tx.Select(&pfi.Databases, "SELECT database_id from project__databases WHERE project_id = $1", pfi.Id)
	if err != nil {
		log.Println(err)
		return
	}

	// Layers WMS
	err = tx.Select(&pfi.Layers, "SELECT ml.id, ml.type FROM project__map_layer pml LEFT JOIN map_layer ml ON pml.map_layer_id = ml.id WHERE pml.project_id = $1", pfi.Id)
	if err != nil {
		log.Println(err)
		return
	}

	// Layers Shapefile
	err = tx.Select(&pfi.Layers, "SELECT s.id from project__shapefile ps LEFT JOIN shapefile s ON ps.shapefile_id = s.id WHERE ps.project_id = $1", pfi.Id)
	if err != nil {
		log.Println(err)
		return
	}

	return
}
