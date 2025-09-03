package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/opd-ai/desktop-companion/internal/artifact"
)

const (
	defaultArtifactsDir = "build/artifacts"
	version             = "1.0.0"
)

var (
	artifactsDir = flag.String("dir", defaultArtifactsDir, "Artifacts directory")
	verbose      = flag.Bool("verbose", false, "Enable verbose output")
	showVersion  = flag.Bool("version", false, "Show version information")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] COMMAND [ARGS...]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Artifact Management Tool for DDS Character Binaries\n\n")
		fmt.Fprintf(os.Stderr, "COMMANDS:\n")
		fmt.Fprintf(os.Stderr, "  store CHARACTER PLATFORM ARCH FILE   Store an artifact with metadata\n")
		fmt.Fprintf(os.Stderr, "  list [CHARACTER] [PLATFORM] [ARCH]   List stored artifacts\n")
		fmt.Fprintf(os.Stderr, "  stats                                Show artifact statistics\n")
		fmt.Fprintf(os.Stderr, "  cleanup POLICY                       Clean up expired artifacts\n")
		fmt.Fprintf(os.Stderr, "  compress POLICY                      Compress old artifacts\n")
		fmt.Fprintf(os.Stderr, "  policies                             List available retention policies\n")
		fmt.Fprintf(os.Stderr, "\nOPTIONS:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEXAMPLES:\n")
		fmt.Fprintf(os.Stderr, "  %s store default linux amd64 build/default_linux_amd64\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s list default\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s cleanup development\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s compress production\n", os.Args[0])
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("DDS Artifact Manager v%s\n", version)
		return
	}

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	// Initialize artifact manager
	manager, err := artifact.NewManager(*artifactsDir)
	if err != nil {
		log.Fatalf("Failed to initialize artifact manager: %v", err)
	}

	command := flag.Arg(0)
	switch command {
	case "store":
		handleStore(manager)
	case "list":
		handleList(manager)
	case "stats":
		handleStats(manager)
	case "cleanup":
		handleCleanup(manager)
	case "compress":
		handleCompress(manager)
	case "policies":
		handlePolicies(manager)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		flag.Usage()
		os.Exit(1)
	}
}

func handleStore(manager *artifact.Manager) {
	if flag.NArg() != 5 {
		fmt.Fprintf(os.Stderr, "Usage: store CHARACTER PLATFORM ARCH FILE\n")
		os.Exit(1)
	}

	character := flag.Arg(1)
	platform := flag.Arg(2)
	arch := flag.Arg(3)
	filePath := flag.Arg(4)

	// Validate file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Fatalf("Artifact file does not exist: %s", filePath)
	}

	// Create metadata
	metadata := map[string]string{
		"stored_by": "artifact-manager",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	info, err := manager.StoreArtifact(filePath, character, platform, arch, metadata)
	if err != nil {
		log.Fatalf("Failed to store artifact: %v", err)
	}

	fmt.Printf("✓ Stored artifact: %s\n", info.Name)
	if *verbose {
		fmt.Printf("  Character: %s\n", info.Character)
		fmt.Printf("  Platform: %s/%s\n", info.Platform, info.Architecture)
		fmt.Printf("  Size: %d bytes\n", info.Size)
		fmt.Printf("  Checksum: %s\n", info.Checksum)
		fmt.Printf("  Created: %s\n", info.CreatedAt.Format(time.RFC3339))
	}
}

func handleList(manager *artifact.Manager) {
	var character, platform, arch string

	// Parse optional arguments
	if flag.NArg() >= 2 {
		character = flag.Arg(1)
	}
	if flag.NArg() >= 3 {
		platform = flag.Arg(2)
	}
	if flag.NArg() >= 4 {
		arch = flag.Arg(3)
	}

	artifacts, err := manager.ListArtifacts(character, platform, arch)
	if err != nil {
		log.Fatalf("Failed to list artifacts: %v", err)
	}

	if len(artifacts) == 0 {
		fmt.Println("No artifacts found")
		return
	}

	fmt.Printf("Found %d artifacts:\n\n", len(artifacts))

	for _, info := range artifacts {
		fmt.Printf("• %s\n", info.Name)
		fmt.Printf("  Character: %s\n", info.Character)
		fmt.Printf("  Platform: %s/%s\n", info.Platform, info.Architecture)
		fmt.Printf("  Size: %s\n", formatSize(info.Size))
		fmt.Printf("  Created: %s\n", info.CreatedAt.Format("2006-01-02 15:04:05"))
		if info.Compressed {
			fmt.Printf("  Status: Compressed\n")
		}
		if *verbose && len(info.Metadata) > 0 {
			fmt.Printf("  Metadata:\n")
			for key, value := range info.Metadata {
				fmt.Printf("    %s: %s\n", key, value)
			}
		}
		fmt.Println()
	}
}

func handleStats(manager *artifact.Manager) {
	stats, err := manager.GetArtifactStats()
	if err != nil {
		log.Fatalf("Failed to get artifact statistics: %v", err)
	}

	fmt.Printf("Artifact Statistics:\n\n")
	fmt.Printf("Total Artifacts: %d\n", stats["total_artifacts"])
	fmt.Printf("Total Size: %s\n", formatSize(stats["total_size"].(int64)))
	fmt.Printf("Compressed: %d\n", stats["compressed"])

	characters := stats["characters"].(map[string]int)
	if len(characters) > 0 {
		fmt.Printf("\nBy Character:\n")
		for char, count := range characters {
			fmt.Printf("  %s: %d artifacts\n", char, count)
		}
	}

	platforms := stats["platforms"].(map[string]int)
	if len(platforms) > 0 {
		fmt.Printf("\nBy Platform:\n")
		for plat, count := range platforms {
			fmt.Printf("  %s: %d artifacts\n", plat, count)
		}
	}
}

func handleCleanup(manager *artifact.Manager) {
	if flag.NArg() != 2 {
		fmt.Fprintf(os.Stderr, "Usage: cleanup POLICY\n")
		os.Exit(1)
	}

	policy := flag.Arg(1)

	fmt.Printf("Cleaning up artifacts with policy: %s\n", policy)

	if err := manager.CleanupArtifacts(policy); err != nil {
		log.Fatalf("Failed to cleanup artifacts: %v", err)
	}

	fmt.Printf("✓ Cleanup completed\n")
}

func handleCompress(manager *artifact.Manager) {
	if flag.NArg() != 2 {
		fmt.Fprintf(os.Stderr, "Usage: compress POLICY\n")
		os.Exit(1)
	}

	policy := flag.Arg(1)

	fmt.Printf("Compressing old artifacts with policy: %s\n", policy)

	if err := manager.CompressOldArtifacts(policy); err != nil {
		log.Fatalf("Failed to compress artifacts: %v", err)
	}

	fmt.Printf("✓ Compression completed\n")
}

func handlePolicies(manager *artifact.Manager) {
	policies := artifact.DefaultRetentionPolicies()

	fmt.Printf("Available Retention Policies:\n\n")

	for name, policy := range policies {
		fmt.Printf("• %s\n", name)
		fmt.Printf("  Retention Period: %s\n", formatDuration(policy.RetentionPeriod))
		fmt.Printf("  Max Count: %s\n", formatMaxCount(policy.MaxCount))
		fmt.Printf("  Compress After: %s\n", formatDuration(policy.CompressAfter))
		fmt.Printf("  Cleanup Interval: %s\n", formatDuration(policy.CleanupInterval))
		fmt.Println()
	}
}

// formatSize formats a byte size into a human-readable string
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatDuration formats a duration into a human-readable string
func formatDuration(d time.Duration) string {
	if d < time.Hour {
		return d.String()
	}

	hours := int(d.Hours())
	if hours < 24 {
		return fmt.Sprintf("%d hours", hours)
	}

	days := hours / 24
	if days < 7 {
		return fmt.Sprintf("%d days", days)
	}

	weeks := days / 7
	if weeks < 52 {
		return fmt.Sprintf("%d weeks", weeks)
	}

	years := weeks / 52
	return fmt.Sprintf("%d years", years)
}

// formatMaxCount formats the max count setting
func formatMaxCount(count int) string {
	if count < 0 {
		return "unlimited"
	}
	return fmt.Sprintf("%d", count)
}
