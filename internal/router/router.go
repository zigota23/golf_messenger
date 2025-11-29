package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/yourusername/golf_messenger/internal/handler"
	"github.com/yourusername/golf_messenger/internal/middleware"
	"go.uber.org/zap"
)

type Router struct {
	mux               *mux.Router
	authHandler       *handler.AuthHandler
	userHandler       *handler.UserHandler
	ttrHandler        *handler.TTRHandler
	invitationHandler *handler.InvitationHandler
	logger            *zap.Logger
	jwtSecret         string
	corsOrigins       []string
}

func NewRouter(
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	ttrHandler *handler.TTRHandler,
	invitationHandler *handler.InvitationHandler,
	logger *zap.Logger,
	jwtSecret string,
	corsOrigins []string,
) *Router {
	return &Router{
		mux:               mux.NewRouter(),
		authHandler:       authHandler,
		userHandler:       userHandler,
		ttrHandler:        ttrHandler,
		invitationHandler: invitationHandler,
		logger:            logger,
		jwtSecret:         jwtSecret,
		corsOrigins:       corsOrigins,
	}
}

func (rt *Router) SetupRoutes() http.Handler {
	api := rt.mux.PathPrefix("/api/v1").Subrouter()

	authRoutes := api.PathPrefix("/auth").Subrouter()
	authRoutes.HandleFunc("/register", rt.authHandler.Register).Methods("POST")
	authRoutes.HandleFunc("/login", rt.authHandler.Login).Methods("POST")
	authRoutes.HandleFunc("/refresh", rt.authHandler.Refresh).Methods("POST")
	authRoutes.HandleFunc("/logout", rt.authHandler.Logout).Methods("POST")

	userRoutes := api.PathPrefix("/users").Subrouter()
	userRoutes.Use(middleware.Auth(rt.jwtSecret))
	userRoutes.HandleFunc("/me", rt.userHandler.GetMe).Methods("GET")
	userRoutes.HandleFunc("/me", rt.userHandler.UpdateMe).Methods("PUT")
	userRoutes.HandleFunc("/me/password", rt.userHandler.ChangePassword).Methods("PUT")
	userRoutes.HandleFunc("/me/avatar", rt.userHandler.UploadAvatar).Methods("POST")
	userRoutes.HandleFunc("/me/avatar", rt.userHandler.DeleteAvatar).Methods("DELETE")
	userRoutes.HandleFunc("/{id}", rt.userHandler.GetUserByID).Methods("GET")
	userRoutes.HandleFunc("", rt.userHandler.SearchUsers).Methods("GET")

	ttrRoutes := api.PathPrefix("/ttrs").Subrouter()
	ttrRoutes.Use(middleware.Auth(rt.jwtSecret))
	ttrRoutes.HandleFunc("", rt.ttrHandler.CreateTTR).Methods("POST")
	ttrRoutes.HandleFunc("", rt.ttrHandler.SearchTTRs).Methods("GET")
	ttrRoutes.HandleFunc("/{id}", rt.ttrHandler.GetTTR).Methods("GET")
	ttrRoutes.HandleFunc("/{id}", rt.ttrHandler.UpdateTTR).Methods("PUT")
	ttrRoutes.HandleFunc("/{id}", rt.ttrHandler.DeleteTTR).Methods("DELETE")
	ttrRoutes.HandleFunc("/{id}/co-captains", rt.ttrHandler.AddCoCaptain).Methods("POST")
	ttrRoutes.HandleFunc("/{id}/co-captains/{userId}", rt.ttrHandler.RemoveCoCaptain).Methods("DELETE")
	ttrRoutes.HandleFunc("/{id}/join", rt.ttrHandler.JoinTTR).Methods("POST")
	ttrRoutes.HandleFunc("/{id}/leave", rt.ttrHandler.LeaveTTR).Methods("POST")
	ttrRoutes.HandleFunc("/{id}/players", rt.ttrHandler.GetPlayers).Methods("GET")
	ttrRoutes.HandleFunc("/{id}/players/{userId}", rt.ttrHandler.UpdatePlayerStatus).Methods("PUT")

	invitationRoutes := api.PathPrefix("/invitations").Subrouter()
	invitationRoutes.Use(middleware.Auth(rt.jwtSecret))
	invitationRoutes.HandleFunc("", rt.invitationHandler.CreateInvitation).Methods("POST")
	invitationRoutes.HandleFunc("/me", rt.invitationHandler.GetMyInvitations).Methods("GET")
	invitationRoutes.HandleFunc("/{id}", rt.invitationHandler.GetInvitation).Methods("GET")
	invitationRoutes.HandleFunc("/{id}/respond", rt.invitationHandler.RespondToInvitation).Methods("PUT")
	invitationRoutes.HandleFunc("/{id}", rt.invitationHandler.CancelInvitation).Methods("DELETE")

	handler := middleware.ErrorRecovery(rt.logger)(rt.mux)
	handler = middleware.Logging(rt.logger)(handler)
	handler = middleware.CORS(rt.corsOrigins)(handler)

	return handler
}
