package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/tarasova-school/internal/types"
	"github.com/tarasova-school/pkg/infrastruct"
	"github.com/tarasova-school/pkg/logger"
	"net/http"
)

func (h *Handlers) VKCallback(w http.ResponseWriter, r *http.Request) {
	errVK := r.FormValue("error")
	errVKDesc := r.FormValue("error_description")

	if errVK != "" {
		logger.LogError(fmt.Errorf("err with auth vk, err:%s\n%s", errVK, errVKDesc))
	}

	code := r.FormValue("code")
	accessURL := fmt.Sprintf("https://oauth.vk.com/access_token?client_id=%s&client_secret=%s&redirect_uri=https://tarasova-school.ru/api/vk/callback&code=%s", h.soc.VKAppID, h.soc.VKSecretKey, code)
	res, err := http.Get(accessURL)
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with http.Get accessURL"))
		apiErrorEncode(w, infrastruct.ErrorInternalServerError)
		return
	}

	accessToken := struct {
		AccessToken string `json:"access_token"`
		Email       string `json:"email"`
	}{}

	if err = json.NewDecoder(res.Body).Decode(&accessToken); err != nil {
		logger.LogError(errors.Wrap(err, "err with Decode access token"))
		apiErrorEncode(w, infrastruct.ErrorInternalServerError)
		return
	}

	getUserURL := fmt.Sprintf("https://api.vk.com/method/users.get?fields=bdate&access_token=%s&v=5.124", accessToken.AccessToken)

	res, err = http.Get(getUserURL)
	if err != nil {
		logger.LogError(errors.Wrap(err, "err with Get getUserURL"))
		apiErrorEncode(w, infrastruct.ErrorInternalServerError)
		return
	}

	firstName := struct {
		Response []struct {
			FirstName string `json:"first_name"`
		}
	}{}

	if err = json.NewDecoder(res.Body).Decode(&firstName); err != nil {
		logger.LogError(errors.Wrap(err, "err with Decode first name"))
		apiErrorEncode(w, infrastruct.ErrorInternalServerError)
		return
	}

	auth := &types.AuthorizeVK{Email: accessToken.Email}
	if len(firstName.Response) > 0 {
		auth.Firstname = firstName.Response[0].FirstName
	}

	token, err := h.srv.AuthorizeVK(auth)
	if err != nil {
		apiErrorEncode(w, err)
		return
	}

	apiResponseEncoder(w, token)
}
