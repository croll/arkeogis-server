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
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	db "github.com/croll/arkeogis-server/db"
	"github.com/croll/arkeogis-server/model"
	"github.com/croll/arkeogis-server/webserver/routes"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

const arkeoproxyheaderurl = "arkeoproxyurl"

var anyproxy *httputil.ReverseProxy

func newAnyHostReverseProxy() *httputil.ReverseProxy {
	director := func(req *http.Request) {
		newurlstr := req.Header.Get(arkeoproxyheaderurl)
		req.Header.Del(arkeoproxyheaderurl)

		index1 := strings.Index(newurlstr, "?")
		index2 := strings.Index(newurlstr, "&")

		if index2 > -1 && (index1 == -1 || index1 > index2) {
			newurlstr = newurlstr[:index2] + "?" + newurlstr[index2+1:]
		}

		newurl, err := url.Parse(newurlstr)
		if err != nil {
			fmt.Println("aie, newurlstr: ", newurlstr, err)
			req.URL = nil
		} else {
			req.URL = newurl
			req.Host = newurl.Host
		}

		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
		fmt.Println("req: ", req)
	}
	return &httputil.ReverseProxy{Director: director}
}

func initproxy(router *mux.Router) {
	anyproxy = newAnyHostReverseProxy()

	router.HandleFunc("/proxy/", func(w http.ResponseWriter, r *http.Request) {
		url := r.RequestURI[8:] // parsed url, we remove /proxy/? from the beggining
		fmt.Println("uri: ", url)

		// Open a transaction to load the user from db
		tx, err := db.DB.Beginx()
		if err != nil {
			log.Panicln("Can't start transaction for loading user")
			routes.ServerError(w, 500, "Can't start transaction for loading user")
			return
		}

		s := routes.LoadSessionFromRequest(tx, r)
		_p, _ := s.Get("user")
		user := _p.(model.User)

		fmt.Println("user: ", user)

		// Check global permsissions
		perm_proxy, _ := user.HavePermissions(tx, "request map")
		perm_fullproxy, _ := user.HavePermissions(tx, "manage all wms/wmts")
		log.Println(perm_fullproxy, perm_proxy)

		err = tx.Commit()
		if err != nil {
			if err, ok := err.(*pq.Error); ok {
				log.Println("commit while getting session user failed, pq error:", err.Code.Name())
			} else {
				log.Println("commit while getting session user failed !", err)
			}
			routes.ServerError(w, 500, "Can't commit transaction")
			return
		}

		// hack
		//perm_proxy = true
		//perm_fullproxy = true

		if !perm_proxy {
			routes.ServerError(w, 403, "No permission to use proxy")
			return
		}

		// search the good proxy
		layers := []model.Map_layer{}
		err = db.DB.Select(&layers, "SELECT * FROM map_layer WHERE published='t'")
		if err != nil {
			log.Println("Can't find layers : ", err)
			return
		}

		layerfound := false
		for _, layer := range layers {
			if strings.HasPrefix(url, layer.Url) {
				fmt.Println("layer match: ", layer)
				layerfound = true
			}
		}

		if layerfound || perm_fullproxy {
			r.Header.Add(arkeoproxyheaderurl, url)
			anyproxy.ServeHTTP(w, r)
		} else {
			routes.ServerError(w, 403, "No permission to use proxy")
			return
		}

	})
}
