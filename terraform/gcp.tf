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


provider "google" {
  credentials = file("./gcp-service-acc.json")
  project     = "${var.project_id}"
  region      = "${var.region}"
}

resource "google_sql_database_instance" "postgres_db" {
  database_version = "POSTGRES_15"
  name             = "postgres-db"
  project          = "${var.project_id}"
  region           = "${var.region}"

  settings {
    activation_policy = "ALWAYS"
    availability_type = "ZONAL"

    backup_configuration {
      backup_retention_settings {
        retained_backups = 7
        retention_unit   = "COUNT"
      }

      enabled                        = true
      location                       = "${var.region}"
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
  project      = "${var.project_id}"
}

resource "google_project_iam_member" "cloud_sql_proxy_permission" {
  project = "${var.project_id}"
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.cloud_sql_proxy.email}"
}

resource "google_container_cluster" "project_cluster" {
  database_encryption {
    state = "DECRYPTED"
  }

  datapath_provider         = "ADVANCED_DATAPATH"

  default_snat_status {
    disabled = false
  }

  dns_config {
    cluster_dns        = "CLOUD_DNS"
    cluster_dns_domain = "cluster.local"
    cluster_dns_scope  = "CLUSTER_SCOPE"
  }

  enable_autopilot            = true

  location = "${var.region}"

  logging_config {
    enable_components = ["SYSTEM_COMPONENTS", "WORKLOADS"]
  }

  master_auth {
    client_certificate_config {
      issue_client_certificate = false
    }
  }

  monitoring_config {
    enable_components = ["SYSTEM_COMPONENTS"]

    managed_prometheus {
      enabled = true
    }
  }

  name    = "project-cluster"

  node_config {
    disk_size_gb = 100
    disk_type    = "pd-standard"
    image_type   = "COS_CONTAINERD"
    machine_type = "e2-small"

    metadata = {
      disable-legacy-endpoints = "true"
    }

    oauth_scopes    = ["https://www.googleapis.com/auth/devstorage.read_only", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/service.management.readonly", "https://www.googleapis.com/auth/servicecontrol", "https://www.googleapis.com/auth/trace.append"]
    service_account = "default"

    shielded_instance_config {
      enable_integrity_monitoring = true
      enable_secure_boot          = true
    }

    taint {
      effect = "NO_SCHEDULE"
      key    = "cloud.google.com/gke-quick-remove"
      value  = "true"
    }
  }

  node_locations = ["${var.region}-a", "${var.region}-b", "${var.region}-c"]

  notification_config {
    pubsub {
      enabled = false
    }
  }

  private_cluster_config {
    enable_private_endpoint = false

    master_global_access_config {
      enabled = false
    }
  }

  project = "${var.project_id}"

  release_channel {
    channel = "REGULAR"
  }

}

resource "google_compute_global_address" "project_ip" {
  name         = "project-ip"
  address_type = "EXTERNAL"
  ip_version   = "IPV4"
  project      = "${var.project_id}"
}

resource "google_dns_managed_zone" "projectsv_org" {
  name          = "projectsv-org"
  visibility    = "public"
  description = "DNS zone for domain: projectsv.org"
  dns_name    = "projectsv.org."
  project       = "${var.project_id}"
}

resource "google_dns_record_set" "frontend" {
  name = "online-store.${google_dns_managed_zone.projectsv_org.dns_name}"
  type = "A"
  ttl  = 300

  managed_zone = google_dns_managed_zone.projectsv_org.name

  rrdatas = [google_compute_global_address.project_ip.address]
}


