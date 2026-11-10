package cli

import (
	"bytes"
	"runtime/debug"
	"strings"
	"testing"
)

func TestFillVersionFromBuildInfo(t *testing.T) {
	info := &debug.BuildInfo{}
	info.Main.Version = "v1.2.3"
	info.Settings = []debug.BuildSetting{
		{Key: "vcs.revision", Value: "abcdef1234567890"},
		{Key: "vcs.time", Value: "2026-11-10T02:00:00Z"},
	}

	version, commit, date := fillVersionFromBuildInfo("dev", "unknown", "unknown", info)
	if version != "v1.2.3" || commit != "abcdef1" || date != "2026-11-10T02:00:00Z" {
		t.Fatalf("got %s %s %s", version, commit, date)
	}

	version, commit, date = fillVersionFromBuildInfo("v9.9.9", "deadbee", "2020-01-01T00:00:00Z", info)
	if version != "v9.9.9" || commit != "deadbee" || date != "2020-01-01T00:00:00Z" {
		t.Fatalf("ldflags were overwritten: %s %s %s", version, commit, date)
	}

	info.Main.Version = "(devel)"
	version, _, _ = fillVersionFromBuildInfo("dev", "abc", "2020-01-01T00:00:00Z", info)
	if version != "dev" {
		t.Fatalf("devel build changed version to %s", version)
	}
}

func TestFormatVersion(t *testing.T) {
	got := FormatVersion("v1.2.3", "abc1234", "2026-11-10T00:00:00Z")
	want := "v1.2.3 (commit: abc1234, built: 2026-11-10T00:00:00Z)"
	if got != want {
		t.Fatalf("FormatVersion() = %q, want %q", got, want)
	}

	if got := FormatVersion("v1.2.3", "unknown", "unknown"); got != "v1.2.3" {
		t.Fatalf("FormatVersion() = %q, want version only", got)
	}
}

func TestVersionCommand(t *testing.T) {
	origVersion := versionInfo
	origCommit := commitInfo
	origDate := dateInfo
	origRootVersion := rootCmd.Version
	t.Cleanup(func() {
		versionInfo = origVersion
		commitInfo = origCommit
		dateInfo = origDate
		rootCmd.Version = origRootVersion
	})

	SetVersionInfo("v1.2.3", "abc1234", "2026-11-10T00:00:00Z")

	var buf bytes.Buffer
	cmd := newVersionCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "v1.2.3 (commit: abc1234, built: 2026-11-10T00:00:00Z)") {
		t.Fatalf("version command printed %q", buf.String())
	}
}
