terraform {
  required_version = ">= 1.5"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# 必要な GCP API を有効化
resource "google_project_service" "services" {
  for_each = toset([
    "run.googleapis.com",
    "cloudscheduler.googleapis.com",
    "secretmanager.googleapis.com",
    "artifactregistry.googleapis.com",
    "firestore.googleapis.com",
    "sqladmin.googleapis.com",
  ])
  service            = each.value
  disable_on_destroy = false
}

# Artifact Registry リポジトリ（Docker イメージ保存先）
resource "google_artifact_registry_repository" "itsumo" {
  location      = var.region
  repository_id = "itsumo"
  format        = "DOCKER"
  description   = "itsumo container images"
  depends_on    = [google_project_service.services]
}

locals {
  image_base = "${var.region}-docker.pkg.dev/${var.project_id}/itsumo"
  job_base_url = "https://${var.region}-run.googleapis.com/v2/projects/${var.project_id}/locations/${var.region}/jobs"
}
