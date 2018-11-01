package websupport

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/radoslavoleksak/go-websupport/websupport"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceWebsupportRecord() *schema.Resource {
	return &schema.Resource{
		Create: resourceWebsupportRecordCreate,
		Read:   resourceWebsupportRecordRead,
		Update: resourceWebsupportRecordUpdate,
		Delete: resourceWebsupportRecordDelete,
		Importer: &schema.ResourceImporter{
			State: resourceWebsupportRecordImport,
		},

		Schema: map[string]*schema.Schema{
			"zone": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"zone_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true, // TODO optional if root - edit comment
				ForceNew: true,
				DiffSuppressFunc: func(k, oldV, newV string, d *schema.ResourceData) bool {
					// Records for zone (not subdomain)
					zone := d.Get("zone").(string)
					if oldV == zone && newV == "@" { // TODO uistit sa ci nemam posielat "@" podla https://rest.websupport.sk/docs/v1.zone#post-record
						return true
					}

					return oldV == newV
				},
			},

			"fqdn": &schema.Schema{ // TODO resp. hostname
				Type:     schema.TypeString,
				Computed: true,
			},

			"type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"value": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				/* TODO uncomment ak treba
				DiffSuppressFunc: func(k, oldV, newV string, d *schema.ResourceData) bool {
					recordType := d.Get("type").(string)
					if recordType == "CNAME" || recordType == "NS" || recordType == "MX" {
						// We expect FQDN here, which may or may not have a trailing dot
						if !strings.HasSuffix(oldV, ".") {
							oldV += "."
						}
						if !strings.HasSuffix(newV, ".") {
							newV += "."
						}
					}

					return oldV == newV
				},*/
			},

			"ttl": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true, // TODO bud pouzijem default tu napr  Default: "3600", alebo sa pouzije default 600 z tadialto https://rest.websupport.sk/docs/v1.zone#post-record
			},

			/* TODO priority and other attributes
			"priority": &schema.Schema{ // TODO co to priority vlastne je (v doc sa to spomina prio pri MX a naviac (prio, weight, port) pri SRV zaznamoch) https://rest.websupport.sk/docs/v1.zone#post-record sa pise
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},*/
		},
	}
}

func resourceWebsupportRecordCreate(d *schema.ResourceData, meta interface{}) error {
	provider := meta.(*Client)

	// Create the new record
	newRecord := &websupport.DNSRecord{
		Name:    d.Get("name").(string),
		Type: 	 d.Get("type").(string),
		Content: d.Get("value").(string),
	}
	if attr, ok := d.GetOk("ttl"); ok {
		newRecord.TTL, _ = strconv.Atoi(attr.(string))
	}

	/* TODO priority and other attributes
	if attr, ok := d.GetOk("priority"); ok {
		newRecord.Priority, _ = strconv.Atoi(attr.(string))
	}
	*/

	log.Printf("[DEBUG] Websupport Record create configuration: %#v", newRecord)

	// check if user has access to zone
	listAllDNSZonesResp, err := provider.client.DNS.ListAllDNSZones(provider.loggedUser.Id)

	findZoneString := d.Get("zone").(string)

	var selectedDNSZone *websupport.DNSZone // TODO ma to byt ako pointer? stale som si neisty
	for _, dnsZone := range listAllDNSZonesResp.Items {
		if findZoneString == dnsZone.Name {
			selectedDNSZone = &dnsZone // TODO dobre je ten zavinac? :P
		}
		//dnsZone.Id, dnsZone.Name
	}

	if selectedDNSZone == nil {
		return fmt.Errorf("Logged User: %s (userId: %d) cannot access DNS zone: %s", provider.loggedUser.Login, provider.loggedUser.Id, findZoneString)
	}

	// create record request
	createDNSResp, err := provider.client.DNS.CreateDNSRecord(provider.loggedUser.Id, selectedDNSZone.Name, newRecord)
	if err != nil {
		return fmt.Errorf("Failed to create Websupport DNS Record: %s", err)
	}

	// response validation errors
	if createDNSResp.Status == "error" {
		return fmt.Errorf("Failed to create Websupport DNS Record. Validation errors: %#v", createDNSResp.Errors)
		// TODO dobre som spravil printf toString? - createDNSResp.Errors je typu: map[string]interface{}
	}

	d.SetId(strconv.Itoa(createDNSResp.Item.Id))
	log.Printf("[INFO] Websupport DNS Record Id: %s", d.Id())

	return resourceWebsupportRecordRead(d, meta)
}

func resourceWebsupportRecordRead(d *schema.ResourceData, meta interface{}) error {
	provider := meta.(*Client)

	recordID, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Error converting record ID: %s", err)
	}

	// TODO check na d.Get("zone").(string) ako je v resourceWebsupportRecordCreate()
	resp, err := provider.client.DNS.GetDNSRecordDetail(provider.loggedUser.Id, d.Get("zone").(string), recordID)
	// TODO uistit sa ci v pripade 404 bude err naplneny
	if err != nil {
		if err != nil && strings.Contains(err.Error(), "404") {
			log.Printf("Websupport DNS Record Not Found - Refreshing from State")
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Couldn't find Websupport DNS Record: %s", err)
	}

	record := resp
	d.Set("zone_id", strconv.Itoa(record.Zone.Id))
	d.Set("name", record.Name)
	d.Set("type", record.Type)
	d.Set("value", record.Content)
	d.Set("ttl", strconv.Itoa(record.TTL))

	// TODO namapuj este dalsie povinne atributy
	// d.Set("priority", strconv.Itoa(record.Prio))
	d.Set("fqdn", getFQDNByRecord(&record))

	return nil
}

func resourceWebsupportRecordUpdate(d *schema.ResourceData, meta interface{}) error {
	provider := meta.(*Client)

	recordID, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Error converting Record ID: %s", err)
	}

	updateRecord := &websupport.DNSRecord{
		Name:    d.Get("name").(string),
		Type: 	 d.Get("type").(string),
		Content: d.Get("value").(string),
	}
	if attr, ok := d.GetOk("ttl"); ok {
		updateRecord.TTL, _ = strconv.Atoi(attr.(string))
	}

	/* TODO priority and other attributes
	if attr, ok := d.GetOk("priority"); ok {
		updateRecord.Priority, _ = strconv.Atoi(attr.(string))
	}
	*/

	log.Printf("[DEBUG] Websupport DNS Record update configuration: %#v", updateRecord)

	// TODO check na d.Get("zone").(string) ako je v resourceWebsupportRecordCreate()
	_, err = provider.client.DNS.UpdateDNSRecord(provider.loggedUser.Id, d.Get("zone").(string), recordID, updateRecord)
	// TODO pozriet success, error status codes v dokumentacii
	if err != nil {
		return fmt.Errorf("Failed to update Websupport DNS Record: %s", err)
	}

	return resourceWebsupportRecordRead(d, meta)
}

func resourceWebsupportRecordDelete(d *schema.ResourceData, meta interface{}) error {
	provider := meta.(*Client)

	log.Printf("[INFO] Deleting Websupport DNS Record: %s, %s", d.Get("zone").(string), d.Id())

	recordID, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Error converting Record ID: %s", err)
	}

	// TODO check na d.Get("zone").(string) ako je v resourceWebsupportRecordCreate()
	_, err = provider.client.DNS.DeleteDNSRecord(provider.loggedUser.Id, d.Get("zone").(string), recordID)
	// TODO pozriet success, error status codes v dokumentacii
	if err != nil {
		return fmt.Errorf("Error deleting Websupport DNS Record: %s", err)
	}

	return nil
}