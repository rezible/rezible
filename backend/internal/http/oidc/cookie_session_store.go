package oidc

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/sessions"
	rez "github.com/rezible/rezible"
)

const (
	sessionCookieName = "oidc_session"
)

var SessionSecretKey []byte

func configureSessionStore() *sessions.CookieStore {
	maxAge := 86400 * 30 // 30 days

	store := sessions.NewCookieStore(SessionSecretKey)
	store.MaxAge(maxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.SameSite = http.SameSiteStrictMode
	store.Options.Secure = true

	if rez.Config.DebugMode() {
		store.Options.SameSite = http.SameSiteLaxMode
		if appUrl, urlErr := url.Parse(rez.Config.AppUrl()); urlErr == nil {
			store.Options.Domain = appUrl.Host
		}
	}

	return store
}

type session struct {
	State string `json:"state"`
}

func (s *session) encode() ([]byte, error) {
	sessVal, jsonErr := json.Marshal(s)
	if jsonErr != nil {
		return nil, fmt.Errorf("encoding session: %w", jsonErr)
	}
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, writeErr := gz.Write(sessVal); writeErr != nil {
		return nil, fmt.Errorf("writing: %w", writeErr)
	}
	if flushErr := gz.Flush(); flushErr != nil {
		return nil, fmt.Errorf("flushing: %w", flushErr)
	}
	if closeErr := gz.Close(); closeErr != nil {
		return nil, fmt.Errorf("closing: %w", closeErr)
	}
	return b.Bytes(), nil
}

func decodeSession(data any) (*session, error) {
	strData, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("invalid session data")
	}
	rdata := strings.NewReader(strData)
	rdr, gzipErr := gzip.NewReader(rdata)
	if gzipErr != nil {
		return nil, fmt.Errorf("gzip reader: %w", gzipErr)
	}
	val, readErr := io.ReadAll(rdr)
	if readErr != nil || val == nil {
		return nil, fmt.Errorf("read: %w", readErr)
	}
	var sess session
	if jsonErr := json.Unmarshal(val, &sess); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal session: %w", jsonErr)
	}
	return &sess, nil
}

func setSession(w http.ResponseWriter, r *http.Request, store sessions.Store, key string, sess *session) error {
	reg, regErr := store.New(r, sessionCookieName)
	if regErr != nil {
		return fmt.Errorf("new session: %w", regErr)
	}

	val, valErr := sess.encode()
	if valErr != nil {
		return fmt.Errorf("encoding session: %w", valErr)
	}
	reg.Values[key] = string(val)

	if saveErr := reg.Save(r, w); saveErr != nil {
		return fmt.Errorf("saving session: %w", saveErr)
	}

	return nil
}

func getSession(r *http.Request, store sessions.Store, key string) (*session, error) {
	reg, regErr := store.Get(r, sessionCookieName)
	if regErr != nil {
		return nil, fmt.Errorf("store get: %w", regErr)
	}
	val, ok := reg.Values[key]
	if !ok || val == nil {
		return nil, nil
	}
	sess, decodeErr := decodeSession(val)
	if decodeErr != nil {
		return nil, fmt.Errorf("decode session: %w", decodeErr)
	}
	return sess, nil
}

func clearSession(w http.ResponseWriter, r *http.Request, store sessions.Store) error {
	sess, sessErr := store.Get(r, sessionCookieName)
	if sessErr != nil {
		return sessErr
	}
	sess.Options.MaxAge = -1
	sess.Values = make(map[interface{}]interface{})
	if saveErr := sess.Save(r, w); saveErr != nil {
		return fmt.Errorf("failed to save cleared session: %w", saveErr)
	}
	return nil
}
