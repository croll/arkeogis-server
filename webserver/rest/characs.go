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
	"encoding/csv"
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

type CharacGetParams struct {
	Id         int `min:"1" error:"Charac Id is mandatory"`
	Project_id int
}

type CharacRootsParams struct {
	Project_id int
}

type CharacListCsvParams struct {
	Id      int    `min:"1" error:"Charac Id is mandatory"`
	Isocode string `json:"isocode"`
	Dl      string `json:"dl"`
	Html    int    `json:"html"`
}

type CharacSetHiddensParams struct {
	Id         int `min:"1" error:"Charac Id is mandatory"`
	Project_id int
}

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:        "/api/characs/flat",
			Func:        CharacsAll,
			Description: "Get all characs in all languages",
			Method:      "GET",
		},
		&routes.Route{
			Path:        "/api/characs",
			Func:        CharacsRoots,
			Description: "Get all root characs in all languages",
			Method:      "GET",
			Params:      reflect.TypeOf(CharacRootsParams{}),
		},
		&routes.Route{
			Path:        "/api/characs/csv",
			Func:        CharacListCsv,
			Description: "Get all characs as csv",
			Params:      reflect.TypeOf(CharacListCsvParams{}),
			Method:      "GET",
			Permissions: []string{},
		},
		&routes.Route{
			Path:        "/api/characs",
			Description: "Create/Update a charac",
			Func:        CharacsUpdate,
			Method:      "POST",
			Json:        reflect.TypeOf(CharacsUpdateStruct{}),
			Permissions: []string{},
		},
		&routes.Route{
			Path:        "/api/characs/{id:[0-9]+}",
			Func:        CharacsGetTree,
			Description: "Get a charac in all languages",
			Method:      "GET",
			Params:      reflect.TypeOf(CharacGetParams{}),
		},
		&routes.Route{
			Path:        "/api/characs/{id:[0-9]+}",
			Description: "Delete a charac",
			Func:        CharacsDelete,
			Method:      "DELETE",
			Permissions: []string{
				"adminusers",
			},
			Params: reflect.TypeOf(CharacGetParams{}),
		},
		&routes.Route{
			Path:        "/api/characs/{id:[0-9]+}/hiddens/{project_id:[0-9]+}",
			Description: "",
			Func:        CharacSetHiddens,
			Method:      "POST",
			Params:      reflect.TypeOf(CharacSetHiddensParams{}),
			Json:        reflect.TypeOf(CharacSetHiddensStruct{}),
		},
	}
	routes.RegisterMultiple(Routes)
}

func CharacsAll(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	type row struct {
		Parent_id int                 `db:"parent_id" json:"parent_id"`
		Id        int                 `db:"id" json:"id"`
		Tr        sqlx_types.JSONText `db:"tr" json:"tr"`
	}

	characs := []row{}

	//err := db.DB.Select(&characs, "select parent_id, id, to_json((select array_agg(charac_tr.*) from charac_tr where charac_tr.charac_id = charac.id)) as tr FROM charac order by parent_id, \"order\", id")
	transquery := model.GetQueryTranslationsAsJSONObject("charac_tr", "tbl.charac_id = charac.id", "", false, "name")
	q := "select parent_id, id, (" + transquery + ") as tr FROM charac order by parent_id, \"order\", id"
	fmt.Println("q: ", q)
	err := db.DB.Select(&characs, q)
	fmt.Println("characs: ", characs)
	if err != nil {
		fmt.Println("err: ", err)
		return
	}

	j, err := json.Marshal(characs)
	w.Write(j)
}

// CharacsRoots write all root characs in all languages
func CharacsRoots(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	type row struct {
		model.Charac_root
		model.Charac
		Name         map[string]string `json:"name"`
		Description  map[string]string `json:"description"`
		HiddensCount int               `json:"hiddens_count"`
		//UsersInGroup []model.User      `json:"users_in_group" ignore:"true"` // read-only, used to display users of the group
		//Author model.User `json:"author" ignore:"true"` // read-only, used to display users of the group
	}

	params := proute.Params.(*CharacRootsParams)

	characs := []*row{}

	// transaction begin...
	tx, err := db.DB.Beginx()
	if err != nil {
		log.Println(err)
		userSqlError(w, err)
		return
	}

	// if Project_id is specified, verify that we are the owner of the project
	if params.Project_id != 0 {
		// get the user
		_user, ok := proute.Session.Get("user")
		if !ok {
			log.Println("CharacsUpdate: can't get user in session...", _user)
			_ = tx.Rollback()
			return
		}

		user, ok := _user.(model.User)
		if !ok {
			log.Println("CharacsUpdate: can't cast user...", _user)
			_ = tx.Rollback()
			return
		}

		// check if the user is the owner of the project
		count := 0
		tx.Get(&count, "SELECT count(*) FROM project WHERE id = "+strconv.Itoa(params.Project_id)+" AND user_id = "+strconv.Itoa(user.Id))
		if count != 1 {
			log.Println("CharacSetHiddens: user is not the owner...", user, params.Project_id)
			_ = tx.Rollback()
			return
		}
	}

	// load all roots
	err = db.DB.Select(&characs, "SELECT charac_root.* FROM charac_root LEFT JOIN charac on charac_root.root_charac_id = charac.id ORDER BY charac.order")
	if err != nil {
		log.Println(err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	// load all root characs
	for _, charac := range characs {
		charac.Charac.Id = charac.Charac_root.Root_charac_id
		err = charac.Charac.Get(tx)
		if err != nil {
			log.Println(err)
			_ = tx.Rollback()
			userSqlError(w, err)
			return
		}

		// load translations
		tr := []model.Charac_tr{}
		err = tx.Select(&tr, "SELECT * FROM charac_tr WHERE charac_id = "+strconv.Itoa(charac.Charac.Id))
		if err != nil {
			log.Println(err)
			userSqlError(w, err)
			_ = tx.Rollback()
			return
		}
		charac.Name = model.MapSqlTranslations(tr, "Lang_isocode", "Name")
		charac.Description = model.MapSqlTranslations(tr, "Lang_isocode", "Description")

		// load custom modified characs

		// get the author user
		/*		err = charac.Author.Get(tx)
				charac.Author.Password = ""
				if err != nil {
					userSqlError(w, err)
					_ = tx.Rollback()
					return
				}
		*/
		if params.Project_id > 0 {
			hidden_count := 0
			err = tx.Get(&hidden_count, `WITH RECURSIVE subcharac(id, parent_id, charac_id, project_id) AS (
                                          SELECT id, parent_id, phc.charac_id, phc.project_id
                                          FROM charac LEFT JOIN project_hidden_characs phc ON phc.charac_id = charac.id WHERE id = $1
                                         UNION ALL
                                          SELECT c2.id, c2.parent_id, phc2.charac_id, phc2.project_id
                                          FROM subcharac AS sc, charac AS c2 LEFT JOIN project_hidden_characs phc2 ON phc2.charac_id = c2.id
                                          WHERE c2.parent_id = sc.id
                                         )
                                         SELECT count(*) FROM subcharac WHERE project_id=$2`, charac.Id, params.Project_id)
			if err != nil {
				log.Println(err)
				userSqlError(w, err)
				_ = tx.Rollback()
				return
			}
			charac.HiddensCount = hidden_count
		}
	}

	// commit...
	err = tx.Commit()
	if err != nil {
		log.Println("commit failed")
		userSqlError(w, err)
		return
	}

	j, err := json.Marshal(characs)
	w.Write(j)
}

type CharacTreeStruct struct {
	model.Charac
	Name        map[string]string  `json:"name"`
	Description map[string]string  `json:"description"`
	Content     []CharacTreeStruct `json:"content"`
	Hidden      bool               `json:"hidden"`
}

// CharacsUpdateStruct structure (json)
type CharacsUpdateStruct struct {
	model.Charac_root
	CharacTreeStruct
	UsersInGroup []model.User `json:"users_in_group" ignore:"true"` // read-only, used to display users of the group
	Author       model.User   `json:"author" ignore:"true"`         // read-only, used to display users of the group
}

// update charac recursively
func setCharacRecursive(tx *sqlx.Tx, charac *CharacTreeStruct, parent *CharacTreeStruct) error {
	var err error = nil

	// if we are the root, we have no parent id
	if parent != nil {
		charac.Parent_id = parent.Id
	} else {
		charac.Parent_id = 0
	}

	// save charac...
	if charac.Id > 0 {
		err = charac.Update(tx)
		if err != nil {
			return err
		}
	} else {
		err = charac.Create(tx)
		if err != nil {
			return err
		}
	}

	//log.Println("c: ", charac)

	// delete any translations
	_, err = tx.Exec("DELETE FROM charac_tr WHERE charac_id = $1", charac.Id)
	if err != nil {
		return err
	}

	// create a map of translations for name...
	tr := map[string]*model.Charac_tr{}
	for isocode, name := range charac.Name {
		tr[isocode] = &model.Charac_tr{
			Charac_id:    charac.Id,
			Lang_isocode: isocode,
			Name:         name,
		}
	}

	// continue to update this map with descriptions...
	for isocode, description := range charac.Description {
		m, ok := tr[isocode]
		if ok {
			m.Description = description
		} else {
			tr[isocode] = &model.Charac_tr{
				Charac_id:    charac.Id,
				Lang_isocode: isocode,
				Description:  description,
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
	ids := []int{} // this array will be usefull to delete others charac of this sub level that does not exists anymore
	for _, sub := range charac.Content {
		err = setCharacRecursive(tx, &sub, charac)
		if err != nil {
			return err
		}
		ids = append(ids, sub.Charac.Id)
	}

	// search any charac that should be deleted
	ids_to_delete := []int{} // the array of characs id to delete
	err = tx.Select(&ids_to_delete, "SELECT id FROM charac WHERE id NOT IN ("+model.IntJoin(ids, true)+") AND parent_id = "+strconv.Itoa(charac.Charac.Id))
	if err != nil {
		return err
	}

	// delete translations of the characs that should be deleted
	_, err = tx.Exec("DELETE FROM charac_tr WHERE charac_id IN (" + model.IntJoin(ids_to_delete, true) + ")")
	if err != nil {
		return err
	}

	// delete characs itselfs...
	_, err = tx.Exec("DELETE FROM charac WHERE id IN (" + model.IntJoin(ids_to_delete, true) + ")")
	if err != nil {
		return err
	}

	return err
}

// CharacsUpdate Create/Update a charac
func CharacsUpdate(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	// get the post
	c := proute.Json.(*CharacsUpdateStruct)

	// transaction begin...
	tx, err := db.DB.Beginx()
	if err != nil {
		userSqlError(w, err)
		return
	}

	// get the user
	_user, ok := proute.Session.Get("user")
	if !ok {
		log.Println("CharacsUpdate: can't get user in session...", _user)
		_ = tx.Rollback()
		return
	}
	user, ok := _user.(model.User)
	if !ok {
		log.Println("CharacsUpdate: can't cast user...", _user)
		_ = tx.Rollback()
		return
	}
	err = user.Get(tx)
	user.Password = "" // immediatly erase password field, we don't need it
	if err != nil {
		log.Println("CharacsUpdate: can't load user...", _user)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	// boolean create, true if we are creating a totaly new charac
	var create bool
	if c.Charac.Id > 0 {
		create = false
		// @TODO: check that you are in group of this charac when updating one
	} else {
		create = true
	}

	// save recursively this charac
	err = setCharacRecursive(tx, &c.CharacTreeStruct, nil)
	if err != nil {
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	//log.Println("geom: ", c.Charac_root.Geom)

	// save the charac_root row, but search/create it's group first
	c.Charac_root.Root_charac_id = c.Charac.Id
	if create {
		// when creating, we also must create it's working group
		group := model.Group{
			Type: "charac",
		}
		err = group.Create(tx)
		if err != nil {
			userSqlError(w, err)
			_ = tx.Rollback()
			return
		}

		// also save group name in langs...
		for isocode, name := range c.CharacTreeStruct.Name {
			group_tr := model.Group_tr{
				Group_id:     group.Id,
				Lang_isocode: isocode,
				Name:         name,
			}
			err = group_tr.Create(tx)
		}

		// create the charac root
		c.Charac_root.Admin_group_id = group.Id
		err = c.Charac_root.Create(tx)
		if err != nil {
			userSqlError(w, err)
			_ = tx.Rollback()
			return
		}

	} else {
		// search the characroot to verify permissions
		characroot := model.Charac_root{
			Root_charac_id: c.Charac.Id,
		}
		err = characroot.Get(tx)
		if err != nil {
			userSqlError(w, err)
			_ = tx.Rollback()
			return
		}

		// take the group
		group := model.Group{
			Id: characroot.Admin_group_id,
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
		for isocode, name := range c.CharacTreeStruct.Name {
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
		//characroot.Credits = c.Charac_root.Credits
		//characroot.Active = c.Charac_root.Active
		//characroot.Geom = c.Charac_root.Geom
		//characroot.Author_user_id = c.Charac_root.Author_user_id
		//characroot.Admin_group_id = c.Charac_root.Admin_group_id
		characroot.Cached_langs = c.Charac_root.Cached_langs

		err = characroot.Update(tx)
		if err != nil {
			log.Println("characroot update failed")
			userSqlError(w, err)
			_ = tx.Rollback()
			return
		}
	}

	answer, err := characsGetTree(w, tx, c.Id, 0, user)

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

func getCharacRecursive(tx *sqlx.Tx, charac *CharacTreeStruct, project_id int) error {
	var err error = nil

	// load translations
	tr := []model.Charac_tr{}
	err = tx.Select(&tr, "SELECT * FROM charac_tr WHERE charac_id = "+strconv.Itoa(charac.Id))
	if err != nil {
		return err
	}
	charac.Name = model.MapSqlTranslations(tr, "Lang_isocode", "Name")
	charac.Description = model.MapSqlTranslations(tr, "Lang_isocode", "Description")

	// check if enabled in project
	if project_id > 0 {
		hiddenCount := 0
		tx.Get(&hiddenCount, "SELECT count(*) FROM project_hidden_characs WHERE project_id = "+strconv.Itoa(project_id)+" AND charac_id = "+strconv.Itoa(charac.Id))
		if hiddenCount > 0 {
			charac.Hidden = true
			log.Println("found hidden : ", charac.Id)
		}
	}

	// get the childs of this charac from the db
	childs, err := charac.Charac.Childs(tx)
	if err != nil {
		return err
	}

	// recurse
	charac.Content = make([]CharacTreeStruct, len(childs))
	for i, child := range childs {
		charac.Content[i].Charac = child
		err = getCharacRecursive(tx, &charac.Content[i], project_id)
		if err != nil {
			return err
		}
	}

	return nil
}

func characsGetTree(w http.ResponseWriter, tx *sqlx.Tx, id int, project_id int, user model.User) (answer *CharacsUpdateStruct, err error) {

	// answer structure that will be printed when everything is done
	answer = &CharacsUpdateStruct{}

	// get the charac_root row
	answer.Charac_root.Root_charac_id = id
	err = answer.Charac_root.Get(tx)
	if err != nil {
		userSqlError(w, err)
		_ = tx.Rollback()
		return nil, err
	}

	// get the charac (root)
	answer.Charac.Id = id
	err = answer.Charac.Get(tx)
	if err != nil {
		userSqlError(w, err)
		_ = tx.Rollback()
		return nil, err
	}

	// now get the charac translations and all childrens
	err = getCharacRecursive(tx, &answer.CharacTreeStruct, project_id)
	if err != nil {
		userSqlError(w, err)
		_ = tx.Rollback()
		return nil, err
	}

	// get users of the charac group
	group := model.Group{
		Id: answer.Charac_root.Admin_group_id,
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

	// get the author user
	/*	answer.Author.Id = answer.Charac_root.Author_user_id
		err = answer.Author.Get(tx)
		answer.Author.Password = ""
		if err != nil {
			userSqlError(w, err)
			_ = tx.Rollback()
			return nil, err
		}*/ // no author in characs vs chrono

	return answer, nil
}

func CharacsGetTree(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	params := proute.Params.(*CharacGetParams)

	// transaction begin...
	tx, err := db.DB.Beginx()
	if err != nil {
		userSqlError(w, err)
		return
	}

	// get the user
	_user, ok := proute.Session.Get("user")
	if !ok {
		log.Println("CharacsGetTree: can't get user in session...", _user)
		_ = tx.Rollback()
		return
	}
	user, ok := _user.(model.User)
	if !ok {
		log.Println("CharacsGetTree: can't cast user...", _user)
		_ = tx.Rollback()
		return
	}
	err = user.Get(tx)
	user.Password = "" // immediatly erase password field, we don't need it
	if err != nil {
		log.Println("CharacsGetTree: can't load user...", _user)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	answer, err := characsGetTree(w, tx, params.Id, params.Project_id, user)

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

func characsDeleteRecurse(charac CharacTreeStruct, tx *sqlx.Tx) error {
	var err error
	for _, charac := range charac.Content {
		err = characsDeleteRecurse(charac, tx)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec("DELETE FROM charac_tr WHERE charac_id = " + strconv.Itoa(charac.Id))
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM charac WHERE id = " + strconv.Itoa(charac.Id))
	if err != nil {
		return err
	}

	return nil
}

func CharacsDelete(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	params := proute.Params.(*CharacGetParams)

	// transaction begin...
	tx, err := db.DB.Beginx()
	if err != nil {
		userSqlError(w, err)
		return
	}

	// get the user
	_user, ok := proute.Session.Get("user")
	if !ok {
		log.Println("CharacsUpdate: can't get user in session...", _user)
		_ = tx.Rollback()
		return
	}
	user, ok := _user.(model.User)
	if !ok {
		log.Println("CharacsUpdate: can't cast user...", _user)
		_ = tx.Rollback()
		return
	}
	err = user.Get(tx)
	user.Password = "" // immediatly erase password field, we don't need it
	if err != nil {
		log.Println("CharacsUpdate: can't load user...", _user)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	// get the full characlogie tree
	answer, err := characsGetTree(w, tx, params.Id, 0, user)

	// delete charac_root
	err = answer.Charac_root.Delete(tx)
	if err != nil {
		log.Println("delete Charac root", err)
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

	// recursively delete charac...
	err = characsDeleteRecurse(answer.CharacTreeStruct, tx)
	if err != nil {
		log.Println("characsDeleteRecurse failed", err)
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

// CharacSetHiddensStruct structure (json)
type CharacSetHiddensStruct struct {
	HiddenIds []int `json:"hidden_ids"`
}

func CharacSetHiddens(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	params := proute.Params.(*CharacSetHiddensParams)
	c := proute.Json.(*CharacSetHiddensStruct)

	log.Println("c: ", c)

	// transaction begin...
	tx, err := db.DB.Beginx()
	if err != nil {
		userSqlError(w, err)
		return
	}

	// get the user
	_user, ok := proute.Session.Get("user")
	if !ok {
		log.Println("CharacsUpdate: can't get user in session...", _user)
		_ = tx.Rollback()
		return
	}

	user, ok := _user.(model.User)
	if !ok {
		log.Println("CharacsUpdate: can't cast user...", _user)
		_ = tx.Rollback()
		return
	}

	// check if the user is the owner of the project
	count := 0
	tx.Get(&count, "SELECT count(*) FROM project WHERE id = "+strconv.Itoa(params.Project_id)+" AND user_id = "+strconv.Itoa(user.Id))
	if count != 1 {
		log.Println("CharacSetHiddens: user is not the owner...", user, params.Project_id)
		_ = tx.Rollback()
		return
	}

	// delete previous settings
	_, err = tx.Exec(`DELETE FROM project_hidden_characs WHERE charac_id in (
		               WITH RECURSIVE subcharac(id, parent_id, charac_id, project_id) AS (
					    SELECT id, parent_id, phc.charac_id, phc.project_id
						FROM charac LEFT JOIN project_hidden_characs phc ON phc.charac_id = charac.id WHERE id = $1
					   UNION ALL
						SELECT c2.id, c2.parent_id, phc2.charac_id, phc2.project_id
						FROM subcharac AS sc, charac AS c2 LEFT JOIN project_hidden_characs phc2 ON phc2.charac_id = c2.id
						WHERE c2.parent_id = sc.id
					   )
					   SELECT id FROM subcharac WHERE project_id=$2)`, params.Id, params.Project_id)
	if err != nil {
		log.Println(err)
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	for _, id := range c.HiddenIds {
		log.Println("deleting: ", id)
		tx.Exec("INSERT INTO project_hidden_characs (project_id, charac_id) VALUES (" + strconv.Itoa(params.Project_id) + "," + strconv.Itoa(id) + ")")
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

func CharacListCsv(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	params := proute.Params.(*CharacListCsvParams)

	// transaction begin...
	tx, err := db.DB.Beginx()
	if err != nil {
		userSqlError(w, err)
		return
	}

	// get the user
	_user, ok := proute.Session.Get("user")
	if !ok {
		log.Println("CharacsUpdate: can't get user in session...", _user)
		_ = tx.Rollback()
		return
	}

	user, ok := _user.(model.User)
	if !ok {
		log.Println("CharacsUpdate: can't cast user...", _user)
		_ = tx.Rollback()
		return
	}

	answer, err := characsGetTree(w, tx, params.Id, 0, user)

	table := [][]string{}
	table = append(table, []string{
		"IDArkeoGIS",
		"CARAC_NAME",
		"CARAC_LVL1",
		"CARAC_LVL2",
		"CARAC_LVL3",
		"CARAC_LVL4",
		"IdArk",
		"IdPactols",
	})

	lvl0 := answer.CharacTreeStruct
	lvl0Name := "LANGNOTFOUND"
	if name, ok := lvl0.Name[params.Isocode]; ok {
		lvl0Name = name
	}

	for _, lvl1 := range lvl0.Content {
		lvl1Name := "LANGNOTFOUND"
		if name, ok := lvl1.Name[params.Isocode]; ok {
			lvl1Name = name
		}

		if params.Html == 1 && len(lvl1.Ark_id) > 0 {
			lvl1.Ark_id = "<a href=\"" + lvl1.Ark_id + "\">" + lvl1.Ark_id + "</a>"
		}

		table = append(table, []string{
			strconv.Itoa(lvl1.Id),
			lvl0Name,
			lvl1Name,
			"",
			"",
			"",
			lvl1.Ark_id,
			lvl1.Pactols_id,
		})

		for _, lvl2 := range lvl1.Content {
			lvl2Name := "LANGNOTFOUND"
			if name, ok := lvl2.Name[params.Isocode]; ok {
				lvl2Name = name
			}

			if params.Html == 1 && len(lvl2.Ark_id) > 0 {
				lvl2.Ark_id = "<a href=\"" + lvl2.Ark_id + "\">" + lvl2.Ark_id + "</a>"
			}

			table = append(table, []string{
				strconv.Itoa(lvl2.Id),
				lvl0Name,
				lvl1Name,
				lvl2Name,
				"",
				"",
				lvl2.Ark_id,
				lvl2.Pactols_id,
			})

			for _, lvl3 := range lvl2.Content {
				lvl3Name := "LANGNOTFOUND"
				if name, ok := lvl3.Name[params.Isocode]; ok {
					lvl3Name = name
				}

				if params.Html == 1 && len(lvl3.Ark_id) > 0 {
					lvl3.Ark_id = "<a href=\"" + lvl3.Ark_id + "\">" + lvl3.Ark_id + "</a>"
				}

				table = append(table, []string{
					strconv.Itoa(lvl3.Id),
					lvl0Name,
					lvl1Name,
					lvl2Name,
					lvl3Name,
					"",
					lvl3.Ark_id,
					lvl3.Pactols_id,
				})

				for _, lvl4 := range lvl3.Content {
					lvl4Name := "LANGNOTFOUND"
					if name, ok := lvl4.Name[params.Isocode]; ok {
						lvl4Name = name
					}

					if params.Html == 1 && len(lvl4.Ark_id) > 0 {
						lvl4.Ark_id = "<a href=\"" + lvl4.Ark_id + "\">" + lvl4.Ark_id + "</a>"
					}

					table = append(table, []string{
						strconv.Itoa(lvl4.Id),
						lvl0Name,
						lvl1Name,
						lvl2Name,
						lvl3Name,
						lvl4Name,
						lvl4.Ark_id,
						lvl4.Pactols_id,
					})

				}

			}

		}
	}

	// commit...
	err = tx.Commit()
	if err != nil {
		log.Println("commit failed")
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	csvW := csv.NewWriter(w)
	csvW.Comma = ';'
	csvW.WriteAll(table)
	csvW.Flush()
}

/*
func OldCharacListCsv(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	params := proute.Params.(*CharacListCsvParams)
	q := `WITH RECURSIVE
	nodes_cte(id, ark_id, "order", path)
	AS (
	  SELECT id, ark_id, "order", cat.name::::TEXT AS path
	  FROM charac AS ca
	  LEFT JOIN charac_tr cat
		ON ca.id = cat.charac_id
	  LEFT JOIN lang
	  ON cat.lang_isocode = lang.isocode
	  WHERE lang.isocode = :isocode
	  AND ca.id = (
		SELECT ca.id
		FROM charac ca
		LEFT JOIN charac_tr cat
		ON ca.id = cat.charac_id
		LEFT JOIN lang
		ON lang.isocode = cat.lang_isocode
		  WHERE lang.isocode = :isocode
		  AND lower(cat.name) = lower(:name)
		  AND ca.parent_id = 0
	  )
	  UNION ALL
	  SELECT ca.id, ca.ark_id, ca."order", (p.path || ';' || cat.name)
	  FROM
		nodes_cte AS p,
		charac AS ca
	  LEFT JOIN charac_tr cat
		ON ca.id = cat.charac_id
	  LEFT JOIN lang
		ON cat.lang_isocode = lang.isocode
		WHERE lang.isocode = :isocode
		AND ca.parent_id = p.id
	)
	SELECT * FROM nodes_cte AS n ORDER BY n.Order ASC
	`

	if params.Name == "" {
		http.Error(w, "Please provide a charac name in url", 500)
		return
	}
	if params.Isocode == "" {
		http.Error(w, "Please provide an isocode in url", 500)
		return
	}
	list := []struct {
		Id     int
		Ark_id string
		Order  int
		Path   string
	}{}
	stmt, err := db.DB.PrepareNamed(q)
	outp := "IDArkeoGIS;IdArk;Order;CARAC_NAME;CARAC_LVL1;CARAC_LVL2;CARAC_LVL3;CARAC_LVL4\n"
	if err != nil {
		log.Println("error while preparing query", err)
		http.Error(w, "INTERNAL SERVER ERROR", 500)
	}
	err = stmt.Select(&list, params)
	if err != nil {
		log.Println("error in select", err)
		http.Error(w, "INTERNAL SERVER ERROR", 500)
	}
	for _, charac := range list {
		num := 4 - strings.Count(charac.Path, ";")
		if num < 4 {
			outp += strconv.Itoa(charac.Id) + ";" + charac.Ark_id + ";" + strconv.Itoa(charac.Order) + ";" + charac.Path + strings.Repeat(";", num) + "\n"
		}
	}

	if params.Dl != "" {
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	}

	w.Write([]byte(outp))
}
*/
