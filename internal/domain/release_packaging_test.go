package domain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test X36-A–H: Native Release Packaging, Authenticity & Installation Lifecycle
func TestReleasePackaging_X36_InstallationLifecycle(t *testing.T) {
	tempDir := t.TempDir()
	archivePath1 := filepath.Join(tempDir, "rewrotron-0.1.0-linux-amd64.tar.gz")
	archivePath2 := filepath.Join(tempDir, "rewrotron-0.1.0-linux-amd64-copy.tar.gz")

	binRewrotron := filepath.Join(tempDir, "rewrotron")
	_ = os.WriteFile(binRewrotron, []byte("#!/bin/sh\necho rewrotron\n"), 0755)

	binMCP := filepath.Join(tempDir, "rewrotron-mcp")
	_ = os.WriteFile(binMCP, []byte("#!/bin/sh\necho rewrotron-mcp\n"), 0755)

	cfg := ReleasePackagingConfig{
		Version:      "0.1.0",
		Architecture: "amd64",
		OS:           "linux",
		CommitHash:   "abc123def456",
		BinRewrotron: binRewrotron,
		BinMCP:       binMCP,
	}

	// 1. Packaging & Reproducibility (X36-H)
	manifest1, err := BuildReleaseArchive(cfg, archivePath1)
	if err != nil {
		t.Fatalf("failed to build archive 1: %v", err)
	}

	manifest2, err := BuildReleaseArchive(cfg, archivePath2)
	if err != nil {
		t.Fatalf("failed to build archive 2: %v", err)
	}

	if manifest1.ArchiveSHA256 != manifest2.ArchiveSHA256 {
		t.Fatalf("reproducibility failure (X36-H): builds produced different checksums: %s vs %s", manifest1.ArchiveSHA256, manifest2.ArchiveSHA256)
	}

	// 2. Clean install into user-local layout (X36-A, X36-D)
	homeDir := filepath.Join(tempDir, "home_user")
	layout := DefaultInstallationLayout(homeDir, "0.1.0")

	err = ExtractAndInstall(archivePath1, manifest1.ArchiveSHA256, layout, false)
	if err != nil {
		t.Fatalf("clean installation failed (X36-A): %v", err)
	}

	// Verify managed current link resolves to releases/0.1.0
	targetRelease, err := os.Readlink(layout.CurrentLinkPath)
	if err != nil || targetRelease != layout.ReleasesDir {
		t.Fatalf("current link target mismatch: got %s, expected %s", targetRelease, layout.ReleasesDir)
	}

	// Verify entrypoint symlinks exist and point to current
	entryTarget, err := os.Readlink(layout.RewrotronBin)
	if err != nil || !strings.Contains(entryTarget, "current/bin/rewrotron") {
		t.Fatalf("entrypoint symlink target mismatch: got %s", entryTarget)
	}

	// 3. Falsification: Authenticity / checksum mismatch refusal (X36-B)
	err = ExtractAndInstall(archivePath1, "sha256:0000000000000000000000000000000000000000000000000000000000000000", layout, false)
	if err == nil || !strings.Contains(err.Error(), "ERR_AUTHENTICITY_MISMATCH") {
		t.Fatalf("expected ERR_AUTHENTICITY_MISMATCH on bad digest, got: %v", err)
	}

	// 4. Interrupted install preserves existing installation (X36-C)
	v2Layout := DefaultInstallationLayout(homeDir, "0.2.0")
	cfg2 := cfg
	cfg2.Version = "0.2.0"
	archivePathV2 := filepath.Join(tempDir, "rewrotron-0.2.0-linux-amd64.tar.gz")
	manifestV2, err := BuildReleaseArchive(cfg2, archivePathV2)
	if err != nil {
		t.Fatal(err)
	}

	// Simulate interruption during v2 install
	err = ExtractAndInstall(archivePathV2, manifestV2.ArchiveSHA256, v2Layout, true)
	if err == nil || !strings.Contains(err.Error(), "ERR_INTERRUPTED_INSTALLATION") {
		t.Fatalf("expected ERR_INTERRUPTED_INSTALLATION on simulated interruption, got: %v", err)
	}

	// Current link must still point to 0.1.0
	currTarget, err := os.Readlink(layout.CurrentLinkPath)
	if err != nil || currTarget != layout.ReleasesDir {
		t.Fatalf("interruption corrupted current pointer: got %s, expected %s", currTarget, layout.ReleasesDir)
	}

	// 5. Complete upgrade to 0.2.0 and rollback to 0.1.0 (X36-E)
	err = ExtractAndInstall(archivePathV2, manifestV2.ArchiveSHA256, v2Layout, false)
	if err != nil {
		t.Fatalf("upgrade to 0.2.0 failed: %v", err)
	}
	upgradedTarget, _ := os.Readlink(layout.CurrentLinkPath)
	if upgradedTarget != v2Layout.ReleasesDir {
		t.Fatalf("upgrade failed to update current link: got %s", upgradedTarget)
	}

	// Rollback current to 0.1.0
	err = RollbackInstallation(layout, "0.1.0")
	if err != nil {
		t.Fatalf("rollback to 0.1.0 failed (X36-E): %v", err)
	}
	rolledBackTarget, _ := os.Readlink(layout.CurrentLinkPath)
	if rolledBackTarget != layout.ReleasesDir {
		t.Fatalf("rollback failed to restore current link: got %s", rolledBackTarget)
	}

	// 6. Preserving uninstall (X36-F)
	// Create customer workspace mock
	customerWorkspace := filepath.Join(homeDir, "my-customer-project")
	_ = os.MkdirAll(customerWorkspace, 0755)
	_ = os.WriteFile(filepath.Join(customerWorkspace, "rewrotron.yaml"), []byte("project: test\n"), 0644)

	err = PreservingUninstall(layout, "0.2.0", false)
	if err != nil {
		t.Fatalf("uninstall 0.2.0 failed: %v", err)
	}
	// Verify 0.2.0 release dir is removed
	if _, err := os.Stat(v2Layout.ReleasesDir); !os.IsNotExist(err) {
		t.Fatalf("uninstalled release directory still exists")
	}
	// Verify customer workspace and files are preserved
	if _, err := os.Stat(filepath.Join(customerWorkspace, "rewrotron.yaml")); err != nil {
		t.Fatalf("customer workspace was deleted during uninstall (X36-F violation)")
	}
}
