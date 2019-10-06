package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCToF(t *testing.T) {
	testCases := []struct {
		C     float64
		WantF float64
	}{
		{0, 32},
		{1, 33.8},
		{10, 50},
		{21, 69.8},
		{35, 95},
		{100, 212},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%gC", tc.C), func(t *testing.T) {
			f := cToF(tc.C)
			assert.Equal(t, tc.WantF, f)
		})
	}
}
