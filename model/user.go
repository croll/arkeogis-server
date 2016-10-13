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
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"

	db "github.com/croll/arkeogis-server/db"

	"github.com/jmoiron/sqlx"
)

/*
 * User Object
 */

// Get the user from the database
func (u *User) Get(tx *sqlx.Tx) error {
	var q = "SELECT * FROM \"user\" WHERE "
	if len(u.Username) > 0 {
		q += "username=:username"
	} else {
		q += "id=:id"
	}
	stmt, err := tx.PrepareNamed(q)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(u, u)
}

// Create the user by inserting it in the database
func (u *User) Create(tx *sqlx.Tx) error {
	//stmt, err := tx.PrepareNamed("INSERT INTO \"user\" (username, firstname, lastname, email, password, description, active, city_geonameid, first_lang_id, second_lang_id, created_at, updated_at) VALUES (:username, :firstname, :lastname, :email, :password, :description, :active, :city_geonameid, 1, 1, now(), now()) RETURNING id")
	stmt, err := tx.PrepareNamed("INSERT INTO \"user\" (" + User_InsertStr + ") VALUES (" + User_InsertValuesStr + ") RETURNING id")
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(&u.Id, u)
}

// Update the user in the database
func (u *User) Update(tx *sqlx.Tx) error {
	//_, err := tx.NamedExec("UPDATE \"user\" SET username=:username, firstname=:firstname, lastname=:lastname, email=:email, password=:password, description=:description, active=:active, city_geonameid=:city_geonameid, first_lang_id=:first_lang_idfirst_lang_id, second_lang_id=:second_lang_id, updated_at=now()", u)
	log.Println("update : ", "UPDATE \"user\" SET "+User_UpdateStr+" WHERE id=:id", u)
	_, err := tx.NamedExec("UPDATE \"user\" SET "+User_UpdateStr+" WHERE id=:id", u)
	return err
}

// Delete the user from the database
func (u *User) Delete(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("DELETE FROM \"user\" WHERE id=:id", u)
	return err
}

/*
 * Group Object
 */

// Get the group from the database
func (g *Group) Get(tx *sqlx.Tx) error {
	stmt, err := tx.PrepareNamed("SELECT * FROM \"group\" WHERE id=:id")
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(g, g)
}

// Create the group by inserting it in the database
func (g *Group) Create(tx *sqlx.Tx) error {
	stmt, err := tx.PrepareNamed("INSERT INTO \"group\" (" + Group_InsertStr + ") VALUES (" + Group_InsertValuesStr + ") RETURNING id")
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(&g.Id, g)
}

// Update the group in the database
func (g *Group) Update(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("UPDATE \"group\" SET "+Group_UpdateStr+" WHERE id=:id", g)
	return err
}

// Delete the group from the database
func (g *Group) Delete(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("DELETE FROM \"group\" WHERE id=:id", g)
	return err
}

/*
 * User <-> Group
 */

// GetGroups return an array of groups of the User
func (u *User) GetGroups(tx *sqlx.Tx) (groups []Group, err error) {
	stmt, err := tx.PrepareNamed("SELECT g.* FROM \"group\" g, \"user__group\" ug WHERE ug.user_id = :id AND ug.group_id = g.id")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	err = stmt.Select(&groups, u)
	return groups, err
}

// HaveGroups return true if the user have all the wanted groups
func (u *User) HaveGroups(tx *sqlx.Tx, groups ...Group) (ok bool, err error) {
	var idsGroups = make([]int, len(groups))
	for i, group := range groups {
		idsGroups[i] = group.Id
	}

	count := 0
	err = tx.Get(&count, "SELECT count(*) FROM user__group ug WHERE ug.user_id = "+strconv.Itoa(u.Id)+" AND ug.group_id in ("+IntJoin(idsGroups, true)+")")
	if err != nil {
		return false, err
	}
	if count == len(groups) {
		return true, err
	}
	return false, err
}

func searchString(a []string, search string) int {
	for i, v := range a {
		if v == search {
			return i
		}
	}
	return -1
}

func removeString(a []string, search string) []string {
	i := searchString(a, search)
	if i >= 0 {
		return append(a[:i], a[i+1:]...)
	}
	return a
}

// SetGroups set user groups
func (u *User) SetGroups(tx *sqlx.Tx, groups []Group) error {
	ids := make([]string, len(groups))
	for i, group := range groups {
		ids[i] = fmt.Sprintf("%d", group.Id)
	}

	_, err := tx.NamedExec("DELETE FROM \"user__group\" WHERE user_id=:id AND group_id NOT IN ("+strings.Join(ids, ",")+")", u)
	if err != nil {
		return err
	}

	rows, err := tx.Queryx("SELECT group_id FROM user__group WHERE user_id = $1 AND group_id IN ("+strings.Join(ids, ",")+")", u.Id)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		rows.Scan(&id)
		ids = removeString(ids, id)
	}

	for _, groupid := range ids {
		_, err := tx.Exec("INSERT INTO user__group (user_id, group_id) VALUES ($1, $2)", u.Id, groupid)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetCompanies return an array of companies of the User
func (u *User) GetCompanies(tx *sqlx.Tx) (companies []Company, err error) {
	stmt, err := tx.PrepareNamed("SELECT g.* FROM \"company\" g, \"user__company\" ug WHERE ug.user_id = :id AND ug.company_id = g.id")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	err = stmt.Select(&companies, u)
	return companies, err
}

// SetCompanies set user companies
func (u *User) SetCompanies(tx *sqlx.Tx, companies []Company) error {
	ids := make([]string, len(companies)+1)
	for i, company := range companies {
		ids[i] = fmt.Sprintf("%d", company.Id)
	}
	ids[len(companies)] = "-1" // empty list is a problem

	_, err := tx.NamedExec("DELETE FROM \"user__company\" WHERE user_id=:id AND company_id NOT IN ("+strings.Join(ids, ",")+")", u)
	if err != nil {
		return err
	}

	rows, err := tx.Queryx("SELECT company_id FROM user__company WHERE user_id = $1 AND company_id IN ("+strings.Join(ids, ",")+")", u.Id)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		rows.Scan(&id)
		ids = removeString(ids, id)
	}
	ids = removeString(ids, "-1")

	for _, companyid := range ids {
		_, err := tx.Exec("INSERT INTO user__company (user_id, company_id) VALUES ($1, $2)", u.Id, companyid)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetProjectId get the project of the user
func (u *User) GetProjectId(tx *sqlx.Tx) (projectID int, err error) {
	err = tx.Get(&projectID, "SELECT id FROM project WHERE user_id = $1", u.Id)
	if err == sql.ErrNoRows {
		err = nil
		projectID = 0
	}
	return
}

// GetUsers return an array of groups of the User
func (g *Group) GetUsers(tx *sqlx.Tx) (users []User, err error) {
	stmt, err := tx.PrepareNamed("SELECT u.* FROM \"user\" u, \"user__group\" ug WHERE ug.group_id = :id AND ug.user_id = u.id")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	err = stmt.Select(&users, g)
	for _, user := range users {
		user.Password = ""
	}
	return users, err
}

// SetUsers of the group
func (g *Group) SetUsers(tx *sqlx.Tx, users []User) error {
	ids := make([]string, len(users)+1)
	for i, user := range users {
		ids[i] = fmt.Sprintf("%d", user.Id)
	}
	ids[len(users)] = "-1" // empty list is a problem

	_, err := tx.NamedExec("DELETE FROM \"user__group\" WHERE group_id=:id AND user_id NOT IN ("+strings.Join(ids, ",")+")", g)
	if err != nil {
		return err
	}

	rows, err := tx.Queryx("SELECT user_id FROM user__group WHERE group_id = $1 AND user_id IN ("+strings.Join(ids, ",")+")", g.Id)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		rows.Scan(&id)
		ids = removeString(ids, id)
	}
	ids = removeString(ids, "-1")

	for _, userid := range ids {
		_, err := tx.Exec("INSERT INTO user__group (group_id, user_id) VALUES ($1, $2)", g.Id, userid)
		if err != nil {
			return err
		}
	}
	return nil
}

// Login test the username/password couple, and return true if it is ok, false if not
/*
func (u *User) Login(tx *sqlx.Tx) (ok bool, err error) {
	var q = "SELECT count(*) FROM \"user\" WHERE password=:password AND "
	if len(u.Username) > 0 {
		q += "username=:username"
	} else {
		q += "id=:id"
	}

	stmt, err := tx.PrepareNamed(q)
	var res int
	err = stmt.Get(&res, u)
	if err != nil {
		return false, err
	}
	defer stmt.Close()
	if res == 1 {
		return true, err
	}
	return false, err
}
*/

// Login test the username/password couple, and return true if it is ok, false if not
func (u *User) Login(password string) (ok bool) {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))

	if err == nil {
		return true
	} else {
		log.Println("Login err: ", err)
		return false
	}
}

// MakeNewPassword
func (u *User) MakeNewPassword(password string) (err error) {
	pass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err == nil {
		u.Password = string(pass)
	}
	return err
}

// GetPermissions return an array of Permissions that the user have
func (u *User) GetPermissions(tx *sqlx.Tx) (permissions []Permission, err error) {
	permissions = make([]Permission, 0)
	stmt, err := tx.PrepareNamed("SELECT p.* FROM permission p,user__group ug, group__permission gp WHERE ug.user_id = :id AND ug.group_id = gp.group_id AND gp.permission_id = p.id GROUP BY p.id")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	err = stmt.Select(&permissions, u)
	return permissions, err
}

// HavePermissions return true if the user have all the wanted permissions
func (u *User) HavePermissions(tx *sqlx.Tx, permissions ...string) (ok bool, err error) {
	if len(permissions) == 0 {
		return true, nil
	}
	query, args, err := sqlx.In("SELECT count(distinct(p.id)) FROM permission p,user__group ug, group__permission gp WHERE ug.user_id = ? AND ug.group_id = gp.group_id AND gp.permission_id = p.id AND p.name in (?)", u.Id, permissions)
	if err != nil {
		return false, err
	}
	query = db.DB.Rebind(query)
	var count int
	err = tx.Get(&count, query, args...)
	if count == len(permissions) {
		return true, err
	}
	return false, err
}

// HavePermissions return true if the user have all the wanted permissions
func (u *User) HaveAtLeastOnePermission(tx *sqlx.Tx, permissions ...string) (ok bool, err error) {
	if len(permissions) == 0 {
		return true, nil
	}
	query, args, err := sqlx.In("SELECT count(distinct(p.id)) FROM permission p,user__group ug, group__permission gp WHERE ug.user_id = ? AND ug.group_id = gp.group_id AND gp.permission_id = p.id AND p.name in (?)", u.Id, permissions)
	if err != nil {
		return false, err
	}
	query = db.DB.Rebind(query)
	var count int
	err = tx.Get(&count, query, args...)
	if count > 0 {
		return true, err
	}
	return false, err
}

// GetPermissions return an array of Permissions that the group have
func (g *Group) GetPermissions(tx *sqlx.Tx) (permissions []Permission, err error) {
	permissions = []Permission{}
	log.Println("TODO: Group.GetPermissions() => This function was not tested. If it work fine, please, remove this log comment!")
	stmt, err := tx.PrepareNamed("SELECT p.* FROM permission p, group__permission gp WHERE gp.group_id = :id AND gp.permission_id = p.id GROUP BY p.id")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	err = stmt.Select(&permissions, g)
	return permissions, err
}

/*
 * Photo Object
 */

// Get the photo from the database
func (u *Photo) Get(tx *sqlx.Tx) error {
	var q = "SELECT * FROM \"photo\" WHERE id=:id"
	stmt, err := tx.PrepareNamed(q)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(u, u)
}

// Create the photo by inserting it in the database
func (u *Photo) Create(tx *sqlx.Tx) error {
	stmt, err := tx.PrepareNamed("INSERT INTO \"photo\" (" + Photo_InsertStr + ") VALUES (" + Photo_InsertValuesStr + ") RETURNING id")
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(&u.Id, u)
}

// Update the photo in the database
func (u *Photo) Update(tx *sqlx.Tx) error {
	log.Println("update : ", "UPDATE \"photo\" SET "+Photo_UpdateStr+" WHERE id=:id")
	_, err := tx.NamedExec("UPDATE \"photo\" SET "+Photo_UpdateStr+" WHERE id=:id", u)
	return err
}

// Delete the photo from the database
func (u *Photo) Delete(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("DELETE FROM \"photo\" WHERE id=:id", u)
	return err
}

/*
 * Group_tr Object
 */

// Get the group_tr from the database
func (u *Group_tr) Get(tx *sqlx.Tx) error {
	var q = "SELECT * FROM \"group_tr\" WHERE group_id=:group_id"
	stmt, err := tx.PrepareNamed(q)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(u, u)
}

// Create the group_tr by inserting it in the database
func (u *Group_tr) Create(tx *sqlx.Tx) error {
	log.Println("saving : ", u)
	stmt, err := tx.PrepareNamed("INSERT INTO \"group_tr\" (" + Group_tr_InsertStr + ", group_id, lang_isocode) VALUES (" + Group_tr_InsertValuesStr + ", :group_id, :lang_isocode) RETURNING group_id")
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.Get(&u.Group_id, u)
}

// Update the group_tr in the database
func (u *Group_tr) Update(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("UPDATE \"group_tr\" SET "+Group_tr_UpdateStr+" WHERE group_id=:group_id AND lang_isocode=:lang_isocode", u)
	return err
}

// Delete the group_tr from the database
func (u *Group_tr) Delete(tx *sqlx.Tx) error {
	_, err := tx.NamedExec("DELETE FROM \"group_tr\" WHERE group_id=:group_id", u)
	return err
}
