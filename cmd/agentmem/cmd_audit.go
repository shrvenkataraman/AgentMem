package main

import (
	"fmt"
	"math"
	"os"
	"regexp"

	"agentmem/internal/redact"
)

// suspectRun matches a high-entropy-looking run: 40+ chars of the base64-ish alphabet.
var suspectRun = regexp.MustCompile(`[A-Za-z0-9+/_=-]{40,}`)

// maxAuditFindings caps how many finding lines are printed; the totals always count everything.
const maxAuditFindings = 50

// cmdAudit re-scans every stored chunk for secrets that survived redaction. It never prints a
// secret: leaks are reported by id only, suspects by id plus a masked sample of the run.
func cmdAudit(args []string) int {
	db, err := openDB()
	if err != nil {
		fmt.Fprintln(os.Stderr, "agentmem audit:", err)
		return 1
	}
	defer db.Close()
	rows, err := db.Query("SELECT id, text FROM chunks ORDER BY id")
	if err != nil {
		fmt.Fprintln(os.Stderr, "agentmem audit:", err)
		return 1
	}
	defer rows.Close()
	var findings []string
	var leaks, suspects int
	for rows.Next() {
		var id int64
		var text string
		if err := rows.Scan(&id, &text); err != nil {
			fmt.Fprintln(os.Stderr, "agentmem audit:", err)
			return 1
		}
		if redact.Text(text) != text {
			leaks++
			findings = append(findings, fmt.Sprintf("leak %d", id))
		}
		if run := suspiciousRun(text); run != "" {
			suspects++
			findings = append(findings, fmt.Sprintf("suspect %d %s…%s", id, run[:4], run[len(run)-2:]))
		}
	}
	if err := rows.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "agentmem audit:", err)
		return 1
	}
	fmt.Printf("leaks=%d suspects=%d\n", leaks, suspects)
	for i, f := range findings {
		if i >= maxAuditFindings {
			break
		}
		fmt.Println(f)
	}
	if leaks > 0 {
		return 1
	}
	return 0
}

// suspiciousRun returns the first run in text that looks like a secret (>= 40 chars of the
// base64-ish alphabet and Shannon entropy >= 4.0 bits/char), or "" when there is none.
func suspiciousRun(text string) string {
	for _, run := range suspectRun.FindAllString(text, -1) {
		if entropyBits(run) >= 4.0 {
			return run
		}
	}
	return ""
}

// entropyBits returns the Shannon entropy of s in bits per byte.
func entropyBits(s string) float64 {
	var counts [256]int
	for i := 0; i < len(s); i++ {
		counts[s[i]]++
	}
	n := float64(len(s))
	var h float64
	for _, c := range counts {
		if c == 0 {
			continue
		}
		p := float64(c) / n
		h -= p * math.Log2(p)
	}
	return h
}
