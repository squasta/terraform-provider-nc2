// Command terraform-provider-nc2 is the Terraform provider plugin for
// Nutanix Cloud Clusters (NC2). It registers itself with the
// terraform-plugin-framework's providerserver and serves the protocol
// over the framework's standard plugin transport.
//
// The address `registry.terraform.io/nutanix/nc2` is registered with
// the Terraform Registry; the `nutanix/nc2` namespace MAY change at
// the first release based on the Registry's allocation.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/nutanix/terraform-provider-nc2/internal/provider"
)

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/nutanix/nc2",
		Debug:   debug,
	}
	if err := providerserver.Serve(context.Background(), provider.New, opts); err != nil {
		log.Fatal(err)
	}
}
