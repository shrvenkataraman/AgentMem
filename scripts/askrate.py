#!/usr/bin/env python3
"""Count agentmem asks per agent and session."""

import collections
import datetime
import json
import sqlite3
import sys


def main():
    if len(sys.argv) != 3:
        sys.exit("usage: askrate.py QUERY_LOG INDEX_DB")
    query_log, index_db = sys.argv[1], sys.argv[2]

    entries = []
    with open(query_log, encoding="utf-8") as fh:
        for line in fh:
            line = line.strip()
            if not line:
                continue
            entries.append(json.loads(line))

    start = 0
    for entry in entries:
        ts = entry.get("ts")
        if not ts:
            continue
        ms = int(
            datetime.datetime.fromisoformat(ts.replace("Z", "+00:00")).timestamp()
            * 1000
        )
        if start == 0 or ms < start:
            start = ms

    # Without a start time, every session in the index would count as measured.
    if start == 0:
        print("no data")
        return

    db = sqlite3.connect("file:" + index_db + "?mode=ro", uri=True)
    sessions = {}
    for agent, count in db.execute(
        "SELECT agent, count(DISTINCT session_id) FROM chunks WHERE ts >= ? GROUP BY agent",
        (start,),
    ):
        sessions[agent] = count
    db.close()

    asks, asking, unattributed = collections.Counter(), collections.defaultdict(set), 0
    for entry in entries:
        agent = entry.get("agent", "")
        asks[agent] += 1
        session_id = entry.get("session_id")
        if session_id:
            asking[agent].add(session_id)
        else:
            unattributed += 1

    for a in sorted(set(asks) | set(sessions)):
        print(f"agent={a} sessions={sessions.get(a, 0)} asking={len(asking[a])} asks={asks[a]}")
    print(f"unattributed={unattributed}")


if __name__ == "__main__":
    main()
