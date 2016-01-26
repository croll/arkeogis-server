/* ArkeoGIS - The Geographic Information System for Archaeologists
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

package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	//"strconv"
	"strings"

	"mime"
	"mime/multipart"

	db "github.com/croll/arkeogis-server/db"
	"github.com/croll/arkeogis-server/model"
	"github.com/croll/arkeogis-server/webserver/filters"
	session "github.com/croll/arkeogis-server/webserver/session"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

// Route structure that is used for registering a new Arkeogis Route
type Route struct {
	Path         string
	Func         func(rw http.ResponseWriter, r *http.Request, o interface{}, s *session.Session)
	Method       string
	Queries      []string
	Json         reflect.Type
	Permissions  []string
	ParamFilters []filters.Filter
}

// MuxRouter is the gorilla mux router initialized here for Arkeogis
var MuxRouter *mux.Router

func init() {
	// router
	MuxRouter = mux.NewRouter()
}

/*
func Register(path string, f func(http.ResponseWriter, *http.Request), method string) error {
	if path == "" || method == "" {
		return errors.New("Unable to register route. Invalid params")
	}
	MuxRouter.HandleFunc(path, f).Methods(method)
	return nil
}
*/
type File struct {
	Name    string
	Content string
}

// Register a new Arkeogis route
func Register(myroute *Route) error {
	// Setup the fonction that will handle the route request
	m := MuxRouter.HandleFunc(myroute.Path, func(rw http.ResponseWriter, r *http.Request) {
		// session
		token := r.Header.Get("Authorization")
		log.Println("token ", token)
		s := session.GetSession(token)

		log.Print("user id from session : ", s.GetIntDef("user_id", -1))

		// Retrieve user id from session
		user := model.User{}
		user.Id = s.GetIntDef("user_id", 0)

		// Open a transaction to load the user from db
		tx, err := db.DB.Beginx()
		if err != nil {
			log.Panicln("Can't start transaction for creating a new user")
			return
		}

		// Retrieve the user from db
		user.Get(tx)
		log.Println("user is : ", user)
		s.Values["user"] = user

		// Check global permsissions
		permok := true
		ok, err := user.HavePermissions(tx, myroute.Permissions...)
		if err != nil {
			log.Printf("user.HavePermissions failed : ", err)
			permok = false
		} else if ok == false {
			log.Printf("user has no permissions : ", myroute.Permissions)
			permok = false
		}

		// Check filters
		errstr := ""
		if permok {
			ok, errstr = filters.CheckAll(tx, myroute.ParamFilters, rw, r, s)
			if !ok {
				permok = false
				log.Printf("filters says no ! ", errstr)
			}
		}

		// Close the transaction
		err = tx.Commit()
		if err != nil {
			if err, ok := err.(*pq.Error); ok {
				log.Println("commit while getting session user failed, pq error:", err.Code.Name())
			} else {
				log.Println("commit while getting session user failed !", err)
			}
			return
		}

		// Print a log
		log.Printf("[%s] %s %s ; authorized: %t\n", user.Username, myroute.Method, myroute.Path, permok)
		if !permok {
			return
		}

		// decode json from request
		if myroute.Json != nil {
			fmt.Println("Json : ", myroute.Json)
			v := reflect.New(myroute.Json)
			o := v.Interface()
			// Check if multipart
			mt, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if err != nil {
				log.Println("error parsing header:", err)
			}
			if strings.HasPrefix(mt, "multipart/") {
				mr := multipart.NewReader(r.Body, params["boundary"])
				// For each part
				for {
					p, err := mr.NextPart()
					if err == io.EOF {
						// EOF we call the Func passed as argument
						myroute.Func(rw, r, o, s)
						return
					}
					if err != nil {
						log.Println("error getting part of multipart form", err)
						return
					}
					j, err := ioutil.ReadAll(p)
					if err != nil {
						log.Println("error: unable to get content of the file")
					}
					// Is it a file ?
					if p.FileName() != "" {
						// Check if target interface has a FileName field
						ta := reflect.Indirect(reflect.ValueOf(o)).FieldByName("File")
						if !ta.IsValid() {
							log.Println("error: target interface does not have File field")
							return
						}
						// Check if target interface "File" field is type of routes.File
						if ta.Type() != reflect.TypeOf(&File{}) {
							log.Println("error: target interface field \"File\" is not of type routes.File")
							return
						}
						// Assign file to structure
						fs := reflect.New(reflect.TypeOf(File{}))
						fsi := fs.Interface()
						ee := reflect.Indirect(reflect.ValueOf(fsi))
						ee.FieldByName("Name").SetString(p.FileName())
						ee.FieldByName("Content").SetString(string(j[:]))
						ta.Set(fs)
					} else {
						// Unmarshall datas into structure
						json.Unmarshal(j, o)
					}
				}
			} else {
				decoder := json.NewDecoder(r.Body)
				//fmt.Printf("t : %t\n", o)
				//fmt.Println("o : ", o)
				err := decoder.Decode(o)
				if err != nil {
					log.Panicln("decode failed", err)
				}
				myroute.Func(rw, r, o, s)
			}
		} else {
			myroute.Func(rw, r, nil, s)
		}
	})

	// Setup the request method
	m.Methods(myroute.Method)
	//m := MuxRouter.HandleFunc(r.Path, r.Func).Methods(r.Method)

	// Setup the Queries
	for _, q := range myroute.Queries {
		m.Queries(q)
	}

	// end success
	return nil
}

// RegisterMultiple will register multiple Arkeogis Routes
func RegisterMultiple(routes []*Route) error {
	for _, route := range routes {
		err := Register(route)
		if err != nil {
			return err
		}
	}
	return nil
}
