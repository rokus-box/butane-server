provider "aws" {
  region = "eu-central-1"
}

variable "functions" {
  description = "Map of functions to create"
  type        = map(map(string))
  default     = {
    oauth = {
      handler     = "bootstrap"
      runtime     = "provided.al2"
      description = "OAuth handler to authenticate users with external OAuth providers"
    }
    vault = {
      handler     = "bootstrap"
      runtime     = "provided.al2"
      description = "Vault handler to retrieve Vaults of the authenticated user"
    }
    secret = {
      handler     = "bootstrap"
      runtime     = "provided.al2"
      description = "Secret handler to retrieve Secrets of the authenticated user"
    }
    sessions = {
      handler     = "app.handler"
      runtime     = "nodejs18.x"
      description = "Greets the world"
    }
  }
}

variable "endpoints" {
  type        = list(string)
  description = "List of API Gateway HTTP endpoints to create. The format is '<function_name>;<METHOD> /<endpoint>'"
  default     = [
    # OAuth
    "oauth;POST /oauth/{provider}/authenticate",
    # Vault
    "vault;GET /vault",
    "vault;POST /vault",
    "vault;PATCH /vault/{id}",
    # Secret
    "secret;GET /vault/{vaultId}/secret",
    "secret;POST /vault/{vaultId}/secret",
    "secret;PATCH /vault/{vaultId}/secret/{id}",
    "secret;DELETE /vault/{vaultId}/secret/{id}",
    # Session
    "session;DELETE /session",
  ]
}
