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
		Tr sqlx_types.JSONText `db:"tr" json:"tr"`
	}

	chronologies := []row{}

	transquery := model.GetQueryTranslationsAsJSONObject("chronology_tr", "tbl.chronology_id = chronology.id", "", false, "name", "description")
	q := "select chronology_root.*, chronology.*, (" + transquery + ") as tr FROM chronology_root LEFT JOIN chronology ON chronology_root.root_chronology_id = chronology.id order by id"
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
}

// now update recursively
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

	log.Println("c: ", chrono)

	// delete any translations
	_, err = tx.Exec("DELETE FROM chronology_tr WHERE chronology_id = $1", chrono.Id)
	if err != nil {
		return err
	}

	// create a map of translations for name...
	tr := map[string]model.Chronology_tr{}
	for isocode, name := range chrono.Name {
		tr[isocode] = model.Chronology_tr{
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
			tr[isocode] = model.Chronology_tr{
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

func ChronologiesUpdate(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	c := proute.Json.(*ChronologiesUpdateStruct)

	// transaction begin...
	tx, err := db.DB.Beginx()
	if err != nil {
		userSqlError(w, err)
		return
	}

	// boolean create, true if we are creating a totaly new chronology
	var create bool
	if c.Chronology.Id > 0 {
		create = true
		// @TODO: check that you are in group of this chrono when updating one
	} else {
		create = false
	}

	// save recursively this chronology
	err = setChronoRecursive(tx, &c.ChronologyTreeStruct, nil)
	if err != nil {
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	// save the chronology_root row
	c.Chronology_root.Root_chronology_id = c.Chronology.Id
	if create {
		err = c.Chronology_root.Create(tx)
		if err != nil {
			userSqlError(w, err)
			_ = tx.Rollback()
			return
		}
	} else {
		err = c.Chronology_root.Update(tx)
		if err != nil {
			userSqlError(w, err)
			_ = tx.Rollback()
			return
		}
	}

	// commit...
	err = tx.Commit()
	if err != nil {
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

}
