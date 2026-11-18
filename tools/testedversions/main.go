// Command testedversions renders the Kubernetes versions covered by the CI test
// matrix into the README's "Tested Kubernetes Versions" section, between the
// TESTED_VERSIONS_START and TESTED_VERSIONS_END markers.
//
// Versions come from .github/tested-k8s-versions.json, so the section is fully
// determined by committed files. Run from the repository root to rewrite it:
//
//	go run ./tools/testedversions
//
// With -check it writes nothing and exits non-zero when the section is stale,
// reporting how to regenerate it. CI uses -check so a stale README fails the
// build rather than being committed by a bot.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	startMarker = "<!-- TESTED_VERSIONS_START -->"
	endMarker   = "<!-- TESTED_VERSIONS_END -->"
)

func main() {
	readmePath := flag.String("readme", "README.md", "path to the README file")
	versionsPath := flag.String("versions", ".github/tested-k8s-versions.json", "path to the JSON list of tested k8s versions")
	check := flag.Bool("check", false, "verify the README section is up to date instead of rewriting it")
	flag.Parse()

	if err := run(*readmePath, *versionsPath, *check); err != nil {
		fmt.Fprintln(os.Stderr, "testedversions:", err)
		os.Exit(1)
	}
}

func run(readmePath, versionsPath string, check bool) error {
	versions, err := readVersions(versionsPath)
	if err != nil {
		return err
	}

	readme, err := os.ReadFile(readmePath)
	if err != nil {
		return err
	}

	updated, err := replaceSection(string(readme), renderBlock(versions))
	if err != nil {
		return err
	}

	if check {
		if updated != string(readme) {
			return fmt.Errorf(
				"%s is out of date in the tested-versions section\n"+
					"regenerate it by running from the repository root:\n\n"+
					"    go run ./tools/testedversions\n\n"+
					"then commit the updated %s",
				readmePath, readmePath)
		}
		return nil
	}

	if updated == string(readme) {
		return nil
	}
	return os.WriteFile(readmePath, []byte(updated), 0o644)
}

// readVersions returns the k3s version tags sorted by Kubernetes version ascending.
func readVersions(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var versions []string
	if err := json.Unmarshal(data, &versions); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	if len(versions) == 0 {
		return nil, fmt.Errorf("%s lists no versions", path)
	}
	sort.Slice(versions, func(i, j int) bool {
		return lessVersion(versions[i], versions[j])
	})
	return versions, nil
}

// renderBlock returns the Markdown between the markers: one bullet per version.
func renderBlock(versions []string) string {
	lines := make([]string, len(versions))
	for i, v := range versions {
		lines[i] = "- " + displayVersion(v)
	}
	return strings.Join(lines, "\n")
}

// displayVersion drops the k3s build suffix, e.g. "v1.31.6-k3s1" -> "v1.31.6".
func displayVersion(v string) string {
	if i := strings.IndexByte(v, '-'); i >= 0 {
		return v[:i]
	}
	return v
}

var sectionRe = regexp.MustCompile("(?s)" + regexp.QuoteMeta(startMarker) + ".*?" + regexp.QuoteMeta(endMarker))

// replaceSection swaps the content between the markers for block, leaving the
// rest of the README untouched. It errors when the markers are absent.
func replaceSection(readme, block string) (string, error) {
	if !sectionRe.MatchString(readme) {
		return "", fmt.Errorf("could not find %q .. %q markers in README", startMarker, endMarker)
	}
	replacement := startMarker + "\n" + block + "\n" + endMarker
	return sectionRe.ReplaceAllStringFunc(readme, func(string) string { return replacement }), nil
}

// lessVersion orders two k3s tags by Kubernetes major.minor.patch, breaking ties
// by string compare so the order is stable.
func lessVersion(a, b string) bool {
	pa, pb := parseVersion(a), parseVersion(b)
	for i := range pa {
		if pa[i] != pb[i] {
			return pa[i] < pb[i]
		}
	}
	return a < b
}

func parseVersion(v string) [3]int {
	s := strings.TrimPrefix(displayVersion(v), "v")
	var out [3]int
	for i, part := range strings.SplitN(s, ".", 3) {
		n, _ := strconv.Atoi(part)
		out[i] = n
	}
	return out
}
