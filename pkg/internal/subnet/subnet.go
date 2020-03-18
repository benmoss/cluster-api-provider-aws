package subnet

import (
	"math"
	"net"

	"github.com/apparentlymart/go-cidr/cidr"
	infrav1 "sigs.k8s.io/cluster-api-provider-aws/api/v1alpha3"
)

// If n=len(zones), divides the given CIDR into n+1 subnets, and then subdivides the
// first subnet into n sub-subnets to create n public and n private subnets
// Inspired by https://github.com/cloudposse/terraform-aws-dynamic-subnets
func FromZones(networkCidr string, zones []string) (infrav1.Subnets, error) {
	_, network, err := net.ParseCIDR(networkCidr)
	if err != nil {
		return nil, err
	}
	var subnets infrav1.Subnets
	publicSubnet, err := calculateSubnet(network, zones, 0)
	if err != nil {
		return nil, err
	}
	for i, zone := range zones {
		public, err := calculateSubnet(publicSubnet, zones, i)
		if err != nil {
			return nil, err
		}
		private, err := calculateSubnet(network, zones, i+1)
		if err != nil {
			return nil, err
		}

		subnets = append(subnets, &infrav1.SubnetSpec{
			IsPublic:         true,
			CidrBlock:        public.String(),
			AvailabilityZone: zone,
		})
		subnets = append(subnets, &infrav1.SubnetSpec{
			IsPublic:         false,
			CidrBlock:        private.String(),
			AvailabilityZone: zone,
		})
	}
	return subnets, nil
}

func calculateSubnet(network *net.IPNet, zones []string, index int) (*net.IPNet, error) {
	return cidr.Subnet(network, int(math.Max(1.0, math.Ceil(math.Log2(float64(len(zones)))))), index)
}
