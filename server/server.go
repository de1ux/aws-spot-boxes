package server

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"time"

	"github.com/de1ux/aws-spot-boxes/common"
	"github.com/de1ux/aws-spot-boxes/generated/api"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	verifier "github.com/futurenda/google-auth-id-token-verifier"
	"google.golang.org/grpc"
)

var (
	c *common.Config
)

type server struct{
	provider common.Provider
}

func init() {
	var err error
	c, err = common.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
}

func (s *server) StartBox(ctx context.Context, request *api.StartBoxRequest) (*api.StartBoxResponse, error) {
	awsConfig := aws.Config{Region: aws.String(c.ServerAWSRegion), Credentials: credentials.NewStaticCredentials(c.ServerAWSAccessID, c.ServerAWSSecretKey, "")}

	session, err := session.NewSession(&awsConfig)
	if err != nil {
		return nil, err
	}

	ec2Svc := ec2.New(session)

	input := s.provider.GetSpotFleetInput()
	log.Println(input)
	result, err := ec2Svc.RequestSpotFleet(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return nil, err
	}

	fmt.Println(result)

	runningFleet := false
	spotInstanceRequestId := aws.String("")
	println("Waiting for fleet to become ready...")
	for !runningFleet {
		describe, err := ec2Svc.DescribeSpotFleetInstances(&ec2.DescribeSpotFleetInstancesInput{
			SpotFleetRequestId: result.SpotFleetRequestId,
		})
		if err != nil {
			panic(err)
		}

		if len(describe.ActiveInstances) > 0 {
			runningFleet = true
			spotInstanceRequestId = describe.ActiveInstances[0].SpotInstanceRequestId
			break
		}
		time.Sleep(time.Second * 2)
	}
	println("Waiting for fleet to become ready...done")

	println("Waiting for instance to become ready...")
	runningSpot := false
	for !runningSpot {
		describeResponse, err := ec2Svc.DescribeSpotInstanceRequests(&ec2.DescribeSpotInstanceRequestsInput{
			SpotInstanceRequestIds: []*string{spotInstanceRequestId},
		})
		if err != nil {
			panic(err)
		}

		if len(describeResponse.SpotInstanceRequests) == 0 {
			continue
		}

		println(*describeResponse.SpotInstanceRequests[0].State)
		if *describeResponse.SpotInstanceRequests[0].State == "open" {
			runningSpot = true
			break
		}

		time.Sleep(time.Second * 2)
	}
	println("Waiting for instance to become ready...done")



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

func Run(p common.Provider) error {
	c, err := common.GetConfig()
	if err != nil {
		return err
	}

	lis, err := net.Listen("tcp", ":"+c.ServerPort)
	if err != nil {
		return err
	}

	s := grpc.NewServer()
//		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(authorizeMiddleware)),
//		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(authorizeMiddleware)))
	api.RegisterAWSSpotBoxesServer(s, &server{provider: p})

	log.Print("Server running...")
	if err := s.Serve(lis); err != nil {
		return err
	}
	return nil
}
