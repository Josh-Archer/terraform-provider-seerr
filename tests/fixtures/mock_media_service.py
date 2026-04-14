import json
import os
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer


SERVICE_KIND = os.environ.get("SERVICE_KIND", "generic").strip().lower()
PORT = int(os.environ.get("PORT", "8080"))
EXPECTED_API_KEY = os.environ.get("EXPECTED_API_KEY", "").strip()


def json_bytes(payload):
    return json.dumps(payload).encode("utf-8")


def xml_bytes(payload):
    return payload.encode("utf-8")


class Handler(BaseHTTPRequestHandler):
    server_version = "SeerrFixture/1.0"

    def _send(self, status, body, content_type="application/json", headers=None):
        self.send_response(status)
        self.send_header("Content-Type", content_type)
        self.send_header("Content-Length", str(len(body)))
        if headers:
            for key, value in headers.items():
                self.send_header(key, value)
        self.end_headers()
        self.wfile.write(body)

    def _ok_json(self, payload, headers=None):
        self._send(200, json_bytes(payload), headers=headers)

    def _ok_xml(self, payload, headers=None):
        self._send(200, xml_bytes(payload), content_type="application/xml", headers=headers)

    def _unauthorized(self):
        self._send(401, json_bytes({"error": "unauthorized"}))

    def _api_key(self):
        return self.headers.get("X-Api-Key", "").strip()

    def _path(self):
        return self.path.split("?", 1)[0]

    def _handle_arr(self):
        if EXPECTED_API_KEY and self._api_key() != EXPECTED_API_KEY:
            self._unauthorized()
            return

        if self._path() == "/api/v3/system/status":
            self._ok_json(
                {
                    "appName": SERVICE_KIND.capitalize(),
                    "instanceName": f"{SERVICE_KIND}-fixture",
                    "version": "0.0.0-test",
                }
            )
            return

        if self._path() == "/api/v3/qualityprofile":
            self._ok_json(
                [
                    {"id": 1, "name": "HD-1080p"},
                    {"id": 2, "name": "Ultra-HD"},
                ]
            )
            return

        self._ok_json({"status": "ok", "service": SERVICE_KIND})

    def _handle_plex(self):
        headers = {
            "X-Plex-Machine-Identifier": "plex-fixture-machine-id",
            "X-Plex-Protocol": "1.0",
            "X-Plex-Version": "1.0.0-test",
            "X-Plex-Product": "Plex Media Server",
        }
        payload = (
            '<?xml version="1.0" encoding="UTF-8"?>'
            '<MediaContainer size="0" friendlyName="Plex Fixture" '
            'machineIdentifier="plex-fixture-machine-id" version="1.0.0-test" />'
        )
        self._ok_xml(payload, headers=headers)

    def _handle_jellyfin(self):
        if EXPECTED_API_KEY and self._api_key() != EXPECTED_API_KEY:
            self._unauthorized()
            return

        self._ok_json(
            {
                "ServerName": "Jellyfin Fixture",
                "Id": "jellyfin-fixture-id",
                "Version": "10.8.0-test",
                "StartupWizardCompleted": True,
            }
        )

    def _handle_tautulli(self):
        api_key = self.path.split("apikey=", 1)[1].split("&", 1)[0] if "apikey=" in self.path else self._api_key()
        if EXPECTED_API_KEY and api_key.strip() != EXPECTED_API_KEY:
            self._unauthorized()
            return

        self._ok_json(
            {
                "response": {
                    "result": "success",
                    "message": None,
                    "data": {
                        "tautulli_version": "2.13.0-test",
                        "server_name": "Tautulli Fixture",
                    },
                }
            }
        )

    def _handle_generic(self):
        self._ok_json({"status": "ok", "service": SERVICE_KIND, "path": self.path})

    def do_GET(self):
        if SERVICE_KIND in {"radarr", "sonarr"}:
            self._handle_arr()
        elif SERVICE_KIND == "plex":
            self._handle_plex()
        elif SERVICE_KIND == "jellyfin":
            self._handle_jellyfin()
        elif SERVICE_KIND == "tautulli":
            self._handle_tautulli()
        else:
            self._handle_generic()

    def do_POST(self):
        self.do_GET()

    def do_PUT(self):
        self.do_GET()

    def do_DELETE(self):
        self.do_GET()

    def log_message(self, fmt, *args):
        print(f"[{SERVICE_KIND}] {self.address_string()} - {fmt % args}", flush=True)


if __name__ == "__main__":
    server = ThreadingHTTPServer(("0.0.0.0", PORT), Handler)
    print(f"starting {SERVICE_KIND} fixture on :{PORT}", flush=True)
    server.serve_forever()
