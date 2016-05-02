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
	sqlx_types "github.com/jmoiron/sqlx/types"
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
	Limit  int    `default:"10" min:"1" max:"100" error:"limit over boundaries"`
	Page   int    `default:"1" min:"1" error:"page over boundaries"`
	Order  string `default:"u.created_at" enum:"u.created_at,-u.created_at,u.updated_at,-u.updated_at,u.username,-u.username,u.firstname,-u.firstname,u.lastname,-u.lastname,u.email,-u.email" error:"bad order"`
	Filter string `default:""`
}

// UserCreate structure (json)
type Usercreate struct {
	model.User
	CityAndCountry model.CityAndCountry_wtr `json:"city_and_country"`
	Companies      []Company                `json:"companies"`
	File           *routes.File
	Groups         []model.Group `json:"groups"`

	// overrides
	Password string `json:"password"`
}

type Userinfos struct {
	Usercreate

	// overrides
	Password string `json:"-"`
}

// Userlogin structure (json)
type Userlogin struct {
	Username string
	Password string
}

type UserGetParams struct {
	Id int `min:"0" error:"User Id is mandatory"`
}

type PhotoGetParams struct {
	Id int `min:"0" error:"Photo Id is mandatory"`
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
			Path:        "/api/users/{id:[0-9]+}",
			Description: "Delete an arkeogis user",
			Func:        UserDelete,
			Method:      "DELETE",
			Permissions: []string{
			//"AdminUsers",
			},
			Params: reflect.TypeOf(UserGetParams{}),
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
			Path:        "/api/users/photo/{id:[0-9]+}",
			Description: "get user photo (jpg)",
			Func:        UserPhoto,
			Method:      "GET",
			Permissions: []string{
			//"AdminUsers",
			},
			Params: reflect.TypeOf(PhotoGetParams{}),
		},
	}
	routes.RegisterMultiple(Routes)
}

func selectTranslated(tabletr string, coltr string, collang string, where string, lang1 int, lang2 int) string {
	return "(" +
		//		" SELECT \"" + coltr + "\" " +
		" SELECT \"" + tabletr + "\" " +
		//" SELECT name " +
		" FROM \"" + tabletr + "\" " +
		" WHERE \"" + collang + "\" IN (" + strconv.Itoa(lang1) + "," + strconv.Itoa(lang2) + ",0) " +
		" AND " + where + " " +
		" ORDER BY idx(array[" + strconv.Itoa(lang1) + "," + strconv.Itoa(lang2) + ",0], \"" + collang + "\") " +
		" LIMIT 1" +
		")"
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
		" AND (city_tr.lang_id = " + strconv.Itoa(langid) + " or city_tr.lang_id=0) " +
		" AND (country_tr.lang_id = " + strconv.Itoa(langid) + " or country_tr.lang_id=0) " +
		"ORDER by city_tr.lang_id desc, country_tr.lang_id desc " +
		"LIMIT 1"
}

func selectCityAndCountryAsJson(city_geonameid string, langid int) string {
	return "COALESCE((select row_to_json(t) from(" + selectCityAndCountry(city_geonameid, langid) + ") t), '[]'::json)"
}

func selectGroupAsJson(group_type string, langid int) string {
	return "" +
		"SELECT " +
		//" jsonb_agg(" + selectTranslated("group_tr", "name", "lang_id", "group_id = g.id", langid, 0) + ") " +
		" to_jsonb(array_agg(" + selectTranslated("group_tr", "name", "lang_id", "group_id = g.id", langid, 0) + ")) " +
		//" json_agg((g.id," + selectTranslated("group_tr", "name", "lang_id", "group_id = g.id", langid, 0) + ")) " +
		" FROM user__group u_g " +
		" LEFT JOIN \"group\" g ON u_g.group_id = g.id " +
		" WHERE g.type='" + group_type + "' AND u_g.user_id = u.id "
}

func selectGroupAsJsonNotNull(group_type string, langid int) string {
	return "COALESCE((" + selectGroupAsJson(group_type, langid) + "), '[]'::jsonb)"
}

func selectCompany(user_id string) string {
	return "" +
		"SELECT " +
		" array_agg(c.*) " +
		" FROM user__company u_c " +
		" LEFT JOIN company c ON u_c.company_id = c.id " +
		" WHERE u_c.user_id = " + user_id
}

func selectCompanyAsJson(user_id string) string {
	//return "COALESCE((select row_to_json(t) from(" + selectCompany(user_id) + ") t), '[]'::json)"
	return "COALESCE(array_to_json((" + selectCompany(user_id) + ")), '[]'::json)"
}

// UserList List of users. no filets, no args actually...
func UserList(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	type User struct {
		model.User
		Groups_user       sqlx_types.JSONText `json:"groups_user"`
		Groups_chronology sqlx_types.JSONText `json:"groups_chronology"`
		Groups_charac     sqlx_types.JSONText `json:"groups_charac"`
		CountryAndCity    sqlx_types.JSONText `json:"country_and_city"`
		Companies         sqlx_types.JSONText `json:"companies"`

		// overrides
		Password string `json:"-"`
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
			" u.id, u.username, u.created_at, u.updated_at, u.firstname, u.lastname, u.active, u.email, u.photo_id, "+
			" "+selectGroupAsJsonNotNull("user", proute.Lang1.Id)+" as groups_user, "+
			" "+selectGroupAsJsonNotNull("chronology", proute.Lang1.Id)+" as groups_chronology, "+
			" "+selectGroupAsJsonNotNull("charac", proute.Lang1.Id)+" as groups_charac, "+
			" "+selectCityAndCountryAsJson("u.city_geonameid", proute.Lang1.Id)+" as countryandcity, "+
			" "+selectCompanyAsJson("u.id")+" as companies "+
			" FROM \"user\" u "+
			" WHERE (u.username ILIKE $1 OR u.firstname ILIKE $1 OR u.lastname ILIKE $1 OR u.email ILIKE $1) "+
			"  AND u.id > 0"+ // don't list anonymous
			" GROUP BY u.id "+
			" ORDER BY "+order+" "+orderdir+
			" OFFSET $2 "+
			" LIMIT $3",
		"%"+params.Filter+"%", offset, params.Limit)
	if err != nil {
		userSqlError(w, err)
		return
	}

	//log.Println("result: ", answer.Data)

	err = db.DB.Get(&answer.Count, "SELECT count(*) FROM \"user\"")
	if err != nil {
		log.Println("err: ", err)
		return
	}

	//fmt.Println("users: ", users)
	j, err := json.Marshal(answer)
	if err != nil {
		log.Println("marshal failed: ", err)
	}
	log.Println("result: ", string(j))
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

	// hack overrides
	u.User.Password = u.Password

	// hack for city
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
		photo := model.Photo{
			Photo: string(u.File.Content),
		}
		err = photo.Create(tx)
		if err != nil {
			log.Println("1")
			userSqlError(w, err)
			tx.Rollback()
			return
		}
		u.Photo_id = photo.Id
	}

	// save the user
	if create {
		err = u.Create(tx)
	} else {
		tmpuser := model.User{
			Id: u.Id,
		}
		err = tmpuser.Get(tx)
		if err != nil {
			log.Println("can't get user for update", err)
			userSqlError(w, err)
			tx.Rollback()
			return
		}

		// if we don't set a new password, we take it back from the db
		if len(u.User.Password) == 0 {
			u.User.Password = tmpuser.Password
		}

		log.Println("updating user id : ", u.Id, u)
		err = u.Update(tx)
	}
	if err != nil {
		log.Println("2")
		userSqlError(w, err)
		tx.Rollback()
		return
	}

	// save the companies
	var companies []model.Company
	for _, form_company := range u.Companies {
		if form_company.Id > 0 {
			form_company.City_geonameid = form_company.CityAndCountry.City.Geonameid
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
			form_company.City_geonameid = form_company.CityAndCountry.City.Geonameid
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

	// save the groups
	err = u.SetGroups(tx, u.Groups)
	if err != nil {
		log.Println("set groups")
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
	params := proute.Params.(*UserGetParams)
	tx, err := db.DB.Beginx()
	if err != nil {
		log.Println("can't start transaction")
		userSqlError(w, err)
		return
	}

	u := Userinfos{}
	u.Id = params.Id

	err = u.Get(tx)
	if err != nil {
		log.Println("can't get user")
		userSqlError(w, err)
		return
	}

	_, err = tx.Exec("DELETE FROM \"user__group\" where \"user_id\" = $1", u.Id)
	if err != nil {
		log.Println("can't remove user from groups")
		userSqlError(w, err)
		return
	}

	_, err = tx.Exec("DELETE FROM \"user__company\" where \"user_id\" = $1", u.Id)
	if err != nil {
		log.Println("can't remove user from companies")
		userSqlError(w, err)
		return
	}

	_, err = tx.Exec("DELETE FROM \"user\" where id = $1", u.Id)
	if err != nil {
		log.Println("can't delete user")
		userSqlError(w, err)
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println("can't commit")
		userSqlError(w, err)
		return
	}
	j, err := json.Marshal(u)
	w.Write(j)

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
	u := Userinfos{}
	u.Id = params.Id

	err = u.Get(tx)
	if err != nil {
		log.Println("can't get user")
		userSqlError(w, err)
		return
	}

	err = u.CityAndCountry.Get(tx, u.User.City_geonameid, proute.Lang1.Id)
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

		err = mcomp.CityAndCountry.Get(tx, company.City_geonameid, proute.Lang1.Id)
		if err != nil {
			log.Println("can't get company city and country err:", err)
			//userSqlError(w, err)
			//return
		}
		u.Companies = append(u.Companies, mcomp)
	}

	u.Groups, err = u.GetGroups(tx)
	if err != nil {
		log.Println("can't get user groups")
		userSqlError(w, err)
		return
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

	// get langs
	lang1 := model.Lang{
		Id: user.First_lang_id,
	}
	lang2 := model.Lang{
		Id: user.Second_lang_id,
	}

	err = lang1.Get(tx)
	if err != nil {
		lang1.Iso_code = "en"
		err = lang1.Get(tx)
		if err != nil {
			tx.Rollback()
			log.Fatal("can't load lang1 !")
			return
		}
	}

	err = lang2.Get(tx)
	if err != nil {
		lang2.Iso_code = "fr"
		err = lang2.Get(tx)
		if err != nil {
			tx.Rollback()
			log.Fatal("can't load lang2 !")
			return
		}
	}

	log.Println("langs: ", lang1, lang2)

	permissions, err := user.GetPermissions(tx)
	if err != nil {
		tx.Rollback()
		log.Fatal("can't get permissions!")
		return
	}
	log.Println("permissions : ", permissions)

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
		User        model.User
		Token       string
		Lang1       model.Lang         `json:"lang1"`
		Lang2       model.Lang         `json:"lang2"`
		Permissions []model.Permission `json:"permissions"`
	}

	a := answer{
		User:        user,
		Token:       token,
		Lang1:       lang1,
		Lang2:       lang2,
		Permissions: permissions,
	}

	j, err := json.Marshal(a)
	w.Write(j)
}

func UserPhoto(w http.ResponseWriter, r *http.Request, proute routes.Proute) {
	params := proute.Params.(*PhotoGetParams)

	var photo []byte

	err := db.DB.Get(&photo, "SELECT photo FROM \"photo\" u WHERE id=$1", params.Id)

	if err != nil {
		//log.Println("user photo get failed")
		// user photo get failed, so revert to default
		//return
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
