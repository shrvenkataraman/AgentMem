#!/usr/bin/env python3
"""Independent per-ticket proof for the agentmem-local bundle.

usage: verify_ticket.py NN [--out DIR]     one ticket
       verify_ticket.py all [--out DIR]    tickets 01..45

Every proof is recomputed from git objects and a fresh `go test`; nothing is read
from PROGRESS.md. Evidence for ticket NN is written to DIR/NN.txt. The last line of
stdout per ticket is `NN VERDICT=PASS|FAIL <failed proofs>`. Exit 1 if any FAIL.

Proofs (P1..P8 are hard gates, I1..I2 are informational and need a human):
  P1 one merge commit exists for the ticket and the ticket test is in its first-parent diff
  P2 scope: non-test files changed by the merge are all on the ticket's Edit line,
     none is a no-go path, count <= 3 files
  P3 size: added+deleted lines of non-test files <= 100 (160 allowed if a file is new)
  P4 RED: tree at merge^1 + the ticket test (+ harness) fails by assertion, not by build error
  P5 GREEN: tree at the merge passes every Test func of the ticket test
  P6 NOW: tree at HEAD still passes every Test func of the ticket test
  P7 symbols: every `func`/`type` named in the ticket's Interface blocks exists in HEAD non-test source
  P8 literals: every backticked literal in `## Done when` that is a number/string/flag appears
     in the ticket test (informational only: never fails; a human judges each miss)
  I1 the Check command in the ticket equals the canonical form
  I2 test at merge vs canonical copy at HEAD (differences must be explained by a planner commit)
"""
import os, re, subprocess, sys, tempfile, shutil, glob

REPO = subprocess.check_output(["git", "rev-parse", "--show-toplevel"], text=True).strip()
BUNDLE = os.path.join(REPO, ".scratch", "agentmem-local")
NOGO = ("HANDOFF.md", ".scratch/", "testdata/", "go.mod", "go.sum", "vendor/")
ENV = dict(os.environ, GOFLAGS="-mod=vendor", CGO_ENABLED="0")


def sh(cmd, cwd=REPO, env=None, check=False):
    p = subprocess.run(cmd, cwd=cwd, env=env or ENV, text=True, capture_output=True, shell=isinstance(cmd, str))
    if check and p.returncode:
        raise RuntimeError(f"{cmd}: {p.stderr}")
    return p


def git(*a, cwd=REPO):
    return sh(["git", *a], cwd=cwd)


def ticket_file(nn):
    fs = glob.glob(f"{BUNDLE}/issues/{nn}-*.md")
    return fs[0] if len(fs) == 1 else None


def field(txt, name):
    m = re.search(rf"^\*\*{name}:\*\*\s*(.*)$", txt, re.M)
    return m.group(1).strip() if m else ""


def section(txt, name):
    m = re.search(rf"^## {name}\n(.*?)(?=^## |\Z)", txt, re.M | re.S)
    return m.group(1) if m else ""


class Worktree:
    def __init__(self, rev):
        self.dir = tempfile.mkdtemp(prefix="vt-")
        r = git("worktree", "add", "--detach", self.dir, rev)
        if r.returncode:
            raise RuntimeError(r.stderr)

    def close(self):
        git("worktree", "remove", "--force", self.dir)
        shutil.rmtree(self.dir, ignore_errors=True)


def run_tests(rev, nn, test_rel, harness_from, test_from):
    """Check out rev, overlay the ticket test (and harness) from the given revs, run the ticket tests."""
    w = Worktree(rev)
    try:
        for rel, src in ((test_rel, test_from), ("e2e/harness_test.go", harness_from)):
            blob = git("show", f"{src}:{rel}")
            if blob.returncode == 0:
                with open(os.path.join(w.dir, rel), "w") as f:
                    f.write(blob.stdout)
        p = sh(["go", "test", "./e2e", "-run", f"^TestT{nn}_", "-count=1", "-v"], cwd=w.dir)
        out = p.stdout + p.stderr
        return p.returncode, out
    finally:
        w.close()


def verdicts(out):
    passed = set(re.findall(r"^--- PASS: (Test\w+)", out, re.M))
    failed = set(re.findall(r"^--- FAIL: (Test\w+)", out, re.M))
    return passed, failed


def verify(nn, outdir):
    ev = []
    fails = []

    def log(s=""):
        ev.append(s)

    def gate(name, ok, detail):
        log(f"[{'PASS' if ok else 'FAIL'}] {name}: {detail}")
        if not ok:
            fails.append(name)

    tf = ticket_file(nn)
    if not tf:
        gate("P0 ticket", False, f"issues/{nn}-*.md not found exactly once")
        return ev, fails
    txt = open(tf).read()
    edit = [e.strip() for e in field(txt, "Edit").split(",") if e.strip()]
    test_rel = field(txt, "Test")
    check = field(txt, "Check")
    gate("I1 check form", check == f"GOFLAGS=-mod=vendor go test ./e2e -run '^TestT{nn}_'", check)

    # P1
    ms = git("log", "--merges", f"--grep=ticket/{nn}:", "--format=%H").stdout.split()
    gate("P1 one merge", len(ms) == 1, f"{len(ms)} merge commit(s) {ms}")
    if len(ms) != 1:
        return ev, fails
    m = ms[0]
    p1 = git("rev-parse", f"{m}^1").stdout.strip()
    diff = git("diff", "--numstat", p1, m).stdout.splitlines()
    rows = [l.split("\t") for l in diff]
    files = [r[2] for r in rows]
    gate("P1 test in merge diff", test_rel in files, f"{test_rel} in first-parent diff of {m[:9]}")
    log("   merge diff:\n" + "\n".join("     " + "\t".join(r) for r in rows))

    # P2/P3: the merge brings in the ticket branch (red commit + impl); judge non-test files.
    impl = [r for r in rows if not r[2].startswith("e2e/")]
    names = [r[2] for r in impl]
    bad_nogo = [n for n in names if n.startswith(NOGO) or n in NOGO]
    off_list = [n for n in names if n not in edit]
    gate("P2 scope", not bad_nogo and not off_list and len(names) <= 3,
         f"files={names} edit={edit} nogo={bad_nogo} off-list={off_list}")
    added = sum(int(r[0]) for r in impl if r[0] != "-")
    deleted = sum(int(r[1]) for r in impl if r[1] != "-")
    new_file = any(git("cat-file", "-e", f"{p1}:{n}").returncode != 0 for n in names)
    cap = 160 if new_file else 100
    gate("P3 size", added + deleted <= cap,
         f"+{added}/-{deleted} = {added + deleted} (cap {cap}{', new-file allowance' if new_file else ''}"
         f"{'; OVER 100: check PROGRESS.md for an explicit acceptance' if added + deleted > 100 else ''})")

    # P4/P5/P6
    declared = set(re.findall(r"^func (Test\w+)\(", git("show", f"{m}:{test_rel}").stdout, re.M))
    rc, out = run_tests(p1, nn, test_rel, m, m)
    pa, fa = verdicts(out)
    build_broken = "[build failed]" in out or "build failed" in out
    gate("P4 RED at merge^1", rc != 0 and bool(fa) and not build_broken,
         f"rc={rc} failing={sorted(fa)} passing={sorted(pa)} build_failed={build_broken}")
    log("   red tail:\n" + "\n".join("     " + l for l in out.strip().splitlines()[-12:]))
    rc, out = run_tests(m, nn, test_rel, m, m)
    pa, fa = verdicts(out)
    gate("P5 GREEN at merge", rc == 0 and declared and pa == declared and not fa,
         f"rc={rc} declared={len(declared)} passed={len(pa)} failed={sorted(fa)}")
    rc, out = run_tests("HEAD", nn, test_rel, "HEAD", "HEAD")
    pa, fa = verdicts(out)
    gate("P6 GREEN at HEAD", rc == 0 and declared and pa >= declared and not fa,
         f"rc={rc} passed={len(pa)}/{len(declared)} failed={sorted(fa)}")

    # P7
    iface = section(txt, "Interface")
    code = "\n".join(re.findall(r"```go\n(.*?)```", iface, re.S))
    syms = set(re.findall(r"^\s*func (?:\([^)]*\) )?([A-Za-z_]\w*)\(", code, re.M))
    syms |= set(re.findall(r"^\s*type ([A-Za-z_]\w*)\b", code, re.M))
    src = ""
    for root, _, fs in os.walk(REPO):
        if any(x in root for x in ("/vendor", "/.git", "/.scratch")):
            continue
        for f in fs:
            if f.endswith(".go") and not f.endswith("_test.go"):
                src += open(os.path.join(root, f)).read()
    missing = sorted(s for s in syms if not re.search(rf"\b(func (\([^)]*\) )?|type ){s}\b", src))
    gate("P7 interface symbols", not missing, f"checked {sorted(syms)} missing={missing}")

    # P8
    done = section(txt, "Done when")
    lits = [l for l in re.findall(r"`([^`]+)`", done) if not l.startswith("go test") and len(l) <= 60]
    testsrc = git("show", f"{m}:{test_rel}").stdout
    unpinned = [l for l in lits if l not in testsrc]
    log(f"[INFO] P8 done-when literals: {len(lits) - len(unpinned)}/{len(lits)} appear verbatim in the test; "
        f"not found (human must judge, CLI args and prose are expected misses): {unpinned}")

    # I2
    canon = glob.glob(f"{BUNDLE}/tests/t{nn}_*_test.go")
    cur = git("show", f"HEAD:{test_rel}").stdout
    same_head = bool(canon) and open(canon[0]).read() == cur
    same_merge = git("show", f"{m}:{test_rel}").stdout == cur
    log(f"[INFO] I2 canonical==HEAD tracked: {same_head}; merge-time test==HEAD test: {same_merge}"
        f"{'' if same_merge else '  (explain via: git log --oneline -- ' + test_rel + ')'}")
    return ev, fails


def main():
    a = sys.argv[1:]
    if not a:
        print(__doc__)
        sys.exit(2)
    outdir = a[a.index("--out") + 1] if "--out" in a else os.path.join(tempfile.gettempdir(), "verify-evidence")
    os.makedirs(outdir, exist_ok=True)
    nns = [f"{i:02d}" for i in range(1, 46)] if a[0] == "all" else [a[0].zfill(2)]
    bad = 0
    for nn in nns:
        ev, fails = verify(nn, outdir)
        open(os.path.join(outdir, f"{nn}.txt"), "w").write("\n".join(ev) + "\n")
        v = "FAIL " + ",".join(fails) if fails else "PASS"
        bad += bool(fails)
        print(f"{nn} VERDICT={v}", flush=True)
    print(f"SUMMARY tickets={len(nns)} failed={bad} evidence={outdir}")
    sys.exit(1 if bad else 0)


main()
