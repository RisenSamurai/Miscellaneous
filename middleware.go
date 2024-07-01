package handlers

import (
	"Oaks/pkg/session"
	"net/http"
)

func AuthMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		_, err := session.Store.Get(r, "user")
		if err != nil {
			session.CreateSession(w, r)
			http.Redirect(w, r, "/client/sign-in", 303)
			return

		}

		if !Sess.Auth {
			http.Redirect(w, r, "/client/sign-in", 303)
			return

		}

		next.ServeHTTP(w, r)

	})
}
