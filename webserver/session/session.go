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

package session

import (
	"math/rand"
	"time"
)

var sessionDuration time.Duration

// Session struct
type Session struct {
	LastAccess time.Time
	Values     map[string]interface{}
}

// Get return a value of any type. ok is false if no value was found
func (session *Session) Get(key string) (value interface{}, ok bool) {
	val, ok := session.Values[key]
	return val, ok
}

// GetDef return a value of any type. return the "def" parameter if no value was found
func (session *Session) GetDef(key string, def interface{}) interface{} {
	val, ok := session.Values[key]
	if !ok {
		return def
	}
	return val
}

// GetString return a value of string type. ok is false if no value was found
func (session *Session) GetString(key string) (value string, ok bool) {
	val, ok := session.Values[key]
	if ok {
		if v, castok := val.(string); castok {
			return v, castok
		}
		return "", false
	}
	return "", false
}

// GetStringDef return a value of string type. return the "def" parameter if no value was found
func (session *Session) GetStringDef(key string, def string) string {
	val, ok := session.Values[key]
	if ok {
		if v, castok := val.(string); castok {
			return v
		}
		return def
	}
	return def
}

// GetInt return a value of integer type. ok is false if no value was found
func (session *Session) GetInt(key string) (value int, ok bool) {
	val, ok := session.Values[key]
	if ok {
		if v, castok := val.(int); castok {
			return v, castok
		}
		return 0, false
	}
	return 0, false
}

// GetIntDef return a value of integer type. return the "def" parameter if no value was found
func (session *Session) GetIntDef(key string, def int) int {
	val, ok := session.Values[key]
	if ok {
		if v, castok := val.(int); castok {
			return v
		}
		return def
	}
	return def
}

var sessions map[string]*Session

func init() {
	sessionDuration = 20 * time.Minute
	sessions = make(map[string]*Session, 0)
	rand.Seed(time.Now().UnixNano())
}

// GetSession return the Session instance of the given token. if no session was found for this token, a transient session will be given
func GetSession(token string) *Session {
	cleanup()
	if s, ok := sessions[token]; ok {
		s.LastAccess = time.Now()
		return s
	} else {
		s = &Session{
			LastAccess: time.Now(),
			Values:     make(map[string]interface{}),
		}
		// removed bellow, because we don't save a transient session
		//sessions[token] = s
		return s
	}
}

// NewSession return a new Session instance, with a new token.
func NewSession() (token string, s *Session) {
	token = BuildRandomToken()
	s = &Session{
		LastAccess: time.Now(),
		Values:     make(map[string]interface{}),
	}
	// save it
	sessions[token] = s

	return token, s
}

func cleanup() {
	expire := time.Now().Add(-sessionDuration)
	for token, session := range sessions {
		if session.LastAccess.Before(expire) {
			delete(sessions, token)
		}
	}
}

// BuildRandomToken build a random string used for token.
func BuildRandomToken() string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	n := 42
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
