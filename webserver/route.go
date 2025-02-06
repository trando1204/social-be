package webserver

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *WebServer) Route() {
	s.mux.Use(middleware.Recoverer, cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "PdsJwt", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	// The home route notifies that the API is up and running
	s.mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("SOCIALAT API is up and running"))
	})
	s.mux.Get("/socket.io/", s.handleSocket())
	s.mux.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			var authRouter = apiAuth{WebServer: s}
			r.Get("/auth-method", authRouter.getAuthMethod)
			r.Post("/assertion-options", authRouter.AssertionOptions)
			r.Post("/assertion-result", authRouter.AssertionResult)
			r.Get("/check-auth-username", authRouter.CheckAuthUsername)
			r.Get("/gen-random-username", authRouter.GenRandomUsername)
			r.Post("/cancel-register", authRouter.CancelPasskeyRegister)
			r.Post("/register-start", authRouter.StartPasskeyRegister)
			r.Post("/register-finish", authRouter.FinishPasskeyRegister)
			r.Post("/register-transfer-finish", authRouter.FinishPasskeyTransferRegister)
			r.Post("/update-passkey-start", authRouter.UpdatePasskeyStart)
			r.Post("/update-passkey-finish", authRouter.UpdatePasskeyFinish)
			r.Post("/register", authRouter.register)
			r.Post("/login", authRouter.login)
		})
		r.Route("/pds", func(r chi.Router) {
			r.Use(s.loggedInMiddleware)
			var pdsRouter = apiPds{WebServer: s}
			r.Get("/get-timeline", pdsRouter.getPdsTimeline)
			r.Get("/get-pds-session", pdsRouter.getPdsSession)
		})
	})
}
