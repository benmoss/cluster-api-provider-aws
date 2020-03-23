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
			return nil, errors.Wrapf(err, "failed to parse subnet %q", s)
		}
		existingNets = append(existingNets, parsed)
	}

CIDRs:
	for i := 0; i < 100; i++ {
	Offsets:
		for j := 0; j < 100; j++ {
			var result infrav1.Subnets
			publicSubnet, err := calculateSubnet(network, numZones+i, j)
			if err != nil {
				if strings.HasPrefix(err.Error(), "prefix extension") {
					continue CIDRs
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
				public, err := calculateSubnet(publicSubnet, numZones+i, k)
				if err != nil {
					return nil, err
				}
				if size, _ := public.Mask.Size(); size >= maxCIDRMask {
					return nil, errInvalidNetwork
				}
				private, err := calculateSubnet(network, numZones+i, j+k+1)
				if err != nil {
					if strings.HasPrefix(err.Error(), "prefix extension") {
						continue CIDRs
					}
					return nil, err
				}
				newNets := append(existingNets, public)
				newNets = append(newNets, private)
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
	return nil, errors.New("Could not find a valid subnet configuration")
}

func calculateSubnet(network *net.IPNet, numZones int, index int) (*net.IPNet, error) {
	return cidr.Subnet(network, int(math.Max(1.0, math.Ceil(math.Log2(float64(numZones))))), index)
}
