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
	yes := flag.Bool("y", false, "skip confirmation before logout")
	flag.BoolVar(yes, "yes", false, "skip confirmation before logout")
	flag.Parse()

	exitCode := run(*scanOnly, *allSessions, *yes)
	waitBeforeExit()
	os.Exit(exitCode)
}

func run(scanOnly, allSessions, yes bool) int {
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

	valid := make(map[string][]string)
	stale := 0
	for token, sources := range tokenMap {
		user, _ := discord.GetCurrentUser(token)
		if user != nil {
			valid[token] = sources
		} else {
			stale++
		}
	}

	fmt.Printf("\nToken-like strings: %d | Valid (active): %d | Expired or invalid: %d\n\n", len(tokenMap), len(valid), stale)

	for token, sources := range valid {
		user, _ := discord.GetCurrentUser(token)
		who := "?"
		if user != nil {
			who = fmt.Sprintf("%s (%s)", user.Username, user.ID)
		}
		masked := maskToken(token)
		fmt.Printf("  • %s\n", who)
		fmt.Printf("    Source: %s\n", strings.Join(sources, ", "))
		fmt.Printf("    Token: %s\n", masked)
	}

	if scanOnly {
		return 0
	}

	if len(valid) == 0 {
		fmt.Println("\nNo active tokens to log out.")
		return 0
	}

	if !yes {
		fmt.Print("\nThis will log out every token listed above (signed-in apps/browsers will be disconnected).\nType yes and press Enter to continue: ")
		line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		line = strings.TrimSpace(strings.ToLower(line))
		if line != "yes" && line != "y" {
			fmt.Println("Cancelled.")
			return 0
		}
	}

	fmt.Println()
	okCount := 0
	for token := range valid {
		result := discord.RevokeToken(token, allSessions)
		name := "?"
		if result.User != nil {
			name = result.User.Username
		}
		status := "FAIL"
		if result.OK {
			status = "OK"
			okCount++
		}
		fmt.Printf("[%s] %s - %s\n", status, name, result.Message)
	}

	fmt.Printf("\nDone: %d/%d succeeded.\n", okCount, len(valid))
	fmt.Println("To revoke every session on every device: Discord → Settings → Change Password.")

	if okCount == len(valid) {
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
