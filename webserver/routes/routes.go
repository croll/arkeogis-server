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
	"os"
	"reflect"
	"strconv"
	"time"
	//"strconv"
	"strings"

	"mime"
	"mime/multipart"

	config "github.com/croll/arkeogis-server/config"
	db "github.com/croll/arkeogis-server/db"
	"github.com/croll/arkeogis-server/model"
	"github.com/croll/arkeogis-server/webserver/sanitizer"
	session "github.com/croll/arkeogis-server/webserver/session"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// Route structure that is used for registering a new Arkeogis Route
type Route struct {
	Path        string
	Description string // for rest doc generator
	Func        func(rw http.ResponseWriter, r *http.Request, proute Proute)
	Method      string
	Queries     []string
	Json        reflect.Type
	Params      reflect.Type
	Permissions []string
}

type Proute struct {
	Json    interface{}
	Params  interface{}
	Session *session.Session
	Lang1   model.Lang
	Lang2   model.Lang
}

type File struct {
	Name    string
	Content []byte
}

// All routes added here are stored there. This is usefull for building REST doc
var Routes []*Route = []*Route{}

// MuxRouter is the gorilla mux router initialized here for Arkeogis
var MuxRouter *mux.Router
var restlog *os.File

func init() {
	// router
	MuxRouter = mux.NewRouter()

	var err error

	restlog, err = os.OpenFile(config.DistPath+"/logs/rest.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0660)
	if err != nil {
		log.Fatalf("Error opening access log file: %v", err)
	}

}

func decodeContent(myroute *Route, rw http.ResponseWriter, r *http.Request, s *session.Session) interface{} {
	if myroute.Json == nil {
		return nil
	}

	// decode json from request
	//fmt.Println("Json : ", myroute.Json)
	v := reflect.New(myroute.Json)
	o := v.Interface()

	// set to default the new structure
	sanitizer.DefaultStruct(o)

	// Check if multipart
	mt, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		log.Println("error parsing header:", err)
	}
	if strings.HasPrefix(mt, "multipart/") {
		log.Println(" = MULTIPART POST = ")
		mr := multipart.NewReader(r.Body, params["boundary"])
		// For each part
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				// EOF we can stop here
				return o
			}
			if err != nil {
				log.Println("error getting part of multipart form", err)
				return nil
			}
			j, err := ioutil.ReadAll(p)
			if err != nil {
				log.Println("error: unable to get content of the file")
			}
			//log.Println(" => part ... ", string(j))
			// Is it a file ?
			if p.FileName() != "" {
				// Check if target interface has a FileName field
				ta := reflect.Indirect(reflect.ValueOf(o)).FieldByName("File")
				if !ta.IsValid() {
					log.Println("error: target interface does not have File field")
					return nil
				}
				// Check if target interface "File" field is type of routes.File
				if ta.Type() != reflect.TypeOf(&File{}) {
					log.Println("error: target interface field \"File\" is not of type routes.File")
					return nil
				}
				// Assign file to structure
				fs := reflect.New(reflect.TypeOf(File{}))
				fsi := fs.Interface()
				ee := reflect.Indirect(reflect.ValueOf(fsi))
				ee.FieldByName("Name").SetString(p.FileName())
				//ee.FieldByName("Content").SetString(string(j[:]))
				ee.FieldByName("Content").SetBytes(j)
				ta.Set(fs)
			} else {
				// Unmarshall datas into structure
				json.Unmarshal(j, o)
			}
		}
	} else {
		//log.Println(" = normal POST = ")
		decoder := json.NewDecoder(r.Body)
		//fmt.Printf("t : %t\n", o)
		//fmt.Println("o : ", o)
		err := decoder.Decode(o)
		if err != nil {
			log.Panicln("decode failed", err)
		}
		return o
	}
}

func decodeParams(myroute *Route, rw http.ResponseWriter, r *http.Request) interface{} {
	if myroute.Params == nil {
		return nil
	}

	// get form parameters
	err := r.ParseForm()
	if err != nil {
		fmt.Println("ParseForm err: ", err)
		return nil
	}

	// get mux parameters
	muxvars := mux.Vars(r)

	v := reflect.New(myroute.Params)
	params := v.Interface()

	// set all defaults
	sanitizer.DefaultStruct(params)

	st := reflect.TypeOf(params).Elem()
	vt := reflect.ValueOf(params).Elem()

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		value := vt.Field(i)

		paramval := ""
		if val, ok := muxvars[strings.ToLower(field.Name)]; ok {
			paramval = val
		} else {
			paramval = r.FormValue(strings.ToLower(field.Name))
		}

		switch field.Type.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			def, _ := strconv.ParseInt(paramval, 10, 64)
			value.SetInt(def)
		case reflect.Float32, reflect.Float64:
			def, _ := strconv.ParseFloat(paramval, 64)
			value.SetFloat(def)
		case reflect.Bool:
			def, _ := strconv.ParseBool(paramval)
			value.SetBool(def)
		case reflect.String:
			value.SetString(paramval)
		default:
			log.Println("decodeParams on type", field.Type.Name(), "not implemented")
		}
	}

	return params
}

func LoadSessionFromRequest(tx *sqlx.Tx, r *http.Request) *session.Session {
	// session
	token := r.Header.Get("Authorization")
	if len(token) == 0 {
		cook, err := r.Cookie("arkeogis_session_token")
		if err == nil {
			token = cook.Value
		}
	}
	//log.Println("token ", token)
	s := session.GetSession(token)

	//fmt.Println("session: ", s)
	//log.Print("user id from session : ", s.GetIntDef("user_id", -1))

	// Retrieve user id from session
	user := model.User{}
	user.Id = s.GetIntDef("user_id", 0)

	// Retrieve the user from db
	user.Get(tx)
	//log.Println("user is : ", user.Username)
	s.Set("user", user)

	return s
}

func handledRoute(myroute *Route, rw http.ResponseWriter, r *http.Request) {

	start := time.Now()

	// debug transactions count
	/*
		count_before := 0
		db.DB.Get(&count_before, "SELECT count(*) from pg_stat_activity where state = 'idle in transaction'")

		defer func() {
			count_end := 0
			db.DB.Get(&count_end, "SELECT count(*) from pg_stat_activity where state = 'idle in transaction'")

			log.Println("idle in transaction : ", count_before, " => ", count_end)
		}()
	*/

	// Open a transaction to load the user from db
	tx, err := db.DB.Beginx()
	if err != nil {
		log.Panicln("Can't start transaction for loading user")
		ServerError(rw, 500, "Can't start transaction for loading user")
		return
	}

	s := LoadSessionFromRequest(tx, r)
	_p, _ := s.Get("user")
	user := _p.(model.User)

	// get langs
	lang1 := model.Lang{
		Isocode: "en",
	}
	lang2 := model.Lang{
		Isocode: "fr",
	}

	if _lang1, err := r.Cookie("arkeogis_user_lang_1"); err == nil {
		lang1.Isocode = _lang1.Value
	}
	if _lang2, err := r.Cookie("arkeogis_user_lang_2"); err == nil {
		lang2.Isocode = _lang2.Value
	}

	err = lang1.Get(tx)
	if err != nil {
		lang1.Isocode = "en"
		err = lang1.Get(tx)
		if err != nil {
			fmt.Println("can't load lang1 !")
		}
	}

	err = lang2.Get(tx)
	if err != nil {
		lang2.Isocode = "fr"
		err = lang2.Get(tx)
		if err != nil {
			fmt.Println("can't load lang2 !")
		}
	}

	//fmt.Println("langs: ", lang1, lang2)

	restlog.WriteString(fmt.Sprintf(`[%s][%s] %s %s %s %s`+"\n", start.Format(time.RFC3339), user.Username, r.RemoteAddr, r.Method, r.URL.Path, myroute.Method))

	// Check global permsissions
	permok := true
	ok, err := user.HaveAtLeastOnePermission(tx, myroute.Permissions...)
	if err != nil {
		log.Println("user.HaveAtLeastOnePermission failed : ", err)
		permok = false
	} else if ok == false {
		log.Println("user has no permissions : ", myroute.Permissions)
		permok = false
	}

	// Close the transaction
	err = tx.Commit()
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			log.Println("commit while getting session user failed, pq error:", err.Code.Name())
		} else {
			log.Println("commit while getting session user failed !", err)
		}
		ServerError(rw, 500, "Can't commit transaction")
		return
	}

	// Print a log
	log.Printf("[%s] %s %s ; authorized: %t\n", user.Username, myroute.Method, myroute.Path, permok)
	if !permok {
		ServerError(rw, 403, "unauthorized")
		return
	}

	o := decodeContent(myroute, rw, r, s)
	if o != nil {
		errors := sanitizer.SanitizeStruct(o, "json")
		if len(errors) > 0 {
			Errors(rw, errors)
			return
		}
	}

	params := decodeParams(myroute, rw, r)
	if params != nil {
		log.Println("params    : ", params)
		errors := sanitizer.SanitizeStruct(params, "params")
		log.Println("Sanitized : ", params)
		if len(errors) > 0 {
			Errors(rw, errors)
			return
		}
	}

	proute := Proute{
		Json:    o,
		Params:  params,
		Session: s,
		Lang1:   lang1,
		Lang2:   lang2,
	}

	myroute.Func(rw, r, proute)

}

func ServerError(w http.ResponseWriter, code int, message string) {
	type ServerError struct {
		Message string `json:"message"`
	}

	aerr := ServerError{
		Message: message,
	}

	j, err := json.Marshal(aerr)
	if err != nil {
		log.Panicln("err in error, marshaling failed: ", err)
	}
	http.Error(w, (string)(j), code)
}

func FieldError(w http.ResponseWriter, fieldpath string, fieldname string, errorstring string) {
	aerr := []sanitizer.FieldError{
		sanitizer.FieldError{
			FieldPath:   fieldpath,
			FieldName:   fieldname,
			ErrorString: errorstring,
		},
	}
	Errors(w, aerr)
}

func Errors(w http.ResponseWriter, errors []sanitizer.FieldError) {
	type Errors struct {
		Errors []sanitizer.FieldError `json:"errors"`
	}
	aerr := Errors{
		Errors: errors,
	}
	j, err := json.Marshal(aerr)
	if err != nil {
		log.Panicln("err in error, marshaling failed: ", err)
	}
	http.Error(w, (string)(j), 400)
}

// Register a new Arkeogis route
func Register(myroute *Route) error {
	Routes = append(Routes, myroute)

	// Setup the fonction that will handle the route request
	m := MuxRouter.HandleFunc(myroute.Path, func(rw http.ResponseWriter, r *http.Request) {
		handledRoute(myroute, rw, r)
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
