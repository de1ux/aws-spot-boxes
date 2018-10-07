package common

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/route53"
)

type Provider interface {
	GetName() string
	GetSpotFleetInput() *ec2.RequestSpotFleetInput
	GetRoute53RecordSets(string, *ec2.Instance) *route53.ChangeResourceRecordSetsInput
}