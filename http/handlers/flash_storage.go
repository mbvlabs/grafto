package handlers

import (
	"net/http"

	"github.com/gorilla/sessions"
)

type FlashStorage struct {
	store *sessions.CookieStore
}

func NewCookieStore(sessionKey string) FlashStorage {
	store := sessions.NewCookieStore([]byte(sessionKey))
	return FlashStorage{store}
}

func (cs FlashStorage) CreateFlashMsg(
	r *http.Request,
	rw http.ResponseWriter,
	key string,
	args ...string,
) error {
	s, err := cs.store.Get(r, "flashMsg")
	if err != nil {
		return err
	}

	s.AddFlash(key, args...)
	if err := s.Save(r, rw); err != nil {
		return err
	}

	return nil
}

func (cs FlashStorage) GetFlashMessages(
	r *http.Request,
	rw http.ResponseWriter,
	key string,
) ([]string, error) {
	s, err := cs.store.Get(r, "flashMsg")
	if err != nil {
		return nil, err
	}

	var msgs []string
	if key != "" {
		for _, f := range s.Flashes(key) {
			msgs = append(msgs, f.(string))
		}
	} else {
		for _, f := range s.Flashes() {
			msgs = append(msgs, f.(string))
		}
	}

	return msgs, nil
}
