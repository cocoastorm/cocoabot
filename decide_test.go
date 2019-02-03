package main

import "testing"

func TestNormalize(t *testing.T) {
	input := "i, love pirate, game"
	expected := []string{
		"i",
		"love pirate",
		"game",
	}
	actual := normalize(input)

	for i, w := range expected {
		a := actual[i]
		if a == "" {
			t.Error("it is not okay.")
		}

		if w != a {
			t.Errorf("got %s, but expected %s", a, w)
		}
	}
}
