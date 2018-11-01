package websupport

import (
	"fmt"

	"github.com/radoslavoleksak/go-websupport/websupport"
)

func getFQDNByRecord(record *websupport.DNSRecord) string {
	if record.Name == "@" { // TODO zavinac presun do konstanty
		return record.Zone.Name
	} else {
		return fmt.Sprintf("%s.%s", record.Name, record.Zone.Name)
	}
}