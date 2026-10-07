package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const apiBase = "https://discord.com/api/v9"

var httpClient = &http.Client{Timeout: 20 * time.Second}

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

type UserInfo struct {
	ID       string
	Username string
}

type RevokeResult struct {
	OK      bool
	Message string
	User    *UserInfo
}

func authRequest(method, url, token string, body any) (*http.Response, error) {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, url, r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json")
	return httpClient.Do(req)
}

func readBodyPreview(resp *http.Response, n int) string {
	if resp.Body == nil {
		return ""
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, int64(n)))
	return string(b)
}

func GetCurrentUser(token string) (*UserInfo, string) {
	resp, err := authRequest(http.MethodGet, apiBase+"/users/@me", token, nil)
	if err != nil {
		return nil, err.Error()
	}
	if resp.StatusCode == http.StatusUnauthorized {
		readBodyPreview(resp, 256)
		return nil, "token expired or invalid"
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := readBodyPreview(resp, 200)
		return nil, fmt.Sprintf("HTTP %d: %s", resp.StatusCode, msg)
	}
	defer resp.Body.Close()
	var data struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err.Error()
	}
	return &UserInfo{ID: data.ID, Username: data.Username}, ""
}

func logoutToken(token string) (bool, string) {
	resp, err := authRequest(http.MethodPost, apiBase+"/auth/logout", token, map[string]any{})
	if err != nil {
		return false, err.Error()
	}
	if resp.StatusCode == http.StatusUnauthorized {
		readBodyPreview(resp, 256)
		return true, "token already revoked"
	}
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
		readBodyPreview(resp, 256)
		return true, "logout successful"
	}
	msg := readBodyPreview(resp, 300)
	return false, fmt.Sprintf("HTTP %d: %s", resp.StatusCode, msg)
}

func listSessions(token string) ([]map[string]any, string) {
	resp, err := authRequest(http.MethodGet, apiBase+"/auth/sessions", token, nil)
	if err != nil {
		return nil, err.Error()
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := readBodyPreview(resp, 300)
		return nil, fmt.Sprintf("HTTP %d: %s", resp.StatusCode, msg)
	}
	defer resp.Body.Close()
	var body struct {
		UserSessions []map[string]any `json:"user_sessions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err.Error()
	}
	return body.UserSessions, ""
}

func logoutAllSessions(token string) (bool, string) {
	sessions, errMsg := listSessions(token)
	if errMsg != "" {
		return logoutToken(token)
	}
	var hashes []string
	for _, s := range sessions {
		if h, ok := s["id_hash"].(string); ok && h != "" {
			hashes = append(hashes, h)
		} else if h, ok := s["session_id_hash"].(string); ok && h != "" {
			hashes = append(hashes, h)
		}
	}
	if len(hashes) == 0 {
		return logoutToken(token)
	}
	for i := 0; i < len(hashes); i += 64 {
		end := i + 64
		if end > len(hashes) {
			end = len(hashes)
		}
		batch := hashes[i:end]
		resp, err := authRequest(http.MethodPost, apiBase+"/auth/sessions/logout", token, map[string]any{
			"session_id_hashes": batch,
		})
		if err != nil {
			return false, err.Error()
		}
		if resp.StatusCode == http.StatusUnauthorized {
			readBodyPreview(resp, 256)
			return false, "MFA required — log out per token or change your Discord password"
		}
		if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
			msg := readBodyPreview(resp, 300)
			return false, fmt.Sprintf("HTTP %d: %s", resp.StatusCode, msg)
		}
		readBodyPreview(resp, 256)
	}
	return true, fmt.Sprintf("logged out all sessions (%d)", len(hashes))
}

func RevokeToken(token string, allSessions bool) RevokeResult {
	user, uerr := GetCurrentUser(token)
	if user == nil {
		return RevokeResult{OK: false, Message: uerr, User: nil}
	}
	var ok bool
	var msg string
	if allSessions {
		ok, msg = logoutAllSessions(token)
	} else {
		ok, msg = logoutToken(token)
	}
	return RevokeResult{OK: ok, Message: msg, User: user}
}
