package appcontext

import (
	"log"

	"github.com/gocraft/web"
	"github.com/ryankurte/authplz/lib/api"
)

const (
	secondFactorRequestSessionKey = "2fa-request"
	secondFactorActionSessionKey  = "2fa-action"
)

// Bind2FARequest Bind a 2fa request and action for a user
// TODO: the request should probably time-out eventually
func (c *AuthPlzCtx) Bind2FARequest(rw web.ResponseWriter, req *web.Request, userid string, action string) {
	secondFactorSession, _ := c.Global.SessionStore.Get(req.Request, secondFactorRequestSessionKey)

	log.Printf("AuthPlzCtx.Bind2faRequest adding authorization flash for user %s\n", userid)

	secondFactorSession.Values[secondFactorRequestSessionKey] = userid
	secondFactorSession.Values[secondFactorActionSessionKey] = action
	secondFactorSession.Save(req.Request, rw)
}

// Get2FARequest Fetch a 2fa request and action for a user
func (c *AuthPlzCtx) Get2FARequest(rw web.ResponseWriter, req *web.Request) (string, string) {
	u2fSession, _ := c.Global.SessionStore.Get(req.Request, secondFactorRequestSessionKey)

	if u2fSession.Values[secondFactorRequestSessionKey] == nil ||
		u2fSession.Values[secondFactorActionSessionKey] == nil {
		c.WriteApiResult(rw, api.ResultError, "No userid found")
		log.Printf("AuthPlzCtx.Get2FARequest No userid found in session flash")
		return "", ""
	}
	userid := u2fSession.Values[secondFactorRequestSessionKey].(string)
	action := u2fSession.Values[secondFactorActionSessionKey].(string)
	return userid, action
}