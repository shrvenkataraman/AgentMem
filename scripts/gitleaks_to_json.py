"""Convert gitleaks `config/gitleaks.toml` (MIT licence) to the rules JSON agentmem embeds.

Usage: gitleaks_to_json.py IN.toml OUT.json
"""

import json
import sys
import tomllib


def main(argv: list[str]) -> int:
    if len(argv) != 2:
        print("usage: gitleaks_to_json.py IN.toml OUT.json", file=sys.stderr)
        return 2
    src, dst = argv
    with open(src, "rb") as f:
        doc = tomllib.load(f)
    kept: list[dict] = []
    dropped = 0
    for r in doc.get("rules", []):
        if not r.get("regex"):
            dropped += 1
            continue
        kept.append(
            {
                "id": r["id"],
                "regex": r["regex"],
                "secretGroup": int(r.get("secretGroup", 0)),
                "entropy": float(r.get("entropy", 0)),
                "keywords": [k.lower() for k in r.get("keywords", [])],
            }
        )
    with open(dst, "w") as f:
        json.dump(kept, f, indent=1)
        f.write("\n")
    print(f"kept={len(kept)} dropped={dropped}")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
