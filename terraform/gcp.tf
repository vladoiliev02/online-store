variable "project_id" {
  description = "The ID of the project"
  type        = string
  default     = "go-course-project-409711"
}

variable "region" {
  description = "The region of the project"
  type        = string
  default     = "europe-central2"
}

variable "gcp_credentials" {
  description = "GCP credentials"
  type        = string
}

provider "google" {
  credentials = var.gcp_credentials
  project     = var.project_id
  region      = var.region
}

resource "google_sql_database_instance" "postgres_db" {
  database_version = "POSTGRES_15"
  name             = "postgres-db"
  project          = var.project_id
  region           = var.region

  settings {
    activation_policy = "ALWAYS"
    availability_type = "ZONAL"

    backup_configuration {
      backup_retention_settings {
        retained_backups = 7
        retention_unit   = "COUNT"
      }

      enabled                        = true
      location                       = var.region
      start_time                     = "10:00"
      transaction_log_retention_days = 7
    }

    disk_autoresize       = false
    disk_autoresize_limit = 0
    disk_size             = 10
    disk_type             = "PD_SSD"

    ip_configuration {
      ipv4_enabled = true
    }

    pricing_plan = "PER_USE"
    tier         = "db-custom-2-8192"
  }
}

resource "google_service_account" "cloud_sql_proxy" {
  account_id   = "cloud-sql-proxy"
  display_name = "cloud-sql-proxy"
  project      = var.project_id
}

resource "google_project_iam_member" "cloud_sql_proxy_permission" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.cloud_sql_proxy.email}"
}

resource "google_container_cluster" "project_cluster" {
  name = "project-cluster"
  location = var.region
  project = var.project_id
  enable_autopilot = true
}

resource "google_compute_global_address" "project_ip" {
  name         = "project-ip"
  address_type = "EXTERNAL"
  ip_version   = "IPV4"
  project      = var.project_id
}

resource "google_dns_managed_zone" "projectsv_org" {
  name        = "projectsv-org"
  visibility  = "public"
  description = "DNS zone for domain: projectsv.org"
  dns_name    = "projectsv.org."
  project     = var.project_id
}

resource "google_dns_record_set" "frontend" {
  name = "online-store.${google_dns_managed_zone.projectsv_org.dns_name}"
  type = "A"
  ttl  = 300

  managed_zone = google_dns_managed_zone.projectsv_org.name

  rrdatas = [google_compute_global_address.project_ip.address]
}

resource "google_storage_bucket" "terraform_state" {
  name          = "terraform-state-project"
  location      = var.region
  storage_class = "STANDARD"
}
