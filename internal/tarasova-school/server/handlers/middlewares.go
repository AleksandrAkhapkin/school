package handlers

import (
	"fmt"
	"github.com/tarasova-school/internal/types"
	"github.com/tarasova-school/pkg/infrastruct"
	"github.com/tarasova-school/pkg/logger"
	"net/http"
)

func (h *Handlers) RecoverPanic(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {

			r := recover()
			if r == nil {
				return
			}

			err := fmt.Errorf("PANIC:'%v'\nRecovered in: %s", r, infrastruct.IdentifyPanic())
			logger.LogError(err)
			apiErrorEncode(w, infrastruct.ErrorInternalServerError)
		}()

		handler.ServeHTTP(w, r)

	})
}

func (h *Handlers) CheckRoleTeacher(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := infrastruct.GetClaimsByRequest(r, h.secretKey)
		if err != nil {
			apiErrorEncode(w, err)
			return
		}

		if claims.Role != types.RoleTeacher {
			apiErrorEncode(w, infrastruct.ErrorPermissionDenied)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

func (h *Handlers) CheckRoleAdmin(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := infrastruct.GetClaimsByRequest(r, h.secretKey)
		if err != nil {
			apiErrorEncode(w, err)
			return
		}

		if claims.Role != types.RoleAdmin {
			apiErrorEncode(w, infrastruct.ErrorPermissionDenied)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

func (h *Handlers) CheckRoleStudent(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := infrastruct.GetClaimsByRequest(r, h.secretKey)
		if err != nil {
			apiErrorEncode(w, err)
			return
		}

		if claims.Role != types.RoleStudent {
			apiErrorEncode(w, infrastruct.ErrorPermissionDenied)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

func (h *Handlers) CheckRoleAdminAndTeacher(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := infrastruct.GetClaimsByRequest(r, h.secretKey)
		if err != nil {
			apiErrorEncode(w, err)
			return
		}

		if claims.Role != types.RoleAdmin && claims.Role != types.RoleTeacher {
			apiErrorEncode(w, infrastruct.ErrorPermissionDenied)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

func (h *Handlers) RecordRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := infrastruct.GetClaimsByRequest(r, h.secretKey)
		if err != nil {
			handler.ServeHTTP(w, r)
			return
		}

		if err := h.srv.RecordTime(&types.RecordTime{
			UserID:     claims.UserID,
			RequestURL: r.RequestURI,
		}); err != nil {
			logger.LogError(err)
		}

		handler.ServeHTTP(w, r)
	})
}

func (h *Handlers) CheckUserInDBUsers(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := infrastruct.GetClaimsByRequest(r, h.secretKey)
		if err != nil {
			handler.ServeHTTP(w, r)
			return
		}
		if err = h.srv.CheckUserInDBUsers(claims.UserID); err != nil {
			apiErrorEncode(w, err)
			return
		}

		handler.ServeHTTP(w, r)
	})
}
