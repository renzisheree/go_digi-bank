package api

import (
	"github.com/gin-gonic/gin"
	db "tutorial.sqlc.dev/app/db/sqlc"
)

// Server will serve HTTP requests for banking service.
type Server struct {
	store  *db.Store
	router *gin.Engine
}

// Function to create a new server with the given store
func NewServer(store *db.Store) *Server {
	server := &Server{
		store: store}
	router := gin.Default()

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.getListAccount)
	router.PUT("/accounts", server.updateAccount)
	router.DELETE("/accounts/:id", server.deleteAccount)
	server.router = router
	return server
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}
func (server *Server) Stop() {
}
