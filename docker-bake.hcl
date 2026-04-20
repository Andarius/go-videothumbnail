variable "VERSION" {
  default = "dev"
}

variable "REGISTRY" {
  default = "andarius/go-videothumbnail"
}

variable "TAGS" {
  default = []
}

group "default" {
  targets = ["default", "sentry"]
}

target "default" {
  context    = "."
  dockerfile = "Dockerfile"
  args = {
    VERSION    = VERSION
    BUILD_TAGS = ""
  }
  tags   = TAGS
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
  tags = [for tag in TAGS : "${tag}-sentry"]
}