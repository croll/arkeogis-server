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
	"log"
	"net/http"
	"reflect"
	"strings"

	db "github.com/croll/arkeogis-server/db"
	"github.com/croll/arkeogis-server/model"
	"github.com/croll/arkeogis-server/webserver/routes"
)

// GroupListParams is params struct for GroupList query
type GroupListParams struct {
	Type   string `default:"user" enum:"user,charac,chronology"`
	Limit  int    `default:"10" min:"1" max:"100" error:"limit over boundaries"`
	Page   int    `default:"1" min:"1" error:"page over boundaries"`
	Order  string `default:"g_tr.name" enum:"g.created_at,-g.created_at,g.updated_at,-g.updated_at,g_tr.name,-g_tr.name" error:"bad order"`
	Filter string `default:""`
}

type GroupGetParams struct {
	Id int `min:"0" error:"id over boundaries"`
}

type GroupSetPost struct {
	model.Group
	Users []model.User `json:"users" ignore:"true"`
}

func init() {

	Routes := []*routes.Route{
		/*&routes.Route{
			Path:        "/api/groups",
			Description: "Create a new arkeogis group",
			Func:        GroupCreate,
			Method:      "POST",
			Json:        reflect.TypeOf(Groupcreate{}),
			Permissions: []string{
			//"AdminGroups",
			},
		},*/
		&routes.Route{
			Path:        "/api/groups",
			Description: "List arkeogis groups",
			Func:        GroupList,
			Method:      "GET",
			Permissions: []string{
			//"AdminGroups",
			},
			Params: reflect.TypeOf(GroupListParams{}),
		},
		&routes.Route{
			Path:        "/api/groups/{id:[0-9]+}",
			Description: "Get an arkeogis group",
			Func:        GroupGet,
			Method:      "GET",
			Permissions: []string{
			//"AdminGroups",
			},
			Params: reflect.TypeOf(GroupGetParams{}),
		},
		&routes.Route{
			Path:        "/api/groups/{id:[0-9]+}",
			Description: "Update an arkeogis group",
			Func:        GroupSet,
			Method:      "POST",
			Json:        reflect.TypeOf(GroupSetPost{}),
			Permissions: []string{
			//"AdminGroups",
			},
		},
		/*&routes.Route{
			Path:        "/api/groups/{id:[0-9]+}",
			Description: "Update an arkeogis group",
			Func:        GroupUpdate,
			Method:      "POST",
			Json:        reflect.TypeOf(Groupcreate{}),
			Permissions: []string{
			//"AdminGroups",
			},
		},*/
		/*&routes.Route{
			Path:        "/api/groups",
			Description: "Delete an arkeogis group",
			Func:        GroupDelete,
			Method:      "DELETE",
			Permissions: []string{
			//"AdminGroups",
			},
		},*/
	}
	routes.RegisterMultiple(Routes)
}

// GroupList List the groups.
func GroupList(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	type TRGroup struct {
		model.Group
		model.Group_tr
	}
	type Answer struct {
		Data  []TRGroup `json:"data"`
		Count int       `json:"count"`
	}

	answer := Answer{}

	params := proute.Params.(*GroupListParams)

	// decode order...
	order := params.Order
	orderdir := "ASC"
	if strings.HasPrefix(order, "-") {
		order = order[1:]
		orderdir = "DESC"
	}
	/////

	offset := (params.Page - 1) * params.Limit

	// get groups
	err := db.DB.Select(&answer.Data,
		" SELECT * FROM \"group\" g "+
			" LEFT JOIN \"group_tr\" g_tr ON g.id = g_tr.group_id "+
			" WHERE g_tr.name ILIKE $1 AND g.type=$2 AND g.id > 0"+
			" ORDER BY "+order+" "+orderdir+
			" OFFSET $3 "+
			" LIMIT $4",
		"%"+params.Filter+"%", params.Type, offset, params.Limit)
	if err != nil {
		log.Println("get groups failed", err)
		routes.ServerError(w, 500, "INTERNAL ERROR")
		return
	}

	// get total count
	err = db.DB.Get(&answer.Count, "SELECT count(g.*) FROM \"group\" g LEFT JOIN \"group_tr\" g_tr ON g.id = g_tr.group_id WHERE g_tr.name ILIKE $1 AND g.type=$2", "%"+params.Filter+"%", params.Type)
	if err != nil {
		log.Println("get total count failed", err)
		routes.ServerError(w, 500, "INTERNAL ERROR")
		return
	}

	j, _ := json.Marshal(answer)
	w.Write(j)
}

func groupGet(w http.ResponseWriter, r *http.Request, params GroupGetParams) {
	type Answer struct {
		model.Group
		Users []model.User `json:"users"`
	}

	tx, err := db.DB.Beginx()
	if err != nil {
		log.Println("can't start transaction", err)
		routes.ServerError(w, 500, "INTERNAL ERROR")
		return
	}

	answer := Answer{}
	answer.Id = params.Id

	answer.Group.Get(tx)

	answer.Users, err = answer.Group.GetUsers(tx)

	err = tx.Commit()
	if err != nil {
		log.Println("can't commit", err)
		routes.ServerError(w, 500, "INTERNAL ERROR")
		return
	}

	j, _ := json.Marshal(answer)
	w.Write(j)
}

func GroupGet(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	params := proute.Params.(*GroupGetParams)
	groupGet(w, r, *params)
}

func GroupSet(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	o := proute.Json.(*GroupSetPost)

	tx, err := db.DB.Beginx()
	if err != nil {
		log.Println("can't start transaction", err)
		routes.ServerError(w, 500, "INTERNAL ERROR")
		return
	}

	err = o.Group.Update(tx)
	if err != nil {
		log.Println("can't update group", err)
		routes.ServerError(w, 500, "INTERNAL ERROR")
		return
	}

	err = o.Group.SetUsers(tx, o.Users)
	if err != nil {
		log.Println("can't set group users", err)
		routes.ServerError(w, 500, "INTERNAL ERROR")
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println("can't commit", err)
		routes.ServerError(w, 500, "INTERNAL ERROR")
		return
	}

	groupGet(w, r, GroupGetParams{
		Id: o.Id,
	})
}

func GroupAddUser(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
}
func GroupDelUser(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
}
