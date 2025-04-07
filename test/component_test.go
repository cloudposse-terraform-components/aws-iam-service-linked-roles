package test

import (
	"context"
	"fmt"
	"testing"

	helper "github.com/cloudposse/test-helpers/pkg/atmos/component-helper"
	"github.com/cloudposse/test-helpers/pkg/atmos"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/stretchr/testify/assert"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

type Tags struct {
	Attributes  string `json:"Attributes"`
	Environment string `json:"Environment"`
	Name        string `json:"Name"`
	Namespace   string `json:"Namespace"`
	Stage       string `json:"Stage"`
	Tenant      string `json:"Tenant"`
}

type Role struct {
	ARN            string `json:"arn"`
	AWSServiceName string `json:"aws_service_name"`
	CreateDate     string `json:"create_date"`
	CustomSuffix   string `json:"custom_suffix"`
	Description    string `json:"description"`
	ID             string `json:"id"`
	Name           string `json:"name"`
	Path           string `json:"path"`
	Tags           Tags   `json:"tags"`
	TagsAll        Tags   `json:"tags_all"`
	UniqueID       string `json:"unique_id"`
}

type ComponentSuite struct {
	helper.TestSuite
}

func (s *ComponentSuite) TestBasic() {
	const component = "iam-service-linked-roles/basic"
	const stack = "default-test"
	const awsRegion = "us-east-2"

	inputs := map[string]interface{}{}

	defer s.DestroyAtmosComponent(s.T(), component, stack, &inputs)
	options, _ := s.DeployAtmosComponent(s.T(), component, stack, &inputs)
	assert.NotNil(s.T(), options)

	accountID := aws.GetAccountId(s.T())

	var role map[string]Role

	atmos.OutputStruct(s.T(), options, "service_linked_roles", &role)
	assert.Equal(s.T(), role["spot_amazonaws_com_test"].Name, "AWSServiceRoleForLexBots")
	assert.Equal(s.T(), role["spot_amazonaws_com_test"].ARN,  fmt.Sprintf("arn:aws:iam::%s:role/aws-service-role/lex.amazonaws.com/AWSServiceRoleForLexBots", accountID))
	assert.Equal(s.T(), role["spot_amazonaws_com_test"].AWSServiceName, "lex.amazonaws.com")
	assert.Equal(s.T(), role["spot_amazonaws_com_test"].Description, "AWSServiceRoleForEC2Spot Service-Linked Role for EC2 Spot")
	assert.Equal(s.T(), role["spot_amazonaws_com_test"].Tags.Environment, "ue2")
	assert.True(s.T(), strings.HasPrefix(role["spot_amazonaws_com_test"].Tags.Name, "eg-default-ue2-test-"))
	assert.Equal(s.T(), role["spot_amazonaws_com_test"].Tags.Namespace, "eg")
	assert.Equal(s.T(), role["spot_amazonaws_com_test"].Tags.Stage, "test")
	assert.Equal(s.T(), role["spot_amazonaws_com_test"].Tags.Tenant, "default")
	assert.NotEmpty(s.T(), role["spot_amazonaws_com_test"].UniqueID)

	// Initialize AWS SDK v2 client
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(awsRegion))
	assert.NoError(s.T(), err)

	iamClient := iam.NewFromConfig(cfg)

	roleName := "AWSServiceRoleForLexBots"
	roleOutput, err := iamClient.GetRole(context.TODO(), &iam.GetRoleInput{
		RoleName: &roleName,
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), *roleOutput.Role.Arn, role["spot_amazonaws_com_test"].ARN)

	s.DriftTest(component, stack, &inputs)
}

func (s *ComponentSuite) TestEnabledFlag() {
	const component = "iam-service-linked-roles/disabled"
	const stack = "default-test"
	const awsRegion = "us-east-2"

	s.VerifyEnabledFlag(component, stack, nil)
}

func TestRunSuite(t *testing.T) {
	suite := new(ComponentSuite)
	helper.Run(t, suite)
}
