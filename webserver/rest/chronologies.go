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
	"reflect"
	"strconv"

	"net/http"

	db "github.com/croll/arkeogis-server/db"
	"github.com/croll/arkeogis-server/model"

	routes "github.com/croll/arkeogis-server/webserver/routes"
	"github.com/jmoiron/sqlx"
	sqlx_types "github.com/jmoiron/sqlx/types"
)

type ChronologyGetParams struct {
	Id int `min:"1" error:"Chronology Id is mandatory"`
}

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:        "/api/chronologies/flat",
			Func:        ChronologiesAll,
			Description: "Get all chronologies in all languages",
			Method:      "GET",
		},
		&routes.Route{
			Path:        "/api/chronologies",
			Func:        ChronologiesRoots,
			Description: "Get all root chronologies in all languages",
			Method:      "GET",
		},
		&routes.Route{
			Path:        "/api/chronologies",
			Description: "Create/Update a chronologie",
			Func:        ChronologiesUpdate,
			Method:      "POST",
			Json:        reflect.TypeOf(ChronologiesUpdateStruct{}),
			Permissions: []string{
				"adminusers",
			},
		},
		&routes.Route{
			Path:        "/api/chronologies/{id:[0-9]+}",
			Func:        ChronologiesGetTree,
			Description: "Get a chronology in all languages",
			Method:      "GET",
			Params:      reflect.TypeOf(ChronologyGetParams{}),
		},
		&routes.Route{
			Path:        "/api/chronologies/{id:[0-9]+}",
			Description: "Delete a chronologie",
			Func:        ChronologiesDelete,
			Method:      "DELETE",
			Permissions: []string{
				"adminusers",
			},
			Params: reflect.TypeOf(ChronologyGetParams{}),
		},
	}
	routes.RegisterMultiple(Routes)
}

// ChronologiesAll write all chronologies in all languages in a flat array
func ChronologiesAll(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	type row struct {
		Parent_id int                 `db:"parent_id" json:"parent_id"`
		Id        int                 `db:"id" json:"id"`
		Tr        sqlx_types.JSONText `db:"tr" json:"tr"`
	}

	chronologies := []row{}

	//err := db.DB.Select(&chronologies, "select parent_id, id, to_json((select array_agg(chronology_tr.*) from chronology_tr where chronology_tr.chronology_id = chronology.id)) as tr FROM chronology order by parent_id, \"order\", id")
	transquery := model.GetQueryTranslationsAsJSONObject("chronology_tr", "tbl.chronology_id = chronology.id", "", false, "name")
	q := "select parent_id, id, (" + transquery + ") as tr FROM chronology order by parent_id, \"start_date\", id"
	fmt.Println("q: ", q)
	err := db.DB.Select(&chronologies, q)
	fmt.Println("chronologies: ", chronologies)
	if err != nil {
		fmt.Println("err: ", err)
		return
	}

	j, err := json.Marshal(chronologies)
	w.Write(j)
}

// ChronologiesRoots write all root chronologies in all languages
func ChronologiesRoots(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	type row struct {
		model.Chronology_root
		model.Chronology
		Name         map[string]string `json:"name"`
		Description  map[string]string `json:"description"`
		UsersInGroup []model.User      `json:"users_in_group" ignore:"true"` // read-only, used to display users of the group
	}

	chronologies := []*row{}

	// transaction begin...
	tx, err := db.DB.Beginx()
	if err != nil {
		userSqlError(w, err)
		return
	}

	// load all roots
	err = db.DB.Select(&chronologies, "SELECT * FROM chronology_root")
	if err != nil {
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	// load all root chronologies
	for i, chrono := range chronologies {
		chrono.Chronology.Id = chrono.Chronology_root.Root_chronology_id
		err = chrono.Chronology.Get(tx)
		if err != nil {
			userSqlError(w, err)
			_ = tx.Rollback()
			return
		}

		// load translations
		tr := []model.Chronology_tr{}
		err = tx.Select(&tr, "SELECT * FROM chronology_tr WHERE chronology_id = "+strconv.Itoa(chrono.Chronology.Id))
		if err != nil {
			userSqlError(w, err)
			_ = tx.Rollback()
			return
		}
		chronologies[i].Name = model.MapSqlTranslations(tr, "Lang_isocode", "Name")
		chronologies[i].Description = model.MapSqlTranslations(tr, "Lang_isocode", "Description")
	}

	j, err := json.Marshal(chronologies)
	w.Write(j)
}

type ChronologyTreeStruct struct {
	model.Chronology
	Name        map[string]string      `json:"name"`
	Description map[string]string      `json:"description"`
	Content     []ChronologyTreeStruct `json:"content"`
}

// ChronologiesUpdateStruct structure (json)
type ChronologiesUpdateStruct struct {
	model.Chronology_root
	ChronologyTreeStruct
	UsersInGroup []model.User `json:"users_in_group" ignore:"true"` // read-only, used to display users of the group
}

// update chrono recursively
func setChronoRecursive(tx *sqlx.Tx, chrono *ChronologyTreeStruct, parent *ChronologyTreeStruct) error {
	var err error = nil

	// if we are the root, we have no parent id
	if parent != nil {
		chrono.Parent_id = parent.Id
	} else {
		chrono.Parent_id = 0
	}

	// save chronology...
	if chrono.Id > 0 {
		err = chrono.Update(tx)
		if err != nil {
			return err
		}
	} else {
		err = chrono.Create(tx)
		if err != nil {
			return err
		}
	}

	//log.Println("c: ", chrono)

	// delete any translations
	_, err = tx.Exec("DELETE FROM chronology_tr WHERE chronology_id = $1", chrono.Id)
	if err != nil {
		return err
	}

	// create a map of translations for name...
	tr := map[string]*model.Chronology_tr{}
	for isocode, name := range chrono.Name {
		tr[isocode] = &model.Chronology_tr{
			Chronology_id: chrono.Id,
			Lang_isocode:  isocode,
			Name:          name,
		}
	}

	// continue to update this map with descriptions...
	for isocode, description := range chrono.Description {
		m, ok := tr[isocode]
		if ok {
			m.Description = description
		} else {
			tr[isocode] = &model.Chronology_tr{
				Chronology_id: chrono.Id,
				Lang_isocode:  isocode,
				Description:   description,
			}
		}
	}

	// now insert translations rows in database...
	for _, m := range tr {
		err = m.Create(tx)
		if err != nil {
			return err
		}
	}

	// recursively call to subcontents...
	ids := []int{} // this array will be usefull to delete others chrono of this sub level that does not exists anymore
	for _, sub := range chrono.Content {
		err = setChronoRecursive(tx, &sub, chrono)
		if err != nil {
			return err
		}
		ids = append(ids, sub.Chronology.Id)
	}

	// search any chronology that should be deleted
	ids_to_delete := []int{} // the array of chronologies id to delete
	err = tx.Select(&ids_to_delete, "SELECT id FROM chronology WHERE id NOT IN ("+model.IntJoin(ids, true)+") AND parent_id = "+strconv.Itoa(chrono.Chronology.Id))
	if err != nil {
		return err
	}

	// delete translations of the chronologies that should be deleted
	_, err = tx.Exec("DELETE FROM chronology_tr WHERE chronology_id IN (" + model.IntJoin(ids_to_delete, true) + ")")
	if err != nil {
		return err
	}

	// delete chronologies itselfs...
	_, err = tx.Exec("DELETE FROM chronology WHERE id IN (" + model.IntJoin(ids_to_delete, true) + ")")
	if err != nil {
		return err
	}

	return err
}

// ChronologiesUpdate Create/Update a chronology
func ChronologiesUpdate(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	// get the post
	c := proute.Json.(*ChronologiesUpdateStruct)

	// transaction begin...
	tx, err := db.DB.Beginx()
	if err != nil {
		userSqlError(w, err)
		return
	}

	// get the user
	_user, ok := proute.Session.Get("user")
	if !ok {
		log.Println("ChronologiesUpdate: can't get user in session...", _user)
		_ = tx.Rollback()
		return
	}
	user, ok := _user.(model.User)
	if !ok {
		log.Println("ChronologiesUpdate: can't cast user...", _user)
		_ = tx.Rollback()
		return
	}
	err = user.Get(tx)
	user.Password = "" // immediatly erase password field, we don't need it
	if err != nil {
		log.Println("ChronologiesUpdate: can't load user...", _user)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	// boolean create, true if we are creating a totaly new chronology
	var create bool
	if c.Chronology.Id > 0 {
		create = false
		// @TODO: check that you are in group of this chrono when updating one
	} else {
		create = true
	}

	// save recursively this chronology
	err = setChronoRecursive(tx, &c.ChronologyTreeStruct, nil)
	if err != nil {
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	log.Println("geom: ", c.Chronology_root.Geom)

	// save the chronology_root row, but search/create it's group first
	c.Chronology_root.Root_chronology_id = c.Chronology.Id
	if create {
		// when creating, we also must create it's working group
		group := model.Group{
			Type: "chronology",
		}
		err = group.Create(tx)
		if err != nil {
			userSqlError(w, err)
			_ = tx.Rollback()
			return
		}

		// also save group name in langs...
		for isocode, name := range c.ChronologyTreeStruct.Name {
			group_tr := model.Group_tr{
				Group_id:     group.Id,
				Lang_isocode: isocode,
				Name:         name,
			}
			err = group_tr.Create(tx)
		}

		// create the chronology root
		c.Chronology_root.Admin_group_id = group.Id
		err = c.Chronology_root.Create(tx)
		if err != nil {
			userSqlError(w, err)
			_ = tx.Rollback()
			return
		}

	} else {
		// search the chronoroot to verify permissions
		chronoroot := model.Chronology_root{
			Root_chronology_id: c.Chronology.Id,
		}
		err = chronoroot.Get(tx)
		if err != nil {
			userSqlError(w, err)
			_ = tx.Rollback()
			return
		}

		// take the group
		group := model.Group{
			Id: chronoroot.Admin_group_id,
		}
		err = group.Get(tx)
		if err != nil {
			userSqlError(w, err)
			_ = tx.Rollback()
			return
		}

		// update translations of the group
		_, err = tx.Exec("DELETE FROM group_tr WHERE group_id = " + strconv.Itoa(group.Id))
		if err != nil {
			userSqlError(w, err)
			_ = tx.Rollback()
			return
		}
		for isocode, name := range c.ChronologyTreeStruct.Name {
			group_tr := model.Group_tr{
				Group_id:     group.Id,
				Lang_isocode: isocode,
				Name:         name,
			}
			err = group_tr.Create(tx)
		}

		// check that the user is in the group
		var ok bool
		ok, err = user.HaveGroups(tx, group)
		if err != nil {
			userSqlError(w, err)
			_ = tx.Rollback()
			return
		}

		if !ok {
			/*
				routes.ServerError(w, 403, "unauthorized")
				_ = tx.Rollback()
				return
			*/
		}

		// only theses fields can be modified
		chronoroot.Credits = c.Chronology_root.Credits
		chronoroot.Active = c.Chronology_root.Active
		chronoroot.Geom = c.Chronology_root.Geom
		//chronoroot.Author_user_id = c.Chronology_root.Author_user_id
		//chronoroot.Admin_group_id = c.Chronology_root.Admin_group_id

		err = chronoroot.Update(tx)
		if err != nil {
			log.Println("chronoroot update failed")
			userSqlError(w, err)
			_ = tx.Rollback()
			return
		}
	}

	answer, err := chronologiesGetTree(w, tx, c.Id, user)

	// commit...
	err = tx.Commit()
	if err != nil {
		log.Println("commit failed")
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	j, err := json.Marshal(answer)
	if err != nil {
		log.Println("marshal failed: ", err)
	}
	//log.Println("result: ", string(j))
	w.Write(j)
}

/*
type ChronologyTreeStruct struct {
	model.Chronology
	Name        map[string]string      `json:"name"`
	Description map[string]string      `json:"description"`
	Content     []ChronologyTreeStruct `json:"content"`
}

// ChronologiesUpdateStruct structure (json)
type ChronologiesUpdateStruct struct {
	model.Chronology_root
	ChronologyTreeStruct
	UsersInGroup []model.User `json:"users_in_group"` // read-only, used to display users of the group
}
*/

func getChronoRecursive(tx *sqlx.Tx, chrono *ChronologyTreeStruct) error {
	var err error = nil

	// load translations
	tr := []model.Chronology_tr{}
	err = tx.Select(&tr, "SELECT * FROM chronology_tr WHERE chronology_id = "+strconv.Itoa(chrono.Id))
	if err != nil {
		return err
	}
	chrono.Name = model.MapSqlTranslations(tr, "Lang_isocode", "Name")
	chrono.Description = model.MapSqlTranslations(tr, "Lang_isocode", "Description")

	// get the childs of this chronology from the db
	childs, err := chrono.Chronology.Childs(tx)
	if err != nil {
		return err
	}

	// recurse
	chrono.Content = make([]ChronologyTreeStruct, len(childs))
	for i, child := range childs {
		chrono.Content[i].Chronology = child
		err = getChronoRecursive(tx, &chrono.Content[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func chronologiesGetTree(w http.ResponseWriter, tx *sqlx.Tx, id int, user model.User) (answer *ChronologiesUpdateStruct, err error) {

	// answer structure that will be printed when everything is done
	answer = &ChronologiesUpdateStruct{}

	// get the chronology_root row
	answer.Chronology_root.Root_chronology_id = id
	err = answer.Chronology_root.Get(tx)
	if err != nil {
		userSqlError(w, err)
		_ = tx.Rollback()
		return nil, err
	}

	// get the chronology (root)
	answer.Chronology.Id = id
	err = answer.Chronology.Get(tx)
	if err != nil {
		userSqlError(w, err)
		_ = tx.Rollback()
		return nil, err
	}

	// now get the chronology translations and all childrens
	err = getChronoRecursive(tx, &answer.ChronologyTreeStruct)
	if err != nil {
		userSqlError(w, err)
		_ = tx.Rollback()
		return nil, err
	}

	// get users of the chrono group
	group := model.Group{
		Id: answer.Chronology_root.Admin_group_id,
	}
	err = group.Get(tx)
	if err != nil {
		userSqlError(w, err)
		_ = tx.Rollback()
		return nil, err
	}
	answer.UsersInGroup, err = group.GetUsers(tx)
	if err != nil {
		userSqlError(w, err)
		_ = tx.Rollback()
		return nil, err
	}

	for i := range answer.UsersInGroup {
		answer.UsersInGroup[i].Password = ""
	}

	return answer, nil
}

func ChronologiesGetTree(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	params := proute.Params.(*ChronologyGetParams)

	// transaction begin...
	tx, err := db.DB.Beginx()
	if err != nil {
		userSqlError(w, err)
		return
	}

	// get the user
	_user, ok := proute.Session.Get("user")
	if !ok {
		log.Println("ChronologiesUpdate: can't get user in session...", _user)
		_ = tx.Rollback()
		return
	}
	user, ok := _user.(model.User)
	if !ok {
		log.Println("ChronologiesUpdate: can't cast user...", _user)
		_ = tx.Rollback()
		return
	}
	err = user.Get(tx)
	user.Password = "" // immediatly erase password field, we don't need it
	if err != nil {
		log.Println("ChronologiesUpdate: can't load user...", _user)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	answer, err := chronologiesGetTree(w, tx, params.Id, user)

	// commit...
	err = tx.Commit()
	if err != nil {
		log.Println("commit failed")
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	j, err := json.Marshal(answer)
	if err != nil {
		log.Println("marshal failed: ", err)
	}
	//log.Println("result: ", string(j))
	w.Write(j)

}

func chronologiesDeleteRecurse(chrono ChronologyTreeStruct, tx *sqlx.Tx) error {
	var err error
	for _, chrono := range chrono.Content {
		err = chronologiesDeleteRecurse(chrono, tx)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec("DELETE FROM chronology_tr WHERE chronology_id = " + strconv.Itoa(chrono.Id))
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM chronology WHERE id = " + strconv.Itoa(chrono.Id))
	if err != nil {
		return err
	}

	return nil
}

func ChronologiesDelete(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	params := proute.Params.(*ChronologyGetParams)

	// transaction begin...
	tx, err := db.DB.Beginx()
	if err != nil {
		userSqlError(w, err)
		return
	}

	// get the user
	_user, ok := proute.Session.Get("user")
	if !ok {
		log.Println("ChronologiesUpdate: can't get user in session...", _user)
		_ = tx.Rollback()
		return
	}
	user, ok := _user.(model.User)
	if !ok {
		log.Println("ChronologiesUpdate: can't cast user...", _user)
		_ = tx.Rollback()
		return
	}
	err = user.Get(tx)
	user.Password = "" // immediatly erase password field, we don't need it
	if err != nil {
		log.Println("ChronologiesUpdate: can't load user...", _user)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	// get the full chronologie tree
	answer, err := chronologiesGetTree(w, tx, params.Id, user)

	// delete chronology_root
	err = answer.Chronology_root.Delete(tx)
	if err != nil {
		log.Println("delete Chronology root", err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	// delete admin gruop in user__group...
	_, err = tx.Exec("DELETE FROM user__group WHERE group_id = " + strconv.Itoa(answer.Admin_group_id))
	if err != nil {
		log.Println("delete admin users group failed", err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	// delete admin gruop in group_tr...
	_, err = tx.Exec("DELETE FROM \"group_tr\" WHERE group_id = " + strconv.Itoa(answer.Admin_group_id))
	if err != nil {
		log.Println("delete admin group_tr failed", err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	// delete admin gruop in group...
	_, err = tx.Exec("DELETE FROM \"group\" WHERE id = " + strconv.Itoa(answer.Admin_group_id))
	if err != nil {
		log.Println("delete admin group failed", err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	// recursively delete chronology...
	err = chronologiesDeleteRecurse(answer.ChronologyTreeStruct, tx)
	if err != nil {
		log.Println("chronologiesDeleteRecurse failed", err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	// commit...
	err = tx.Commit()
	if err != nil {
		log.Println("commit failed")
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}
}
