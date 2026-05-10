output "artifact_registry_url" {
  description = "Docker イメージのプッシュ先 URL"
  value       = "${var.region}-docker.pkg.dev/${var.project_id}/itsumo"
}

output "jobs_service_account" {
  description = "Cloud Run Jobs のサービスアカウント"
  value       = google_service_account.jobs.email
}

output "push_commands" {
  description = "イメージをビルド＆プッシュするコマンド"
  value = <<-EOT
    gcloud auth configure-docker ${var.region}-docker.pkg.dev

    docker build -f cmd/jobs/fetch-prices/Dockerfile \
      -t ${var.region}-docker.pkg.dev/${var.project_id}/itsumo/fetch-prices:latest .
    docker push ${var.region}-docker.pkg.dev/${var.project_id}/itsumo/fetch-prices:latest

    docker build -f cmd/jobs/fetch-earnings/Dockerfile \
      -t ${var.region}-docker.pkg.dev/${var.project_id}/itsumo/fetch-earnings:latest .
    docker push ${var.region}-docker.pkg.dev/${var.project_id}/itsumo/fetch-earnings:latest
  EOT
}
