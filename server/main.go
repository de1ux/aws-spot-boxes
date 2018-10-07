package main

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"net"

	"github.com/de1ux/aws-spot-boxes/common"
	"github.com/de1ux/aws-spot-boxes/generated/api"

	"google.golang.org/grpc"
	"github.com/grpc-ecosystem/go-grpc-middleware/auth"
	verifier "github.com/futurenda/google-auth-id-token-verifier"
)

var (
	c *common.Config
)

type server struct{}

func init() {
	var err error
	c, err = common.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
}

func (s *server) StartBox(ctx context.Context, request *api.StartBoxRequest) (*api.StartBoxResponse, error) {
	return &api.StartBoxResponse{}, nil
}

func (s *server) KeepAlive(stream api.AWSSpotBoxes_KeepAliveServer) error {
	return nil
}

func authorizeMiddleware(ctx context.Context) (context.Context, error) {
	m, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, status.Error(codes.Unauthenticated, "No metadata present")
	}

	tokens, ok := m["id_token"]
	if !ok {
		return ctx, status.Error(codes.Unauthenticated, "No id_token present")
	}

	token := tokens[0]

	v := verifier.Verifier{}
	err := v.VerifyIDToken(token, []string{
		c.ClientId,
	})
	if err != nil {
		log.Printf("Failed to verify id_token: %s", err)
		return ctx, status.Error(codes.Unauthenticated, "Failed to verify id_token")
	}

	claimSet, err := verifier.Decode(token)
	if err != nil {
		log.Printf("Failed to decode id_token: %s", err)
		return ctx, status.Error(codes.Unauthenticated, "Failed to verify id_token")
	}

	return context.WithValue(ctx, "claimSet", *claimSet), nil
}

func main() {
	c, err := common.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen("tcp", ":"+c.ServerPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(authorizeMiddleware)),
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(authorizeMiddleware)))
	api.RegisterAWSSpotBoxesServer(s, &server{})

	log.Print("Server running...")
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
	log.Print("Server running...done")
}
