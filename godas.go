package godas

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

// Status represents possible domain statuses
type Status string

const (
	StatusActive    Status = "active"
	StatusAvailable Status = "available"
	StatusInvalid   Status = "invalid"
	StatusError     Status = "error"
)

// Config contains the host and port settings for the server.
type Config struct {
	ServerAddr string // e.g., "das.domain.fi:715"
	Timeout    time.Duration
}

// Response represents parsed response
type Response struct {
	DomainName string
	Status     Status
	RawXML     string
}

// domainResponse is internal XML mapping structure
type domainResponse struct {
	XMLName      xml.Name `xml:"domain"`
	DomainName   string   `xml:"domainName"`
	Status       struct {
		Active    *struct{} `xml:"active"`
		Available *struct{} `xml:"available"`
		Invalid   *struct{} `xml:"invalid"`
	} `xml:"status"`
}

// Lookup performs a domain check for the given domain name
func Lookup(cfg Config, domain string) (*Response, error) {
	xmlRequest := buildRequest(domain)

	// Resolve address
	raddr, err := net.ResolveUDPAddr("udp", cfg.ServerAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial UDP: %w", err)
	}
	defer conn.Close()

	// Set timeout
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	conn.SetReadDeadline(time.Now().Add(timeout))

	// Send request
	_, err = conn.Write([]byte(xmlRequest))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Read response
	buf := make([]byte, 2048)
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	raw := string(buf[:n])

	// Parse response
	var parsed domainResponse
	if err := xml.Unmarshal(buf[:n], &parsed); err != nil {
		return &Response{
			DomainName: domain,
			Status:     StatusError,
			RawXML:     raw,
		}, nil
	}

	// Determine status
	var status Status
	switch {
	case parsed.Status.Active != nil:
		status = StatusActive
	case parsed.Status.Available != nil:
		status = StatusAvailable
	case parsed.Status.Invalid != nil:
		status = StatusInvalid
	default:
		status = StatusError
	}

	return &Response{
		DomainName: parsed.DomainName,
		Status:     status,
		RawXML:     raw,
	}, nil
}

func buildRequest(domain string) string {
	domain = strings.TrimSpace(domain)
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<iris1:request xmlns:iris1="urn:ietf:params:xml:ns:iris1">
  <iris1:searchSet>
    <iris1:lookupEntity registryType="dchk1" entityClass="domain-name" entityName="%s"/>
  </iris1:searchSet>
</iris1:request>`, domain)
}


