resource "sigma_source_swap_policy" "example" {
  type               = "deployment"
  name               = "Swap warehouse"
  from_connection_id = "connection-id"
  swaps_json = jsonencode({
    toConnection = {
      swapType        = "attribute"
      userAttributeId = "user-attribute-id"
    }
    deploymentSwaps = []
  })
}
