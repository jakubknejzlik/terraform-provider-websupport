package websupport

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"strconv"
)

func resourceWebsupportRecordImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	provider := meta.(*Client)

	values := strings.Split(d.Id(), "/")

	if len(values) != 2 {
		return nil, fmt.Errorf("invalid id provided, expected format: {userId}/{zone}")
	}

	userId, err := strconv.Atoi(values[0])
	if err != nil {
		return nil, fmt.Errorf("Error converting userId: %s", err)
	}

	recordZone := values[1]

	listAllDNSRecordsResp, err := provider.client.DNS.ListAllDNSRecords(userId, recordZone)
	if err != nil {
		return nil, fmt.Errorf("Error importing Websupport DNS Records. Reason: %s", err)
	}

	lenOfItems := len(listAllDNSRecordsResp.Items)
	results := make([]*schema.ResourceData, lenOfItems, lenOfItems)
	if lenOfItems > 0 {
		for index, dnsRecord := range listAllDNSRecordsResp.Items {
			d.SetId(strconv.Itoa(dnsRecord.Id))
			d.Set("zone", dnsRecord.Zone.Name)
			d.Set("zone_id", strconv.Itoa(dnsRecord.Zone.Id))
			d.Set("name", dnsRecord.Name)
			d.Set("value", dnsRecord.Content)
			d.Set("type", dnsRecord.Type)
			d.Set("fqdn", getFQDNByRecord(&dnsRecord))
			d.Set("ttl", strconv.Itoa(dnsRecord.TTL))
			// TODO other attributes

			results[index] = d
		}
	}

	return results, nil
}