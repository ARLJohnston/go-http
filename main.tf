terraform {
  required_providers {
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 3.0.1"
    }
    # azurerm = {
    #   source  = "hashicorp/azurerm"
    #   version = "~> 3.0.2"
    # }
  }
  required_version = ">= 1.1.0"
}

variable "username" {
  description = "GHCR Username"
  type        = string
}

variable "password" {
  description = "GHCR password (PAT)"
  type        = string
  sensitive   = true
}

provider "docker" {
  registry_auth {
    address  = "https://ghcr.io"
    username = var.username
    password = var.password
  }
}

resource "docker_image" "http" {
  name         = "ghcr.io/arljohnston/go-http"
  keep_locally = false
}

resource "docker_container" "http" {
  image = docker_image.http.image_id
  name  = "http"

  ports {
    internal = 8080
    external = 8080
  }
}
