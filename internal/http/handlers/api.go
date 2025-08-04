package handlers

import (
	"net/http"

	"github.com/andriihomiak/wallabago/internal/http/middleware"
	"github.com/andriihomiak/wallabago/internal/http/response"
)

type API struct{}

func NewAPI() *API {
	return &API{}
}

func (a *API) AuthInfo(w http.ResponseWriter, r *http.Request) {
	token := middleware.MustGetAccessToken(r)
	response.RespondOKJSON(w, r, token)
}
