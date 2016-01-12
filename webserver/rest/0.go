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
	//"fmt"
	//"github.com/lib/pq"
	//db "github.com/croll/arkeogis-server/db"
	//routes "github.com/croll/arkeogis-server/webserver/routes"
	"log"
	"net/http"

	"github.com/emicklei/go-restful"
)

// Rest is the interface to implement for registering new rest routes
type Rest interface {
	Name() string
	Register(container *restful.Container)
}

var rests = []Rest{}

// Register all the routes from this rest package
func Register(container *restful.Container) {
	for _, rest := range rests {
		log.Println("REST init : " + rest.Name())
		rest.Register(container)
	}
}

func register(rest Rest) {
	rests = append(rests, rest)
}

/*
 * errors
 */

/*
type ArkeoError struct {
	HttpCode int
	Name     string
}


ArkkeoErrors := map[string]ArkeoError {
	"DUP"
}
*/

func ArkeoError(w http.ResponseWriter, code string, description string) {
	type ArkeoError struct {
		Code        string
		Description string
	}
	aerr := ArkeoError{code, description}
	j, err := json.Marshal(aerr)
	if err != nil {
		log.Panicln("err in error, marshaling failed: ", err)
	}
	http.Error(w, (string)(j), 409)
}
