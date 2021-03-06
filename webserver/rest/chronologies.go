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

	"encoding/csv"
	"net/http"

	db "github.com/croll/arkeogis-server/db"
	"github.com/croll/arkeogis-server/model"
	"github.com/croll/arkeogis-server/translate"

	routes "github.com/croll/arkeogis-server/webserver/routes"
	"github.com/jmoiron/sqlx"
	sqlx_types "github.com/jmoiron/sqlx/types"
)

type ChronologyGetParams struct {
	Id     int `min:"1" error:"Chronology Id is mandatory"`
	Active bool
}

type ChronologyListCsvParams struct {
	Isocode string `json:"isocode"`
	Id      int    `json:"id"`
	Dl      string `json:"dl"`
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
			Params:      reflect.TypeOf(ChronologiesRootsParams{}),
		},
		&routes.Route{
			Path:        "/api/chronologies",
			Description: "Create/Update a chronologie",
			Func:        ChronologiesUpdate,
			Method:      "POST",
			Json:        reflect.TypeOf(ChronologiesUpdateStruct{}),
			Permissions: []string{
				"user can edit some chronology",
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
				"user can edit some chronology",
			},
			Params: reflect.TypeOf(ChronologyGetParams{}),
		},
		&routes.Route{
			Path:        "/api/chronologies/csv",
			Func:        ChronologiesListCsv,
			Description: "Get a chronologie as csv",
			Params:      reflect.TypeOf(ChronologyListCsvParams{}),
			Method:      "GET",
			Permissions: []string{},
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
	// fmt.Println("q: ", q)
	err := db.DB.Select(&chronologies, q)
	// fmt.Println("chronologies: ", chronologies)
	if err != nil {
		fmt.Println("err: ", err)
		return
	}

	j, err := json.Marshal(chronologies)
	w.Write(j)
}

// ChronologiesRootsStruct holds get params passed to ChronologiesRoots
type ChronologiesRootsParams struct {
	Bounding_box string
	Active       bool `json:"active"`
	Start_date   int  `json:"start_date"`
	End_date     int  `json:"end_date"`
	Check_dates  bool `json:"check_dates"`
}

// ChronologiesRoots write all root chronologies in all languages
func ChronologiesRoots(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	type row struct {
		model.Chronology_root
		model.Chronology
		Name        map[string]string `json:"name"`
		Description map[string]string `json:"description"`
		//UsersInGroup []model.User      `json:"users_in_group" ignore:"true"` // read-only, used to display users of the group
		Author model.User `json:"author" ignore:"true"` // read-only, used to display users of the group
	}

	// get the params
	params := proute.Params.(*ChronologiesRootsParams)

	// get the user logged
	_user, _ := proute.Session.Get("user")
	user := _user.(model.User)

	chronologies := []*row{}
	returnedChronologies := []*row{}

	// transaction begin...
	tx, err := db.DB.Beginx()
	if err != nil {
		userSqlError(w, err)
		return
	}

	// load all roots yes condition is always true
	q := "SELECT *,ST_AsGeoJSON(geom) as geom FROM chronology_root WHERE 1 = 1"

	if params.Bounding_box != "" {
		q += " AND (ST_Contains(ST_GeomFromGeoJSON(:bounding_box), geom::::geometry) OR ST_Contains(geom::::geometry, ST_GeomFromGeoJSON(:bounding_box)) OR ST_Overlaps(ST_GeomFromGeoJSON(:bounding_box), geom::::geometry))"
	}

	viewUnpublished, err := user.HavePermissions(tx, "manage all databases")
	if err != nil {
		userSqlError(w, err)
		return
	}

	if params.Active || !viewUnpublished {
		q += " AND active = 't'"
	}

	stmt, err := db.DB.PrepareNamed(q)

	if err != nil {
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	err = stmt.Select(&chronologies, params)

	if err != nil {
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	// load all root chronologies
	for _, chrono := range chronologies {
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
		chrono.Name = model.MapSqlTranslations(tr, "Lang_isocode", "Name")
		chrono.Description = model.MapSqlTranslations(tr, "Lang_isocode", "Description")

		// get the author user
		chrono.Author.Id = chrono.Chronology_root.Author_user_id
		err = chrono.Author.Get(tx)
		chrono.Author.Password = ""
		if err != nil {
			userSqlError(w, err)
			_ = tx.Rollback()
			return
		}

		// check if chronology is in requested date bounds
		if params.Check_dates {
			if chrono.Start_date >= params.Start_date && chrono.End_date <= params.End_date {
				returnedChronologies = append(returnedChronologies, chrono)
			}
		} else {
			returnedChronologies = append(returnedChronologies, chrono)
		}
	}

	// commit...
	err = tx.Commit()
	if err != nil {
		log.Println("commit failed")
		_ = tx.Rollback()
		return
	}

	j, _ := json.Marshal(returnedChronologies)
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
	Author       model.User   `json:"author" ignore:"true"`         // read-only, used to display users of the group
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
	_user, _ := proute.Session.Get("user")
	user := _user.(model.User)

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
		chronoroot.Author_user_id = c.Chronology_root.Author_user_id
		//chronoroot.Admin_group_id = c.Chronology_root.Admin_group_id
		chronoroot.Cached_langs = c.Chronology_root.Cached_langs

		err = chronoroot.Update(tx)
		if err != nil {
			log.Println("chronoroot update failed")
			userSqlError(w, err)
			_ = tx.Rollback()
			return
		}
	}

	answer, err := chronologiesGetTree(tx, c.Id, user)
	if err != nil {
		_ = tx.Rollback()
		userSqlError(w, err)
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

func chronologiesGetTree(tx *sqlx.Tx, id int, user model.User) (answer *ChronologiesUpdateStruct, err error) {

	// answer structure that will be printed when everything is done
	answer = &ChronologiesUpdateStruct{}

	// get the chronology_root row
	answer.Chronology_root.Root_chronology_id = id
	err = answer.Chronology_root.Get(tx)
	if err != nil {
		return nil, err
	}

	// get the chronology (root)
	answer.Chronology.Id = id
	err = answer.Chronology.Get(tx)
	if err != nil {
		return nil, err
	}

	// now get the chronology translations and all childrens
	err = getChronoRecursive(tx, &answer.ChronologyTreeStruct)
	if err != nil {
		return nil, err
	}

	// get users of the chrono group
	group := model.Group{
		Id: answer.Chronology_root.Admin_group_id,
	}
	err = group.Get(tx)
	if err != nil {
		return nil, err
	}
	answer.UsersInGroup, err = group.GetUsers(tx)
	if err != nil {
		return nil, err
	}

	for i := range answer.UsersInGroup {
		answer.UsersInGroup[i].Password = ""
	}

	// get the author user
	answer.Author.Id = answer.Chronology_root.Author_user_id
	err = answer.Author.Get(tx)
	answer.Author.Password = ""
	if err != nil {
		return nil, err
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
	_user, _ := proute.Session.Get("user")
	user := _user.(model.User)

	answer, err := chronologiesGetTree(tx, params.Id, user)
	if err != nil {
		_ = tx.Rollback()
		userSqlError(w, err)
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

	if params.Active && !answer.Active {
		routes.FieldError(w, "Active", "Active", "CHRONO.SERVER_ERROR.T_NOT_ACTIVE")
		return
	}

	j, err := json.Marshal(answer)
	if err != nil {
		log.Println("marshal failed: ", err)
		return
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
	answer, err := chronologiesGetTree(tx, params.Id, user)
	if err != nil {
		_ = tx.Rollback()
		userSqlError(w, err)
		return
	}

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

func ChronologiesListCsv(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	params := proute.Params.(*ChronologyListCsvParams)

	// fmt.Println("params : ", params)

	// transaction begin...
	tx, err := db.DB.Beginx()
	if err != nil {
		userSqlError(w, err)
		return
	}

	// get the user
	_user, _ := proute.Session.Get("user")
	user := _user.(model.User)

	answer, err := chronologiesGetTree(tx, params.Id, user)
	if err != nil {
		_ = tx.Rollback()
		userSqlError(w, err)
		return
	}

	// fmt.Println("answer: ", answer)

	// commit...
	err = tx.Commit()
	if err != nil {
		log.Println("commit failed")
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	if params.Dl != "" {
		w.Header().Set("Content-Type", "text/csv")
	}

	csvwriter := csv.NewWriter(w)
	csvwriter.Comma = ';'
	csvwriter.Write([]string{
		translate.T(params.Isocode, "CHRONODITOR.CSVEXPORT.T_PERIOD_NAME_L1"),
		translate.T(params.Isocode, "CHRONODITOR.CSVEXPORT.T_PERIOD_START_L1"),
		translate.T(params.Isocode, "CHRONODITOR.CSVEXPORT.T_PERIOD_END_L1"),
		translate.T(params.Isocode, "CHRONODITOR.CSVEXPORT.T_PERIOD_NAME_L2"),
		translate.T(params.Isocode, "CHRONODITOR.CSVEXPORT.T_PERIOD_START_L2"),
		translate.T(params.Isocode, "CHRONODITOR.CSVEXPORT.T_PERIOD_END_L2"),
		translate.T(params.Isocode, "CHRONODITOR.CSVEXPORT.T_PERIOD_NAME_L3"),
		translate.T(params.Isocode, "CHRONODITOR.CSVEXPORT.T_PERIOD_START_L3"),
		translate.T(params.Isocode, "CHRONODITOR.CSVEXPORT.T_PERIOD_END_L3"),
		translate.T(params.Isocode, "CHRONODITOR.CSVEXPORT.T_PERIOD_NAME_L4"),
		translate.T(params.Isocode, "CHRONODITOR.CSVEXPORT.T_PERIOD_START_L4"),
		translate.T(params.Isocode, "CHRONODITOR.CSVEXPORT.T_PERIOD_END_L4"),
	})
	row := make([]string, 12)
	recurseprint(&answer.ChronologyTreeStruct, csvwriter, &row, params.Isocode, 0, 0)
	csvwriter.Write(row)

	csvwriter.Write([]string{
		translate.T(params.Isocode, "CHRONODITOR.CSVEXPORT.T_NAME") + ": " + answer.Name[params.Isocode],
	})

	csvwriter.Write([]string{
		translate.T(params.Isocode, "CHRONODITOR.CSVEXPORT.T_DESCRIPTION") + ": " + answer.Description[params.Isocode],
	})

	csvwriter.Write([]string{
		translate.T(params.Isocode, "CHRONODITOR.CSVEXPORT.T_CREDITS") + ": " + answer.Credits,
	})

	csvwriter.Flush()

	//w.Write([]byte(outp))
}

func dateToHuman(date int) int {
	if date <= 0 {
		return date - 1
	} else {
		return date
	}
}

func recurseprint(elem *ChronologyTreeStruct, csvwriter *csv.Writer, row *[]string, isocode string, level int, index int) {
	if index > 0 {
		csvwriter.Write(*row)
		*row = make([]string, 12)
	}
	if level > 0 {
		(*row)[(level-1)*3+0] = elem.Name[isocode]
		(*row)[(level-1)*3+1] = strconv.Itoa(dateToHuman(elem.Start_date))
		(*row)[(level-1)*3+2] = strconv.Itoa(dateToHuman(elem.End_date))
	}
	for i, e := range elem.Content {
		recurseprint(&e, csvwriter, row, isocode, level+1, i)
	}
}

/*
func ChronologiesListCsv(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	params := proute.Params.(*ChronologyListCsvParams)
	q := `WITH RECURSIVE nodes_cte(id, path) AS (
		   SELECT id, chrono_tr.name::::TEXT AS path FROM chronology AS chrono
		   LEFT JOIN chronology_tr chrono_tr ON chrono.id = chrono_tr.chronology_id
		   LEFT JOIN lang ON chrono_tr.lang_isocode = lang.isocode
		   WHERE lang.isocode = :isocode AND chrono.id = (
			SELECT chrono.id FROM chronology chrono
			LEFT JOIN chronology_tr chrono_tr ON chrono.id = chrono_tr.chronology_id
			LEFT JOIN lang ON lang.isocode = chrono_tr.lang_isocode
			WHERE lang.isocode = :isocode AND lower(chrono_tr.name) = lower(:name) AND chrono.parent_id = 0
		   )
		   UNION ALL
		   SELECT chrono.id, (p.path || ';' || chrono_tr.name)
		   FROM nodes_cte AS p, chronology AS chrono
		   LEFT JOIN chronology_tr chrono_tr ON chrono.id = chrono_tr.chronology_id
		   LEFT JOIN lang ON chrono_tr.lang_isocode = lang.isocode
		   WHERE lang.isocode = :isocode AND chrono.parent_id = p.id
		  )
		  SELECT * FROM nodes_cte AS n ORDER BY n.id ASC;`
	if params.Name == "" {
		http.Error(w, "Please provide a chronology name in url", 500)
		return
	}
	if params.Isocode == "" {
		http.Error(w, "Please provide an isocode in url", 500)
		return
	}
	list := []struct {
		Id   int
		Path string
	}{}
	stmt, err := db.DB.PrepareNamed(q)
	outp := ""
	if err != nil {
		log.Println(err)
		http.Error(w, "INTERNAL SERVER ERROR", 500)
	}
	err = stmt.Select(&list, params)
	if err != nil {
		log.Println(err)
		http.Error(w, "INTERNAL SERVER ERROR", 500)
	}
	for _, chronology := range list {
		num := 4 - strings.Count(chronology.Path, ";")
		if num < 4 {
			outp += chronology.Path + strings.Repeat(";", num) + "\n"
		}
	}

	if params.Dl != "" {
		w.Header().Set("Content-Type", "text/csv")
	}

	w.Write([]byte(outp))
}
*/
