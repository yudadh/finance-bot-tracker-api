package pagination

import (
	"testing"
)

func TestParamsNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    Params
		expected Params
	}{
		{
			name: "valid params",
			input: Params{
				Page:    2,
				PerPage: 20,
			},
			expected: Params{
				Page:    2,
				PerPage: 20,
			},
		},
		{
			name: "page less than one",
			input: Params{
				Page:    0,
				PerPage: 20,
			},
			expected: Params{
				Page:    1,
				PerPage: 20,
			},
		},
		{
			name: "per page less than minimum",
			input: Params{
				Page:    2,
				PerPage: 5,
			},
			expected: Params{
				Page:    2,
				PerPage: 10,
			},
		},
		{
			name:  "zero values",
			input: Params{},
			expected: Params{
				Page:    1,
				PerPage: 10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.Normalize()

			if result != tt.expected {
				t.Fatalf("expected %+v, got %+v", tt.expected, result)
			}
		})
	}
}

func TestParamsOffset(t *testing.T) {
	tests := []struct {
		name     string
		params   Params
		expected int
	}{
		{
			name:     "first page",
			params:   Params{Page: 1, PerPage: 10},
			expected: 0,
		},
		{
			name:     "second page",
			params:   Params{Page: 2, PerPage: 10},
			expected: 10,
		},
		{
			name:     "third page with twenty per page",
			params:   Params{Page: 3, PerPage: 20},
			expected: 40,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.params.Offset()

			if result != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestNewMeta(t *testing.T) {
	tests := []struct {
		name       string
		params     Params
		total      int64
		totalPages int
	}{
		{
			name: "exact division",
			params: Params{
				Page:    1,
				PerPage: 10,
			},
			total:      100,
			totalPages: 10,
		},
		{
			name: "has remainder",
			params: Params{
				Page:    1,
				PerPage: 10,
			},
			total:      101,
			totalPages: 11,
		},
		{
			name: "less than one page",
			params: Params{
				Page:    1,
				PerPage: 10,
			},
			total:      7,
			totalPages: 1,
		},
		{
			name: "empty result",
			params: Params{
				Page:    1,
				PerPage: 10,
			},
			total:      0,
			totalPages: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := NewMeta(tt.params, tt.total)

			if meta.TotalPages != tt.totalPages {
				t.Fatalf(
					"expected total pages %d, got %d",
					tt.totalPages,
					meta.TotalPages,
				)
			}
		})
	}
}