# ── サービスアカウント ─────────────────────────────────────────────────────────

# Cloud Run Jobs 実行用
resource "google_service_account" "jobs" {
  account_id   = "itsumo-jobs"
  display_name = "itsumo Jobs SA"
}

resource "google_project_iam_member" "jobs_firestore" {
  project = var.project_id
  role    = "roles/datastore.user"
  member  = "serviceAccount:${google_service_account.jobs.email}"
}

# Cloud Scheduler 起動用
resource "google_service_account" "scheduler" {
  account_id   = "itsumo-scheduler"
  display_name = "itsumo Scheduler SA"
}

# Scheduler SA が Cloud Run Jobs を実行できる権限
resource "google_project_iam_member" "scheduler_run_developer" {
  project = var.project_id
  role    = "roles/run.developer"
  member  = "serviceAccount:${google_service_account.scheduler.email}"
}

# ── Cloud Run Jobs ─────────────────────────────────────────────────────────────

resource "google_cloud_run_v2_job" "fetch_prices" {
  name     = "fetch-prices"
  location = var.region

  template {
    template {
      service_account = google_service_account.jobs.email
      max_retries     = 1

      containers {
        name  = "fetch-prices"
        image = "${local.image_base}/fetch-prices:latest"

        depends_on = ["cloud-sql-proxy"]

        env {
          name  = "GCP_PROJECT_ID"
          value = var.project_id
        }
        env {
          name  = "DB_DSN"
          value = "host=127.0.0.1 port=5432 user=itsumo_app dbname=itsumo sslmode=disable"
        }
        env {
          name = "DB_PASSWORD"
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.db_password.secret_id
              version = "latest"
            }
          }
        }
        env {
          name = "JQUANTS_EMAIL"
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.jquants_email.secret_id
              version = "latest"
            }
          }
        }
        env {
          name = "JQUANTS_PASSWORD"
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.jquants_password.secret_id
              version = "latest"
            }
          }
        }

        resources {
          limits = {
            cpu    = "1"
            memory = "512Mi"
          }
        }
      }

      containers {
        name  = "cloud-sql-proxy"
        image = "gcr.io/cloud-sql-connectors/cloud-sql-proxy:2"
        args = [
          "--structured-logs",
          "--port=5432",
          "${var.project_id}:${var.region}:itsumo-prices",
        ]
        resources {
          limits = {
            cpu    = "0.5"
            memory = "128Mi"
          }
        }
      }
    }
  }

  depends_on = [
    google_artifact_registry_repository.itsumo,
    google_project_service.services,
    google_sql_database_instance.prices,
  ]
}

resource "google_cloud_run_v2_job" "fetch_earnings" {
  name     = "fetch-earnings"
  location = var.region

  template {
    template {
      service_account = google_service_account.jobs.email
      max_retries     = 1

      containers {
        image = "${local.image_base}/fetch-earnings:latest"

        env {
          name  = "GCP_PROJECT_ID"
          value = var.project_id
        }
        env {
          name = "JQUANTS_EMAIL"
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.jquants_email.secret_id
              version = "latest"
            }
          }
        }
        env {
          name = "JQUANTS_PASSWORD"
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.jquants_password.secret_id
              version = "latest"
            }
          }
        }

        resources {
          limits = {
            cpu    = "1"
            memory = "512Mi"
          }
        }
      }
    }
  }

  depends_on = [
    google_artifact_registry_repository.itsumo,
    google_project_service.services,
  ]
}

# ── Cloud Scheduler ───────────────────────────────────────────────────────────

# 平日 17:30 JST（UTC 08:30）に株価取得
resource "google_cloud_scheduler_job" "fetch_prices" {
  name      = "fetch-prices-daily"
  region    = var.region
  schedule  = "30 8 * * 1-5"
  time_zone = "Asia/Tokyo"

  http_target {
    http_method = "POST"
    uri         = "${local.job_base_url}/fetch-prices:run"

    oauth_token {
      service_account_email = google_service_account.scheduler.email
    }
  }

  depends_on = [
    google_cloud_run_v2_job.fetch_prices,
    google_project_service.services,
  ]
}

# 平日 22:00 JST（UTC 13:00）に決算情報取得
resource "google_cloud_scheduler_job" "fetch_earnings" {
  name      = "fetch-earnings-daily"
  region    = var.region
  schedule  = "0 13 * * 1-5"
  time_zone = "Asia/Tokyo"

  http_target {
    http_method = "POST"
    uri         = "${local.job_base_url}/fetch-earnings:run"

    oauth_token {
      service_account_email = google_service_account.scheduler.email
    }
  }

  depends_on = [
    google_cloud_run_v2_job.fetch_earnings,
    google_project_service.services,
  ]
}
