package main

import (
	"github.com/haf/terraform-provider-sealedsecrets/sealedsecrets"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	opts := &plugin.ServeOpts{ProviderFunc: sealedsecrets.Provider}
	plugin.Serve(opts)
}
