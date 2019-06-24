package routes

import (
	"context"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"

	"gosample/server/conf"
	"gosample/server/controllers"
)

type customerWriter struct {
	http.ResponseWriter
	status int
}

func (cw customerWriter) WriteHeader(code int) {
	cw.status = code
	cw.ResponseWriter.WriteHeader(code)
}

// Logger for print time diff between req and res with route info
func loggingHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		cw := &customerWriter{w, http.StatusOK}
		next.ServeHTTP(cw, r)
		t2 := time.Now()

		log.Printf("[%s] %d %q %v", r.Method, cw.status, r.URL.String(), t2.Sub(t1))
	}

	return http.HandlerFunc(fn)
}

// Prevent abnormal shutdown while panic
func recoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %+v", err)
				log.Print(string(debug.Stack()))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// Auth check before calling controller action
func sessionHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		session, _ := controllers.SessionStore.Get(r, conf.CookieName)

		if _, ok := session.Values["loggedInUser"]; !ok {
			log.Print("Session expired...")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// Reset expiration time to rolling for each request
		session.Options.MaxAge = conf.SessionTimeout
		session.Save(r, w)

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// Put params in context for sharing them between handlers
func wrapHandler(next http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := context.WithValue(r.Context(), "params", ps)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func RouterConfig() (router *httprouter.Router) {

	/// Middle ware chain ///
	indexHandlers := alice.New(loggingHandler, recoverHandler)
	//commonHandlers := alice.New(context.ClearHandler, loggingHandler, recoverHandler, sessionHandler) // Auth check if needed
	commonHandlers := alice.New(loggingHandler, recoverHandler)

	router = httprouter.New()

	/// User ///
	router.GET("/users", wrapHandler(commonHandlers.ThenFunc(controllers.ListUser)))
	router.GET("/activeusers", wrapHandler(commonHandlers.ThenFunc(controllers.ActiveUsers)))
	router.POST("/user", wrapHandler(commonHandlers.ThenFunc(controllers.SaveUser)))
	router.PUT("/user/:id", wrapHandler(commonHandlers.ThenFunc(controllers.UpdateUser)))
	router.GET("/user/:id", wrapHandler(commonHandlers.ThenFunc(controllers.GetUser)))
	router.PUT("/user/:id/activate", wrapHandler(commonHandlers.ThenFunc(controllers.ActivateUser)))
	router.PUT("/user/:id/inactivate", wrapHandler(commonHandlers.ThenFunc(controllers.InactivateUser)))
	router.PUT("/user/:id/image", wrapHandler(commonHandlers.ThenFunc(controllers.SaveUserImage)))
	router.GET("/user/:id/image", wrapHandler(commonHandlers.ThenFunc(controllers.GetUserImage)))
	router.POST("/login", wrapHandler(commonHandlers.ThenFunc(controllers.Authenticate)))
	router.GET("/loggedInUser", wrapHandler(commonHandlers.ThenFunc(controllers.GetLoggedInUser)))

	/// Serve public ///
	router.ServeFiles("/public/*filepath", http.Dir("public"))
	router.GET("/", wrapHandler(indexHandlers.Then(http.FileServer(http.Dir("public")))))

	return
}
