package server

import (
	"context"
	"github.com/go-chi/chi/v5"
	"net/http"
	"simpleHttpRequest/handler"
	"simpleHttpRequest/middlewares"
	"time"
)

type Server struct {
	chi.Router
	server *http.Server
}

const (
	readTimeout       = 5 * time.Minute
	readHeaderTimeout = 30 * time.Second
	writeTimeout      = 5 * time.Minute
)

// SetupRoutes provides all the routes that can be used
func SetupRoutes() *Server {
	router := chi.NewRouter()
	router.Route("/todo", func(todo chi.Router) {
		todo.Post("/health", handler.Health)
		todo.Route("/public", func(public chi.Router) {
			public.Post("/register", handler.CreateUser)
			public.Post("/login", handler.LoginUser)
		})

		todo.Route("/users", func(private chi.Router) {
			private.Use(middlewares.AuthMiddleWareJwt)
			private.Delete("/logout", handler.LogoutUser)
			private.Get("/", handler.SearchUserWithName)
			private.Route("/task", func(task chi.Router) {
				task.Get("/", handler.GetAllTaskForTheUser)
				task.Post("/", handler.CreateNewTask)
				task.Get("/completed", handler.GetAllCompletedTaskForTheUser)
				task.Put("/status", handler.SetTaskStatus)
				task.Route("/{taskId}", func(taskID chi.Router) {
					taskID.Delete("/", handler.DeleteTask)
				})
			})
		})
	})
	return &Server{Router: router}
}

func (svc *Server) Run(port string) error {
	svc.server = &http.Server{
		Addr:              port,
		Handler:           svc.Router,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
	}
	return svc.server.ListenAndServe()
}

func (svc *Server) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return svc.server.Shutdown(ctx)
}
