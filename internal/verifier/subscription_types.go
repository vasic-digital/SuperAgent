// Package verifier provides integration with LLMsVerifier for model verification,
// scoring, and health monitoring capabilities.
package verifier

import (
	"net/http"
	"time"
)

// SubscriptionType represents the billing model for a provider
type SubscriptionType string

const (
	// SubTypeFree is for genuinely free forever providers (Ollama, Groq free tier)
	SubTypeFree SubscriptionType = "free"

	// SubTypeFreeCredits is for time-limited credit providers (OpenAI $5/3mo, DeepSeek 5M/30d)
	SubTypeFreeCredits SubscriptionType = "free_credits"

	// SubTypeFreeTier is for free with rate restrictions (Mistral 2RPM, Cohere trial)
	SubTypeFreeTier SubscriptionType = "free_tier"

	// SubTypePayAsYouGo is for per-token billing
	SubTypePayAsYouGo SubscriptionType = "pay_as_you_go"

	// SubTypeMonthly is for monthly subscription (HuggingFace PRO, ZAI GLM Coding)
	SubTypeMonthly SubscriptionType = "monthly"

	// SubTypeEnterprise is for custom contract providers
	SubTypeEnterprise SubscriptionType = "enterprise"
)

// AuthMechanism describes how a specific provider authenticates API requests
type AuthMechanism struct {
	// HeaderName is the HTTP header used for authentication (e.g., "Authorization", "x-api-key")
	HeaderName string `json:"header_name"`

	// HeaderPrefix is the value prefix (e.g., "Bearer " with trailing space)
	HeaderPrefix string `json:"header_prefix,omitempty"`

	// QueryParam is for providers that accept auth via query parameter (e.g., "key" for Gemini)
	QueryParam string `json:"query_param,omitempty"`

	// ExtraHeaders are additional required headers (e.g., {"anthropic-version": "2023-06-01"})
	ExtraHeaders map[string]string `json:"extra_headers,omitempty"`

	// NoAuth indicates providers that require no authentication (Ollama, Zen anonymous)
	NoAuth bool `json:"no_auth,omitempty"`

	// DeviceIDHeader is the header name for device-based auth (e.g., "X-Device-ID" for Zen)
	DeviceIDHeader string `json:"device_id_header,omitempty"`
}

// SubscriptionInfo contains detected/configured subscription details for a provider
type SubscriptionInfo struct {
	// Type is the detected subscription type
	Type SubscriptionType `json:"type"`

	// AvailableTiers lists all subscription tiers this provider offers
	AvailableTiers []SubscriptionType `json:"available_tiers"`

	// DetectedAt is when the subscription was detected
	DetectedAt time.Time `json:"detected_at,omitempty"`

	// DetectionSource indicates how the subscription was detected: "api", "rate_limit_headers", "static"
	DetectionSource string `json:"detection_source"`

	// CreditsRemaining is the remaining credit balance (nil if unknown)
	CreditsRemaining *float64 `json:"credits_remaining,omitempty"`

	// CreditsCurrency is the currency of credits (e.g., "USD")
	CreditsCurrency string `json:"credits_currency,omitempty"`

	// CreditsExpiresAt is when credits expire (nil if unknown or non-expiring)
	CreditsExpiresAt *time.Time `json:"credits_expires_at,omitempty"`

	// RateLimits contains parsed rate limit information from response headers
	RateLimits *RateLimitInfo `json:"rate_limits,omitempty"`

	// PlanName is the provider-reported plan name (e.g., "free", "pro", "enterprise")
	PlanName string `json:"plan_name,omitempty"`

	// Restrictions lists known restrictions for this subscription
	Restrictions []string `json:"restrictions,omitempty"`
}

// RateLimitInfo contains rate limit data parsed from HTTP response headers
type RateLimitInfo struct {
	// Request rate limits
	RequestsLimit     int       `json:"requests_limit,omitempty"`
	RequestsRemaining int       `json:"requests_remaining,omitempty"`
	RequestsReset     time.Time `json:"requests_reset,omitempty"`

	// Token rate limits
	TokensLimit     int       `json:"tokens_limit,omitempty"`
	TokensRemaining int       `json:"tokens_remaining,omitempty"`
	TokensReset     time.Time `json:"tokens_reset,omitempty"`

	// Daily rate limits (some providers use daily quotas)
	DailyLimit     int `json:"daily_limit,omitempty"`
	DailyRemaining int `json:"daily_remaining,omitempty"`

	// UpdatedAt is when the rate limit info was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// ProviderAccessConfig holds the full access configuration for a provider
type ProviderAccessConfig struct {
	// ProviderType is the provider identifier (e.g., "openai", "anthropic")
	ProviderType string `json:"provider_type"`

	// AuthMechanisms lists all supported authentication mechanisms
	AuthMechanisms []AuthMechanism `json:"auth_mechanisms"`

	// PrimaryAuth is the primary/recommended authentication mechanism
	PrimaryAuth AuthMechanism `json:"primary_auth"`

	// SubscriptionCheckURL is the API endpoint for checking subscription status
	SubscriptionCheckURL string `json:"subscription_check_url,omitempty"`

	// BillingCheckURL is the API endpoint for checking billing/credit status
	BillingCheckURL string `json:"billing_check_url,omitempty"`

	// ModelsURL is the API endpoint for listing available models
	ModelsURL string `json:"models_url,omitempty"`

	// RateLimitHeaders maps provider-specific rate limit header names
	RateLimitHeaders *RateLimitHeaderNames `json:"rate_limit_headers,omitempty"`

	// DefaultSubscription is the default/assumed subscription type
	DefaultSubscription SubscriptionType `json:"default_subscription"`

	// AvailableTiers lists all subscription tiers this provider offers
	AvailableTiers []SubscriptionType `json:"available_tiers"`
}

// RateLimitHeaderNames maps the provider-specific header names for rate limit information
type RateLimitHeaderNames struct {
	RequestsLimit     string `json:"requests_limit,omitempty"`
	RequestsRemaining string `json:"requests_remaining,omitempty"`
	RequestsReset     string `json:"requests_reset,omitempty"`
	TokensLimit       string `json:"tokens_limit,omitempty"`
	TokensRemaining   string `json:"tokens_remaining,omitempty"`
	TokensReset       string `json:"tokens_reset,omitempty"`
	DailyLimit        string `json:"daily_limit,omitempty"`
	DailyRemaining    string `json:"daily_remaining,omitempty"`
}

// HasRateLimitHeaders returns true if the provider has known rate limit header mappings
func (pac *ProviderAccessConfig) HasRateLimitHeaders() bool {
	return pac.RateLimitHeaders != nil && pac.RateLimitHeaders.RequestsLimit != ""
}

// HasSubscriptionCheckAPI returns true if the provider exposes a subscription check API
func (pac *ProviderAccessConfig) HasSubscriptionCheckAPI() bool {
	return pac.SubscriptionCheckURL != ""
}

// ApplyAuth applies the primary auth mechanism to an HTTP request
func (am *AuthMechanism) ApplyAuth(req *http.Request, apiKey string) {
	if am.NoAuth {
		if am.DeviceIDHeader != "" {
			req.Header.Set(am.DeviceIDHeader, apiKey)
		}
		return
	}

	if am.HeaderName != "" && apiKey != "" {
		req.Header.Set(am.HeaderName, am.HeaderPrefix+apiKey)
	}

	if am.QueryParam != "" && apiKey != "" {
		q := req.URL.Query()
		q.Set(am.QueryParam, apiKey)
		req.URL.RawQuery = q.Encode()
	}

	for k, v := range am.ExtraHeaders {
		req.Header.Set(k, v)
	}
}

// IsValid returns true if the subscription type is a known valid type
func (st SubscriptionType) IsValid() bool {
	switch st {
	case SubTypeFree, SubTypeFreeCredits, SubTypeFreeTier,
		SubTypePayAsYouGo, SubTypeMonthly, SubTypeEnterprise:
		return true
	default:
		return false
	}
}

// IsFree returns true if the subscription type is any kind of free tier
func (st SubscriptionType) IsFree() bool {
	return st == SubTypeFree || st == SubTypeFreeCredits || st == SubTypeFreeTier
}

// IsPaid returns true if the subscription type is a paid tier
func (st SubscriptionType) IsPaid() bool {
	return st == SubTypePayAsYouGo || st == SubTypeMonthly || st == SubTypeEnterprise
}
