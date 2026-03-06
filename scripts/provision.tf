terraform {
  backend "local" {
    path = "/terraform-data/terraform.tfstate"
  }
}
