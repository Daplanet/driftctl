// GENERATED, DO NOT EDIT THIS FILE
package aws

const AwsInternetGatewayResourceType = "aws_internet_gateway"

type AwsInternetGateway struct {
	Arn     *string           `cty:"arn" computed:"true"`
	Id      string            `cty:"id" computed:"true"`
	OwnerId *string           `cty:"owner_id" computed:"true"`
	Tags    map[string]string `cty:"tags"`
	VpcId   *string           `cty:"vpc_id"`
}

func (r *AwsInternetGateway) TerraformId() string {
	return r.Id
}

func (r *AwsInternetGateway) TerraformType() string {
	return AwsInternetGatewayResourceType
}
