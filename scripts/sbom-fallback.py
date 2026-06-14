#!/usr/bin/env python3
"""Minimal CycloneDX 1.5 SBOM fallback when syft is not installed.
Reads `go list -m -json all` and emits a CycloneDX JSON document.
"""
import json
import os
import subprocess
import sys
import uuid
from datetime import datetime, timezone


def main() -> int:
    env = os.environ.copy()
    env.pop("GOROOT", None)
    out = subprocess.check_output(["go", "list", "-m", "-json", "all"], text=True, env=env)
    decoder = json.JSONDecoder()
    idx = 0
    components: list[dict[str, object]] = []
    text = out.strip()
    while idx < len(text):
        while idx < len(text) and text[idx] in " \n\r\t":
            idx += 1
        if idx >= len(text):
            break
        obj, end = decoder.raw_decode(text, idx)
        idx = end
        if obj.get("Main"):
            continue
        path = obj.get("Path") or ""
        version = obj.get("Version") or ""
        if not path or not version:
            continue
        components.append(
            {
                "type": "library",
                "name": path,
                "version": version,
                "purl": f"pkg:golang/{path}@{version}",
                "bom-ref": f"{path}@{version}",
            }
        )

    doc = {
        "bomFormat": "CycloneDX",
        "specVersion": "1.5",
        "serialNumber": f"urn:uuid:{uuid.uuid4()}",
        "version": 1,
        "metadata": {
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "component": {
                "type": "application",
                "name": "harness",
                "bom-ref": "harness",
            },
            "tools": [{"vendor": "HarnessX", "name": "scripts/sbom-fallback.py", "version": "1"}],
        },
        "components": components,
    }
    json.dump(doc, sys.stdout, indent=2)
    sys.stdout.write("\n")
    return 0


if __name__ == "__main__":
    sys.exit(main())
