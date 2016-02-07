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
	"strings"
	"time"

	db "github.com/croll/arkeogis-server/db"
	model "github.com/croll/arkeogis-server/model"
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

// UserListParams is params struct for UserList query
type UserListParams struct {
	Limit  int    `default:"10" min:"1" max:"100" error:"limit over boundaries"`
	Page   int    `default:"1" min:"1" error:"page over boundaries"`
	Order  string `default:"u.created_at" enum:"u.created_at,-u.created_at,u.updated_at,-u.updated_at,u.username,-u.username,u.firstname,-u.firstname,u.lastname,-u.lastname,u.email,-u.email" error:"bad order"`
	Filter string `default:""`
}

type Usercreate struct {
	model.User
	City     Valuedisplay `json:"city"`
	Company1 Company      `json:"company1"`
	Company2 Company      `json:"company2"`
}

func init() {

	Routes := []*routes.Route{
		&routes.Route{
			Path:        "/api/users",
			Description: "Create a new arkeogis user",
			Func:        UserCreate,
			Method:      "POST",
			Json:        reflect.TypeOf(Usercreate{}),
			Permissions: []string{
			//"AdminUsers",
			},
		},
		&routes.Route{
			Path:        "/api/users",
			Description: "List arkeogis users",
			Func:        UserList,
			Method:      "GET",
			Permissions: []string{
			//"AdminUsers",
			},
			Params: reflect.TypeOf(UserListParams{}),
		},
		&routes.Route{
			Path:        "/api/users",
			Description: "Update an arkeogis user",
			Func:        UserUpdate,
			Method:      "PUT",
		},
		&routes.Route{
			Path:        "/api/users",
			Description: "Delete an arkeogis user",
			Func:        UserDelete,
			Method:      "DELETE",
		},
		&routes.Route{
			Path:        "/api/login",
			Description: "Login to an arkeogis session",
			Func:        UserLogin,
			Method:      "POST",
			Json:        reflect.TypeOf(Userlogin{}),
		},
	}
	fmt.Println("routes : ", Routes[4])
	fmt.Println("routes : ", Routes[2])
	routes.RegisterMultiple(Routes)
}

// UserList List of users. no filets, no args actually...
func UserList(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	type Answer struct {
		Data  []model.User `json:"data"`
		Count int          `json:"count"`
	}

	answer := Answer{}

	params := proute.Params.(*UserListParams)

	// decode order...
	order := params.Order
	orderdir := "ASC"
	if strings.HasPrefix(order, "-") {
		order = order[1:]
		orderdir = "DESC"
	}
	if order == "u.lastname" {
		order = "u.lastname " + orderdir + ", u.firstname"
	}
	/////

	offset := (params.Page - 1) * params.Limit

	err := db.DB.Select(&answer.Data, "SELECT * FROM \"user\" u WHERE (u.username ILIKE $1 OR u.firstname ILIKE $1 OR u.lastname ILIKE $1 OR u.email ILIKE $1) ORDER BY "+order+" "+orderdir+" OFFSET $2 LIMIT $3", "%"+params.Filter+"%", offset, params.Limit)
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

func userSqlError(w http.ResponseWriter, err error) {
	if pgerr, ok := err.(*pq.Error); ok {
		log.Printf("pgerr: %#v\n", pgerr)
		switch pgerr.Code.Name() {
		case "foreign_key_violation":
			switch pgerr.Constraint {
			case "c_user.first_lang_id":
				routes.FieldError(w, "user.first_lang_id", "user.first_lang_id", "USERS.FIELD_LANG.S_ERROR_BADLANG")
			case "user_ibfk_2":
				routes.FieldError(w, "user.second_lang_id", "user.second_lang_id", "USERS.FIELD_LANG.S_ERROR_BADLANG")
			default:
				routes.ServerError(w, 500, "INTERNAL ERROR")
			}
		case "unique_violation":
			switch pgerr.Constraint {
			case "user_idx_4":
				routes.FieldError(w, "user.username", "username", "USERS.FIELD_USERNAME.S_ERROR_ALREADYEXISTS")
			default:
				routes.ServerError(w, 500, "INTERNAL ERROR")
			}
		default:
			log.Printf("unhandled postgresql error ! : %#v\n", pgerr)
			routes.ServerError(w, 500, "INTERNAL ERROR")
		}
	} else {
		log.Println("not a postgresql error !", err)
		routes.ServerError(w, 500, "INTERNAL ERROR")
	}
}

// UserCreate Create a user, see usercreate struct inside this function for json content
func UserCreate(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	u := proute.Json.(*Usercreate)

	// hack
	u.City_geonameid = u.City.Value

	tx, err := db.DB.Beginx()
	if err != nil {
		userSqlError(w, err)
		return
	}

	err = u.Create(tx)
	if err != nil {
		userSqlError(w, err)
		return
	}

	err = tx.Commit()
	if err != nil {
		userSqlError(w, err)
		return
	}

	j, err := json.Marshal("ok")
	w.Write(j)
}

// UserUpdate update an user.
func UserUpdate(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	//params := mux.Vars(r)
	//uid := params["id"]
	//email := r.FormValue("email")
}

// UserDelete delete an user.
func UserDelete(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	fmt.Println("delete")
}

// UserInfos return detailed infos on an user
func UserInfos(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	w.Header().Set("Allow", "DELETE,GET,HEAD,OPTIONS,POST,PUT")
}

// Userlogin structure (json)
type Userlogin struct {
	Username string
	Password string
}

// UserLogin Check Login
func UserLogin(w http.ResponseWriter, r *http.Request, proute routes.Proute) {

	time.Sleep(1 * time.Second) // limit rate

	l := proute.Json.(*Userlogin)
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
