variable "VERSION" {
  default = "dev"
}

group "default" {
  targets = ["default", "sentry"]
}

target "docker-metadata-action" {
  tags = []
}

target "default" {
  inherits   = ["docker-metadata-action"]
  context    = "."
  dockerfile = "Dockerfile"
  args = {
    VERSION    = VERSION
    BUILD_TAGS = ""
  }
  labels = {
    "organization" = "Obitrain"
  }
}

target "sentry" {
  inherits = ["default"]
  args = {
    VERSION    = VERSION
    BUILD_TAGS = "sentry"
  }
  tags = [for tag in target.docker-metadata-action.tags : "${tag}-sentry"]
}