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
	"log"
	"math"
	"net/http"
	"reflect"

	db "github.com/croll/arkeogis-server/db"
	"github.com/jmoiron/sqlx"

	"github.com/croll/arkeogis-server/model"
	routes "github.com/croll/arkeogis-server/webserver/routes"
)

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:        "/api/project",
			Description: "Get project infos",
			Func:        GetProject,
			Permissions: []string{},
			Params:      reflect.TypeOf(GetProjectParams{}),
			Method:      "GET",
		},
		&routes.Route{
			Path:        "/api/project",
			Description: "Save project layer",
			Func:        SaveProject,
			Permissions: []string{},
			Json:        reflect.TypeOf(SaveProjectParams{}),
			Method:      "POST",
		},
	}
	routes.RegisterMultiple(Routes)
}

type GetProjectParams struct {
	User_id int `json:"user_id"`
	Id      int `json:"id"`
}

func GetProject(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Json.(*GetProjectParams)

	var projectID int

	_user, _ := proute.Session.Get("user")
	user := _user.(model.User)

	if params.Id == 0 && params.User_id == 0 {
		http.Error(w, "Unable to get project project. No project id and no user id provided", 500)
		return
	}

	project := &model.ProjectFullInfos{}

	tx, err := db.DB.Beginx()
	if err != nil {
		http.Error(w, "Error creating transaction getting project: "+err.Error(), http.StatusBadRequest)
		return
	}

	if params.User_id != 0 {
		projectID, err = user.GetProjectId(tx)

		if err != nil {
			tx.Rollback()
			log.Fatal("can't get project!")
			userSqlError(w, err)
			return
		}
	} else {
		projectID = params.Id
	}

	if projectID > 0 {
		project.Id = projectID
		err = project.Get(tx)
		if err != nil {
			tx.Rollback()
			log.Fatal("can't get project!")
			userSqlError(w, err)
			return
		}
	} else {
		project.Id = 0
	}

	// Commit
	err = tx.Commit()
	if err != nil {
		log.Println("commit failed")
		userSqlError(w, err)
		return
	}

	j, err := json.Marshal(project)
	w.Write(j)

}

type SaveProjectParams struct {
	Name         string                 `json:"name"`
	Id           int                    `default:"0" json:"id"`
	User_id      int                    `json:"-"`
	Start_date   int                    `json:"start_date"`
	End_date     int                    `json:"end_date"`
	Geom         string                 `json:"geom"`
	Chronologies []int                  `json:"chronologies"`
	Layers       []model.LayerFullInfos `json:"layers"`
	Databases    []int                  `json:"databases"`
	Characs      []int                  `json:"characs"`
}

func SaveProject(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Json.(*SaveProjectParams)

	var err error

	_user, _ := proute.Session.Get("user")
	user := _user.(model.User)

	params.User_id = user.Id

	if params.Start_date == 0 && params.End_date == 0 {
		params.Start_date = math.MinInt32
		params.End_date = math.MaxInt32
	}

	tx, err := db.DB.Beginx()
	if err != nil {
		http.Error(w, "Save Project: Error creating transaction saving project: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Insert or update project
	var stmtProject *sqlx.NamedStmt

	if params.Id == 0 {
		stmtProject, err = tx.PrepareNamed("INSERT INTO \"project\" (\"name\", \"user_id\", \"created_at\", \"updated_at\", \"start_date\", \"end_date\", \"geom\") VALUES (:name, :user_id, now(), now(), :start_date, :end_date, ST_geomFromGeoJSON(:geom)) RETURNING id")
		if err != nil {
			log.Println("Save Project: error preparing insert of project", err)
			_ = tx.Rollback()
			userSqlError(w, err)
			return
		}
		err = stmtProject.Get(&params.Id, params)
		if err != nil {
			log.Println("Save Project: error inserting project", err)
			_ = tx.Rollback()
			userSqlError(w, err)
			return
		}
	} else {
		stmtProject, err = tx.PrepareNamed("UPDATE \"project\" SET \"name\" = :name, \"user_id\" = :user_id, \"updated_at\" = now(), \"start_date\" = :start_date, \"end_date\" = :end_date, \"geom\" = ST_geomFromGeoJSON(:geom) WHERE id = :id")
		_, err = stmtProject.Exec(params)
		if err != nil {
			log.Println("Save Project: error updateing project", err)
			_ = tx.Rollback()
			userSqlError(w, err)
			return
		}
		// Delete links
		_, err = tx.NamedExec("DELETE FROM \"project__chronology\" WHERE project_id=:id", params)
		if err != nil {
			log.Println("Save Project: Error deleting project__chronology", err)
			_ = tx.Rollback()
			userSqlError(w, err)
			return
		}
		_, err = tx.NamedExec("DELETE FROM \"project__charac\" WHERE project_id=:id", params)
		if err != nil {
			log.Println("Save Project: Error deleting project__charac", err)
			_ = tx.Rollback()
			userSqlError(w, err)
			return
		}
		_, err = tx.NamedExec("DELETE FROM \"project__database\" WHERE project_id=:id", params)
		if err != nil {
			log.Println("Save Project: Error deleting project__database", err)
			_ = tx.Rollback()
			userSqlError(w, err)
			return
		}
		_, err = tx.NamedExec("DELETE FROM \"project__map_layer\" WHERE project_id=:id", params)
		if err != nil {
			log.Println("Save Project: Error deleting project__map_layer", err)
			_ = tx.Rollback()
			userSqlError(w, err)
			return
		}
		_, err = tx.NamedExec("DELETE FROM \"project__shapefile\" WHERE project_id=:id", params)
		if err != nil {
			log.Println("Save Project: Error deleting project__shapefile", err)
			_ = tx.Rollback()
			userSqlError(w, err)
			return
		}
	}
	// Insert chronologies

	stmtChronos, err := tx.PrepareNamed("INSERT INTO \"project__chronology\" (project_id, root_chronology_id) VALUES (:project_id, :id)")
	if err != nil {
		log.Println("Save Project: Error inserting chronologies", err)
		_ = tx.Rollback()
		userSqlError(w, err)
		return
	}

	for _, chronoId := range params.Chronologies {
		_, err = stmtChronos.Exec(struct {
			Id         int
			Project_id int
		}{Id: chronoId, Project_id: params.Id})
		if err != nil {
			log.Println(err)
			_ = tx.Rollback()
			userSqlError(w, err)
			return
		}
	}

	// Insert characs

	stmtCharacs, err := tx.PrepareNamed("INSERT INTO \"project__charac\" (project_id, root_charac_id) VALUES (:project_id, :id)")
	if err != nil {
		log.Println("Save Project: Error inserting chronologies", err)
		_ = tx.Rollback()
		userSqlError(w, err)
		return
	}

	for _, characId := range params.Characs {
		_, err = stmtCharacs.Exec(struct {
			Id         int
			Project_id int
		}{Id: characId, Project_id: params.Id})
		if err != nil {
			log.Println(err)
			_ = tx.Rollback()
			userSqlError(w, err)
			return
		}
	}

	// Insert layers

	stmtLayersSHP, err := tx.PrepareNamed("INSERT INTO \"project__shapefile\" (project_id, shapefile_id) VALUES (:project_id, :id)")
	if err != nil {
		log.Println("Save Project: Error inserting shapefile", err)
		_ = tx.Rollback()
		userSqlError(w, err)
		return
	}

	stmtLayersWMS, err := tx.PrepareNamed("INSERT INTO \"project__map_layer\" (project_id, map_layer_id) VALUES (:project_id, :id)")
	if err != nil {
		log.Println("Save Project: Error inserting wms layer", err)
		_ = tx.Rollback()
		userSqlError(w, err)
		return
	}

	for _, layer := range params.Layers {
		if layer.Type == "shp" {
			_, err = stmtLayersSHP.Exec(struct {
				Id         int
				Project_id int
			}{Id: layer.Id, Project_id: params.Id})
		} else if layer.Type == "wms" || layer.Type == "wmts" {
			_, err = stmtLayersWMS.Exec(struct {
				Id         int
				Project_id int
			}{Id: layer.Id, Project_id: params.Id})
		}
		if err != nil {
			log.Println(err)
			_ = tx.Rollback()
			userSqlError(w, err)
			return
		}
	}

	// Insert databases

	stmtDatabases, err := tx.PrepareNamed("INSERT INTO \"project__database\" (project_id, database_id) VALUES (:project_id, :id)")
	if err != nil {
		log.Println("Save Project: Error inserting databases", err)
		_ = tx.Rollback()
		userSqlError(w, err)
		return
	}

	for _, databaseId := range params.Databases {
		_, err = stmtDatabases.Exec(struct {
			Id         int
			Project_id int
		}{Id: databaseId, Project_id: params.Id})
		if err != nil {
			log.Println(err)
			_ = tx.Rollback()
			userSqlError(w, err)
			return
		}

	}
	// Commit
	err = tx.Commit()
	if err != nil {
		log.Println("commit failed")
		userSqlError(w, err)
		return
	}

	j, err := json.Marshal(struct {
		Project_id int `json:"project_id"`
	}{params.Id})
	w.Write(j)

}
