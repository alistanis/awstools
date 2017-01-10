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
		log.Printf("Searching for an appropriate instance to snapshot... filters: %s\n", filters)
		// this assumes your instancesOutput are tagged with a Name key
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
			if !req.ChooseFirstResult {
				return nil, errors.New("Mutliple results found without specifying to choose the first result. Please choose more restrictive filters.")
			}
		}

		instance := instances[0]

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
		fmt.Println(name)
		input.Name = aws.String(name)

		output, err := service.CreateImage(&calculate.CreateImageInput{AwsInput:&ec2.CreateImageInput{InstanceId:instance.InstanceId}})
		if err != nil {
			return nil, err
		}


		// TODO - finish this section
		/*var finished bool
		for {
			fmt.Println("Checking images...")
			finished, err = ec2.CheckImages(imageIds)
			if err != nil {
				log.Fatal(err)
			}
			if finished {
				fmt.Println("Images have finished!")
				break
			}
			fmt.Println("Sleeping for 5 seconds before checking again...")
			time.Sleep(time.Duration(time.Second * 5))
		}
		*/

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

func CheckImages(ids []string) (bool, error) {
	idptrs := util.StringSliceToPointers(ids)
	images, err := service.DescribeImages(&ec2.DescribeImagesInput{ImageIds: idptrs})
	if err != nil {
		return false, err
	}
	for _, s := range images.Images {
		if *s.State != "available" {
			return false, nil
		}
	}
	return true, nil
}