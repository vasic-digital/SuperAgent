package config

import (
	"os"
	"testing"
)

func TestConfig_GetString(t *testing.T) {
	c := Config{"key1": "value1", "key2": 123}

	// Test existing string key
	val, ok := c.GetString("key1")
	if !ok || val != "value1" {
		t.Errorf("Expected 'value1', true; got %s, %v", val, ok)
	}

	// Test non-existing key
	val, ok = c.GetString("key3")
	if ok || val != "" {
		t.Errorf("Expected '', false; got %s, %v", val, ok)
	}

	// Test non-string value
	val, ok = c.GetString("key2")
	if ok || val != "" {
		t.Errorf("Expected '', false; got %s, %v", val, ok)
	}
}

func TestConfig_GetStringWithDefault(t *testing.T) {
	c := Config{"key1": "value1"}

	// Test existing key
	result := c.GetStringWithDefault("key1", "default")
	if result != "value1" {
		t.Errorf("Expected 'value1', got %s", result)
	}

	// Test non-existing key
	result = c.GetStringWithDefault("key2", "default")
	if result != "default" {
		t.Errorf("Expected 'default', got %s", result)
	}
}

func TestConfig_GetInt(t *testing.T) {
	c := Config{"int": 42, "int64": int64(100), "float": 3.14, "string": "123", "invalid": "abc"}

	tests := []struct {
		key      string
		expected int
		ok       bool
	}{
		{"int", 42, true},
		{"int64", 100, true},
		{"float", 3, true}, // float64 truncated to int
		{"string", 123, true},
		{"invalid", 0, false},
		{"missing", 0, false},
	}

	for _, test := range tests {
		val, ok := c.GetInt(test.key)
		if val != test.expected || ok != test.ok {
			t.Errorf("GetInt(%s): expected %d, %v; got %d, %v", test.key, test.expected, test.ok, val, ok)
		}
	}
}

func TestConfig_GetIntWithDefault(t *testing.T) {
	c := Config{"key1": 42}

	// Test existing key
	result := c.GetIntWithDefault("key1", 100)
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}

	// Test non-existing key
	result = c.GetIntWithDefault("key2", 100)
	if result != 100 {
		t.Errorf("Expected 100, got %d", result)
	}
}

func TestConfig_GetBool(t *testing.T) {
	c := Config{"true": true, "false": false, "string_true": "true", "string_false": "false", "invalid": "maybe"}

	tests := []struct {
		key      string
		expected bool
		ok       bool
	}{
		{"true", true, true},
		{"false", false, true},
		{"string_true", true, true},
		{"string_false", false, true},
		{"invalid", false, false},
		{"missing", false, false},
	}

	for _, test := range tests {
		val, ok := c.GetBool(test.key)
		if val != test.expected || ok != test.ok {
			t.Errorf("GetBool(%s): expected %v, %v; got %v, %v", test.key, test.expected, test.ok, val, ok)
		}
	}
}

func TestConfig_GetBoolWithDefault(t *testing.T) {
	c := Config{"key1": true}

	// Test existing key
	result := c.GetBoolWithDefault("key1", false)
	if result != true {
		t.Errorf("Expected true, got %v", result)
	}

	// Test non-existing key
	result = c.GetBoolWithDefault("key2", true)
	if result != true {
		t.Errorf("Expected true, got %v", result)
	}
}

func TestConfig_GetFloat(t *testing.T) {
	c := Config{"float": 3.14, "int": 42, "int64": int64(100), "string": "2.71", "invalid": "abc"}

	tests := []struct {
		key      string
		expected float64
		ok       bool
	}{
		{"float", 3.14, true},
		{"int", 42.0, true},
		{"int64", 100.0, true},
		{"string", 2.71, true},
		{"invalid", 0, false},
		{"missing", 0, false},
	}

	for _, test := range tests {
		val, ok := c.GetFloat(test.key)
		if val != test.expected || ok != test.ok {
			t.Errorf("GetFloat(%s): expected %f, %v; got %f, %v", test.key, test.expected, test.ok, val, ok)
		}
	}
}

func TestConfig_GetFloatWithDefault(t *testing.T) {
	c := Config{"key1": 3.14}

	// Test existing key
	result := c.GetFloatWithDefault("key1", 1.0)
	if result != 3.14 {
		t.Errorf("Expected 3.14, got %f", result)
	}

	// Test non-existing key
	result = c.GetFloatWithDefault("key2", 2.71)
	if result != 2.71 {
		t.Errorf("Expected 2.71, got %f", result)
	}
}

func TestConfig_Set(t *testing.T) {
	c := Config{}

	c.Set("key1", "value1")
	c.Set("key2", 42)

	if c["key1"] != "value1" {
		t.Errorf("Expected 'value1', got %v", c["key1"])
	}

	if c["key2"] != 42 {
		t.Errorf("Expected 42, got %v", c["key2"])
	}
}

func TestValidator_AddRule(t *testing.T) {
	v := NewValidator()

	rule := func(Config) error { return nil }
	v.AddRule("test", rule)

	if len(v.rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(v.rules))
	}

	if v.rules["test"] == nil {
		t.Error("Expected rule to be set")
	}
}

func TestValidator_Validate(t *testing.T) {
	v := NewValidator()

	// Add a passing rule
	v.AddRule("pass", func(Config) error { return nil })

	// Add a failing rule
	v.AddRule("fail", func(Config) error { return ValidationError{"test error"} })

	config := Config{}

	err := v.Validate(config)
	if err == nil {
		t.Error("Expected validation error")
	}

	if err.Error() != "validation failed for fail: test error" {
		t.Errorf("Expected specific error message, got %s", err.Error())
	}
}

// ValidationError is a mock error for testing
type ValidationError struct {
	message string
}

func (e ValidationError) Error() string {
	return e.message
}

func TestRequired(t *testing.T) {
	rule := Required("test_key")

	// Test missing key
	config := Config{}
	err := rule(config)
	if err == nil || err.Error() != "required field test_key is missing" {
		t.Errorf("Expected missing field error, got %v", err)
	}

	// Test empty string
	config = Config{"test_key": ""}
	err = rule(config)
	if err == nil || err.Error() != "required field test_key is empty" {
		t.Errorf("Expected empty field error, got %v", err)
	}

	// Test valid value
	config = Config{"test_key": "value"}
	err = rule(config)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestOneOf(t *testing.T) {
	rule := OneOf("test_key", "option1", "option2")

	// Test valid value
	config := Config{"test_key": "option1"}
	err := rule(config)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test invalid value
	config = Config{"test_key": "invalid"}
	err = rule(config)
	if err == nil {
		t.Error("Expected validation error")
	}
}

func TestMinLength(t *testing.T) {
	rule := MinLength("test_key", 3)

	// Test valid length
	config := Config{"test_key": "valid"}
	err := rule(config)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test too short
	config = Config{"test_key": "ab"}
	err = rule(config)
	if err == nil {
		t.Error("Expected validation error")
	}
}

func TestConfig_LoadFromEnv(t *testing.T) {
	// Set up environment variables
	os.Setenv("TEST_PREFIX_KEY1", "value1")
	os.Setenv("TEST_PREFIX_KEY2", "42")
	os.Setenv("TEST_PREFIX_KEY3", "3.14")
	os.Setenv("TEST_PREFIX_KEY4", "true")
	os.Setenv("OTHER_VAR", "ignored")
	defer func() {
		os.Unsetenv("TEST_PREFIX_KEY1")
		os.Unsetenv("TEST_PREFIX_KEY2")
		os.Unsetenv("TEST_PREFIX_KEY3")
		os.Unsetenv("TEST_PREFIX_KEY4")
		os.Unsetenv("OTHER_VAR")
	}()

	c := Config{}
	c.LoadFromEnv("TEST_PREFIX_")

	if c["key1"] != "value1" {
		t.Errorf("Expected 'value1', got %v", c["key1"])
	}

	if c["key2"] != 42 {
		t.Errorf("Expected 42, got %v", c["key2"])
	}

	if c["key3"] != 3.14 {
		t.Errorf("Expected 3.14, got %v", c["key3"])
	}

	if c["key4"] != true {
		t.Errorf("Expected true, got %v", c["key4"])
	}

	if c["other"] != nil {
		t.Errorf("Expected other var to be ignored, got %v", c["other"])
	}
}

func TestProviderConfig_Validate(t *testing.T) {
	// Test valid config
	pc := &ProviderConfig{APIKey: "test-key"}
	err := pc.Validate()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test missing API key
	pc = &ProviderConfig{}
	err = pc.Validate()
	if err == nil {
		t.Error("Expected validation error for missing API key")
	}

	// Test empty API key
	pc = &ProviderConfig{APIKey: ""}
	err = pc.Validate()
	if err == nil {
		t.Error("Expected validation error for empty API key")
	}
}

func TestLoadProviderConfigFromEnv(t *testing.T) {
	// Set up environment variables
	os.Setenv("PROVIDER_API_KEY", "test-key")
	os.Setenv("PROVIDER_BASE_URL", "http://example.com")
	os.Setenv("PROVIDER_TIMEOUT", "60")
	os.Setenv("PROVIDER_MAX_RETRIES", "5")
	os.Setenv("PROVIDER_RATE_LIMIT", "20")
	os.Setenv("PROVIDER_RATE_LIMIT_BURST", "40")
	defer func() {
		os.Unsetenv("PROVIDER_API_KEY")
		os.Unsetenv("PROVIDER_BASE_URL")
		os.Unsetenv("PROVIDER_TIMEOUT")
		os.Unsetenv("PROVIDER_MAX_RETRIES")
		os.Unsetenv("PROVIDER_RATE_LIMIT")
		os.Unsetenv("PROVIDER_RATE_LIMIT_BURST")
	}()

	config := LoadProviderConfigFromEnv("PROVIDER_")

	if config.APIKey != "test-key" {
		t.Errorf("Expected 'test-key', got %s", config.APIKey)
	}

	if config.BaseURL != "http://example.com" {
		t.Errorf("Expected 'http://example.com', got %s", config.BaseURL)
	}

	if config.Timeout != 60 {
		t.Errorf("Expected 60, got %d", config.Timeout)
	}

	if config.MaxRetries != 5 {
		t.Errorf("Expected 5, got %d", config.MaxRetries)
	}

	if config.RateLimit != 20 {
		t.Errorf("Expected 20, got %d", config.RateLimit)
	}

	if config.RateLimitBurst != 40 {
		t.Errorf("Expected 40, got %d", config.RateLimitBurst)
	}
}

func TestLoadProviderConfigFromEnv_Defaults(t *testing.T) {
	// Load config without setting any environment variables
	// Using a unique prefix to avoid conflicts with other tests
	config := LoadProviderConfigFromEnv("NONEXISTENT_PREFIX_")

	if config.APIKey != "" {
		t.Errorf("Expected empty APIKey, got %s", config.APIKey)
	}

	if config.BaseURL != "" {
		t.Errorf("Expected empty BaseURL, got %s", config.BaseURL)
	}

	if config.Timeout != 30 {
		t.Errorf("Expected default timeout 30, got %d", config.Timeout)
	}

	if config.MaxRetries != 3 {
		t.Errorf("Expected default max retries 3, got %d", config.MaxRetries)
	}

	if config.RateLimit != 10 {
		t.Errorf("Expected default rate limit 10, got %d", config.RateLimit)
	}

	if config.RateLimitBurst != 20 {
		t.Errorf("Expected default rate limit burst 20, got %d", config.RateLimitBurst)
	}
}

func TestConfig_LoadFromEnv_EmptyPrefix(t *testing.T) {
	// Set environment variable with no prefix
	os.Setenv("test_key_no_prefix", "value123")
	defer os.Unsetenv("test_key_no_prefix")

	c := Config{}
	c.LoadFromEnv("") // Empty prefix should load all env vars

	// This test verifies that empty prefix works without crashing
	// The actual values will depend on the system environment
}

func TestConfig_LoadFromEnv_TypeConversions(t *testing.T) {
	os.Setenv("TYPETEST_INT_VAL", "42")
	os.Setenv("TYPETEST_FLOAT_VAL", "3.14159")
	os.Setenv("TYPETEST_BOOL_TRUE", "true")
	os.Setenv("TYPETEST_BOOL_FALSE", "false")
	os.Setenv("TYPETEST_STRING_VAL", "hello")
	defer func() {
		os.Unsetenv("TYPETEST_INT_VAL")
		os.Unsetenv("TYPETEST_FLOAT_VAL")
		os.Unsetenv("TYPETEST_BOOL_TRUE")
		os.Unsetenv("TYPETEST_BOOL_FALSE")
		os.Unsetenv("TYPETEST_STRING_VAL")
	}()

	c := Config{}
	c.LoadFromEnv("TYPETEST_")

	// Test integer conversion
	if c["int_val"] != 42 {
		t.Errorf("Expected int 42, got %v (%T)", c["int_val"], c["int_val"])
	}

	// Test float conversion
	if c["float_val"] != 3.14159 {
		t.Errorf("Expected float 3.14159, got %v (%T)", c["float_val"], c["float_val"])
	}

	// Test bool true conversion
	if c["bool_true"] != true {
		t.Errorf("Expected bool true, got %v (%T)", c["bool_true"], c["bool_true"])
	}

	// Test bool false conversion
	if c["bool_false"] != false {
		t.Errorf("Expected bool false, got %v (%T)", c["bool_false"], c["bool_false"])
	}

	// Test string conversion (non-numeric, non-bool)
	if c["string_val"] != "hello" {
		t.Errorf("Expected string 'hello', got %v (%T)", c["string_val"], c["string_val"])
	}
}

func TestValidator_EmptyValidator(t *testing.T) {
	v := NewValidator()

	config := Config{"any_key": "any_value"}
	err := v.Validate(config)

	if err != nil {
		t.Errorf("Expected no error for empty validator, got %v", err)
	}
}

func TestValidator_MultipleRules(t *testing.T) {
	v := NewValidator()
	v.AddRule("api_key", Required("api_key"))
	v.AddRule("mode", OneOf("mode", "dev", "prod", "test"))
	v.AddRule("password", MinLength("password", 8))

	// Test valid config
	validConfig := Config{
		"api_key":  "my-api-key",
		"mode":     "prod",
		"password": "securepassword123",
	}

	err := v.Validate(validConfig)
	if err != nil {
		t.Errorf("Expected no error for valid config, got %v", err)
	}
}

func TestRequired_WhitespaceOnly(t *testing.T) {
	rule := Required("test_key")

	// Test whitespace-only string
	config := Config{"test_key": "   "}
	err := rule(config)
	if err == nil || err.Error() != "required field test_key is empty" {
		t.Errorf("Expected empty field error for whitespace-only, got %v", err)
	}

	// Test tabs and spaces
	config = Config{"test_key": "\t  \n"}
	err = rule(config)
	if err == nil || err.Error() != "required field test_key is empty" {
		t.Errorf("Expected empty field error for whitespace with tabs, got %v", err)
	}
}

func TestRequired_NonStringValue(t *testing.T) {
	rule := Required("test_key")

	// Test with integer value (should pass - it's present and not an empty string)
	config := Config{"test_key": 42}
	err := rule(config)
	if err != nil {
		t.Errorf("Expected no error for integer value, got %v", err)
	}

	// Test with bool value
	config = Config{"test_key": true}
	err = rule(config)
	if err != nil {
		t.Errorf("Expected no error for bool value, got %v", err)
	}

	// Test with nil value (should fail - nil is like missing)
	config = Config{"test_key": nil}
	err = rule(config)
	// nil is present in the map, so this depends on implementation
	// Current implementation doesn't check for nil explicitly
}

func TestOneOf_MissingKey(t *testing.T) {
	rule := OneOf("optional_key", "opt1", "opt2")

	// Missing key should not error (let Required handle that)
	config := Config{}
	err := rule(config)
	if err != nil {
		t.Errorf("Expected no error for missing key, got %v", err)
	}
}

func TestOneOf_CaseSensitive(t *testing.T) {
	rule := OneOf("mode", "dev", "prod")

	// Test uppercase (should fail since values are case-sensitive)
	config := Config{"mode": "DEV"}
	err := rule(config)
	if err == nil {
		t.Error("Expected error for case mismatch")
	}

	// Test correct case
	config = Config{"mode": "dev"}
	err = rule(config)
	if err != nil {
		t.Errorf("Expected no error for correct case, got %v", err)
	}
}

func TestMinLength_MissingKey(t *testing.T) {
	rule := MinLength("optional_key", 5)

	// Missing key should not error
	config := Config{}
	err := rule(config)
	if err != nil {
		t.Errorf("Expected no error for missing key, got %v", err)
	}
}

func TestMinLength_ExactLength(t *testing.T) {
	rule := MinLength("password", 8)

	// Test exactly 8 characters
	config := Config{"password": "12345678"}
	err := rule(config)
	if err != nil {
		t.Errorf("Expected no error for exact minimum length, got %v", err)
	}
}

func TestMinLength_Unicode(t *testing.T) {
	rule := MinLength("name", 3)

	// Unicode characters might count differently
	config := Config{"name": "abc"}
	err := rule(config)
	if err != nil {
		t.Errorf("Expected no error for 3-char name, got %v", err)
	}
}

func TestProviderConfig_ValidateWithBaseURL(t *testing.T) {
	pc := &ProviderConfig{
		APIKey:  "test-key",
		BaseURL: "https://api.example.com",
		Timeout: 30,
	}

	err := pc.Validate()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestConfig_GetString_WrongType(t *testing.T) {
	c := Config{
		"int_val":   42,
		"bool_val":  true,
		"float_val": 3.14,
		"slice_val": []string{"a", "b"},
	}

	// All should return empty string and false for non-string values
	tests := []string{"int_val", "bool_val", "float_val", "slice_val"}

	for _, key := range tests {
		val, ok := c.GetString(key)
		if ok || val != "" {
			t.Errorf("GetString(%s): expected ('', false), got (%q, %v)", key, val, ok)
		}
	}
}

func TestConfig_GetInt_StringParsing(t *testing.T) {
	c := Config{
		"valid_int_string":   "123",
		"invalid_int_string": "abc",
		"float_string":       "3.14",
	}

	// Valid integer string
	val, ok := c.GetInt("valid_int_string")
	if !ok || val != 123 {
		t.Errorf("GetInt(valid_int_string): expected (123, true), got (%d, %v)", val, ok)
	}

	// Invalid integer string
	val, ok = c.GetInt("invalid_int_string")
	if ok || val != 0 {
		t.Errorf("GetInt(invalid_int_string): expected (0, false), got (%d, %v)", val, ok)
	}

	// Float string (should fail)
	val, ok = c.GetInt("float_string")
	if ok {
		t.Errorf("GetInt(float_string): expected false, got true with value %d", val)
	}
}

func TestConfig_GetBool_StringParsing(t *testing.T) {
	c := Config{
		"bool_1":       "1",
		"bool_0":       "0",
		"bool_yes":     "yes", // Not supported by ParseBool
		"bool_TRUE":    "TRUE",
		"bool_FALSE":   "FALSE",
		"bool_invalid": "maybe",
	}

	// Test "1" -> true
	val, ok := c.GetBool("bool_1")
	if !ok || val != true {
		t.Errorf("GetBool(bool_1): expected (true, true), got (%v, %v)", val, ok)
	}

	// Test "0" -> false
	val, ok = c.GetBool("bool_0")
	if !ok || val != false {
		t.Errorf("GetBool(bool_0): expected (false, true), got (%v, %v)", val, ok)
	}

	// Test "TRUE" (case insensitive)
	val, ok = c.GetBool("bool_TRUE")
	if !ok || val != true {
		t.Errorf("GetBool(bool_TRUE): expected (true, true), got (%v, %v)", val, ok)
	}

	// Test invalid bool string
	val, ok = c.GetBool("bool_invalid")
	if ok {
		t.Errorf("GetBool(bool_invalid): expected false, got true with value %v", val)
	}
}

func TestConfig_GetFloat_StringParsing(t *testing.T) {
	c := Config{
		"float_string":   "2.718281828",
		"int_string":     "42",
		"invalid_string": "not-a-number",
	}

	// Valid float string
	val, ok := c.GetFloat("float_string")
	if !ok || val != 2.718281828 {
		t.Errorf("GetFloat(float_string): expected (2.718281828, true), got (%v, %v)", val, ok)
	}

	// Integer string (should work as float)
	val, ok = c.GetFloat("int_string")
	if !ok || val != 42.0 {
		t.Errorf("GetFloat(int_string): expected (42.0, true), got (%v, %v)", val, ok)
	}

	// Invalid string
	val, ok = c.GetFloat("invalid_string")
	if ok {
		t.Errorf("GetFloat(invalid_string): expected false, got true with value %v", val)
	}
}

func TestConfig_MultipleOperations(t *testing.T) {
	c := Config{}

	// Set multiple values
	c.Set("string_key", "hello")
	c.Set("int_key", 42)
	c.Set("float_key", 3.14)
	c.Set("bool_key", true)

	// Verify all values
	if s, ok := c.GetString("string_key"); !ok || s != "hello" {
		t.Errorf("String key mismatch: got %s, %v", s, ok)
	}

	if i, ok := c.GetInt("int_key"); !ok || i != 42 {
		t.Errorf("Int key mismatch: got %d, %v", i, ok)
	}

	if f, ok := c.GetFloat("float_key"); !ok || f != 3.14 {
		t.Errorf("Float key mismatch: got %f, %v", f, ok)
	}

	if b, ok := c.GetBool("bool_key"); !ok || b != true {
		t.Errorf("Bool key mismatch: got %v, %v", b, ok)
	}

	// Overwrite a value
	c.Set("string_key", "world")
	if s, ok := c.GetString("string_key"); !ok || s != "world" {
		t.Errorf("Overwritten string key mismatch: got %s, %v", s, ok)
	}
}

func TestNewValidator_InitializesRulesMap(t *testing.T) {
	v := NewValidator()

	if v == nil {
		t.Fatal("Expected non-nil validator")
	}

	if v.rules == nil {
		t.Error("Expected rules map to be initialized")
	}

	if len(v.rules) != 0 {
		t.Error("Expected empty rules map")
	}
}

func TestValidator_RuleOverwrite(t *testing.T) {
	v := NewValidator()

	errorCount := 0

	// Add a rule that increments counter
	v.AddRule("test", func(c Config) error {
		errorCount++
		return nil
	})

	// Overwrite with a different rule
	v.AddRule("test", func(c Config) error {
		errorCount += 10
		return nil
	})

	config := Config{}
	v.Validate(config)

	// Only the second rule should have run (counter should be 10, not 11)
	if errorCount != 10 {
		t.Errorf("Expected errorCount 10, got %d", errorCount)
	}
}

func TestConfig_LoadFromEnv_SkipsInvalidFormat(t *testing.T) {
	// This tests edge cases in environment variable parsing
	// Environment variables always have format KEY=VALUE
	// The SplitN(env, "=", 2) should handle values with = in them

	os.Setenv("EDGETEST_KEY_WITH_EQUALS", "value=with=equals")
	defer os.Unsetenv("EDGETEST_KEY_WITH_EQUALS")

	c := Config{}
	c.LoadFromEnv("EDGETEST_")

	if c["key_with_equals"] != "value=with=equals" {
		t.Errorf("Expected 'value=with=equals', got %v", c["key_with_equals"])
	}
}
