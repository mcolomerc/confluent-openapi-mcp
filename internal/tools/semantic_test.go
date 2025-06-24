package tools

import (
	"testing"
)

func TestExtractResourceFromPath_ConfigsExclusion(t *testing.T) {
	testCases := []struct {
		path     string
		expected string
		desc     string
	}{
		{
			path:     "/kafka/v3/clusters/{cluster_id}/topics/{topic_name}/configs",
			expected: "topics",
			desc:     "Topic configs path should extract 'topics', not 'configs'",
		},
		{
			path:     "/kafka/v3/clusters/{cluster_id}/topics/{topic_name}/configs/{name}",
			expected: "topics",
			desc:     "Specific topic config path should extract 'topics', not 'configs'",
		},
		{
			path:     "/kafka/v3/clusters/{cluster_id}/broker-configs",
			expected: "broker-configs",
			desc:     "Broker configs path should extract 'broker-configs'",
		},
		{
			path:     "/kafka/v3/clusters/{cluster_id}/broker-configs/{name}",
			expected: "broker-configs",
			desc:     "Specific broker config path should extract 'broker-configs'",
		},
		{
			path:     "/kafka/v3/clusters/{cluster_id}/topics",
			expected: "topics",
			desc:     "Topics path should extract 'topics'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := ExtractResourceFromPath(tc.path)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s' for path '%s'", tc.expected, result, tc.path)
			}
		})
	}
}

func TestIsLikelyResourceName_ConfigsExclusion(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
		desc     string
	}{
		{
			input:    "configs",
			expected: false,
			desc:     "'configs' should not be considered a resource name",
		},
		{
			input:    "topics",
			expected: true,
			desc:     "'topics' should be considered a resource name",
		},
		{
			input:    "broker-configs",
			expected: true,
			desc:     "'broker-configs' should be considered a resource name",
		},
		{
			input:    "clusters",
			expected: true,
			desc:     "'clusters' should be considered a resource name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := isLikelyResourceName(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for input '%s'", tc.expected, result, tc.input)
			}
		})
	}
}
