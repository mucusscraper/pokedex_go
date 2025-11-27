package main

import "testing"

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "   hello world   ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "PIKACHU bulbasaur ",
			expected: []string{"pikachu", "bulbasaur"},
		},
		{
			input:    "FERaliGATOr treecko tre CKO",
			expected: []string{"feraligator", "treecko", "tre", "cko"},
		},
	}
	for _, c := range cases {
		actual := cleanInput(c.input)
		if len(actual) != len(c.expected) {
			t.Errorf("Not the same size")
		}
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]
			if word != expectedWord {
				t.Errorf("Not the same results")
			}
		}
	}
}
