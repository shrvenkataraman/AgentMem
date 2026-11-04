package redact

import (
	_ "embed"
	"encoding/json"
	"math"
	"regexp"
	"strings"
)

//go:embed gitleaks.json
var gitleaksJSON []byte

// glRule is one gitleaks rule as written by scripts/gitleaks_to_json.py.
type glRule struct {
	ID          string   `json:"id"`
	Regex       string   `json:"regex"`
	SecretGroup int      `json:"secretGroup"`
	Entropy     float64  `json:"entropy"`
	Keywords    []string `json:"keywords"`
	re          *regexp.Regexp
}

// glRules are the compiled gitleaks rules; a rule whose regex does not compile is skipped.
var glRules = loadGitleaks(gitleaksJSON)

func loadGitleaks(b []byte) []glRule {
	var rs []glRule
	if err := json.Unmarshal(b, &rs); err != nil {
		panic(err)
	}
	out := make([]glRule, 0, len(rs))
	for _, r := range rs {
		re, err := regexp.Compile(r.Regex)
		if err != nil {
			continue
		}
		r.re = re
		out = append(out, r)
	}
	return out
}

// shannon is the Shannon entropy in bits per byte of s (0 for "").
func shannon(s string) float64 {
	if s == "" {
		return 0
	}
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

// applyGitleaks masks the secret span of every gitleaks rule match.
func applyGitleaks(s string) string {
	low := strings.ToLower(s)
	for _, r := range glRules {
		if len(r.Keywords) > 0 {
			hit := false
			for _, k := range r.Keywords {
				if strings.Contains(low, k) {
					hit = true
					break
				}
			}
			if !hit {
				continue
			}
		}
		next := r.re.ReplaceAllStringFunc(s, func(m string) string {
			sub := r.re.FindStringSubmatchIndex(m)
			lo, hi := 0, len(m)
			if g := r.SecretGroup; g > 0 && 2*g+1 < len(sub) && sub[2*g] >= 0 {
				lo, hi = sub[2*g], sub[2*g+1]
			} else if len(sub) > 2 && sub[2] >= 0 {
				lo, hi = sub[2], sub[3]
			}
			secret := m[lo:hi]
			if r.Entropy > 0 && shannon(secret) < r.Entropy {
				return m
			}
			return m[:lo] + "[REDACTED:" + r.ID + "]" + m[hi:]
		})
		if next != s {
			s = next
			low = strings.ToLower(s)
		}
	}
	return s
}
