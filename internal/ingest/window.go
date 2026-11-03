package ingest

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"unicode/utf8"

	"agentmem/internal/model"
)

const maxWindow = 2560

// Window splits msgs into chunks (windows of about 2560 bytes at paragraph boundaries) and caps
// tool_result text. ContentHash is hex sha256 of the chunk text.
func Window(msgs []model.Message) []model.Chunk {
	var out []model.Chunk
	for _, m := range msgs {
		if m.Text == "" {
			continue
		}
		base := len(out) // WindowIdx counts from 0 per message.
		window := ""
		has := false
		for _, para := range strings.Split(m.Text, "\n\n") {
			if para == "" {
				continue
			}
			if len(para) > maxWindow {
				if has {
					out = append(out, chunk(m, window, len(out)-base))
					has = false
				}
				for _, piece := range cut(para, maxWindow) {
					out = append(out, chunk(m, piece, len(out)-base))
				}
				continue
			}
			if has && len(window)+2+len(para) <= maxWindow {
				window += "\n\n" + para
				continue
			}
			if has {
				out = append(out, chunk(m, window, len(out)-base))
			}
			window = para
			has = true
		}
		if has {
			out = append(out, chunk(m, window, len(out)-base))
		}
	}
	return out
}

// chunk builds a stored row from message m whose text is the window text.
func chunk(m model.Message, text string, idx int) model.Chunk {
	c := model.Chunk{Message: m}
	c.Text = text
	c.WindowIdx = idx
	sum := sha256.Sum256([]byte(text))
	c.ContentHash = hex.EncodeToString(sum[:])
	return c
}

// cut splits an over-long paragraph into pieces of at most max bytes, backing each cut off to a
// rune boundary.
func cut(s string, max int) []string {
	var pieces []string
	for len(s) > max {
		n := max
		for n > 0 && !utf8.RuneStart(s[n]) {
			n--
		}
		if n == 0 {
			// A single rune longer than max cannot happen for valid UTF-8 (max >= 4); fall back
			// to the first rune boundary after the start.
			n = 1
			for n < len(s) && !utf8.RuneStart(s[n]) {
				n++
			}
		}
		pieces = append(pieces, s[:n])
		s = s[n:]
	}
	if len(s) > 0 {
		pieces = append(pieces, s)
	}
	return pieces
}
