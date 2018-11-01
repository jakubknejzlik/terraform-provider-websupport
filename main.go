package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/radoslavoleksak/terraform-provider-websupport/websupport"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: websupport.Provider})
}