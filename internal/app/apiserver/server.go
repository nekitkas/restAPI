package apiserver

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type Server struct {
	config *Config
	router *mux.Router
	logger *log.Logger
}

func New(config *Config) *Server {
	return &Server{
		config: config,
		logger: log.Default(),
	}
}

func (s *Server) Start() error {
	s.logger.Printf("Server started at port: %v\n", s.config.Port)
	return http.ListenAndServe(s.config.Port, nil)
}
