provider "google" {
  credentials = file("./gcp-service-acc.json")
  project     = "<PROJECT_ID>"
  region      = "europe-central2"
}

resource "google_sql_database_instance" "default" {
  name             = "postgres-db"
  region           = "europe-central2"
  database_version = "POSTGRES_15"

  settings {
    tier = "db-f1-micro"
  }
}

resource "google_container_cluster" "primary" {
  name     = "project-cluster"
  location = "europe-central2"

  initial_node_count = 3

  master_auth {
    username = ""
    password = ""

    client_certificate_config {
      issue_client_certificate = false
    }
  }
}

resource "google_container_node_pool" "primary_preemptible_nodes" {
  name       = "my-node-pool"
  location   = "europe-central2"
  cluster    = google_container_cluster.primary.name
  node_count = 1

  node_config {
    preemptible  = true
    machine_type = "e2-medium"

    metadata = {
      disable-legacy-endpoints = "true"
    }

    oauth_scopes = [
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring",
    ]
  }
}

resource "kubernetes_ingress" "example" {
  metadata {
    name = "example"
  }

  spec {
    rule {
      host = "www.example.com"

      http {
        path {
          path = "/store/*"

          backend {
            service_name = "example"
            service_port = "8080"
          }
        }
      }
    }
  }
}