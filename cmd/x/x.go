package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func main() {
	// --- Configuration ---
	// The base directory where the 'data' symlink will live.
	baseDir := "/tmp/webhook-certs"
	// The full service name for the certificate.
	host := "sriov-network-operator-webhook-service.openshift-sriov-network-operator.svc"
	// The name of the final symlink.
	symlinkName := "data"
	log.Printf("Starting atomic certificate update in: %s", baseDir)
	// Ensure the base directory exists.
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		log.Fatalf("❌ Failed to create base directory: %v", err)
	}
	// Read the current target of the 'data' symlink to know what to clean up later.
	finalLinkPath := filepath.Join(baseDir, symlinkName)
	oldTimestampDir, err := os.Readlink(finalLinkPath)
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("❌ Failed to read existing symlink: %v", err)
	}
	// --- Step 1: Create a new timestamped directory ---
	timestamp := time.Now().Format("2006_01_02_15_04_05.000000000")
	newDir := filepath.Join(baseDir, timestamp)
	if err := os.MkdirAll(newDir, 0755); err != nil {
		log.Fatalf("❌ Failed to create new timestamp directory: %v", err)
	}
	log.Printf("Step 1: Created new directory: %s", newDir)
	// --- Step 2: Run mkcert inside the new directory ---
	log.Printf("Step 2: Running mkcert for host: %s", host)
	cmd := exec.Command("mkcert", host)
	cmd.Dir = newDir // IMPORTANT: This makes mkcert run in our new directory.
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("❌ mkcert command failed: %v\nOutput:\n%s", err, string(output))
	}
	// --- Step 3: Rename the generated files to a consistent name ---
	generatedCertFile := filepath.Join(newDir, host+".pem")
	finalCertFile := filepath.Join(newDir, "tls.crt")
	if err := os.Rename(generatedCertFile, finalCertFile); err != nil {
		log.Fatalf("❌ Failed to rename cert file: %v", err)
	}
	generatedKeyFile := filepath.Join(newDir, host+"-key.pem")
	finalKeyFile := filepath.Join(newDir, "tls.key")
	if err := os.Rename(generatedKeyFile, finalKeyFile); err != nil {
		log.Fatalf("❌ Failed to rename key file: %v", err)
	}
	log.Println("Step 3: Renamed generated files to tls.crt and tls.key.")
	// --- Step 4: Perform the atomic symlink swap ⚛️ ---
	tempLinkPath := filepath.Join(baseDir, symlinkName+"_tmp")
	// Create a temporary symlink. Target path is relative for portability.
	if err := os.Symlink(timestamp, tempLinkPath); err != nil {
		log.Fatalf("❌ Failed to create temporary symlink: %v", err)
	}
	// Atomically rename the temp link to the final link.
	if err := os.Rename(tempLinkPath, finalLinkPath); err != nil {
		log.Fatalf("❌ Failed to perform atomic rename: %v", err)
	}
	log.Printf("Step 4: Atomically pointed '%s' symlink to new directory.", symlinkName)
	// --- Step 5: Clean up the old directory ---
	if oldTimestampDir != "" {
		oldDirPath := filepath.Join(baseDir, oldTimestampDir)
		log.Printf("Step 5: Cleaning up old directory: %s", oldDirPath)
		if err := os.RemoveAll(oldDirPath); err != nil {
			log.Printf("⚠️ Warning: Failed to clean up old directory: %v", err)
		}
	}
	log.Println("✅ Successfully completed atomic certificate update.")
}
