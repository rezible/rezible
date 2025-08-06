package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	rez "github.com/rezible/rezible"
	"github.com/rs/zerolog/log"
)

func NewWebhooksHandler(chat rez.ChatService) http.Handler {
	mux := chi.NewMux()
	mux.Mount("/slack", chat.GetWebhooksHandler())
	mux.NotFound(func(w http.ResponseWriter, r *http.Request) {
		log.Warn().Str("path", r.URL.Path).Msg("unhandled webhook request")
	})
	return mux
}
