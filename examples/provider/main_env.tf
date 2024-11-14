terraform {
  required_providers {
    storagegrid = {
      source  = "dmpe/storagegrid"
      version = "" # My strong advice - always pin this provider's version!
    }
  }
}

provider "storagegrid" {
}