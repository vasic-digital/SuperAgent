package security

import (
	"context"
	"regexp"
	"strings"
)

// RegexPIIDetector detects PII using regular expressions
type RegexPIIDetector struct {
	patterns map[PIIType]*regexp.Regexp
	enabled  map[PIIType]bool
}

// NewRegexPIIDetector creates a new PII detector
func NewRegexPIIDetector() *RegexPIIDetector {
	detector := &RegexPIIDetector{
		patterns: make(map[PIIType]*regexp.Regexp),
		enabled:  make(map[PIIType]bool),
	}

	// Email pattern
	detector.patterns[PIITypeEmail] = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	detector.enabled[PIITypeEmail] = true

	// Phone patterns (US and international)
	detector.patterns[PIITypePhone] = regexp.MustCompile(`(\+?1?[-.\s]?)?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}`)
	detector.enabled[PIITypePhone] = true

	// SSN pattern
	detector.patterns[PIITypeSSN] = regexp.MustCompile(`\b\d{3}[-.\s]?\d{2}[-.\s]?\d{4}\b`)
	detector.enabled[PIITypeSSN] = true

	// Credit card patterns (Visa, MasterCard, Amex, Discover)
	detector.patterns[PIITypeCreditCard] = regexp.MustCompile(`\b(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|3[47][0-9]{13}|6(?:011|5[0-9]{2})[0-9]{12})\b`)
	detector.enabled[PIITypeCreditCard] = true

	// IP Address pattern
	detector.patterns[PIITypeIPAddress] = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	detector.enabled[PIITypeIPAddress] = true

	// Date of birth patterns
	detector.patterns[PIITypeDateOfBirth] = regexp.MustCompile(`\b(?:0?[1-9]|1[0-2])[/-](?:0?[1-9]|[12]\d|3[01])[/-](?:19|20)\d{2}\b`)
	detector.enabled[PIITypeDateOfBirth] = true

	// Passport number (basic pattern)
	detector.patterns[PIITypePassport] = regexp.MustCompile(`\b[A-Z]{1,2}[0-9]{6,9}\b`)
	detector.enabled[PIITypePassport] = true

	// Driver's license (basic US pattern)
	detector.patterns[PIITypeDriverLicense] = regexp.MustCompile(`\b[A-Z]{1,2}[0-9]{5,8}\b`)
	detector.enabled[PIITypeDriverLicense] = true

	// Bank account numbers (basic pattern)
	detector.patterns[PIITypeBankAccount] = regexp.MustCompile(`\b\d{8,17}\b`)
	detector.enabled[PIITypeBankAccount] = false // Disabled by default due to false positives

	// API keys (common patterns - sk_*, pk_*, api_* with any suffix containing alphanumerics)
	detector.patterns[PIITypeAPIKey] = regexp.MustCompile(`(?i)\b(sk|pk|api)[_-][a-z0-9_-]{8,}\b`)
	detector.enabled[PIITypeAPIKey] = true

	// Passwords in common formats
	detector.patterns[PIITypePassword] = regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[:=]\s*\S+`)
	detector.enabled[PIITypePassword] = true

	return detector
}

// EnableType enables detection for a PII type
func (d *RegexPIIDetector) EnableType(piiType PIIType) {
	d.enabled[piiType] = true
}

// DisableType disables detection for a PII type
func (d *RegexPIIDetector) DisableType(piiType PIIType) {
	d.enabled[piiType] = false
}

// Detect detects PII in text
func (d *RegexPIIDetector) Detect(ctx context.Context, text string) ([]*PIIDetection, error) {
	var detections []*PIIDetection

	for piiType, pattern := range d.patterns {
		if !d.enabled[piiType] {
			continue
		}

		matches := pattern.FindAllStringIndex(text, -1)
		for _, match := range matches {
			value := text[match[0]:match[1]]

			// Calculate confidence based on pattern specificity
			confidence := d.calculateConfidence(piiType, value)

			detection := &PIIDetection{
				Type:       piiType,
				Value:      value,
				Masked:     d.maskValue(piiType, value),
				StartIndex: match[0],
				EndIndex:   match[1],
				Confidence: confidence,
			}
			detections = append(detections, detection)
		}
	}

	return detections, nil
}

// Mask masks detected PII in text
func (d *RegexPIIDetector) Mask(ctx context.Context, text string) (string, []*PIIDetection, error) {
	detections, err := d.Detect(ctx, text)
	if err != nil {
		return text, nil, err
	}

	// Sort detections by start index in reverse order to avoid offset issues
	for i := len(detections) - 1; i >= 0; i-- {
		det := detections[i]
		text = text[:det.StartIndex] + det.Masked + text[det.EndIndex:]
	}

	return text, detections, nil
}

// Redact removes detected PII from text
func (d *RegexPIIDetector) Redact(ctx context.Context, text string) (string, []*PIIDetection, error) {
	detections, err := d.Detect(ctx, text)
	if err != nil {
		return text, nil, err
	}

	// Sort detections by start index in reverse order
	for i := len(detections) - 1; i >= 0; i-- {
		det := detections[i]
		redacted := "[" + string(det.Type) + "_REDACTED]"
		text = text[:det.StartIndex] + redacted + text[det.EndIndex:]
	}

	return text, detections, nil
}

// calculateConfidence calculates detection confidence
func (d *RegexPIIDetector) calculateConfidence(piiType PIIType, value string) float64 {
	switch piiType {
	case PIITypeEmail:
		// Higher confidence for common email domains
		if strings.Contains(value, "@gmail.com") || strings.Contains(value, "@yahoo.com") {
			return 0.95
		}
		return 0.85

	case PIITypePhone:
		// Higher confidence for properly formatted numbers
		if len(value) >= 10 {
			return 0.8
		}
		return 0.6

	case PIITypeSSN:
		// SSN pattern is quite specific
		return 0.9

	case PIITypeCreditCard:
		// Validate using Luhn algorithm
		if d.validateLuhn(value) {
			return 0.95
		}
		return 0.5

	case PIITypeAPIKey:
		// Higher confidence for longer keys
		if len(value) > 30 {
			return 0.9
		}
		return 0.7

	default:
		return 0.7
	}
}

// maskValue creates a masked version of the value
func (d *RegexPIIDetector) maskValue(piiType PIIType, value string) string {
	switch piiType {
	case PIITypeEmail:
		parts := strings.Split(value, "@")
		if len(parts) == 2 {
			maskedLocal := maskString(parts[0], 2)
			return maskedLocal + "@" + parts[1]
		}
		return maskString(value, 2)

	case PIITypePhone:
		// Show last 4 digits
		cleaned := regexp.MustCompile(`\D`).ReplaceAllString(value, "")
		if len(cleaned) >= 4 {
			return "***-***-" + cleaned[len(cleaned)-4:]
		}
		return "***-***-****"

	case PIITypeSSN:
		return "***-**-" + value[len(value)-4:]

	case PIITypeCreditCard:
		// Show last 4 digits
		cleaned := regexp.MustCompile(`\D`).ReplaceAllString(value, "")
		if len(cleaned) >= 4 {
			return "****-****-****-" + cleaned[len(cleaned)-4:]
		}
		return "****-****-****-****"

	case PIITypeAPIKey, PIITypePassword:
		return "********"

	default:
		return maskString(value, len(value)/4)
	}
}

// maskString masks a string keeping first n characters
func maskString(s string, keepFirst int) string {
	if len(s) <= keepFirst {
		return strings.Repeat("*", len(s))
	}
	return s[:keepFirst] + strings.Repeat("*", len(s)-keepFirst)
}

// validateLuhn validates a number using the Luhn algorithm
func (d *RegexPIIDetector) validateLuhn(number string) bool {
	// Remove non-digits
	cleaned := regexp.MustCompile(`\D`).ReplaceAllString(number, "")
	if len(cleaned) < 13 || len(cleaned) > 19 {
		return false
	}

	sum := 0
	alternate := false

	for i := len(cleaned) - 1; i >= 0; i-- {
		digit := int(cleaned[i] - '0')

		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

// PIIGuardrail wraps PII detector as a guardrail
type PIIGuardrail struct {
	detector PIIDetector
	action   GuardrailAction
	piiTypes []PIIType
}

// NewPIIGuardrail creates a PII detection guardrail
func NewPIIGuardrail(detector PIIDetector, action GuardrailAction, piiTypes []PIIType) *PIIGuardrail {
	if piiTypes == nil {
		piiTypes = []PIIType{
			PIITypeEmail, PIITypePhone, PIITypeSSN, PIITypeCreditCard,
			PIITypeAPIKey, PIITypePassword,
		}
	}

	return &PIIGuardrail{
		detector: detector,
		action:   action,
		piiTypes: piiTypes,
	}
}

func (g *PIIGuardrail) Name() string {
	return "pii_detector"
}

func (g *PIIGuardrail) Type() GuardrailType {
	return GuardrailTypePII
}

func (g *PIIGuardrail) Check(ctx context.Context, content string, metadata map[string]interface{}) (*GuardrailResult, error) {
	detections, err := g.detector.Detect(ctx, content)
	if err != nil {
		return nil, err
	}

	// Filter detections by enabled types
	var filtered []*PIIDetection
	for _, det := range detections {
		for _, enabledType := range g.piiTypes {
			if det.Type == enabledType {
				filtered = append(filtered, det)
				break
			}
		}
	}

	if len(filtered) == 0 {
		return &GuardrailResult{
			Triggered: false,
			Guardrail: g.Name(),
		}, nil
	}

	// Calculate overall confidence
	maxConfidence := 0.0
	piiTypesFound := make([]string, 0)
	for _, det := range filtered {
		if det.Confidence > maxConfidence {
			maxConfidence = det.Confidence
		}
		piiTypesFound = append(piiTypesFound, string(det.Type))
	}

	result := &GuardrailResult{
		Triggered:  true,
		Action:     g.action,
		Guardrail:  g.Name(),
		Reason:     "PII detected in content",
		Confidence: maxConfidence,
		Metadata: map[string]interface{}{
			"pii_count":      len(filtered),
			"pii_types":      piiTypesFound,
			"detections":     filtered,
		},
	}

	// If action is modify, mask the content
	if g.action == GuardrailActionModify {
		masked, _, _ := g.detector.Mask(ctx, content)
		result.ModifiedContent = masked
	}

	return result, nil
}
