terraform {
  required_providers {
    googleforms = {
      source  = "45ck/googleforms"
      version = "~> 0.1"
    }
  }
}

# Uses Application Default Credentials by default.
# Set GOOGLE_CREDENTIALS env var for service account JSON.
provider "googleforms" {}
