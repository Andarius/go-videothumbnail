variable "VERSION" {
  default = "dev"
}

variable "SUFFIX" {
  default = ""
}

group "default" {
  targets = ["default", "sentry"]
}

target "docker-metadata-action" {
  tags = []
}

target "_common" {
  context    = "."
  dockerfile = "Dockerfile"
  labels = {
    "organization" = "Obitrain"
  }
}

target "default" {
  inherits = ["_common", "docker-metadata-action"]
  args = {
    VERSION    = VERSION
    BUILD_TAGS = ""
  }
}

target "sentry" {
  inherits = ["_common", "docker-metadata-action"]
  args = {
    VERSION    = VERSION
    BUILD_TAGS = "sentry"
  }
}