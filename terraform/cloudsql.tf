# ── Cloud SQL (PostgreSQL) ─────────────────────────────────────────────────────

resource "google_sql_database_instance" "prices" {
  name             = "itsumo-prices"
  database_version = "POSTGRES_16"
  region           = var.region
  deletion_protection = true

  settings {
    tier              = "db-f1-micro"
    availability_type = "ZONAL"

    backup_configuration {
      enabled                        = true
      point_in_time_recovery_enabled = true
      start_time                     = "03:00"
    }

    ip_configuration {
      ipv4_enabled = true
    }
  }

  depends_on = [google_project_service.services]
}

resource "google_sql_database" "itsumo" {
  name     = "itsumo"
  instance = google_sql_database_instance.prices.name
}

resource "google_sql_user" "app" {
  name     = "itsumo_app"
  instance = google_sql_database_instance.prices.name
  password = random_password.db_password.result
}

resource "random_password" "db_password" {
  length  = 32
  special = false
}

resource "google_secret_manager_secret" "db_password" {
  secret_id = "cloudsql-db-password"
  replication { auto {} }
  depends_on = [google_project_service.services]
}

resource "google_secret_manager_secret_version" "db_password" {
  secret      = google_secret_manager_secret.db_password.id
  secret_data = random_password.db_password.result
}

# Jobs SA に Cloud SQL 接続権限を付与
resource "google_project_iam_member" "jobs_cloudsql" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.jobs.email}"
}

resource "google_secret_manager_secret_iam_member" "jobs_db_password" {
  secret_id = google_secret_manager_secret.db_password.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.jobs.email}"
}

# NOTE: Cloud Run サービス（app server）の SA ができたら同様の権限付与が必要。
