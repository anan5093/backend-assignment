package validation

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
	MaxURLCount  = 20
	MaxURLLength = 2048
)

func RequiredString(value, field string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New(field + " is required")
	}
	return nil
}

func ValidateURLs(urls []string, fieldName string) error {
	if len(urls) > MaxURLCount {
		return errors.New(fieldName + " cannot contain more than 20 URLs")
	}

	for _, raw := range urls {
		if len(raw) > MaxURLLength {
			return errors.New(fieldName + " exceeds maximum URL length")
		}

		parsed, err := url.Parse(raw)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return errors.New("invalid " + singularURLField(fieldName))
		}

		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return errors.New("invalid " + singularURLField(fieldName))
		}
	}

	return nil
}

func ValidatePagination(limitRaw, offsetRaw string) (int, int, error) {
	limit := DefaultLimit
	offset := 0

	if limitRaw != "" {
		parsed, err := strconv.Atoi(limitRaw)
		if err != nil || parsed < 1 {
			return 0, 0, errors.New("limit must be a positive integer")
		}
		if parsed > MaxLimit {
			parsed = MaxLimit
		}
		limit = parsed
	}

	if offsetRaw != "" {
		parsed, err := strconv.Atoi(offsetRaw)
		if err != nil || parsed < 0 {
			return 0, 0, errors.New("offset must be a non-negative integer")
		}
		offset = parsed
	}

	return limit, offset, nil
}

func singularURLField(fieldName string) string {
	switch fieldName {
	case "image_urls":
		return "image url"
	case "video_urls":
		return "video url"
	default:
		return "url"
	}
}
