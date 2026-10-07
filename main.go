package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"resettoken/internal/console"
	"resettoken/internal/discord"
	"resettoken/internal/paths"
	"resettoken/internal/scanner"
)

func main() {
	console.EnableUTF8()

	scanOnly := flag.Bool("scan-only", false, "list valid tokens without logging out")
	allSessions := flag.Bool("all-sessions", false, "log out all sessions per account (MFA may be required)")
	flag.Parse()

	exitCode := run(*scanOnly, *allSessions)
	waitBeforeExit()
	os.Exit(exitCode)
}

func run(scanOnly, allSessions bool) int {
	targets := paths.AllScanTargets()
	if len(targets) == 0 {
		fmt.Println("No Discord Desktop or browser storage folders found to scan.")
		return 1
	}

	scanTargets := make([]scanner.ScanTarget, len(targets))
	for i, t := range targets {
		scanTargets[i] = scanner.ScanTarget{Label: t.Label, Path: t.Path}
	}

	fmt.Printf("Scanning %d locations (Discord Desktop + web browsers)...\n", len(targets))
	bySource := scanner.ScanAll(scanTargets)
	tokenMap := scanner.MergeUnique(bySource)

	if len(tokenMap) == 0 {
		fmt.Println("No tokens found in Local Storage / Session Storage.")
		fmt.Println("Note: newer Discord builds may encrypt tokens — log out in the app or change your password at discord.com.")
		return 0
	}

	stale := 0
	okCount := 0
	validCount := 0

	fmt.Printf("\nToken-like strings: %d\n\n", len(tokenMap))

	for token, sources := range tokenMap {
		user, _ := discord.GetCurrentUser(token)
		if user == nil {
			stale++
			continue
		}
		validCount++
		masked := maskToken(token)
		fmt.Printf("  • %s (%s)\n", user.Username, user.ID)
		fmt.Printf("    Source: %s\n", strings.Join(sources, ", "))
		fmt.Printf("    Token: %s\n", masked)

		if scanOnly {
			continue
		}

		result := discord.RevokeToken(token, allSessions)
		status := "FAIL"
		if result.OK {
			status = "OK"
			okCount++
		}
		fmt.Printf("    [%s] %s\n", status, result.Message)
	}

	fmt.Printf("\nValid (active): %d | Expired or invalid: %d\n", validCount, stale)

	if scanOnly {
		return 0
	}

	if validCount == 0 {
		fmt.Println("No active tokens to log out.")
		return 0
	}

	fmt.Printf("\nDone: %d/%d logged out.\n", okCount, validCount)
	fmt.Println("To revoke every session on every device: Discord → Settings → Change Password.")

	if okCount == validCount {
		return 0
	}
	return 2
}

func maskToken(token string) string {
	if len(token) <= 12 {
		return token
	}
	return token[:8] + "..." + token[len(token)-4:]
}

func waitBeforeExit() {
	fmt.Print("\nPress Enter to exit...")
	_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
}
