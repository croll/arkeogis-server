/* ArkeoGIS - The Arkeolog Geographical Information Server Program
 * Copyright (C) 2015-2016 CROLL SAS
 *
 * Authors :
 *  Christophe Beveraggi <beve@croll.fr>
 *  Nicolas Dimitrijevic <nicolas@croll.fr>
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

package translate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"sort"
	"strings"

	config "github.com/croll/arkeogis-server/config"
	db "github.com/croll/arkeogis-server/db"
)

var translations map[string]map[string]string // [lang][key]

func init() {
	translations = make(map[string]map[string]string)
}

// parse one level of the json translation file, this func will also parse sublevels by calling
func readLevel(jspath string, dec *json.Decoder, tr map[string]string) {
	if jspath == "" {
		t, err := dec.Token()
		if err != nil {
			log.Fatal("translate error 1: ", err)
		}
		if fmt.Sprintf("%T:%v", t, t) != "json.Delim:{" {
			log.Fatal("translate error 1.1: bad translaton file format")
		}
		//fmt.Printf("%s> %T:%v\n", jspath, t, t)
	}

	for dec.More() {
		t1, err1 := dec.Token()
		t2, err2 := dec.Token()

		if err1 != nil {
			log.Fatal("translate error 2: ", err1)
		}

		if err2 != nil {
			log.Fatal("translate error 3: ", err2)
		}

		//fmt.Printf("%s: %T:%v => %T:%v\n", jspath, t1, t1, t2, t2)

		var newpath string
		if jspath == "" {
			newpath = fmt.Sprintf("%v", t1)
		} else {
			newpath = fmt.Sprintf("%s.%v", jspath, t1)
		}
		if fmt.Sprintf("%T:%v", t2, t2) == "json.Delim:{" {
			readLevel(newpath, dec, tr)
		} else {
			tr[newpath] = fmt.Sprintf("%v", t2)
			//fmt.Printf("%s => %v\n", newpath, t2)
		}
	}

	t, err := dec.Token()
	if err != nil {
		log.Fatal("translate error 3=4: ", err)
	}
	if fmt.Sprintf("%T:%v", t, t) != "json.Delim:}" {
		log.Fatal("translate error 4.1: bad translaton file format")
	}
	//fmt.Printf("%s< %T:%v\n", jspath, t, t)

}

// return a file path to the json file lang. prodonly force to use dist files even in dev mode
func makeFilePath(lang string, side string, prodonly bool) (filename string, err error) {
	var distpath string
	var webpath string

	if prodonly {
		distpath = config.DistPath
		webpath = config.WebPath
	} else {
		distpath = config.CurDistPath
		webpath = config.CurWebPath
	}

	if side == "server" {
		filename = path.Join(distpath, "languages", lang+".json")
	} else if side == "web" {
		filename = path.Join(webpath, "languages", lang+".json")
	} else {
		err = errors.New("Bad side")
		return
	}
	return
}

// readTranslation will load a translation file (json format).
// side is "server" or "web"
func readTranslation(lang string, side string, res map[string]string) error {
	filename, err := makeFilePath(lang, side, true)
	if err != nil {
		return err
	}

	fmt.Println("Reading translation file " + filename)
	c, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println("Error while loading lang file '"+lang+"'", err)
		return err
	}

	dec := json.NewDecoder(bytes.NewReader(c))

	readLevel("", dec, res)

	return err
}

// ReadTranslation will load a translation file (json format).
// side is "server", "web" or "*" meaning server+web merged
func ReadTranslation(lang string, side string) (res map[string]string, err error) {
	res = make(map[string]string)

	switch side {
	case "server", "web":
		err = readTranslation(lang, side, res)
		break
	case "*":
		err = readTranslation(lang, "web", res)
		if err == nil {
			err = readTranslation(lang, "server", res)
		}
	}

	return res, err
}

// load server translations to be useable in server context
func loadServerTranslation(lang string) (res map[string]string, err error) {
	var ok bool
	if res, ok = translations[lang]; !ok {
		res, err = ReadTranslation(lang, "server")
		translations[lang] = res
	}
	return
}

// T will return the translation represented by the key string and the wanted language. Translation can use fmt formats
func T(lang string, key string, a ...interface{}) string {
	trans, err := loadServerTranslation(lang)
	if err != nil {
		return key
	}

	// take the translatiion if it exists and apply fmt
	if r, ok := trans[key]; ok {
		return fmt.Sprintf(r, a...)
	} else {
		return key
	}
}

// MergeIn will merge oldTrans in newTrans, only key in newTrans will exists, others from oldTrans will not be imported
func MergeIn(newTrans map[string]string, oldTrans map[string]string) {
	for key, _ := range newTrans {
		if oldstring, ok := oldTrans[key]; ok {
			newTrans[key] = oldstring
		}
	}
}

func buildElem(key string, subtree map[string]interface{}, splits []string, v string) {
	if len(splits) > 1 { // dir
		key := splits[0]
		if subsubtree, ok := subtree[key]; ok {
			if csubsubtree, ok := subsubtree.(map[string]interface{}); ok {
				buildElem(key, csubsubtree, splits[1:], v)
			} else {
				log.Fatal("Found two conflicting values for key (one is a string, one is an object): ", key, " striong : ", subsubtree)
			}
		} else {
			subsubtree := make(map[string]interface{})
			subtree[key] = subsubtree
			buildElem(key, subsubtree, splits[1:], v)
		}
	} else if len(splits) == 1 { // leaf
		key := splits[0]
		subtree[key] = v
	}
}

// PlateToTree convert a plate map to a tree map
func PlateToTree(trans map[string]string) map[string]interface{} {
	// for sorting
	mk := make([]string, len(trans))
	i := 0
	for k, _ := range trans {
		mk[i] = k
		i++
	}
	sort.Strings(mk)

	tree := make(map[string]interface{})

	for _, key := range mk {
		v := trans[key]
		splits := strings.Split(key, ".")
		buildElem(key, tree, splits, v)
	}

	return tree
}

// BuildJSON return a JSON string representing the map of strings
func BuildJSON(trans map[string]string) string {
	tree := PlateToTree(trans)

	res, err := json.MarshalIndent(tree, "", "\t")
	if err != nil {
		log.Fatal("Marshal of lang failed", err)
	}

	return fmt.Sprintf("%s", res)
}

// WriteJSON writes json where it should be, using map of strings
func WriteJSON(trans map[string]interface{}, lang string, side string) (err error) {
	var filename string
	filename, err = makeFilePath(lang, side, false)
	if err != nil {
		return
	}

	j, err := json.MarshalIndent(trans, "", "\t")
	if err != nil {
		log.Fatal("Marshal of lang failed", err)
		return err
	}

	log.Println("writing file : ", filename)
	err = ioutil.WriteFile(filename, ([]byte)(j), 0777)

	return
}

// GetQueryTranslationsAsJSON load translations from database
func GetQueryTranslationsAsJSON(tableName, where, wrapTo string, fields ...string) string {
	var f = "*"
	if len(fields) > 0 {
		f = strings.Join(fields, ", tbl.")
	}
	return db.AsJSON("SELECT tbl."+f+", la.iso_code FROM "+tableName+" tbl LEFT JOIN lang la ON tbl.lang_id = la.id WHERE "+where, true, wrapTo, true)
}

// GetQueryTranslationsAsJSONObject load translations from database
func GetQueryTranslationsAsJSONObject(tableName, where string, wrapTo string, noBrace bool, fields ...string) (jsonQuery string, err error) {

	jsonQuery = "SELECT '"

	if noBrace == false {
		jsonQuery += "{"
	}
	if wrapTo != "" {
		jsonQuery += "\"" + wrapTo + "\": {' || "
	} else {
		jsonQuery += "' || "

	}
	numFields := len(fields)
	if numFields == 0 {
		return "", errors.New("GetQueryTranslationsAsJSONObject: You have to provide at least one field")
	}
	for k, f := range fields {
		jsonQuery += "'\"" + f + "\": ' || json_object_agg(la.iso_code, tbl." + f + ")"
		if k < numFields-1 {
			jsonQuery += " || ',' || "
		}
	}
	if wrapTo != "" {
		jsonQuery += " || '}'"
	}
	if noBrace == false {
		jsonQuery += " || '}'"
	}
	jsonQuery += " FROM " + tableName + " tbl LEFT JOIN lang la ON tbl.lang_id = la.id WHERE " + where
	fmt.Println(jsonQuery)
	return
}
