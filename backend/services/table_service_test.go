package services

import (
	"iptables-management-backend/utils"
	"testing"
)

func TestParsePacketCount(t *testing.T) {

	testCases := []struct {
		input    string
		expected uint64
		name     string
	}{
		{"231K", 231000, "231K should be 231000"},
		{"21M", 21000000, "21M should be 21000000"},
		{"2319K", 2319000, "2319K should be 2319000"},
		{"213K", 213000, "213K should be 213000"},
		{"9995K", 9995000, "9995K should be 9995000"},
		{"601K", 601000, "601K should be 601000"},
		{"11M", 11000000, "11M should be 11000000"},
		{"10M", 10000000, "10M should be 10000000"},
		{"462K", 462000, "462K should be 462000"},
		{"1G", 1000000000, "1G should be 1000000000"},
		{"2.5K", 2500, "2.5K should be 2500"},
		{"1.2M", 1200000, "1.2M should be 1200000"},
		{"123", 123, "123 should be 123"},
		{"0", 0, "0 should be 0"},
		{"--", 0, "-- should be 0"},
		{"", 0, "empty string should be 0"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := utils.ParsePacketCount(tc.input)
			if result != tc.expected {
				t.Errorf("parsePacketCount(%s) = %d, expected %d", tc.input, result, tc.expected)
			}
		})
	}
}
