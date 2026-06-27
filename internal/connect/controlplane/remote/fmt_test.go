package connectremotecontrolplane

import (
	"testing"
)

func TestFormatEndpoint(t *testing.T) {
	var testCases = []struct {
		input  string
		output string
	}{
		{
			"http://123.123.123.123:51121",
			"http://123.123.123.123:51121",
		},
		{
			"123.123.123.123:51121",
			"http://123.123.123.123:51121",
		},
		{
			"123.123.123.123",
			"http://123.123.123.123:51121",
		},
		{
			"http://123.123.123.123",
			"http://123.123.123.123:51121",
		},
		{
			"http://caas.edgeworx.io:51121",
			"http://caas.edgeworx.io:51121",
		},
		{
			"caas.edgeworx.io:51121",
			"http://caas.edgeworx.io:51121",
		},
		{
			"caas.edgeworx.io",
			"http://caas.edgeworx.io:51121",
		},
		{
			"http://caas.edgeworx.io",
			"http://caas.edgeworx.io:51121",
		},
		{
			"http://caas.edgeworx.io/api/v3",
			"http://caas.edgeworx.io:51121/api/v3",
		},
		{
			"https://caas.edgeworx.io/api/v3",
			"https://caas.edgeworx.io/api/v3",
		},
	}
	for _, c := range testCases {
		u, err := formatEndpoint(c.input)
		if u == nil {
			t.Fatalf("%s %s", c.input, err.Error())
		}
		if err != nil {
			t.Fatalf("%s %s %s", c.input, u.String(), err.Error())
		}
		if u.String() != c.output {
			t.Fatalf("%s %s", c.input, u.String())
		}
	}
}
