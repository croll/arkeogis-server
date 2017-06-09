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

	db "github.com/croll/arkeogis-server/db"
	"github.com/croll/arkeogis-server/model"
	"github.com/gorilla/mux"
)

type ArkProxy struct {
	Layer model.Map_layer
	Proxy *httputil.ReverseProxy
}

var proxies *[]ArkProxy

func InitProxies() {
	layers := []model.Map_layer{}
	err := db.DB.Select(&layers, "SELECT * FROM map_layer WHERE published='t'")
	if err != nil {
		log.Println("Can't init proxy with layers : ", err)
		return
	}
	_proxies := make([]ArkProxy, len(layers))
	for i, layer := range layers {
		u, _ := url.Parse(layer.Url)
		p := httputil.NewSingleHostReverseProxy(u)
		_proxies[i] = ArkProxy{
			Layer: layer,
			Proxy: p,
		}
	}
	proxies = &_proxies
	fmt.Println("proxies inited.", proxies)
}

func initproxy(router *mux.Router) {
	InitProxies()
	router.HandleFunc("/proxy/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("request: ", *r)
		fmt.Println("uri: ", r.RequestURI[8:])
		for _, proxy := range *proxies {
			//url := mux.Vars(r)["url"]
			//url = strings.Replace(url, "http:/", "http://")
			//url = strings.Replace(url, "https:/", "https://")
			//fmt.Println("url: ", url)
			fmt.Println("proxy: ", proxy)
		}
		fmt.Fprint(w, "proxy")
	})
}
