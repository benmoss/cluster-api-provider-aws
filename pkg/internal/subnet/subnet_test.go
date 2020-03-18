package subnet

import (
	"testing"

	infrav1 "sigs.k8s.io/cluster-api-provider-aws/api/v1alpha3"
)

func TestFromZones(t *testing.T) {
	testCases := []struct {
		zones    []string
		cidr     string
		expected infrav1.Subnets
	}{
		{
			zones: []string{"us-east-1a", "us-east-1b", "us-east-1c", "us-east-1d", "us-east-1e", "us-east-1f"},
			cidr:  "10.0.0.0/16",
			expected: infrav1.Subnets{
				{
					IsPublic:         true,
					CidrBlock:        "10.0.0.0/22",
					AvailabilityZone: "us-east-1a",
				},
				{
					IsPublic:         true,
					CidrBlock:        "10.0.4.0/22",
					AvailabilityZone: "us-east-1b",
				},
				{
					IsPublic:         true,
					CidrBlock:        "10.0.8.0/22",
					AvailabilityZone: "us-east-1c",
				},
				{
					IsPublic:         true,
					CidrBlock:        "10.0.12.0/22",
					AvailabilityZone: "us-east-1d",
				},
				{
					IsPublic:         true,
					CidrBlock:        "10.0.16.0/22",
					AvailabilityZone: "us-east-1e",
				},
				{
					IsPublic:         true,
					CidrBlock:        "10.0.20.0/22",
					AvailabilityZone: "us-east-1f",
				},
				{
					IsPublic:         false,
					CidrBlock:        "10.0.32.0/19",
					AvailabilityZone: "us-east-1a",
				},
				{
					IsPublic:         false,
					CidrBlock:        "10.0.64.0/19",
					AvailabilityZone: "us-east-1b",
				},
				{
					IsPublic:         false,
					CidrBlock:        "10.0.96.0/19",
					AvailabilityZone: "us-east-1c",
				},
				{
					IsPublic:         false,
					CidrBlock:        "10.0.128.0/19",
					AvailabilityZone: "us-east-1d",
				},
				{
					IsPublic:         false,
					CidrBlock:        "10.0.160.0/19",
					AvailabilityZone: "us-east-1e",
				},
				{
					IsPublic:         false,
					CidrBlock:        "10.0.192.0/19",
					AvailabilityZone: "us-east-1f",
				},
			},
		},
		{
			zones: []string{"us-east-2a", "us-east-2b", "us-east-2c"},
			cidr:  "192.168.0.0/16",
			expected: infrav1.Subnets{
				{
					IsPublic:         true,
					CidrBlock:        "192.168.0.0/20",
					AvailabilityZone: "us-east-2a",
				},
				{
					IsPublic:         true,
					CidrBlock:        "192.168.16.0/20",
					AvailabilityZone: "us-east-2b",
				},
				{
					IsPublic:         true,
					CidrBlock:        "192.168.32.0/20",
					AvailabilityZone: "us-east-2c",
				},
				{
					IsPublic:         false,
					CidrBlock:        "192.168.64.0/18",
					AvailabilityZone: "us-east-2a",
				},
				{
					IsPublic:         false,
					CidrBlock:        "192.168.128.0/18",
					AvailabilityZone: "us-east-2b",
				},
				{
					IsPublic:         false,
					CidrBlock:        "192.168.192.0/18",
					AvailabilityZone: "us-east-2c",
				},
			},
		},
		{
			zones: []string{"us-east-2a", "us-east-2b", "us-east-2c"},
			cidr:  "192.168.0.0/20",
			expected: infrav1.Subnets{
				{
					IsPublic:         true,
					CidrBlock:        "192.168.0.0/24",
					AvailabilityZone: "us-east-2a",
				},
				{
					IsPublic:         true,
					CidrBlock:        "192.168.1.0/24",
					AvailabilityZone: "us-east-2b",
				},
				{
					IsPublic:         true,
					CidrBlock:        "192.168.2.0/24",
					AvailabilityZone: "us-east-2c",
				},
				{
					IsPublic:         false,
					CidrBlock:        "192.168.4.0/22",
					AvailabilityZone: "us-east-2a",
				},
				{
					IsPublic:         false,
					CidrBlock:        "192.168.8.0/22",
					AvailabilityZone: "us-east-2b",
				},
				{
					IsPublic:         false,
					CidrBlock:        "192.168.12.0/22",
					AvailabilityZone: "us-east-2c",
				},
			},
		},
	}
	for _, c := range testCases {
		actual, err := FromZones(c.cidr, c.zones)
		if err != nil {
			t.Errorf("failed to calculate subnets: %v", err)
			return
		}
		if len(c.expected) != len(actual) {
			t.Errorf("expected to have %d subnets, got %d", len(c.expected), len(actual))
			return
		}
		for _, e := range c.expected {
			var found bool
			for _, a := range actual {
				if e.CidrBlock == a.CidrBlock &&
					e.IsPublic == a.IsPublic &&
					e.AvailabilityZone == a.AvailabilityZone {
					found = true
				}
			}
			if !found {
				var printable []infrav1.SubnetSpec
				for i := 0; i < len(actual); i++ {
					printable = append(printable, *actual[i])
				}

				t.Errorf("failed to find subnet %#v in %#v", e, printable)
				return
			}
		}
	}
}
