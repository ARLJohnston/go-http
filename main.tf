terraform {
  required_providers {
    docker = {
      source = "kreuzwerker/docker"
      version = "~> 3.0.1"
    }
  }
}

variable "username" {
  description = "GHCR Username"
  type = "string"
  sensitive = true
}

variable "password" {
  description = "GHCR password (PAT)"
  type = "string"
  sensitive = true
}

provider "docker" {
  registry_auth {
    address = "https://ghcr.io"
    username = var.username
    password = var.password
  }
}

resource "docker_image" "hello" {
  name = "ghcr.io/arljohnston/hello-world-ghcr"
  keep_locally = false
}

resource "docker_container" "hello" {
  image = docker_image.hello.image_id
  name = "tutorial"

  # ports {
  #   internal = 80
  #   external = 8000
  # }
}
