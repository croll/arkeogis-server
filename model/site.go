/* ArkeoGIS - The Geographic Information System for Archaeologists
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

package model

import (
	"errors"

	"github.com/croll/arkeogis-server/geo"
	"github.com/jmoiron/sqlx"
)

// SiteInfos is a meta struct which stores all the informations about a site
type SiteInfos struct {
	Site
	Site_tr
	//NbSiteRanges     int
	HasError  bool
	Point     *geo.Point
	Latitude  string
	Longitude string
	Altitude  string
	GeonameID string
	Created   bool
	EPSG      int
}

func (s *Site) Get(tx *sqlx.Tx) (err error) {
	stmt, err := tx.PrepareNamed("SELECT *,ST_GeomFromGeoJSON(geom) AS geom, ST_GeomFromGeoJSON(geom_3d) AS geom_3d from \"site\" WHERE id=:id")
	if err != nil {
		err = errors.New("Site::Get: " + err.Error())
		return
	}
	defer stmt.Close()
	err = stmt.Get(s, s)
	if err != nil {
		err = errors.New("Site::Get: " + err.Error())
	}
	return
}

func (s *SiteInfos) Create(tx *sqlx.Tx) (err error) {
	var q string
	if s.EPSG != 4326 {
		q = "INSERT INTO \"site\" (\"code\", \"name\", \"city_name\", \"city_geonameid\", \"geom\", \"geom_3d\", \"centroid\", \"occupation\", \"database_id\", \"created_at\", \"updated_at\") VALUES (:code, :name, :city_name, :city_geonameid, ST_Transform(ST_GeometryFromText(:geom), 4326)::::geography, ST_Transform(ST_GeometryFromText(:geom_3d), 4326)::::geography, :centroid, :occupation, :database_id, now(), now()) RETURNING id"
	} else {
		q = "INSERT INTO \"site\" (\"code\", \"name\", \"city_name\", \"city_geonameid\", \"geom\", \"geom_3d\", \"centroid\", \"occupation\", \"database_id\", \"created_at\", \"updated_at\") VALUES (:code, :name, :city_name, :city_geonameid, ST_GeographyFromText(:geom), ST_GeographyFromText(:geom_3d), :centroid, :occupation, :database_id, now(), now()) RETURNING id"
	}
	stmt, err := tx.PrepareNamed(q)
	if err != nil {
		err = errors.New("SiteInfos::Create: " + err.Error())
		return
	}
	defer stmt.Close()
	err = stmt.Get(&s.Id, s)
	if err != nil {
		err = errors.New("SiteInfos::Create: " + err.Error())
	}
	return
}

func (s *SiteInfos) Update(tx *sqlx.Tx) (err error) {
	var q string
	if s.EPSG != 4326 {
		q = "UPDATE \"site\" SET \"code\" = :code, \"name\" = :name, \"city_name\" = :city_name, \"city_geonameid\" = :city_geonameid, geom = ST_Transform(ST_GeometryFromText(:geom), 4326)::geography, geom_3d = ST_Transform(ST_GeometryFromText(:geom_3d), 4326)::::geography, \"centroid\" = :centroid, \"occupation\" = :occupation, \"database_id\" = :database_id, \"updated_at\" = now() WHERE database_id = :id"
	} else {
		q = "UPDATE \"site\" SET \"code\" = :code, \"name\" = :name, \"city_name\" = :city_name, \"city_geonameid\" = :city_geonameid, geom = ST_GeographyFromText(:geom), geom_3d = ST_GeographyFromText(:geom_3d), \"centroid\" = :centroid, \"occupation\" = :occupation, \"database_id\" = :database_id, \"updated_at\" = now() WHERE id = :id"
	}
	stmt, err := tx.PrepareNamed(q)
	if err != nil {
		err = errors.New("SiteInfos::Update: " + err.Error())
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(s)
	if err != nil {
		err = errors.New("SiteInfos::Update: " + err.Error())
	}
	return
}

func (sr *Site_range) Create(tx *sqlx.Tx) (err error) {
	stmt, err := tx.PrepareNamed("INSERT INTO \"site_range\" (" + Site_range_InsertStr + ") VALUES (" + Site_range_InsertValuesStr + ") RETURNING id")
	if err != nil {
		err = errors.New("Site_rante::Create: " + err.Error())
		return
	}
	defer stmt.Close()
	err = stmt.Get(&sr.Id, sr)
	if err != nil {
		err = errors.New("Site_rante::Create: " + err.Error())
	}
	return
}
