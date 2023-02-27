package middleware

import (
	"net/http"
	"sync"
	"time"
)

var globalSession = make(map[string]*Session)

var globalSessionLock sync.RWMutex

var globalSessionExpireSeconds = 6000.00

// 基于内存session
type Session struct {
	sync.RWMutex
	id            string
	data          map[string]interface{}
	lastTouchTime time.Time
}

func newSession(context Context) *Session {

	id := Guid()
	s := Session{
		id:            id,
		data:          make(map[string]interface{}),
		lastTouchTime: time.Now(),
	}
	context.SetCookie(&http.Cookie{
		Name:     "sessionId",
		Value:    id,
		HttpOnly: true,
	})
	globalSessionLock.Lock()
	globalSession[id] = &s
	globalSessionLock.Unlock()
	return &s
}

func getSession(context Context) *Session {
	s, ok := globalSession[context.GetCookie("sessionId")]
	if ok {
		return s
	}
	return newSession(context)
}

func (t *Session) Set(key string, val interface{}) {
	t.data[key] = val
}

func (t *Session) Get(key string) interface{} {
	return t.data[key]
}

func (t *Session) Id() string {
	return t.id
}

func init() {
	// session过期
	// Schedule("session-expire", 30*60, func() {
	// 	for k, v := range globalSession {
	// 		v.Lock()
	// 		if time.Now().Sub(v.lastTouchTime).Seconds() > globalSessionExpireSeconds {
	// 			delete(globalSession, k)
	// 		}
	// 		v.Unlock()
	// 	}
	// })
}

func (c *Context) SessionSet(key string, value interface{}) {
	getSession(*c).Set(key, value)
}

func (c *Context) SessionGet(key string) interface{} {
	return getSession(*c).Get(key)
}
