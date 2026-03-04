terraform {
  backend "local" {
    path = "/terraform-data/terraform.tfstate"
  }

  required_providers {
    zitadel = {
      source  = "zitadel/zitadel"
      version = "2.9.0"
    }
  }
}

variable "local_dev_host" {
  type        = string
  description = "dev.rezible.com"
}

variable "admin_sa_pat_path" {
  type        = string
  description = "Path inside the terraform container to the admin PAT file"
}

locals {
  auth_domain     = "auth.${var.local_dev_host}"
  login_redirect   = "https://app.${var.local_dev_host}/auth/callback"
}

provider "zitadel" {
  domain                   = local.auth_domain
  insecure                 = false
  insecure_skip_verify_tls = true
  access_token             = trimspace(file(var.admin_sa_pat_path))
}

data "zitadel_orgs" "search" {
  name          = "local-dev"
  name_method   = "TEXT_QUERY_METHOD_CONTAINS_IGNORE_CASE"
  domain        = "auth.dev.rezible.com"
  domain_method = "TEXT_QUERY_METHOD_CONTAINS_IGNORE_CASE"
  state         = "ORG_STATE_ACTIVE"
}

data "zitadel_org" "local_dev" {
  id = data.zitadel_orgs.search.ids[0]
}

resource "zitadel_instance_features" "default" {
  login_default_org   = true
  oidc_token_exchange = true
  user_schema         = false
  improved_performance = []
  debug_oidc_parent_error            = false
  oidc_single_v1_session_termination = true
  enable_back_channel_logout         = true
  login_v2 {
    required = false
    # base_uri = "https://login.example.com"
  }
  permission_check_v2     = true
  console_use_v2_user_api = true
}

resource "zitadel_default_domain_policy" "default" {
  user_login_must_be_domain                   = false
  validate_org_domains                        = false
  smtp_sender_address_matches_instance_domain = false
}

resource "zitadel_project" "rezible" {
  org_id = data.zitadel_org.local_dev.id
  name = "rezible"
  project_role_assertion = true
  has_project_check = true
  project_role_check = true
}

resource "zitadel_application_oidc" "rezible_frontend" {
  org_id = data.zitadel_org.local_dev.id
  project_id = zitadel_project.rezible.id
  name       = "rezible-frontend"

  response_types             = ["OIDC_RESPONSE_TYPE_CODE"]
  grant_types                = ["OIDC_GRANT_TYPE_AUTHORIZATION_CODE"]
  redirect_uris              = [local.login_redirect]
  post_logout_redirect_uris  = [local.login_redirect]

  app_type         = "OIDC_APP_TYPE_WEB"
  auth_method_type = "OIDC_AUTH_METHOD_TYPE_BASIC"
  version          = "OIDC_VERSION_1_0"
  clock_skew       = "0s"
  dev_mode         = true

  access_token_type            = "OIDC_TOKEN_TYPE_BEARER"
  access_token_role_assertion  = false
  id_token_role_assertion      = false
  id_token_userinfo_assertion  = false
  additional_origins           = []
  skip_native_app_success_page = false
}

resource "zitadel_org_idp_oidc" "test" {
  org_id = data.zitadel_org.local_dev.id

  name          = "OIDC Test"
  issuer        = "https://oidc.${var.local_dev_host}/"
  client_id     = "client"
  client_secret = "secret"
  scopes        = ["openid", "profile", "email"]

  is_creation_allowed = true
  is_auto_creation    = true
  is_auto_update      = true
  is_linking_allowed  = true
  auto_linking        = "AUTO_LINKING_OPTION_EMAIL"

  is_id_token_mapping = true
}

resource "zitadel_action" "fix_dev_idp_attributes" {
  org_id = data.zitadel_org.local_dev.id

  name            = "fixDevIDPAttributes"
  allowed_to_fail = true
  script          = <<EOT
    let logger = require("zitadel/log")
    function fixDevIDPAttributes(ctx, api) {
      logger.info("providerInfo: " + JSON.stringify(ctx.v1.providerInfo));
      let email = ctx.v1.providerInfo["email"];
      if (email) {
        if (!ctx.v1.providerInfo["preferred_username"])
          api.setPreferredUsername(email.split("@", 2)[0]);
        api.setEmailVerified(true);
      }
    }
  EOT
  timeout         = "10s"
}

resource "zitadel_trigger_actions" "fix_dev_idp_attributes" {
  org_id       = data.zitadel_org.local_dev.id
  flow_type    = "FLOW_TYPE_EXTERNAL_AUTHENTICATION"
  trigger_type = "TRIGGER_TYPE_POST_AUTHENTICATION"
  action_ids = [zitadel_action.fix_dev_idp_attributes.id]
}

resource "zitadel_login_policy" "org_login" {
  org_id = data.zitadel_org.local_dev.id

  user_login = false
  allow_register = false
  allow_external_idp = true

  force_mfa            = false
  force_mfa_local_only = false

  hide_password_reset = true

  ignore_unknown_usernames = false

  default_redirect_uri = local.login_redirect

  passwordless_type             = "PASSWORDLESS_TYPE_ALLOWED"
  password_check_lifetime       = "240h0m0s"
  external_login_check_lifetime = "240h0m0s"
  multi_factor_check_lifetime   = "24h0m0s"
  mfa_init_skip_lifetime        = "720h0m0s"
  second_factor_check_lifetime  = "24h0m0s"

  idps = [zitadel_org_idp_oidc.test.id]
}

output "org_id" {
  value = data.zitadel_org.local_dev.id
}

output "project_id" {
  value = zitadel_project.rezible.id
}

output "application_id" {
  value = zitadel_application_oidc.rezible_frontend.id
}
