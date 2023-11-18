package backend

import (
	"github.com/gin-contrib/sessions"
	"golang.org/x/oauth2"
)

func GetUserSession(session sessions.Session) (sessionUser SessionUser, ok bool) {
	sessionUser, ok = session.Get(UserKey).(SessionUser)

	return
}

func SetUserSession(session sessions.Session, sessionUser SessionUser) {
	session.Set(UserKey, sessionUser)
}

func GetTokenSession(session sessions.Session) (token oauth2.Token, ok bool) {
	token, ok = session.Get(TokenKey).(oauth2.Token)

	return
}

func SetTokenSession(session sessions.Session, token oauth2.Token) {
	session.Set(TokenKey, token)
}

func ClearTokenSession(session sessions.Session) {
	session.Delete(TokenKey)
}

func GetStateSession(session sessions.Session) (state string, ok bool) {
	state, ok = session.Get(StateKey).(string)

	return
}

func SetStateSession(session sessions.Session, state string) {
	session.Set(StateKey, state)
}

func GetPreviousPathSession(session sessions.Session) (previousPath string, ok bool) {
	previousPath, ok = session.Get(PreviousPathKey).(string)

	return
}

func SetPreviousPathSession(session sessions.Session, previousPath string) {
	session.Set(PreviousPathKey, previousPath)
}