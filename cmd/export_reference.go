package cmd

import (
	"fmt"
	"net/url"
	"strings"
)

const defaultReferenceDomainURL = "https://feishu.cn"

func buildWikiReferenceURL(rawInput, nodeToken, domainURL string) string {
	if normalizedURL, ok := normalizeReferenceInputURL(rawInput); ok {
		return normalizedURL
	}
	base := normalizeDomainURL(domainURL)
	return fmt.Sprintf("%s/wiki/%s", base, nodeToken)
}

func buildDocReferenceURL(rawInput, documentID, domainURL string) string {
	if normalizedURL, ok := normalizeReferenceInputURL(rawInput); ok {
		return normalizedURL
	}
	base := normalizeDomainURL(domainURL)
	return fmt.Sprintf("%s/docx/%s", base, documentID)
}

func prependReferenceQuote(markdown, referenceURL string) string {
	trimmedURL := strings.TrimSpace(referenceURL)
	if trimmedURL == "" {
		return markdown
	}
	trimmedMarkdown := strings.TrimLeft(markdown, "\n")
	if trimmedMarkdown == "" {
		return fmt.Sprintf("> %s\n", trimmedURL)
	}
	return fmt.Sprintf("> %s\n\n%s", trimmedURL, trimmedMarkdown)
}

func normalizeReferenceInputURL(rawInput string) (string, bool) {
	input := strings.TrimSpace(rawInput)
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		return "", false
	}
	parsed, err := url.Parse(input)
	if err != nil || parsed.Host == "" {
		return "", false
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return strings.TrimRight(parsed.String(), "/"), true
}

func normalizeDomainURL(domainURL string) string {
	trimmed := strings.TrimSpace(domainURL)
	if trimmed == "" {
		return defaultReferenceDomainURL
	}
	if !strings.HasPrefix(trimmed, "http://") && !strings.HasPrefix(trimmed, "https://") {
		trimmed = "https://" + trimmed
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Hostname() == "" {
		return defaultReferenceDomainURL
	}
	scheme := parsed.Scheme
	if scheme == "" {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, parsed.Host)
}
