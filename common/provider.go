package common

import "github.com/aws/aws-sdk-go/service/ec2"

type Provider interface {
	GetSpotFleetInput() *ec2.RequestSpotFleetInput

}