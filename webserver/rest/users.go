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

package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	db "github.com/croll/arkeogis-server/db"
	model "github.com/croll/arkeogis-server/model"
	"github.com/croll/arkeogis-server/webserver/filters"
	routes "github.com/croll/arkeogis-server/webserver/routes"
	"github.com/croll/arkeogis-server/webserver/session"
	"github.com/lib/pq"
)

// Structures for create
type Valuedisplay struct {
	Value   int    `json:"value"`
	Display string `json:"display"`
}

type Company struct {
	Name       Valuedisplay `json:"data"`
	SearchName string       `json:"searchname"`
	City       Valuedisplay `json:"city"`
}

type Usercreate struct {
	model.User
	City     Valuedisplay `city`
	Company1 Company      `company1`
	Company2 Company      `company2`
}

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:   "/api/users",
			Func:   UserCreate,
			Method: "POST",
			Json:   reflect.TypeOf(Usercreate{}),
			Permissions: []string{
				"AdminUsers",
			},
		},
		&routes.Route{
			Path:   "/api/users",
			Func:   UserList,
			Method: "GET",
			Permissions: []string{
				"AdminUsers",
			},
			ParamFilters: []filters.Filter{
				filters.ParamFilterIntBoundary{
					ParamFilter: filters.ParamFilter{
						ParamType:   filters.ParamTypeForm,
						ParamName:   "limit",
						ErrorString: "limit over boundaries",
						Permissions: []string{"AdminUsers"},
					},
					Lower: 0,
					Upper: 100,
				},
			},
		},
		&routes.Route{
			Path:   "/api/users",
			Func:   UserUpdate,
			Method: "PUT",
		},
		&routes.Route{
			Path:   "/api/users",
			Func:   UserDelete,
			Method: "DELETE",
		},
		&routes.Route{
			Path:   "/api/login",
			Func:   UserLogin,
			Method: "POST",
			Json:   reflect.TypeOf(Userlogin{}),
		},
	}
	fmt.Println("routes : ", Routes[4])
	fmt.Println("routes : ", Routes[2])
	routes.RegisterMultiple(Routes)
}

// UserList List of users. no filets, no args actually...
func UserList(w http.ResponseWriter, r *http.Request, o interface{}, s *session.Session) {
	type Answer struct {
		Data  []model.User `json:"data"`
		Count int          `json:"count"`
	}

	answer := Answer{}

	err := r.ParseForm()
	if err != nil {
		fmt.Println("ParseForm err: ", err)
		return
	}

	// decode order...
	order := r.FormValue("order")
	orderdir := "ASC"
	if strings.HasPrefix(order, "-") {
		order = order[1:]
		orderdir = "DESC"
	}

	switch {
	case order == "u.username",
		order == "u.id",
		order == "u.created_at",
		order == "u.updated_at",
		order == "u.email",
		order == "u.active":
		// accepted
	case order == "u.lastname, u.firstname":
		order = "u.lastname " + orderdir + ", u.firstname"
	default:
		log.Println("rest(users.UsersList): order denied : ", order)
		order = "u.id"
		orderdir = "ASC"
	}
	/////

	// decode limit / page
	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		limit = 10
	}
	switch {
	case limit == 5,
		limit == 10,
		limit == 15:
		// accepted
	default:
		limit = 10
	}

	page, err := strconv.Atoi(r.FormValue("page"))
	if err != nil || page < 1 {
		page = 1
	}

	offset := (page - 1) * limit
	/////

	err = db.DB.Select(&answer.Data, "SELECT * FROM \"user\" u WHERE (u.username ILIKE $1 OR u.firstname ILIKE $1 OR u.lastname ILIKE $1 OR u.email ILIKE $1) ORDER BY "+order+" "+orderdir+" OFFSET $2 LIMIT $3", "%"+r.FormValue("filter")+"%", offset, limit)
	if err != nil {
		log.Println("err: ", err)
		return
	}

	err = db.DB.Get(&answer.Count, "SELECT count(*) FROM \"user\"")
	if err != nil {
		log.Println("err: ", err)
		return
	}

	//fmt.Println("users: ", users)
	j, err := json.Marshal(answer)
	w.Write(j)
}

// UserCreate Create a user, see usercreate struct inside this function for json content
func UserCreate(w http.ResponseWriter, r *http.Request, o interface{}, s *session.Session) {

	u := o.(*Usercreate)
	// hack
	u.City_geonameid = u.City.Value

	tx, err := db.DB.Beginx()
	if err != nil {
		log.Panicln("Can't start transaction for creating a new user")
		return
	}

	err = u.Create(tx)

	if err != nil {
		if pgerr, ok := err.(*pq.Error); ok {
			log.Println("create user failed, pq error:", pgerr.Code.Name())
			ArkeoError(w, pgerr.Code.Name(), "Mince mince minnnnnce !")
		} else {
			log.Println("create user failed !", err)
		}
		return
	}

	err = tx.Commit()
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			log.Println("commit user failed, pq error:", err.Code.Name())
		} else {
			log.Println("commit user failed !", err)
		}
		return
	}
}

// UserUpdate update an user.
func UserUpdate(w http.ResponseWriter, r *http.Request, o interface{}, s *session.Session) {
	//params := mux.Vars(r)
	//uid := params["id"]
	//email := r.FormValue("email")
}

// UserDelete delete an user.
func UserDelete(w http.ResponseWriter, r *http.Request, o interface{}, s *session.Session) {
	fmt.Println("delete")
}

// UserInfos return detailed infos on an user
func UserInfos(w http.ResponseWriter, r *http.Request, o interface{}, s *session.Session) {
	w.Header().Set("Allow", "DELETE,GET,HEAD,OPTIONS,POST,PUT")
}

// Userlogin structure (json)
type Userlogin struct {
	Username string
	Password string
}

// UserLogin Check Login
func UserLogin(w http.ResponseWriter, r *http.Request, o interface{}, s *session.Session) {

	time.Sleep(1 * time.Second) // limit rate

	l := o.(*Userlogin)
	user := model.User{
		Username: l.Username,
		Password: l.Password,
	}

	log.Println("sploarf : ", user)

	tx, err := db.DB.Beginx()
	if err != nil {
		log.Panicln("Can't start transaction for creating a new user")
		return
	}

	// test login
	ok, err := user.Login(tx)
	if err != nil {
		log.Println("Login Failed with error : ", err)
		tx.Rollback()
		return
	}

	if !ok {
		log.Println("Login failed for user ", l.Username)
		tx.Rollback()
		ArkeoError(w, "401", "Bad Username/Password")
		return
	}

	user.Get(tx)       // retrieve the user
	user.Password = "" // immediatly erase password field

	log.Println("Login ", user.Username, " => ", ok)

	token, s := session.NewSession()
	s.Values["user_id"] = user.Id
	s.Values["user"] = user

	err = tx.Commit()
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			log.Println("commit user failed, pq error:", err.Code.Name())
		} else {
			log.Println("commit user failed !", err)
		}
		return
	}

	type answer struct {
		User  model.User
		Token string
	}

	a := answer{
		User:  user,
		Token: token,
	}

	j, err := json.Marshal(a)
	w.Write(j)
}
