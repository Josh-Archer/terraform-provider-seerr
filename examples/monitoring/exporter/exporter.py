#!/usr/bin/env python3
# Copyright (c) Josh Archer
# SPDX-License-Identifier: MPL-2.0
"""
Lightweight Prometheus exporter for Seerr / Jellyseerr / Overseerr instances
and Terraform / OpenTofu Infrastructure as Code state outputs.
"""

import json
import logging
import os
import subprocess
import time
from http.server import HTTPServer, BaseHTTPRequestHandler
from urllib.request import Request, urlopen
from urllib.error import URLError, HTTPError

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s"
)
logger = logging.getLogger("seerr-exporter")

PORT = int(os.environ.get("EXPORTER_PORT", "9850"))
SEERR_URL = os.environ.get("SEERR_URL", "http://localhost:5055").rstrip("/")
SEERR_API_KEY = os.environ.get("SEERR_API_KEY", "")
TF_STATE_PATH = os.environ.get("TF_STATE_PATH", "")
SCRAPE_INTERVAL = int(os.environ.get("SCRAPE_INTERVAL_SECONDS", "30"))


def fetch_json(endpoint: str) -> dict:
    url = f"{SEERR_URL}{endpoint}"
    req = Request(url, headers={
        "X-Api-Key": SEERR_API_KEY,
        "Accept": "application/json",
        "User-Agent": "terraform-provider-seerr-exporter/1.0"
    })
    with urlopen(req, timeout=10) as resp:
        return json.loads(resp.read().decode("utf-8"))


def load_tf_state_metrics() -> dict:
    """Reads metrics from local terraform state or JSON output if specified."""
    if not TF_STATE_PATH or not os.path.exists(TF_STATE_PATH):
        return {}
    try:
        with open(TF_STATE_PATH, "r", encoding="utf-8") as f:
            data = json.load(f)
            if "metrics_summary" in data:
                return data["metrics_summary"].get("value", {})
            return data
    except Exception as e:
        logger.warning("Failed to parse Terraform state file %s: %s", TF_STATE_PATH, e)
        return {}


def collect_metrics() -> str:
    start_time = time.time()
    lines = []
    success = 1

    try:
        # 1. Server Info & Telemetry
        about = {}
        status = {}
        users_count = 0
        requests_summary = {"pending": 0, "approved": 0, "declined": 0, "total": 0}
        media_count = 0
        jobs = []
        backup_retention = 0

        if SEERR_API_KEY:
            try:
                about = fetch_json("/api/v1/settings/about")
            except Exception as e:
                logger.warning("Failed fetching /api/v1/settings/about: %s", e)

            try:
                status = fetch_json("/api/v1/status")
            except Exception as e:
                logger.warning("Failed fetching /api/v1/status: %s", e)

            try:
                users_resp = fetch_json("/api/v1/user?take=1")
                page_info = users_resp.get("pageInfo", {})
                users_count = page_info.get("results", len(users_resp.get("results", [])))
            except Exception as e:
                logger.warning("Failed fetching /api/v1/user: %s", e)

            try:
                req_resp = fetch_json("/api/v1/request?take=1")
                page_info = req_resp.get("pageInfo", {})
                requests_summary["total"] = page_info.get("results", 0)

                pending_resp = fetch_json("/api/v1/request?filter=pending&take=1")
                requests_summary["pending"] = pending_resp.get("pageInfo", {}).get("results", 0)
                
                approved_resp = fetch_json("/api/v1/request?filter=approved&take=1")
                requests_summary["approved"] = approved_resp.get("pageInfo", {}).get("results", 0)

                declined_resp = fetch_json("/api/v1/request?filter=declined&take=1")
                requests_summary["declined"] = declined_resp.get("pageInfo", {}).get("results", 0)
            except Exception as e:
                logger.warning("Failed fetching requests: %s", e)

            try:
                media_resp = fetch_json("/api/v1/media?take=1")
                media_count = media_resp.get("pageInfo", {}).get("results", 0)
            except Exception as e:
                logger.warning("Failed fetching media: %s", e)

            try:
                jobs = fetch_json("/api/v1/settings/jobs")
                if not isinstance(jobs, list):
                    jobs = []
            except Exception as e:
                logger.warning("Failed fetching jobs: %s", e)

            try:
                backup = fetch_json("/api/v1/settings/backup")
                backup_retention = backup.get("retention", 0)
            except Exception as e:
                logger.warning("Failed fetching backup settings: %s", e)
        else:
            # Fallback to Terraform state file if direct API key not supplied
            tf_data = load_tf_state_metrics()
            if tf_data:
                about = {
                    "version": tf_data.get("version", "unknown"),
                    "totalRequests": tf_data.get("total_requests", 0),
                    "totalMediaItems": tf_data.get("total_media_items", 0),
                    "tz": "UTC"
                }
                status = {
                    "updateAvailable": tf_data.get("update_available", False),
                    "commitsBehind": tf_data.get("commits_behind", 0),
                    "restartRequired": tf_data.get("restart_required", False)
                }
                users_count = tf_data.get("total_users", 0)
                requests_summary["pending"] = tf_data.get("pending_requests", 0)
                requests_summary["total"] = tf_data.get("total_requests", 0)
                backup_retention = tf_data.get("backup_retention", 0)

        # Output Prometheus formatted metrics
        version = about.get("version", "unknown")
        tz = about.get("tz", "UTC")
        commit_tag = status.get("commitTag", "")

        lines.append("# HELP seerr_server_info Information about the running Seerr instance.")
        lines.append("# TYPE seerr_server_info gauge")
        lines.append(f'seerr_server_info{{version="{version}",commit_tag="{commit_tag}",tz="{tz}"}} 1')

        lines.append("# HELP seerr_update_available 1 if an upstream release is available, 0 otherwise.")
        lines.append("# TYPE seerr_update_available gauge")
        lines.append(f'seerr_update_available {1 if status.get("updateAvailable") else 0}')

        lines.append("# HELP seerr_commits_behind Number of commits the build is behind upstream.")
        lines.append("# TYPE seerr_commits_behind gauge")
        lines.append(f'seerr_commits_behind {status.get("commitsBehind", 0)}')

        lines.append("# HELP seerr_restart_required 1 if a server restart is needed, 0 otherwise.")
        lines.append("# TYPE seerr_restart_required gauge")
        lines.append(f'seerr_restart_required {1 if status.get("restartRequired") else 0}')

        lines.append("# HELP seerr_requests_total Total cumulative media requests in Seerr.")
        lines.append("# TYPE seerr_requests_total gauge")
        lines.append(f'seerr_requests_total {about.get("totalRequests", requests_summary["total"])}')

        lines.append("# HELP seerr_media_items_total Total indexed media items in Seerr.")
        lines.append("# TYPE seerr_media_items_total gauge")
        lines.append(f'seerr_media_items_total {about.get("totalMediaItems", media_count)}')

        lines.append("# HELP seerr_users_total Total registered users in Seerr.")
        lines.append("# TYPE seerr_users_total gauge")
        lines.append(f'seerr_users_total {users_count}')

        lines.append("# HELP seerr_requests_pending_total Number of requests currently awaiting approval.")
        lines.append("# TYPE seerr_requests_pending_total gauge")
        lines.append(f'seerr_requests_pending_total {requests_summary["pending"]}')

        lines.append("# HELP seerr_requests_status_total Total requests partitioned by status.")
        lines.append("# TYPE seerr_requests_status_total gauge")
        lines.append(f'seerr_requests_status_total{{status="pending"}} {requests_summary["pending"]}')
        lines.append(f'seerr_requests_status_total{{status="approved"}} {requests_summary["approved"]}')
        lines.append(f'seerr_requests_status_total{{status="declined"}} {requests_summary["declined"]}')

        lines.append("# HELP seerr_backup_retention_days Configured retention period for automated database backups.")
        lines.append("# TYPE seerr_backup_retention_days gauge")
        lines.append(f'seerr_backup_retention_days {backup_retention}')

        lines.append("# HELP seerr_job_running 1 if background job is actively running, 0 otherwise.")
        lines.append("# TYPE seerr_job_running gauge")
        for job in jobs:
            job_id = job.get("id", "unknown")
            job_name = job.get("name", "unknown")
            is_running = 1 if job.get("running") else 0
            lines.append(f'seerr_job_running{{job_id="{job_id}",job_name="{job_name}"}} {is_running}')

        # 2. Terraform State & Drift Metrics
        drift_status = int(os.environ.get("SEERR_TF_DRIFT_STATUS", "0"))
        lines.append("# HELP seerr_terraform_drift_status 0=In Sync, 1=Drift Detected, 2=Plan Error.")
        lines.append("# TYPE seerr_terraform_drift_status gauge")
        lines.append(f'seerr_terraform_drift_status {drift_status}')

        last_backup = int(os.environ.get("SEERR_LAST_STATE_BACKUP_TIME", str(int(time.time()))))
        lines.append("# HELP seerr_last_state_backup_timestamp_seconds Unix timestamp of last state backup.")
        lines.append("# TYPE seerr_last_state_backup_timestamp_seconds gauge")
        lines.append(f'seerr_last_state_backup_timestamp_seconds {last_backup}')

    except Exception as e:
        logger.error("Scrape failed: %s", e)
        success = 0

    duration = time.time() - start_time
    lines.append("# HELP seerr_scrape_duration_seconds Duration of metric collection in seconds.")
    lines.append("# TYPE seerr_scrape_duration_seconds gauge")
    lines.append(f'seerr_scrape_duration_seconds {duration:.4f}')

    lines.append("# HELP seerr_scrape_success 1 if scrape succeeded, 0 otherwise.")
    lines.append("# TYPE seerr_scrape_success gauge")
    lines.append(f'seerr_scrape_success {success}')

    return "\n".join(lines) + "\n"


class MetricsHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path in ("/metrics", "/"):
            metrics_payload = collect_metrics()
            self.send_response(200)
            self.send_header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
            self.end_headers()
            self.wfile.write(metrics_payload.encode("utf-8"))
        elif self.path == "/healthz":
            self.send_response(200)
            self.send_header("Content-Type", "text/plain")
            self.end_headers()
            self.wfile.write(b"OK\n")
        else:
            self.send_response(404)
            self.end_headers()

    def log_message(self, format, *args):
        # Suppress standard HTTP access logs to keep terminal quiet
        pass


def main():
    server_address = ("", PORT)
    httpd = HTTPServer(server_address, MetricsHandler)
    logger.info("Starting Seerr Prometheus Exporter on :%d (Target: %s)", PORT, SEERR_URL)
    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        logger.info("Shutting down exporter...")
        httpd.server_close()


if __name__ == "__main__":
    main()
