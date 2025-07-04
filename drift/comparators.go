package drift

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
)

// compareString compares two string values according to the provided configuration
func compareString(actual, expected string, config AttributeConfig) (bool, string) {
	if config.ComparisonType == FuzzyMatch {
		if config.CaseSensitive {
			return actual == expected, fmt.Sprintf("string comparison (case-sensitive fuzzy): '%s' vs '%s'", actual, expected)
		} else {
			return strings.EqualFold(actual, expected), fmt.Sprintf("string comparison (case-insensitive fuzzy): '%s' vs '%s'", actual, expected)
		}
	}

	// Default to exact match
	if config.CaseSensitive {
		return actual == expected, fmt.Sprintf("string comparison (case-sensitive exact): '%s' vs '%s'", actual, expected)
	} else {
		return strings.EqualFold(actual, expected), fmt.Sprintf("string comparison (case-insensitive exact): '%s' vs '%s'", actual, expected)
	}
}

// compareNumeric compares two numeric values with optional tolerance
func compareNumeric(actual, expected float64, config AttributeConfig) (bool, string) {
	if config.ComparisonType == NumericTolerance && config.Tolerance != nil {
		diff := math.Abs(actual - expected)
		tolerance := *config.Tolerance
		isWithinTolerance := diff <= tolerance
		return isWithinTolerance, fmt.Sprintf("numeric comparison with tolerance %.6f: %.6f vs %.6f (diff: %.6f)", tolerance, actual, expected, diff)
	}

	// Default to exact match
	return actual == expected, fmt.Sprintf("numeric comparison (exact): %.6f vs %.6f", actual, expected)
}

// compareArray compares two arrays/slices according to the provided configuration
func compareArray(actual, expected []interface{}, config AttributeConfig) (bool, string) {
	if len(actual) != len(expected) {
		return false, fmt.Sprintf("array length mismatch: %d vs %d", len(actual), len(expected))
	}

	if config.ComparisonType == ArrayUnordered {
		return compareArrayUnordered(actual, expected)
	}

	// Default to ordered comparison
	return compareArrayOrdered(actual, expected)
}

// compareArrayOrdered compares arrays considering element order
func compareArrayOrdered(actual, expected []interface{}) (bool, string) {
	for i := 0; i < len(actual); i++ {
		if !deepEqual(actual[i], expected[i]) {
			return false, fmt.Sprintf("array element mismatch at index %d: %v vs %v", i, actual[i], expected[i])
		}
	}
	return true, "array comparison (ordered): all elements match"
}

// compareArrayUnordered compares arrays ignoring element order
func compareArrayUnordered(actual, expected []interface{}) (bool, string) {
	// Convert to string slices for sorting
	actualStrs := make([]string, len(actual))
	expectedStrs := make([]string, len(expected))

	for i, v := range actual {
		actualStrs[i] = fmt.Sprintf("%v", v)
	}
	for i, v := range expected {
		expectedStrs[i] = fmt.Sprintf("%v", v)
	}

	sort.Strings(actualStrs)
	sort.Strings(expectedStrs)

	for i := 0; i < len(actualStrs); i++ {
		if actualStrs[i] != expectedStrs[i] {
			return false, fmt.Sprintf("array content mismatch (unordered): %v vs %v", actual, expected)
		}
	}

	return true, "array comparison (unordered): all elements match"
}

// compareMap compares two maps key by key
func compareMap(actual, expected map[string]interface{}, config AttributeConfig) (bool, string) {
	if len(actual) != len(expected) {
		return false, fmt.Sprintf("map size mismatch: %d vs %d keys", len(actual), len(expected))
	}

	// Check all keys in expected map
	for key, expectedValue := range expected {
		actualValue, exists := actual[key]
		if !exists {
			return false, fmt.Sprintf("missing key in actual map: '%s'", key)
		}

		if !deepEqual(actualValue, expectedValue) {
			return false, fmt.Sprintf("map value mismatch for key '%s': %v vs %v", key, actualValue, expectedValue)
		}
	}

	// Check for extra keys in actual map
	for key := range actual {
		if _, exists := expected[key]; !exists {
			return false, fmt.Sprintf("extra key in actual map: '%s'", key)
		}
	}

	return true, "map comparison: all key-value pairs match"
}

// compareNestedObject compares nested objects/structures
func compareNestedObject(actual, expected interface{}, config AttributeConfig) (bool, string) {
	// Handle nil cases
	if actual == nil && expected == nil {
		return true, "both values are nil"
	}
	if actual == nil || expected == nil {
		return false, fmt.Sprintf("nil mismatch: %v vs %v", actual, expected)
	}

	// Use reflection for deep comparison
	actualValue := reflect.ValueOf(actual)
	expectedValue := reflect.ValueOf(expected)

	// Check if types match
	if actualValue.Type() != expectedValue.Type() {
		return false, fmt.Sprintf("type mismatch: %T vs %T", actual, expected)
	}

	// Handle different types
	switch actualValue.Kind() {
	case reflect.String:
		return compareString(actualValue.String(), expectedValue.String(), config)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return compareNumeric(float64(actualValue.Int()), float64(expectedValue.Int()), config)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return compareNumeric(float64(actualValue.Uint()), float64(expectedValue.Uint()), config)
	case reflect.Float32, reflect.Float64:
		return compareNumeric(actualValue.Float(), expectedValue.Float(), config)
	case reflect.Bool:
		isEqual := actualValue.Bool() == expectedValue.Bool()
		return isEqual, fmt.Sprintf("boolean comparison: %t vs %t", actualValue.Bool(), expectedValue.Bool())
	case reflect.Slice, reflect.Array:
		actualSlice := make([]interface{}, actualValue.Len())
		expectedSlice := make([]interface{}, expectedValue.Len())
		for i := 0; i < actualValue.Len(); i++ {
			actualSlice[i] = actualValue.Index(i).Interface()
		}
		for i := 0; i < expectedValue.Len(); i++ {
			expectedSlice[i] = expectedValue.Index(i).Interface()
		}
		return compareArray(actualSlice, expectedSlice, config)
	case reflect.Map:
		actualMap := make(map[string]interface{})
		expectedMap := make(map[string]interface{})
		for _, key := range actualValue.MapKeys() {
			actualMap[fmt.Sprintf("%v", key.Interface())] = actualValue.MapIndex(key).Interface()
		}
		for _, key := range expectedValue.MapKeys() {
			expectedMap[fmt.Sprintf("%v", key.Interface())] = expectedValue.MapIndex(key).Interface()
		}
		return compareMap(actualMap, expectedMap, config)
	case reflect.Struct:
		return compareStruct(actualValue, expectedValue, config)
	case reflect.Ptr:
		if actualValue.IsNil() && expectedValue.IsNil() {
			return true, "both pointers are nil"
		}
		if actualValue.IsNil() || expectedValue.IsNil() {
			return false, fmt.Sprintf("pointer nil mismatch: %v vs %v", actual, expected)
		}
		return compareNestedObject(actualValue.Elem().Interface(), expectedValue.Elem().Interface(), config)
	default:
		// Fallback to deep equal
		isEqual := deepEqual(actual, expected)
		return isEqual, fmt.Sprintf("deep comparison (%s): %v vs %v", actualValue.Kind().String(), actual, expected)
	}
}

// compareStruct compares two struct values field by field
func compareStruct(actualValue, expectedValue reflect.Value, config AttributeConfig) (bool, string) {
	structType := actualValue.Type()

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		actualField := actualValue.Field(i)
		expectedField := expectedValue.Field(i)

		// Skip unexported fields
		if !actualField.CanInterface() {
			continue
		}

		// Create a new config for this field
		fieldConfig := AttributeConfig{
			AttributeName:  field.Name,
			ComparisonType: config.ComparisonType,
			Tolerance:      config.Tolerance,
			CaseSensitive:  config.CaseSensitive,
			Required:       config.Required,
		}

		isEqual, description := compareNestedObject(actualField.Interface(), expectedField.Interface(), fieldConfig)
		if !isEqual {
			return false, fmt.Sprintf("struct field '%s' mismatch: %s", field.Name, description)
		}
	}

	return true, "struct comparison: all fields match"
}

// deepEqual performs a deep equality check between two values
func deepEqual(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

// convertToFloat64 attempts to convert an interface{} to float64
func convertToFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

// convertToString attempts to convert an interface{} to string
func convertToString(value interface{}) string {
	if value == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%v", value)
}

// convertToSlice attempts to convert an interface{} to []interface{}
func convertToSlice(value interface{}) ([]interface{}, error) {
	if value == nil {
		return nil, fmt.Errorf("cannot convert nil to slice")
	}

	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return nil, fmt.Errorf("value is not a slice or array: %T", value)
	}

	result := make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		result[i] = v.Index(i).Interface()
	}

	return result, nil
}

// convertToMap attempts to convert an interface{} to map[string]interface{}
func convertToMap(value interface{}) (map[string]interface{}, error) {
	if value == nil {
		return nil, fmt.Errorf("cannot convert nil to map")
	}

	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Map {
		return nil, fmt.Errorf("value is not a map: %T", value)
	}

	result := make(map[string]interface{})
	for _, key := range v.MapKeys() {
		keyStr := fmt.Sprintf("%v", key.Interface())
		result[keyStr] = v.MapIndex(key).Interface()
	}

	return result, nil
}

// CompareValues is a high-level function that compares two values using the appropriate comparator
func CompareValues(actual, expected interface{}, config AttributeConfig) (bool, string) {
	// Handle nil cases first
	if actual == nil && expected == nil {
		return true, "both values are nil"
	}
	if actual == nil || expected == nil {
		return false, fmt.Sprintf("nil mismatch: %v vs %v", actual, expected)
	}

	// Try to determine the best comparison method based on the types
	actualValue := reflect.ValueOf(actual)
	expectedValue := reflect.ValueOf(expected)

	// If types don't match, try to convert them
	if actualValue.Type() != expectedValue.Type() {
		// Try string conversion first
		actualStr := convertToString(actual)
		expectedStr := convertToString(expected)
		return compareString(actualStr, expectedStr, config)
	}

	// Use the appropriate comparator based on the type
	switch actualValue.Kind() {
	case reflect.String:
		return compareString(actualValue.String(), expectedValue.String(), config)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		actualFloat, err1 := convertToFloat64(actual)
		expectedFloat, err2 := convertToFloat64(expected)
		if err1 != nil || err2 != nil {
			return false, fmt.Sprintf("numeric conversion error: %v, %v", err1, err2)
		}
		return compareNumeric(actualFloat, expectedFloat, config)
	case reflect.Slice, reflect.Array:
		actualSlice, err1 := convertToSlice(actual)
		expectedSlice, err2 := convertToSlice(expected)
		if err1 != nil || err2 != nil {
			return false, fmt.Sprintf("slice conversion error: %v, %v", err1, err2)
		}
		return compareArray(actualSlice, expectedSlice, config)
	case reflect.Map:
		actualMap, err1 := convertToMap(actual)
		expectedMap, err2 := convertToMap(expected)
		if err1 != nil || err2 != nil {
			return false, fmt.Sprintf("map conversion error: %v, %v", err1, err2)
		}
		return compareMap(actualMap, expectedMap, config)
	default:
		return compareNestedObject(actual, expected, config)
	}
}
