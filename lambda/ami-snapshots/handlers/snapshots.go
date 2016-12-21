package handlers

import (
	"encoding/json"
	"log"

	"fmt"

	"github.com/alistanis/awstools/awsregions"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/eawsy/aws-lambda-go/service/lambda/runtime"
)

const (
	InstanceSnapshot    = "instance_snapshot"
	AutoscalingSnapshot = "autoscaling_snapshot"
)

var (
	service *ec2.EC2
)

func init() {
	// This will pick up credentials in the order of precedence specified by the AWS SDK
	service = ec2.New(session.New(&aws.Config{Region: aws.String(awsregions.USEast1)}))
}

type CreateSnapshotRequest struct {
	SnapshotPatterns []*string `json:"snapshot_patterns"` // because the aws sdk uses pointers for everything, and the json unmarshaller is awesome, we'll use pointers to strings
	SnapshotType     string    `json:"snapshot_type"`
}

func (c *CreateSnapshotRequest) Validate() error {
	if len(c.SnapshotPatterns) == 0 || c.SnapshotType == "" {
		return fmt.Errorf("handlers: snapshot_pattern and snapshot_type are required parameters. Got: snapshot_pattern=%s and snapshot_type=%s", c.SnapshotPatterns, c.SnapshotType)
	}
	return nil
}

func CreateSnapshot(evt json.RawMessage, ctx *runtime.Context) (interface{}, error) {
	req := &CreateSnapshotRequest{}
	// I'm curious why the lambda runtime doesn't use io.Reader instead of json.RawMessage... will definitely slow down potentially large payloads from being decoded over the wire
	// worth reaching out to them to see if they're willing to change their interface?
	err := json.Unmarshal(evt, &req)
	if err != nil {
		return nil, err
	}
	if err = req.Validate(); err != nil {
		return nil, err
	}

	switch req.SnapshotType {
	case InstanceSnapshot:
		log.Printf("Searching for an appropriate instance to snapshot... patterns: %s\n", req.SnapshotPatterns)
		// this assumes your instances are tagged with a Name key
		n := "Name"
		input := &ec2.DescribeInstancesInput{Filters: []*ec2.Filter{{Name: &n, Values: req.SnapshotPatterns}}}

		// TODO - wrap this like I have done for work libraries - why this isn't part of the default lib I have no idea
		instances, err := service.DescribeInstances(input)
		if err != nil {
			return nil, err
		}
		for instances.NextToken != nil {
			input.NextToken = instances.NextToken
			i, err := service.DescribeInstances(input)
			if err != nil {
				return nil, err
			}
			instances.Reservations = append(instances.Reservations, i.Reservations...)
			instances.NextToken = i.NextToken
		}
		if len(instances.Reservations == 0) {
			return nil, fmt.Errorf("No instances found matching name tags: %s", req.SnapshotPatterns)
		}

	case AutoscalingSnapshot:
		log.Println("Do autoscaling snapshot things, such as finding an ASG, an instance, snapshotting the instance, updating the ASG, respinning other instances, etc")
	}

	return "Exiting without errors :)", nil
}

func CheckSnapshot(evt json.RawMessage, ctx *runtime.Context) (interface{}, error) {
	return "", nil
}

func DeleteSnapshot(evt json.RawMessage, ctx *runtime.Context) (interface{}, error) {
	return "", nil
}
