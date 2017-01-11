package handlers

import (
	"encoding/json"
	"errors"
	"log"

	"fmt"

	"os"

	"github.com/alistanis/awstools/awsregions"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/eawsy/aws-lambda-go/service/lambda/runtime"
	"github.com/thisisfineio/calculate"
	"strings"
	"github.com/alistanis/util"
	"time"
)

const (
	InstanceSnapshot    = "instance_snapshot"
	AutoscalingSnapshot = "autoscaling_snapshot"
	EBSSnapshot = "ebs_snapshot"
)



var (
	service *calculate.EC2
)

func init() {
	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = awsregions.USEast1
	}
	// This will pick up credentials in the order of precedence specified by the AWS SDK
	service = calculate.NewEC2(awsRegion)
}

type CreateSnapshotRequest struct {
	Filters       map[string][]*string `json:"-"`
	FiltersStrings map[string]string `json:"filters"`
	SnapshotType  string    `json:"snapshot_type"`
	ChooseFirstResult bool `json:"choose_first_result"`
	NoReboot bool `json:"no_reboot"`
}

func (c *CreateSnapshotRequest)splitFilters() {
	m := make(map[string][]*string)
	for k, v := range c.FiltersStrings {
		patterns := strings.Split(v, ",")
		ptrs := util.StringSliceToPointers(patterns)
		m[k] = ptrs
	}
	c.Filters = m
}

func (c *CreateSnapshotRequest) EC2Filters() []*ec2.Filter{
	filters := make([]*ec2.Filter, len(c.Filters))
	for k, v := range c.Filters {
		f := &ec2.Filter{Name: &k, Values: v}
		filters = append(filters, f)
	}
	return filters
}

func (c *CreateSnapshotRequest) Validate() error {
	c.splitFilters()
	if len(c.Filters) == 0 || c.SnapshotType == "" {
		return fmt.Errorf("handlers: filters and snapshot_type are required parameters. Got: filters=%s and snapshot_type=%s", c.FiltersStrings, c.SnapshotType)
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

	filters := req.EC2Filters()

	switch req.SnapshotType {
	case InstanceSnapshot:
		err := createAMISnapshot(filters, req)
		// explicitly return error here because i don't know how lambda treats both a value and an error
		if err != nil {
			return nil, err
		}
		return fmt.Sprintf("Instance snapshot successfully completed at %s", time.Now().Format("2006-01-02-15-04-05")), nil

	case AutoscalingSnapshot:
		log.Println("Do autoscaling snapshot things, such as finding an ASG, an instance, snapshotting the instance, updating the ASG, respinning other instancesOutput, etc")
	case EBSSnapshot:
		log.Println("Do EBS related things here.")
	}

	return "Exiting without errors :)", nil
}

func CheckSnapshot(evt json.RawMessage, ctx *runtime.Context) (interface{}, error) {
	return "", nil
}

func DeleteSnapshot(evt json.RawMessage, ctx *runtime.Context) (interface{}, error) {
	return "", nil
}

func createAMISnapshot(filters []*ec2.Filter, req *CreateSnapshotRequest) error {
	log.Printf("Searching for an appropriate instance to snapshot... filters: %s\n", filters)

	instance, err := findInstance(filters, req.ChooseFirstResult)
	if err != nil {
		return err
	}
	input := &ec2.CreateImageInput{NoReboot: &req.NoReboot}

	curTime := time.Now().Format("2006-01-02-15-04-05")
	input.Description = aws.String(fmt.Sprintf("Created by running lambda on %s. Originating VPC: %s, Original InstanceID: %s", curTime, *instance.VpcId, *instance.ImageId))
	input.InstanceId = instance.InstanceId
	var name string
	for _, t := range instance.Tags {
		if t.Key != nil && t.Value != nil {
			if *t.Key == "Name" {
				name = fmt.Sprintf("%s-image-%s", *t.Value, curTime)
			}
		}
	}
	if name == "" {
		name = fmt.Sprintf("%s-image-%s", *instance.InstanceId, curTime)
	}

	input.Name = aws.String(name)

	output, err := service.CreateImage(&calculate.CreateImageInput{AwsInput:&ec2.CreateImageInput{InstanceId:instance.InstanceId}})
	if err != nil {
		return err
	}

	return checkInstance(*output.AwsOutput.ImageId)
}

func findInstance(filters []*ec2.Filter, chooseFirstResult bool) (*ec2.Instance, error) {
	input := &ec2.DescribeInstancesInput{Filters: filters}

	instancesOutput, err := service.DescribeEC2Instances(input)
	if err != nil {
		return nil, err
	}
	if len(instancesOutput.Reservations) == 0 {
		return nil, fmt.Errorf("No instancesOutput found matching filters: %s", filters)
	}

	instances := make([]*ec2.Instance, 0)
	for _, r := range instancesOutput.Reservations {
		instances = append(instances, r.Instances...)
	}

	if len(instances) > 1 {
		if !chooseFirstResult {
			return nil, errors.New("Mutliple results found without specifying to choose the first result. Please choose more restrictive filters.")
		}
	}

	instance := instances[0]
	return instance, nil
}

func checkInstance(imageID string) error {
	for {
		log.Println("Checking images...")
		image, err := service.DescribeImages(&calculate.DescribeImagesInput{AwsInput:&ec2.DescribeImagesInput{ImageIds:[]*string{&imageID}}})
		if err != nil {
			return err
		}
		state := *image.AwsOutput.Images[0].State
		finished := state == "available"
		if finished {
			log.Println("Image has finished!")
			break
		}
		log.Println("Sleeping for 5 seconds before checking again...")
		time.Sleep(time.Duration(time.Second * 5))
	}
	return nil
}