#!/usr/bin/env python3
"""Score a labelled query file: recall@5 per query class."""

import json
import os
import subprocess
import sys


def main():
    if len(sys.argv) != 2:
        sys.exit("usage: recall_eval.py QUERIES_TSV")
    path = sys.argv[1]
    bin = os.environ.get("AGENTMEM_BIN", "agentmem")
    env = dict(os.environ, AGENTMEM_QUERYLOG="off")

    n, hit = {}, {}

    def note(cls, ok):
        for c in (cls, "all"):
            n[c] = n.get(c, 0) + 1
            hit[c] = hit.get(c, 0) + int(ok)

    with open(path, encoding="utf-8") as f:
        for line, raw in enumerate(f, 1):
            if not raw.strip() or raw.startswith("#"):
                continue
            cls, session, query = raw.rstrip("\n").split("\t", 2)
            cmd = [
                bin, "ask", "--json", "--no-refresh", "--include-current",
                "--limit", "5", *query.split(),
            ]
            out = subprocess.run(
                cmd, capture_output=True, text=True, check=True, env=env
            )
            data = json.loads(out.stdout)
            sessions = [r["session"] for r in (data["results"] or [])[:5]]
            ok = session in sessions
            if not ok:
                print(f"miss line={line}")
            note(cls, ok)

    for c in sorted(k for k in n if k != "all") + ["all"]:
        print(f"class={c} n={n[c]} hit={hit[c]} recall@5={hit[c] / n[c]:.2f}")


if __name__ == "__main__":
    main()
