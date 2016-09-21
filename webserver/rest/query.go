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

package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"

	db "github.com/croll/arkeogis-server/db"
	"github.com/croll/arkeogis-server/model"
	"github.com/croll/arkeogis-server/webserver/routes"
)

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:        "/api/query",
			Description: "Save Map Query",
			Func:        QuerySave,
			Method:      "POST",
			Json:        reflect.TypeOf(QuerySaveParams{}),
			Permissions: []string{
			//"request map",
			},
		},
		&routes.Route{
			Path:        "/api/query/{project_id:[0-9]+}",
			Func:        QueryGet,
			Description: "Get all queries from a project id",
			Method:      "GET",
			Params:      reflect.TypeOf(QueryGetParams{}),
			Permissions: []string{
			//"request map",
			},
		},
	}
	routes.RegisterMultiple(Routes)
}

type QueryGetParams struct {
	Project_id int
}

type QuerySaveParams struct {
	ProjectId int    `json:"project_id" min:"1"`
	Name      string `json:"name" min:"1"`
	Params    string `json:"params" min:"1"`
}

// QuerySave save the query
func QuerySave(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	params := proute.Json.(*QuerySaveParams)

	fmt.Println("params: ", params)

	// get the user
	_user, _ := proute.Session.Get("user")
	user := _user.(model.User)

	tx, err := db.DB.Beginx()
	if err != nil {
		log.Println("can't start transaction")
		userSqlError(w, err)
		return
	}

	// check if project exists and is owned by the current user
	c := 0
	err = tx.Get(&c, `SELECT count(*) FROM "project" WHERE "id"=$1 AND "user_id"=$2`, params.ProjectId, user.Id)
	if err != nil {
		fmt.Println("search project query failed : ", err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}
	if c != 1 {
		routes.FieldError(w, "name", "name", "QUERY.SAVE.T_ERROR_PROJECT_NOT_FOUND")
		tx.Rollback()
		return
	}

	err = tx.Get(&c, `SELECT count(*) FROM "saved_query" WHERE "project_id"=$1 AND "name"=$2`, params.ProjectId, params.Name)
	if err != nil {
		fmt.Println("search saved_query failed : ", err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}
	if c == 1 { // update
		_, err = tx.Exec(`UPDATE "saved_query" SET "params"=$3 WHERE "project_id"=$1 AND "name"=$2`, params.ProjectId, params.Name, params.Params)
		if err != nil {
			fmt.Println("update saved_query failed : ", err)
			userSqlError(w, err)
			_ = tx.Rollback()
			return
		}
	} else {
		_, err = tx.Exec(`INSERT INTO "saved_query" ("project_id", "name", "params") VALUES ($1, $2, $3)`, params.ProjectId, params.Name, params.Params)
		if err != nil {
			fmt.Println("insert saved_query failed : ", err)
			userSqlError(w, err)
			_ = tx.Rollback()
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("can't commit")
		userSqlError(w, err)
		return
	}

	j, err := json.Marshal(params)
	if err != nil {
		log.Println("marshal failed: ", err)
		return
	}
	w.Write(j)

}

// QuerySave save the query
func QueryGet(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	params := proute.Params.(*QueryGetParams)

	fmt.Println("params: ", params)

	// get the user
	_user, _ := proute.Session.Get("user")
	user := _user.(model.User)

	tx, err := db.DB.Beginx()
	if err != nil {
		log.Println("can't start transaction")
		userSqlError(w, err)
		return
	}

	// check if project exists and is owned by the current user
	c := 0
	err = tx.Get(&c, `SELECT count(*) FROM "project" WHERE "id"=$1 AND "user_id"=$2`, params.Project_id, user.Id)
	if err != nil {
		fmt.Println("search project query failed : ", err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}
	if c != 1 {
		routes.FieldError(w, "name", "name", "QUERY.SAVE.T_ERROR_PROJECT_NOT_FOUND")
		fmt.Println("project not found : ", params)
		tx.Rollback()
		return
	}

	res := []model.Saved_query{}
	err = tx.Select(&res, `SELECT * FROM "saved_query" WHERE "project_id"=$1`, params.Project_id)
	if err != nil {
		fmt.Println("search project saved queries failed : ", err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println("can't commit")
		userSqlError(w, err)
		return
	}

	j, err := json.Marshal(res)
	if err != nil {
		log.Println("marshal failed: ", err)
		return
	}
	w.Write(j)

}
