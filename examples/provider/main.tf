terraform {
  required_providers {
    storagegrid = {
      source  = "dmpe/storagegrid"
      version = "" # My strong advice - always pin this provider's version!
    }
  }
}

provider "storagegrid" {
  address   = "https://grid.firm.com:9443"
  username  = "grid"
  password  = "change_me"
  tenant    = "<int>" # Tenant ID
  insecure  = false
}