package filters

import (
	"log"
	"net/http"
	"strconv"

	"github.com/croll/arkeogis-server/model"
	"github.com/croll/arkeogis-server/webserver/session"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

type Filter interface {
	Check(tx *sqlx.Tx, w http.ResponseWriter, r *http.Request, s *session.Session) bool
	CheckPerms(tx *sqlx.Tx, w http.ResponseWriter, r *http.Request, s *session.Session) bool
}

// CheckAll filters from a []Filter array
func CheckAll(tx *sqlx.Tx, filters []Filter, w http.ResponseWriter, r *http.Request, s *session.Session) bool {
	res := true
	for _, filter := range filters {
		if filter.CheckPerms(tx, w, r, s) == false {
			// do nothing, we are not concerned by this filter
		} else {
			res = filter.Check(tx, w, r, s)
			if !res {
				return false
			}
		}
	}
	return res
}

/*
 * ParamFilter
 */

type ParamType int

const (
	ParamTypeForm ParamType = iota + 1
	ParamTypeMux
)

type ParamFilter struct {
	Filter
	ParamType   ParamType
	ParamName   string
	ErrorString string
	Permissions []string
}

func (ff ParamFilter) CheckPerms(tx *sqlx.Tx, w http.ResponseWriter, r *http.Request, s *session.Session) bool {
	_user, found := s.Get("user")
	if found != true {
		log.Fatalln("we should have a user in session, but this was not the case !")
		return false
	}
	user := _user.(model.User)

	ok, err := user.HavePermissions(tx, ff.Permissions...)
	if err != nil {
		log.Fatalln("Permission check failed ! ", err)
		return false
	}

	return ok
}

func (ff ParamFilter) GetParamValue(w http.ResponseWriter, r *http.Request) string {
	switch ff.ParamType {
	case ParamTypeForm:
		err := r.ParseForm()
		if err != nil {
			log.Println("ParseForm err: ", err)
			return ""
		}

		return r.FormValue(ff.ParamName)

	case ParamTypeMux:
		vars := mux.Vars(r)
		return vars[ff.ParamName]
	}
	return ""
}

/*
 * ParamFilterIntBoundary
 */

type ParamFilterIntBoundary struct {
	ParamFilter
	Lower int
	Upper int
}

func (ff ParamFilterIntBoundary) Check(tx *sqlx.Tx, w http.ResponseWriter, r *http.Request, s *session.Session) bool {
	v := ff.ParamFilter.GetParamValue(w, r)
	vi, _ := strconv.Atoi(v)

	if vi < ff.Lower || vi > ff.Upper {
		return true
	}

	return false
}
