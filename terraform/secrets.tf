# J-Quants 認証情報を Secret Manager で管理する
# 実際の値は terraform apply 後に手動で登録:
#   gcloud secrets versions add jquants-email --data-file=<(echo -n "your@email.com")
#   gcloud secrets versions add jquants-password --data-file=<(echo -n "yourpassword")

resource "google_secret_manager_secret" "jquants_email" {
  secret_id = "jquants-email"
  replication {
    auto {}
  }
  depends_on = [google_project_service.services]
}

resource "google_secret_manager_secret" "jquants_password" {
  secret_id = "jquants-password"
  replication {
    auto {}
  }
  depends_on = [google_project_service.services]
}

# Jobs SA に Secret へのアクセス権を付与
resource "google_secret_manager_secret_iam_member" "jobs_jquants_email" {
  secret_id = google_secret_manager_secret.jquants_email.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.jobs.email}"
}

resource "google_secret_manager_secret_iam_member" "jobs_jquants_password" {
  secret_id = google_secret_manager_secret.jquants_password.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.jobs.email}"
}
