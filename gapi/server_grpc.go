package gapi

import (
	"fmt"

	db "github.com/nhoc20170861/simple-bank/db/sqlc"
	"github.com/nhoc20170861/simple-bank/pb"
	"github.com/nhoc20170861/simple-bank/token"
	"github.com/nhoc20170861/simple-bank/util"
)

// Server servers HTTP requests for our banking service
// This has two main parts: the "store" which is the database
// and the "router" which is the web server
type Server struct {
	pb.UnimplementedSimpleBankServiceServer
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
}

// NewServer creates a new gRPC server
func NewServer(config util.Config, store db.Store) (*Server, error) {
	token, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: token,
	}
	return server, nil
}
