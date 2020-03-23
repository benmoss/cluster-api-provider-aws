package subnet

import (
	"math"
	"net"
	"strings"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/pkg/errors"
	infrav1 "sigs.k8s.io/cluster-api-provider-aws/api/v1alpha3"
)

const maxCIDRMask = 28

var errInvalidNetwork = errors.New("Could not find a valid subnet configuration")

// If n=len(zones), divides the given CIDR into n+1 subnets, and then subdivides the
// first subnet into n sub-subnets to create n public and n private subnets
// Inspired by https://github.com/cloudposse/terraform-aws-dynamic-subnets
func FromZones(existingSubnets []string, networkCidr string, zones []string) (infrav1.Subnets, error) {
	_, network, err := net.ParseCIDR(networkCidr)
	if err != nil {
		return nil, err
	}
	numZones := len(zones)
	var existingNets []*net.IPNet
	for _, s := range existingSubnets {
		_, parsed, err := net.ParseCIDR(s)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to parse subnet %q", s)
		}
		existingNets = append(existingNets, parsed)
	}

	// First, the inner loop tries to use offsets to find the largest subnets
	// that can fit with existing subnets
	// If that fails, start increasing the number of zones to force the algorithm
	// to select smaller subnets
	// Errors if the mask size of one of the subnets would be >=28
Outer:
	for i := 0; ; i++ {
	Offsets:
		for j := 0; j < 50; j++ {
			var result infrav1.Subnets
			publicSubnet, err := calculateSubnet(network, numZones+i, j)
			if err != nil {
				if strings.HasPrefix(err.Error(), "prefix extension") {
					continue Outer
				}
				return nil, err
			}
			if size, _ := publicSubnet.Mask.Size(); size >= maxCIDRMask {
				return nil, errInvalidNetwork
			}
			newNets := append(existingNets, publicSubnet)
			if err := cidr.VerifyNoOverlap(newNets, network); err != nil {
				continue
			}

			for k, zone := range zones {
				// carve the public network into smaller subnets
				public, err := calculateSubnet(publicSubnet, numZones+i, k)
				if err != nil {
					return nil, err
				}
				if size, _ := public.Mask.Size(); size >= maxCIDRMask {
					return nil, errInvalidNetwork
				}
				// offset by 1 to avoid the already allocated public subnet
				private, err := calculateSubnet(network, numZones+i, j+k+1)
				if err != nil {
					if strings.HasPrefix(err.Error(), "prefix extension") {
						continue Outer
					}
					return nil, err
				}
				// we already know the public subnet is not overlapping
				newNets := append(newNets, private)
				if err := cidr.VerifyNoOverlap(newNets, network); err != nil {
					continue Offsets
				}

				result = append(result, &infrav1.SubnetSpec{
					IsPublic:         true,
					CidrBlock:        public.String(),
					AvailabilityZone: zone,
				})
				result = append(result, &infrav1.SubnetSpec{
					IsPublic:         false,
					CidrBlock:        private.String(),
					AvailabilityZone: zone,
				})
			}
			return result, nil
		}
	}
}

// Takes an existing network and calculates the number of new bits needed to
// divide into at least numZones subnetworks. Returns the sub-network specified
// by the given network number.
func calculateSubnet(network *net.IPNet, numZones int, num int) (*net.IPNet, error) {
	return cidr.Subnet(network, int(math.Max(1.0, math.Ceil(math.Log2(float64(numZones))))), num)
}
