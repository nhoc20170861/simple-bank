package gapi

import (
	"context"
	"log"

	"github.com/lib/pq"
	db "github.com/nhoc20170861/simple-bank/db/sqlc"
	"github.com/nhoc20170861/simple-bank/pb"
	"github.com/nhoc20170861/simple-bank/util"
	val "github.com/nhoc20170861/simple-bank/validator"
	"github.com/nhoc20170861/simple-bank/worker"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateUser creates a new user with the given username, password, full_name, and email.
// This endpoint is idempotent.
func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	// Validate the request
	violations := validateCreateUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	// Hash the password
	hashedPassword, err := util.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
	}

	// Create the user
	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Email:          req.Email,
	}

	log.Printf("creating user with username: %s, full_name: %s, email: %s\n",
		arg.Username, arg.FullName, arg.Email)

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				log.Printf("error creating user: %v\n", err)
				return nil, status.Errorf(codes.AlreadyExists, "user already exists: %v", err)
			}
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	// Distribute the task to send the verification email
	taskPayload := &worker.PayloadSendVerifyEmail{
		Username: user.Username,
	}
	err = server.taskDistributor.DistributeTaskSendVerifyEmail(ctx, taskPayload, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to distribute task: %v", err)
	}

	resp := &pb.CreateUserResponse{
		User: convertUser(user),
	}

	return resp, nil
}

func validateCreateUserRequest(req *pb.CreateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(req.Username); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}
	if err := val.ValidatePassword(req.Password); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}
	if err := val.ValidateFullName(req.FullName); err != nil {
		violations = append(violations, fieldViolation("full_name", err))
	}
	if err := val.ValidateEmail(req.Email); err != nil {
		violations = append(violations, fieldViolation("email", err))
	}
	return violations
}
