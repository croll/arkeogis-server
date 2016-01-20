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
	"net/http"

	db "github.com/croll/arkeogis-server/db"
	"github.com/croll/arkeogis-server/model"
	routes "github.com/croll/arkeogis-server/webserver/routes"
	"github.com/croll/arkeogis-server/webserver/session"
)

func init() {
	Routes := []*routes.Route{
		&routes.Route{
			Path:   "/api/database",
			Func:   DatabaseCreate,
			Method: "POST",
		},
		&routes.Route{
			Path:   "/api/database",
			Func:   DatabasesList,
			Method: "GET",
		},
		&routes.Route{
			Path:   "/api/database",
			Func:   DatabaseUpdate,
			Method: "PUT",
		},
		&routes.Route{
			Path:   "/api/database",
			Func:   DatabaseDelete,
			Method: "DELETE",
		},
		&routes.Route{
			Path:   "/api/database/geographical_extent",
			Func:   DatabaseDelete,
			Method: "DELETE",
		},
	}
	routes.RegisterMultiple(Routes)
}

func DatabasesList(w http.ResponseWriter, r *http.Request, o interface{}, s *session.Session) {
	databases := []model.Database{}
	err := db.DB.Select(&databases, "SELECT * FROM \"database\"")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	l, _ := json.Marshal(databases)
	w.Header().Set("Content-Type", "application/json")
	w.Write(l)
}

func DatabaseCreate(w http.ResponseWriter, r *http.Request, o interface{}, s *session.Session) {
}

func DatabaseUpdate(w http.ResponseWriter, r *http.Request, o interface{}, s *session.Session) {
}

func DatabaseDelete(w http.ResponseWriter, r *http.Request, o interface{}, s *session.Session) {
}
