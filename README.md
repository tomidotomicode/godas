# godas

Go implementation of Traficom Domain Availability Service (DAS) 

```
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/tomidotomicode/godas"
)

func main() {
	cfg := godas.Config{
		ServerAddr: "das.domain.fi:715",
		Timeout:    5 * time.Second,
	}

	resp, err := godas.Lookup(cfg, "ficora.fi")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	fmt.Printf("Domain: %s\nStatus: %s\n", resp.DomainName, resp.Status)
	fmt.Println("Raw response:")
	fmt.Println(resp.RawXML)
}
```
