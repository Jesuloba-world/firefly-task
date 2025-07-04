package drift

import (
	"testing"
)

func TestCompareString(t *testing.T) {
	tests := []struct {
		name                string
		actual              string
		expected            string
		config              AttributeConfig
		wantEqual           bool
		descriptionContains string
	}{
		{
			name:                "exact match case sensitive",
			actual:              "test",
			expected:            "test",
			config:              AttributeConfig{ComparisonType: ExactMatch, CaseSensitive: true},
			wantEqual:           true,
			descriptionContains: "case-sensitive exact",
		},
		{
			name:                "exact match case insensitive",
			actual:              "Test",
			expected:            "test",
			config:              AttributeConfig{ComparisonType: ExactMatch, CaseSensitive: false},
			wantEqual:           true,
			descriptionContains: "case-insensitive exact",
		},
		{
			name:                "case sensitive mismatch",
			actual:              "Test",
			expected:            "test",
			config:              AttributeConfig{ComparisonType: ExactMatch, CaseSensitive: true},
			wantEqual:           false,
			descriptionContains: "case-sensitive exact",
		},
		{
			name:                "fuzzy match case insensitive",
			actual:              "TEST",
			expected:            "test",
			config:              AttributeConfig{ComparisonType: FuzzyMatch, CaseSensitive: false},
			wantEqual:           true,
			descriptionContains: "case-insensitive fuzzy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEqual, description := compareString(tt.actual, tt.expected, tt.config)
			if gotEqual != tt.wantEqual {
				t.Errorf("compareString() equal = %v, want %v", gotEqual, tt.wantEqual)
			}
			if !contains(description, tt.descriptionContains) {
				t.Errorf("compareString() description = %v, should contain %v", description, tt.descriptionContains)
			}
		})
	}
}

func TestCompareNumeric(t *testing.T) {
	tolerance := 0.1
	tests := []struct {
		name      string
		actual    float64
		expected  float64
		config    AttributeConfig
		wantEqual bool
	}{
		{
			name:      "exact match",
			actual:    5.0,
			expected:  5.0,
			config:    AttributeConfig{ComparisonType: ExactMatch},
			wantEqual: true,
		},
		{
			name:      "exact mismatch",
			actual:    5.0,
			expected:  5.1,
			config:    AttributeConfig{ComparisonType: ExactMatch},
			wantEqual: false,
		},
		{
			name:      "within tolerance",
			actual:    5.05,
			expected:  5.0,
			config:    AttributeConfig{ComparisonType: NumericTolerance, Tolerance: &tolerance},
			wantEqual: true,
		},
		{
			name:      "outside tolerance",
			actual:    5.2,
			expected:  5.0,
			config:    AttributeConfig{ComparisonType: NumericTolerance, Tolerance: &tolerance},
			wantEqual: false,
		},
		{
			name:      "tolerance with negative difference",
			actual:    4.95,
			expected:  5.0,
			config:    AttributeConfig{ComparisonType: NumericTolerance, Tolerance: &tolerance},
			wantEqual: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEqual, _ := compareNumeric(tt.actual, tt.expected, tt.config)
			if gotEqual != tt.wantEqual {
				t.Errorf("compareNumeric() = %v, want %v", gotEqual, tt.wantEqual)
			}
		})
	}
}

func TestCompareArray(t *testing.T) {
	tests := []struct {
		name      string
		actual    []interface{}
		expected  []interface{}
		config    AttributeConfig
		wantEqual bool
	}{
		{
			name:      "ordered arrays match",
			actual:    []interface{}{"a", "b", "c"},
			expected:  []interface{}{"a", "b", "c"},
			config:    AttributeConfig{ComparisonType: ArrayOrdered},
			wantEqual: true,
		},
		{
			name:      "ordered arrays different order",
			actual:    []interface{}{"a", "c", "b"},
			expected:  []interface{}{"a", "b", "c"},
			config:    AttributeConfig{ComparisonType: ArrayOrdered},
			wantEqual: false,
		},
		{
			name:      "unordered arrays same content",
			actual:    []interface{}{"a", "c", "b"},
			expected:  []interface{}{"a", "b", "c"},
			config:    AttributeConfig{ComparisonType: ArrayUnordered},
			wantEqual: true,
		},
		{
			name:      "different lengths",
			actual:    []interface{}{"a", "b"},
			expected:  []interface{}{"a", "b", "c"},
			config:    AttributeConfig{ComparisonType: ArrayOrdered},
			wantEqual: false,
		},
		{
			name:      "empty arrays",
			actual:    []interface{}{},
			expected:  []interface{}{},
			config:    AttributeConfig{ComparisonType: ArrayOrdered},
			wantEqual: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEqual, _ := compareArray(tt.actual, tt.expected, tt.config)
			if gotEqual != tt.wantEqual {
				t.Errorf("compareArray() = %v, want %v", gotEqual, tt.wantEqual)
			}
		})
	}
}

func TestCompareMap(t *testing.T) {
	tests := []struct {
		name      string
		actual    map[string]interface{}
		expected  map[string]interface{}
		wantEqual bool
	}{
		{
			name: "identical maps",
			actual: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
			wantEqual: true,
		},
		{
			name: "different values",
			actual: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": 43,
			},
			wantEqual: false,
		},
		{
			name: "missing key in actual",
			actual: map[string]interface{}{
				"key1": "value1",
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
			wantEqual: false,
		},
		{
			name: "extra key in actual",
			actual: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
				"key3": "extra",
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
			wantEqual: false,
		},
		{
			name:      "empty maps",
			actual:    map[string]interface{}{},
			expected:  map[string]interface{}{},
			wantEqual: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := AttributeConfig{ComparisonType: MapComparison}
			gotEqual, _ := compareMap(tt.actual, tt.expected, config)
			if gotEqual != tt.wantEqual {
				t.Errorf("compareMap() = %v, want %v", gotEqual, tt.wantEqual)
			}
		})
	}
}

func TestCompareNestedObject(t *testing.T) {
	tests := []struct {
		name      string
		actual    interface{}
		expected  interface{}
		config    AttributeConfig
		wantEqual bool
	}{
		{
			name:      "both nil",
			actual:    nil,
			expected:  nil,
			config:    AttributeConfig{ComparisonType: ExactMatch},
			wantEqual: true,
		},
		{
			name:      "one nil",
			actual:    nil,
			expected:  "not nil",
			config:    AttributeConfig{ComparisonType: ExactMatch},
			wantEqual: false,
		},
		{
			name:      "different types",
			actual:    "string",
			expected:  42,
			config:    AttributeConfig{ComparisonType: ExactMatch},
			wantEqual: false,
		},
		{
			name:      "same strings",
			actual:    "test",
			expected:  "test",
			config:    AttributeConfig{ComparisonType: ExactMatch, CaseSensitive: true},
			wantEqual: true,
		},
		{
			name:      "same integers",
			actual:    42,
			expected:  42,
			config:    AttributeConfig{ComparisonType: ExactMatch},
			wantEqual: true,
		},
		{
			name:      "same booleans",
			actual:    true,
			expected:  true,
			config:    AttributeConfig{ComparisonType: ExactMatch},
			wantEqual: true,
		},
		{
			name:      "different booleans",
			actual:    true,
			expected:  false,
			config:    AttributeConfig{ComparisonType: ExactMatch},
			wantEqual: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEqual, _ := compareNestedObject(tt.actual, tt.expected, tt.config)
			if gotEqual != tt.wantEqual {
				t.Errorf("compareNestedObject() = %v, want %v", gotEqual, tt.wantEqual)
			}
		})
	}
}

func TestConvertToFloat64(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		want      float64
		wantError bool
	}{
		{"float64", float64(3.14), 3.14, false},
		{"float32", float32(3.14), 3.140000104904175, false}, // float32 precision
		{"int", int(42), 42.0, false},
		{"int8", int8(42), 42.0, false},
		{"int16", int16(42), 42.0, false},
		{"int32", int32(42), 42.0, false},
		{"int64", int64(42), 42.0, false},
		{"uint", uint(42), 42.0, false},
		{"uint8", uint8(42), 42.0, false},
		{"uint16", uint16(42), 42.0, false},
		{"uint32", uint32(42), 42.0, false},
		{"uint64", uint64(42), 42.0, false},
		{"string", "not a number", 0, true},
		{"bool", true, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToFloat64(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("convertToFloat64() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && got != tt.want {
				t.Errorf("convertToFloat64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertToString(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  string
	}{
		{"nil", nil, "<nil>"},
		{"string", "test", "test"},
		{"int", 42, "42"},
		{"float", 3.14, "3.14"},
		{"bool", true, "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToString(tt.value); got != tt.want {
				t.Errorf("convertToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertToSlice(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		want      []interface{}
		wantError bool
	}{
		{
			name:      "slice of strings",
			value:     []string{"a", "b", "c"},
			want:      []interface{}{"a", "b", "c"},
			wantError: false,
		},
		{
			name:      "slice of ints",
			value:     []int{1, 2, 3},
			want:      []interface{}{1, 2, 3},
			wantError: false,
		},
		{
			name:      "array of strings",
			value:     [3]string{"a", "b", "c"},
			want:      []interface{}{"a", "b", "c"},
			wantError: false,
		},
		{
			name:      "nil",
			value:     nil,
			want:      nil,
			wantError: true,
		},
		{
			name:      "not a slice",
			value:     "string",
			want:      nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToSlice(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("convertToSlice() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && !slicesEqual(got, tt.want) {
				t.Errorf("convertToSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertToMap(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		want      map[string]interface{}
		wantError bool
	}{
		{
			name: "string map",
			value: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			want: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			wantError: false,
		},
		{
			name: "int key map",
			value: map[int]string{
				1: "value1",
				2: "value2",
			},
			want: map[string]interface{}{
				"1": "value1",
				"2": "value2",
			},
			wantError: false,
		},
		{
			name:      "nil",
			value:     nil,
			want:      nil,
			wantError: true,
		},
		{
			name:      "not a map",
			value:     "string",
			want:      nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToMap(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("convertToMap() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && !mapsEqual(got, tt.want) {
				t.Errorf("convertToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareValues(t *testing.T) {
	tests := []struct {
		name      string
		actual    interface{}
		expected  interface{}
		config    AttributeConfig
		wantEqual bool
	}{
		{
			name:      "both nil",
			actual:    nil,
			expected:  nil,
			config:    AttributeConfig{ComparisonType: ExactMatch},
			wantEqual: true,
		},
		{
			name:      "one nil",
			actual:    nil,
			expected:  "test",
			config:    AttributeConfig{ComparisonType: ExactMatch},
			wantEqual: false,
		},
		{
			name:      "same strings",
			actual:    "test",
			expected:  "test",
			config:    AttributeConfig{ComparisonType: ExactMatch, CaseSensitive: true},
			wantEqual: true,
		},
		{
			name:      "different types converted to string",
			actual:    42,
			expected:  "42",
			config:    AttributeConfig{ComparisonType: ExactMatch, CaseSensitive: true},
			wantEqual: true,
		},
		{
			name:      "numeric comparison",
			actual:    42,
			expected:  42,
			config:    AttributeConfig{ComparisonType: ExactMatch},
			wantEqual: true,
		},
		{
			name:      "slice comparison",
			actual:    []string{"a", "b"},
			expected:  []string{"a", "b"},
			config:    AttributeConfig{ComparisonType: ArrayOrdered},
			wantEqual: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEqual, _ := CompareValues(tt.actual, tt.expected, tt.config)
			if gotEqual != tt.wantEqual {
				t.Errorf("CompareValues() = %v, want %v", gotEqual, tt.wantEqual)
			}
		})
	}
}

// Helper functions for tests

func slicesEqual(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func mapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
