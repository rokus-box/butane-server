provider "aws" {
  region = "eu-central-1"
}

variable "functions" {
  description = "Map of functions to create"
  type        = map(map(string))
  default     = {
    oauth = {
      handler     = "bootstrap"
      runtime     = "provided.al2023"
      description = "OAuth handler to authenticate users with external OAuth providers"
    }
    vault = {
      handler     = "bootstrap"
      runtime     = "provided.al2023"
      description = "Handles CRUD operations for Vaults of the authenticated user"
    }
    secret = {
      handler     = "bootstrap"
      runtime     = "provided.al2023"
      description = "Handles CRUD operations for Secrets of the authenticated user"
    }
    audit_log = {
      handler     = "bootstrap"
      runtime     = "provided.al2023"
      description = "Handles CRUD operations for Audit Logs of the authenticated user"
    }
    session = {
      handler     = "app.handler"
      runtime     = "nodejs20.x"
      description = "Handles CRUD operations for Sessions of the authenticated user"
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
    # Audit Log
    "audit_log;GET /audit-log",
  ]
}
