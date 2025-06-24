package main

import (
	"encoding/xml"
	"fmt"
	"net"
	"time"
)

// Request structure for IRIS
const requestXML = `<?xml version="1.0" encoding="UTF-8"?>
<iris1:request xmlns:iris1="urn:ietf:params:xml:ns:iris1">
  <iris1:searchSet>
    <iris1:lookupEntity registryType="dchk1" entityClass="domain-name" entityName="ficora.fi"/>
  </iris1:searchSet>
</iris1:request>`

// Basic response parsing
type DomainResponse struct {
	XMLName      xml.Name `xml:"domain"`
	DomainName   string   `xml:"domainName"`
	Status       Status   `xml:"status"`
	RegistryType string   `xml:"registryType,attr"`
	EntityClass  string   `xml:"entityClass,attr"`
	EntityName   string   `xml:"entityName,attr"`
	Authority    string   `xml:"authority,attr"`
}

type Status struct {
	Active   *struct{} `xml:"active"`
	Available *struct{} `xml:"available"`
	Invalid  *struct{} `xml:"invalid"`
}

func main() {
	serverAddr := "das.domain.fi:715"

	// Resolve UDP address
	raddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		panic(fmt.Sprintf("Failed to resolve address: %v", err))
	}

	// Connect to UDP server
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		panic(fmt.Sprintf("Failed to dial UDP: %v", err))
	}
	defer conn.Close()

	// Set read timeout
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// Send request
	_, err = conn.Write([]byte(requestXML))
	if err != nil {
		panic(fmt.Sprintf("Failed to send request: %v", err))
	}

	// Receive response
	buf := make([]byte, 2048)
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		panic(fmt.Sprintf("Failed to read response: %v", err))
	}

	response := buf[:n]
	fmt.Println("Raw response:")
	fmt.Println(string(response))

	// Parse XML response
	var domain DomainResponse
	err = xml.Unmarshal(response, &domain)
	if err != nil {
		fmt.Printf("Failed to parse XML: %v\n", err)
		return
	}

	// Determine status
	status := "unknown"
	switch {
	case domain.Status.Active != nil:
		status = "active"
	case domain.Status.Available != nil:
		status = "available"
	case domain.Status.Invalid != nil:
		status = "invalid"
	}

	// Output parsed info
	fmt.Printf("\nParsed response:\n")
	fmt.Printf("Domain: %s\nStatus: %s\n", domain.DomainName, status)
}

