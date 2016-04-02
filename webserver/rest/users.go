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
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	config "github.com/croll/arkeogis-server/config"
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
	model.Company
	CityAndCountry model.CityAndCountry_wtr `json:"city_and_country"`
	SearchName     string                   `json:"searchname"`
}

// UserListParams is params struct for UserList query
type UserListParams struct {
	Limit   int    `default:"10" min:"1" max:"100" error:"limit over boundaries"`
	Page    int    `default:"1" min:"1" error:"page over boundaries"`
	Order   string `default:"u.created_at" enum:"u.created_at,-u.created_at,u.updated_at,-u.updated_at,u.username,-u.username,u.firstname,-u.firstname,u.lastname,-u.lastname,u.email,-u.email" error:"bad order"`
	Filter  string `default:""`
	Lang_id int    `default:"1"`
}

// UserCreate structure (json)
type Usercreate struct {
	model.User
	CityAndCountry model.CityAndCountry_wtr `json:"city_and_country"`
	Companies      []Company                `json:"companies"`
	File           *routes.File
}

// Userlogin structure (json)
type Userlogin struct {
	Username string
	Password string
}

type UserGetParams struct {
	Id int `min:"0" error:"User Id is mandatory"`
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
			Path:        "/api/users/{id:[0-9]+}",
			Description: "Get an arkeogis user",
			Func:        UserInfos,
			Method:      "GET",
			Permissions: []string{
			//"AdminUsers",
			},
			Params: reflect.TypeOf(UserGetParams{}),
		},
		&routes.Route{
			Path:        "/api/users/{id:[0-9]+}",
			Description: "Update an arkeogis user",
			Func:        UserUpdate,
			Method:      "POST",
			Json:        reflect.TypeOf(Usercreate{}),
			Permissions: []string{
			//"AdminUsers",
			},
		},
		&routes.Route{
			Path:        "/api/users",
			Description: "Delete an arkeogis user",
			Func:        UserDelete,
			Method:      "DELETE",
			Permissions: []string{
			//"AdminUsers",
			},
		},
		&routes.Route{
			Path:        "/api/login",
			Description: "Login to an arkeogis session",
			Func:        UserLogin,
			Method:      "POST",
			Json:        reflect.TypeOf(Userlogin{}),
			Permissions: []string{
			//"AdminUsers",
			},
		},
		&routes.Route{
			Path:        "/api/users/{id:[0-9]+}/photo",
			Description: "get user photo (jpg)",
			Func:        UserPhoto,
			Method:      "GET",
			Permissions: []string{
			//"AdminUsers",
			},
			Params: reflect.TypeOf(UserGetParams{}),
		},
	}
	routes.RegisterMultiple(Routes)
}

func selectCityAndCountry(city_geonameid string, langid int) string {
	return "" +
		"SELECT " +
		" city.geonameid as city_geonameid, city_tr.lang_id as city_lang_id, city_tr.name as city_name, " +
		" country.geonameid as country_geonameid, country.iso_code as country_iso_code, country_tr.lang_id as country_lang_id, country_tr.name as country_name " +
		"from city " +
		"LEFT JOIN city_tr ON city_tr.city_geonameid=city.geonameid " +
		"LEFT JOIN country ON country.geonameid=city.country_geonameid " +
		"LEFT JOIN country_tr ON country_tr.country_geonameid = country.geonameid " +
		"WHERE city.geonameid=" + city_geonameid +
		" AND (city_tr.lang_id = " + strconv.Itoa(langid) + " or city_tr.lang_id=1) " +
		" AND (country_tr.lang_id = " + strconv.Itoa(langid) + " or country_tr.lang_id=1) " +
		"ORDER by city_tr.lang_id desc, country_tr.lang_id desc " +
		"LIMIT 1"
}

// UserList List of users. no filets, no args actually...
func UserList(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	type User struct {
		model.User
		Groups_user       string `json:"groups_user"`
		Groups_chronology string `json:"groups_chronology"`
		Groups_charac     string `json:"groups_charac"`
		CountryAndCity    string `json:"country_and_city"`
	}
	type Answer struct {
		Data  []User `json:"data"`
		Count int    `json:"count"`
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

	err := db.DB.Select(&answer.Data,
		"SELECT "+
			" *, "+
			" COALESCE((SELECT array_to_json(array_agg(group_tr.*)) FROM user__group u_g LEFT JOIN \"group\" g ON u_g.group_id = g.id LEFT JOIN group_tr ON g.id = group_tr.group_id WHERE g.type='user' AND u_g.user_id = u.id), '[]') as groups_user,"+
			" COALESCE((SELECT array_to_json(array_agg(group_tr.*)) FROM user__group u_g LEFT JOIN \"group\" g ON u_g.group_id = g.id LEFT JOIN group_tr ON g.id = group_tr.group_id WHERE g.type='chronology' AND u_g.user_id = u.id), '[]') as groups_chronology,"+
			" COALESCE((SELECT array_to_json(array_agg(group_tr.*)) FROM user__group u_g LEFT JOIN \"group\" g ON u_g.group_id = g.id LEFT JOIN group_tr ON g.id = group_tr.group_id WHERE g.type='charac' AND u_g.user_id = u.id), '[]') as groups_charac, "+
			" COALESCE((select row_to_json(t) from("+selectCityAndCountry("u.city_geonameid", params.Lang_id)+") t), '{}') as countryandcity"+
			" FROM \"user\" u WHERE (u.username ILIKE $1 OR u.firstname ILIKE $1 OR u.lastname ILIKE $1 OR u.email ILIKE $1) GROUP BY u.id ORDER BY "+order+" "+orderdir+" OFFSET $2 LIMIT $3",
		"%"+params.Filter+"%", offset, params.Limit)
	if err != nil {
		userSqlError(w, err)
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
	log.Printf("paf: %#v\n", err)
	if pgerr, ok := err.(*pq.Error); ok {
		log.Printf("pgerr: %#v\n", pgerr)
		switch pgerr.Code.Name() {
		case "foreign_key_violation":
			switch pgerr.Constraint {
			case "c_user.first_lang_id":
				routes.FieldError(w, "json.first_lang_id", "first_lang_id", "USER.FIELD_LANG.T_CHECK_BADLANG")
			case "c_user.second_lang_id":
				routes.FieldError(w, "json.second_lang_id", "second_lang_id", "USER.FIELD_LANG.T_CHECK_BADLANG")
			case "c_user.city_geonameid":
				routes.FieldError(w, "json.searchTextCity", "searchTextCity", "USER.FIELD_CITY.T_CHECK_MANDATORY")
			default:
				routes.ServerError(w, 500, "INTERNAL ERROR")
			}
		case "unique_violation":
			switch pgerr.Constraint {
			case "i_user.username":
				routes.FieldError(w, "json.username", "username", "USER.FIELD_USERNAME.T_CHECK_ALREADYEXISTS")
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

// serach a company in an array of companies, using the Id for key search
func companyIndex(id int, slice []model.Company) int {
	for i, v := range slice {
		if v.Id == id {
			return i
		}
	}
	return -1
}

// userSet is for UserCreate or UserUpdate
func userSet(w http.ResponseWriter, r *http.Request, proute routes.Proute, create bool) {

	u := proute.Json.(*Usercreate)

	// hack
	u.City_geonameid = u.CityAndCountry.City.Geonameid
	log.Println("city : ", u.City_geonameid)

	tx, err := db.DB.Beginx()
	if err != nil {
		log.Println("1")
		userSqlError(w, err)
		return
	}

	// photo...
	if u.File != nil {
		u.Photo = string(u.File.Content)
	}

	// save the user
	if create {
		err = u.Create(tx)
	} else {
		err = u.Update(tx)
	}
	if err != nil {
		log.Println("2")
		userSqlError(w, err)
		tx.Rollback()
		return
	}

	var companies []model.Company
	for _, form_company := range u.Companies {
		if form_company.Id > 0 {
			form_company.City_geonameid = form_company.CityAndCountry.City.City_geonameid
			log.Println("updating company : ", form_company.Company)
			err = form_company.Update(tx)
			if err != nil {
				log.Println("error while updating a company", err)
				tx.Rollback()
				userSqlError(w, err)
				return
			}
			companies = append(companies, form_company.Company)
		} else if len(form_company.Name) > 0 {
			form_company.City_geonameid = form_company.CityAndCountry.City.City_geonameid
			log.Println("creating company : ", form_company.Company)
			err = form_company.Create(tx)
			if err != nil {
				log.Println("error while adding a company", err)
				tx.Rollback()
				userSqlError(w, err)
				return
			}
			companies = append(companies, form_company.Company)
		}
	}

	err = u.SetCompanies(tx, companies)
	if err != nil {
		log.Println("7")
		tx.Rollback()
		userSqlError(w, err)
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println("8")
		userSqlError(w, err)
		return
	}

	j, err := json.Marshal("ok")
	w.Write(j)
}

// UserCreate Create a user, see usercreate struct inside this function for json content
func UserCreate(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	userSet(w, r, proute, true)
}

// UserUpdate update an user.
func UserUpdate(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	userSet(w, r, proute, false)
}

// UserDelete delete an user.
func UserDelete(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	fmt.Println("delete")
}

// UserInfos return detailed infos on an user
func UserInfos(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	params := proute.Params.(*UserGetParams)
	tx, err := db.DB.Beginx()
	if err != nil {
		log.Println("can't start transaction")
		userSqlError(w, err)
		return
	}
	u := Usercreate{}
	u.Id = params.Id

	err = u.Get(tx)
	if err != nil {
		log.Println("can't get user")
		userSqlError(w, err)
		return
	}

	err = u.CityAndCountry.Get(tx, u.User.City_geonameid, 48) // todo: take good lang
	if err != nil {
		log.Println("can't get user city and country", err)
		//userSqlError(w, err)
		//return
	}
	//log.Println("city and country : ", u.CityAndCountry)

	companies, err := u.GetCompanies(tx)
	if err != nil {
		log.Println("can't get user companies")
		userSqlError(w, err)
		return
	}
	for _, company := range companies {
		mcomp := Company{}
		mcomp.Id = company.Id
		mcomp.Name = company.Name

		err = mcomp.CityAndCountry.Get(tx, company.City_geonameid, 48) // todo: take good lang
		if err != nil {
			log.Println("can't get company city and country", err)
			//userSqlError(w, err)
			//return
		}
		u.Companies = append(u.Companies, mcomp)
	}

	//log.Println("user id : ", params.Id, "user : ", u)
	err = tx.Commit()
	if err != nil {
		log.Println("can't commit")
		userSqlError(w, err)
		return
	}
	j, err := json.Marshal(u)
	w.Write(j)
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

func UserPhoto(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	params := proute.Params.(*UserGetParams)

	var photo []byte

	err := db.DB.Get(&photo, "SELECT photo FROM \"user\" u WHERE id=$1", params.Id)

	if err != nil {
		log.Println("user photo get failed")
		return
	}

	if len(photo) == 0 {
		photo, err = ioutil.ReadFile(config.WebPath + "/img/default-user-photo.jpg")
		if err != nil {
			log.Println("user default photo load failed")
			return
		}
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(photo)))
	w.Write(photo)
}
