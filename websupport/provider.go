package websupport

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("WEBSUPPORT_USERNAME", nil),
				Description: "A Websupport username.",
			},

			"password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("WEBSUPPORT_PASSWORD", nil),
				Description: "The Websupport password.",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"websupport_record": resourceWebsupportRecord(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		Username:     d.Get("username").(string),
		Password:     d.Get("password").(string),
	}

	return config.Client()
}