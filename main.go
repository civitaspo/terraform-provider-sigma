package main

import (
	"context"
	"log"

	sigmaprovider "github.com/civitaspo/terraform-provider-sigma/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

const providerAddress = "registry.terraform.io/civitaspo/sigma"

var version = "dev"

func main() {
	if err := providerserver.Serve(context.Background(), sigmaprovider.New(version), providerserver.ServeOpts{
		Address: providerAddress,
	}); err != nil {
		log.Fatal(err)
	}
}
