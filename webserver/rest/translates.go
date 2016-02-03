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

	translate "github.com/croll/arkeogis-server/translate"
	routes "github.com/croll/arkeogis-server/webserver/routes"
)

type TranslateListParams struct {
	Lang   string `default:"en" min:"2" max:"2"` // todo: sanitize using regular expression
	Side   string `default:"web" enum:"server,web,*"`
	Domain string
}

type TranslateSaveParams struct {
	Lang   string `default:"en" min:"2" max:"2"` // todo: sanitize using regular expression
	Side   string `default:"web" enum:"server,web,*"`
	Domain string
}

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:   "/api/translates",
			Func:   TranslatesSave,
			Method: "PUT",
			Params: reflect.TypeOf(TranslateSaveParams{}),
			Json:   reflect.TypeOf(make(map[string]interface{}, 0)),
			/*
				Permissions: []string{
					"PermTranslatesAdmin",
				},
			*/
		},
		&routes.Route{
			Path:   "/api/translates",
			Func:   TranslatesList,
			Method: "GET",
			Params: reflect.TypeOf(TranslateListParams{}),
		},
	}
	routes.RegisterMultiple(Routes)
}

// TranslatesList List root translations...
func TranslatesList(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Params.(*TranslateListParams)

	trans, err := translate.ReadTranslation(params.Lang, params.Side)
	if err != nil {
		ArkeoError(w, "404", err.Error())
		return
	}

	tree := translate.PlateToTree(trans)

	if params.Domain == "" {
		res := make([]string, len(tree))
		i := 0
		for k, _ := range tree {
			res[i] = k
			i++
		}
		j, err := json.Marshal(res)
		if err != nil {
			log.Fatal("Marshal of lang failed", err)
		}
		log.Printf("j: %s\n", j)
		w.Write(j)
		return
	} else {
		subtree := tree[params.Domain]
		j, err := json.Marshal(subtree)
		if err != nil {
			log.Fatal("Marshal of lang failed", err)
		}
		log.Printf("j: %s\n", j)
		w.Write(j)
		return
	}
	/*
		j := translate.BuildJSON(trans)
		log.Println("j: ", j)
		fmt.Fprint(w, j)
	*/
}

// UserCreate Create a user, see usercreate struct inside this function for json content
func TranslatesSave(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	params := proute.Params.(*TranslateSaveParams)
	newtrans := proute.Json.(*map[string]interface{})

	trans, err := translate.ReadTranslation(params.Lang, params.Side)
	if err != nil {
		ArkeoError(w, "404", err.Error())
		return
	}

	tree := translate.PlateToTree(trans)
	domain := r.FormValue("domain")

	tree[domain] = *newtrans

	err = translate.WriteJSON(tree, params.Lang, params.Side)
	if err != nil {
		fmt.Println("WriteJSON failed: ", err)
		return
	}

	log.Println("theoriquement saved")
}
