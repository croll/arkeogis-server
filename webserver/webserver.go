/* ArkeoGIS - The Arkeolog Geographical Information Server Program
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

package webserver

import (
	"fmt"
	"net/http"
	"strconv"

	config "github.com/croll/arkeogis-server/config"
	rest "github.com/croll/arkeogis-server/webserver/rest"
	routes "github.com/croll/arkeogis-server/webserver/routes"
	"github.com/codegangsta/negroni"
)

func StartServer() {
	fmt.Println("starting web server...")
	rest.P()
	Negroni := negroni.New(
		negroni.NewRecovery(),
		negroni.HandlerFunc(authMiddleware),
		negroni.HandlerFunc(crossDomainMiddleware),
		negroni.NewLogger(),
		negroni.NewStatic(http.Dir(config.WebPath)),
	)
	Negroni.UseHandler(routes.MuxRouter)
	Negroni.Run(":" + strconv.Itoa(config.Main.Server.Port))
}

func crossDomainMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	next(w, r)
}

// AuthMiddleware ...
func authMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	err := getUserFromAuth(r.Header.Get("Authorization"))
	if err != nil {
		w.WriteHeader(401)
		return
	}
	next(w, r)
}

func getUserFromAuth(code string) interface{} {
	return nil
}
