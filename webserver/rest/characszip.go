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
	"fmt"
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"net/http"

	db "github.com/croll/arkeogis-server/db"
	"github.com/croll/arkeogis-server/model"
	"github.com/lib/pq"

	routes "github.com/croll/arkeogis-server/webserver/routes"
)

const CharacCsvColumnId = 0
const CharacCsvColumnName = 1
const CharacCsvColumnPath0 = 1
const CharacCsvColumnPath1 = 2
const CharacCsvColumnPath2 = 3
const CharacCsvColumnPath3 = 4
const CharacCsvColumnPath4 = 5
const CharacCsvColumnArkId = 6
const CharacCsvColumnPactolsId = 7
const CharacCsvColumnAatId = 8

type CharacsZipUpdateStruct struct {
	CharacId   int    `json:"characId"`
	ZipContent []byte `json:"zipContent"`
}

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:        "/api/characszip",
			Description: "Create/Update a charac",
			Func:        CharacsUpdateZip,
			Method:      "POST",
			Json:        reflect.TypeOf(CharacsZipUpdateStruct{}),
			Permissions: []string{
				"request map",
			},
		},
	}
	routes.RegisterMultiple(Routes)
}

// CharacsUpdateZip Create/Update all characs from a zip file containing multiple languages
func CharacsUpdateZip(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	// get the post
	c := proute.Json.(*CharacsZipUpdateStruct)

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
	/*
		var create bool
		if c.CharacId > 0 {
			create = false
			// @TODO: check that you are in group of this charac when updating one
		} else {
			create = true
		}
	*/

	// search the characroot to verify permissions
	characroot := model.Charac_root{
		Root_charac_id: c.CharacId,
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

	/**
	*
	* Job Here
	*
	**/

	/**** unzip ****/

	contents, err := readZipCharacs(c.ZipContent)
	if err != nil {
		log.Println("CharacsUpdateZip: can't load zip...", err)
		//fmt.Println("zip content : ", c.ZipContent)
		_ = tx.Rollback()
		routes.FieldError(w, "json.zipcontent", "zipcontent", err.Error())
		return
	}

	if len(contents) != 4 {
		log.Println("CharacsUpdateZip: contents != 4", len(contents))
		_ = tx.Rollback()
		routes.FieldError(w, "json.zipcontent", "zipcontent", "CHARAC.FIELD_CONTENT.T_BADCOUNT")
		return
	}

	//fmt.Println("decoded[0]: ", contents["fr"].Decoded)

	/************************************/

	answer, err := characsGetTree(w, tx, c.CharacId, 0, false, user)
	if err != nil {
		log.Println("CharacsUpdateZip: characsGetTree failed...", err)
		_ = tx.Rollback()
		routes.FieldError(w, "json.zipcontent", "internal error", err.Error())
		return
	}

	err = csvzipDoTheMix(answer, contents)
	if err != nil {
		log.Println("CharacsUpdateZip: csvzipDoTheMix failed...", err)
		_ = tx.Rollback()
		characzipError(w, err)
		return
	}

	// save recursively this charac
	err = setCharacRecursive(tx, &answer.CharacTreeStruct, nil)
	if err != nil {
		userSqlError(w, err)
		_ = tx.Rollback()
		return
	}

	// commit...
	err = tx.Commit()
	if err != nil {
		log.Println("commit failed")
		characzipError(w, err)
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

func characzipError(w http.ResponseWriter, err error) {
	log.Printf("paf: %#v\n", err)
	if pgerr, ok := err.(*pq.Error); ok {
		log.Printf("pgerr: %#v\n", pgerr.Code.Name())
		switch pgerr.Code.Name() {
		case "foreign_key_violation":
			routes.FieldError(w, "json.zipcontent", "error", "Unable to update or delete a row, the row may be in use by a database")
			break
		default:
			log.Printf("unhandled postgresql error ! : %#v\n", pgerr)
			routes.FieldError(w, "json.zipcontent", "database error", err.Error())
		}
	} else {
		log.Println("not a postgresql error !", err)
		routes.FieldError(w, "json.zipcontent", "internal error", err.Error())
	}
}


type ZipContent struct {
	IsoCode string
	CSV     string
	Decoded [][]string
}

func readZipCharacs(zipContent []byte) (map[string]ZipContent, error) {
	reader, err := zip.NewReader(bytes.NewReader(zipContent), int64(len(zipContent)))
	if err != nil {
		log.Println("CharacsUpdateZip: can't load zip...", err)
		return nil, err
	}

	zipContents := map[string]ZipContent{}

	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		content := ZipContent{}

		re := regexp.MustCompile(`-([a-z]{2})\.csv$`)
		matches := re.FindStringSubmatch(file.Name)
		if len(matches) != 2 {
			return nil, errors.New("Bad file name inside zip. Must be for example : \"Something-en.csv\" for english, etc.")
		}
		content.IsoCode = matches[1]

		fileReader, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer fileReader.Close()

		buf := new(strings.Builder)
		_, err = io.Copy(buf, fileReader)
		if err != nil {
			return nil, err
		}
		content.CSV = buf.String()
		content.Decoded, err = csvDecodeCharacs(content.CSV)
		if err != nil {
			return nil, err
		}

		// strip every fields
		for y, _ := range content.Decoded {
			for x, _ := range content.Decoded[y] {
				content.Decoded[y][x] = strings.TrimSpace(content.Decoded[y][x])
			}
		}

		zipContents[content.IsoCode] = content
	}

	return zipContents, nil

}

func csvDecodeCharacs(in string) ([][]string, error) {
	r := csv.NewReader(strings.NewReader(in))
	r.Comma = ';'
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	return records, nil
}

func csvzipSearchCharacByPath(elem *CharacTreeStruct, lang string, path []string, levelsize int) (*CharacTreeStruct, int) {

	if name, ok := elem.Name[lang]; ok {
		if name == path[0] {
			if len(path) > 1 {
				for i, _ := range elem.Content {
					search, newlevelsize := csvzipSearchCharacByPath(&elem.Content[i], lang, path[1:], levelsize+1)
					if search != nil {
						return search, newlevelsize
					}
				}
				return nil, 0
			} else {
				return elem, levelsize + 1
			}
		}
	}

	return nil, 0
}

func csvzipSearchCharacByID(elem *CharacTreeStruct, id int, levelsize int) (*CharacTreeStruct, int) {
	if elem.Id == id {
		return elem, levelsize + 1
	}

	for i, _ := range elem.Content {
		found, foundlevelsize := csvzipSearchCharacByID(&elem.Content[i], id, levelsize+1)
		if found != nil {
			return found, foundlevelsize
		}
	}

	return nil, 0
}

/*
func printC(str string, c *CharacTreeStruct) {
	log.Println(str)
	for i, elem := range c.Content {
		log.Printf("%d : id %d\n", i, elem.Id)
	}
}
*/

func csvzipRemoveCharacByID(elem *CharacTreeStruct, id int, parent *CharacTreeStruct, parentidx int) (*CharacTreeStruct) {
	if elem.Id == id {
		copy := *elem
		// remove from parent
		parent.Content = append(parent.Content[:parentidx], parent.Content[parentidx+1:]...)
		return &copy
	}

	for i, _ := range elem.Content {
		found := csvzipRemoveCharacByID(&elem.Content[i], id, elem, i)
		if found != nil {
			return found
		}
	}

	return nil
}

func printSlice(s []string) {
	fmt.Printf("len=%d cap=%d %v\n", len(s), cap(s), s)
}

func csvzipDoTheMix(actual *CharacsUpdateStruct, newcontent map[string]ZipContent) error {

	//fmt.Println("actual: ", actual)

	firstlang := "en"
	for lang, _ := range newcontent {
		firstlang = lang
		break
	}
	totalcount := len(newcontent[firstlang].Decoded)

	toremove := CharacTreeStruct{}

	for linenum := 1; linenum < totalcount; linenum++ {
		ids := map[string]string{}
		arkIds := map[string]string{}
		pactolsIds := map[string]string{}
		aatIds := map[string]string{}
		paths := map[string][]string{}

		for lang, zipContent := range newcontent {
			if (len(zipContent.Decoded[linenum]) != (CharacCsvColumnAatId + 1)) {
				fmt.Println("line : ", zipContent.Decoded[linenum])
				printSlice(zipContent.Decoded[linenum])
				return errors.New(
					"Line " + strconv.Itoa(linenum) +
					" for language " + lang +
					" have " + strconv.Itoa(len(zipContent.Decoded[linenum])) +
					" columns instead of " + strconv.Itoa(CharacCsvColumnAatId + 1))
			}

			ids[lang] = zipContent.Decoded[linenum][CharacCsvColumnId]
			arkIds[lang] = zipContent.Decoded[linenum][CharacCsvColumnArkId]
			pactolsIds[lang] = zipContent.Decoded[linenum][CharacCsvColumnPactolsId]
			aatIds[lang] = zipContent.Decoded[linenum][CharacCsvColumnAatId]
			path := []string{}

			for y, val := range zipContent.Decoded[linenum][CharacCsvColumnPath0 : CharacCsvColumnPath4+1] {
				if len(val) == 0 {
					break
				}
				path = append(path, zipContent.Decoded[linenum][CharacCsvColumnPath0+y])
			}
			paths[lang] = path
		}

		// check if every ids are identical
		for _, id := range ids {
			if id != ids[firstlang] {
				return errors.New("IDs on line " + strconv.Itoa(linenum) + " are not identical on all languages")
			}
		}

		// check if every arkIds are identical
		for _, arkId := range arkIds {
			if arkId != arkIds[firstlang] {
				return errors.New("arkIds on line " + strconv.Itoa(linenum) + " are not identical on all languages")
			}
		}

		// check if every pactolsIds are identical
		for _, pactolsId := range pactolsIds {
			if pactolsId != pactolsIds[firstlang] {
				return errors.New("pactolsIds on line " + strconv.Itoa(linenum) + " are not identical on all languages : '" + pactolsIds[firstlang] + "' != '" + pactolsId + "'")
			}
		}

		// check if every arkIds are identical
		for _, aatId := range aatIds {
			if aatId != aatIds[firstlang] {
				return errors.New("aatIds on line " + strconv.Itoa(linenum) + " are not identical on all languages : '" + aatIds[firstlang] + "' != '" + aatId + "'")
			}
		}

		// check if every paths heve same size
		for _, path := range paths {
			if len(path) != len(paths[firstlang]) {
				return errors.New("characs on line " + strconv.Itoa(linenum) + " do not have same levels on all languages")
			}
		}

		// do the update/insert
		if len(ids[firstlang]) > 0 {
			// we have an id, so it's an update action
			id, err := strconv.Atoi(ids[firstlang])
			if err != nil {
				return errors.New("bad ID on line " + strconv.Itoa(linenum))
			}

			// if id < 0, this is a delete action
			if id < 0 { // DELETE ACTION
				found := csvzipRemoveCharacByID(&actual.CharacTreeStruct, -id, nil, -1)
				if found != nil {
					log.Println("removed " + strconv.Itoa(found.Id))
					toremove.Content = append(toremove.Content, *found)
				} else {
					leaffound := csvzipRemoveCharacByID(&toremove, -id, nil, -1)
					if leaffound != nil {
						log.Println("removed2 " + strconv.Itoa(leaffound.Id))
						toremove.Content = append(toremove.Content, *leaffound)
					} else {
						return errors.New("characs on line " + strconv.Itoa(linenum) + " with id "+strconv.Itoa(-id)+" was not found for removing")
					}
				}
				//if found == nil {
				//	return errors.New("characs on line " + strconv.Itoa(linenum) + " with id "+strconv.Itoa(-id)+" was not found for removing")
				//}
			} else { // UPDATE ACTION
				elem, levelsize := csvzipSearchCharacByID(&actual.CharacTreeStruct, id, 0)
				if elem != nil && levelsize != len(paths[firstlang]) {
					return errors.New("characs on line " + strconv.Itoa(linenum) + " have a different level count from what exists actually in database " + strconv.Itoa(len(paths[firstlang])) + " != " + strconv.Itoa(levelsize))
				}
				if elem == nil {
					return errors.New("characs on line " + strconv.Itoa(linenum) + " with id "+strconv.Itoa(id)+" was not found for updating")
				}
	
				elem.Charac.Order = linenum * 10
				elem.Charac.Ark_id = arkIds[firstlang]
				elem.Charac.Pactols_id = pactolsIds[firstlang]
				elem.Charac.Aat_id = aatIds[firstlang]
	
				for lang, _ := range newcontent {
					if _, ok := elem.Name[lang]; ok {
	
						if elem.Name[lang] != paths[lang][len(paths[lang])-1] {
							//fmt.Println("update["+lang+"] : ", elem.Name[lang], " => ", paths[lang][len(paths[lang])-1])
						}
	
						elem.Name[lang] = paths[lang][len(paths[lang])-1] // this is the update
					}
				}
			}
		} else {
			// we no not have an id, so it's an insert action

			// create a new element
			subelem := CharacTreeStruct{}
			subelem.Charac.Order = linenum * 10
			subelem.Charac.Ark_id = arkIds[firstlang]
			subelem.Charac.Pactols_id = pactolsIds[firstlang]
			subelem.Charac.Aat_id = aatIds[firstlang]

			subelem.Name = map[string]string{}
			subelem.Description = map[string]string{}

			for lang, _ := range newcontent {
				subelem.Name[lang] = paths[lang][len(paths[lang])-1]
				subelem.Description[lang] = ""
			}

			if len(paths[firstlang]) > 1 {
				// search the parent
				parent, levelsize := csvzipSearchCharacByPath(&actual.CharacTreeStruct, firstlang, paths[firstlang][:len(paths[firstlang])-1], 0)

				if parent != nil && levelsize != len(paths[firstlang]) { // found
					subelem.Charac.Parent_id = parent.Id
				} else {
					return errors.New("Parent element of line " + strconv.Itoa(linenum) + " not found")
				}

				//fmt.Println("INSERT: ", subelem)

				parent.Content = append(parent.Content, subelem)
			} else {

				//fmt.Println("INSERT ROOT: ", subelem)

				actual.Content = append(actual.Content, subelem)
			}

		}

	}

	// check that toremove does not contain leafs anymore
	for _, elem := range toremove.Content {
		if len(elem.Content) > 0 {
			return errors.New("You are trying to remove a line without removing all it's leafs. id : " + strconv.Itoa(elem.Id))
		}
	}

	return nil // no error !
}
