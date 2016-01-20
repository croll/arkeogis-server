package filters

import (
	"log"
	"net/http"

	"github.com/croll/arkeogis-server/model"
	"github.com/croll/arkeogis-server/webserver/session"
)

type Filter interface {
	Check(w http.ResponseWriter, r *http.Request, s *session.Session) bool
}

// CheckAll filters from a []Filter array
func CheckAll(filters []Filter, w http.ResponseWriter, r *http.Request, s *session.Session) bool {
	for _, filter := range filters {
		if filter.Check(w, r, s) == false {
			return false
		}
	}
	return true
}

/*
 * FormFilter
 */

type FormFilter struct {
	Filter
	FieldName   string
	ErrorString string
	Permissions []string
}

func (ff FormFilter) CheckPerms(w http.ResponseWriter, r *http.Request, s *session.Session) bool {
	_user, found := s.Get("user")
	if found != true {
		log.Fatalln("we should have a user in session, but this was not the case !")
		return false
	}
	user := _user.(model.User)

	ok, err := user.HavePermissions(nil, ff.Permissions...)
	if err != nil {
		log.Fatalln("Permission check failed ! ", err)
		return false
	}

	return ok
}

/*
 * FormFilterIntBoundary
 */

type FormFilterIntBoundary struct {
	FormFilter
	Lower int
	Upper int
}

func (ff FormFilterIntBoundary) Check(w http.ResponseWriter, r *http.Request, s *session.Session) bool {

	return false
}
