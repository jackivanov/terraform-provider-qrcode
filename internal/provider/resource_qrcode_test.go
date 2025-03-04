package provider

import (
	"math/rand"
	"os"
	"path/filepath"
	"time"
	"testing"
  "fmt"
  "crypto/sha256"
  "encoding/hex"
  "io"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
  "github.com/hashicorp/terraform-plugin-testing/terraform"
)

// randomTempFileName generates a random temporary file name.
func randomTempFileName() string {
	rand.Seed(time.Now().UnixNano())
	tmpDir := os.TempDir()
	return filepath.Join(tmpDir, fmt.Sprintf("tmp-%d", rand.Uint64()))
}

// calculateSHA256 computes the SHA-256 checksum of a file.
func calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// TestAccQRCodeResource verifies the qrcode_generate resource
func TestAccQRCodeResource(t *testing.T) {
	filePath := randomTempFileName()
  expectedChecksum := "21489894b9e5f457473da5025741a7ce935c14d4a6ca9e29a72eec324c5fd743"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "qrcode" {}

					resource "qrcode_generate" "test" {
						text = "qrcode"
						file = "` + filePath + `"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the file is generated
					func(s *terraform.State) error {
						if _, err := os.Stat(filePath); os.IsNotExist(err) {
							return fmt.Errorf("file %s does not exist", filePath)
						}
						return nil
					},

					// Verify the SHA-256 checksum is as expected
					func(s *terraform.State) error {
						actualChecksum, err := calculateSHA256(filePath)
						if err != nil {
							return fmt.Errorf("failed to calculate SHA-256 checksum: %s", err)
						}
						if actualChecksum != expectedChecksum {
							return fmt.Errorf("expected SHA-256 checksum %s, got %s", expectedChecksum, actualChecksum)
						}
						return nil
					},

					// Verify the SHA-256 checksum is as expected
					resource.TestCheckResourceAttr(
						"qrcode_generate.test", "sha256",
						expectedChecksum,
					),
				),
			},
		},
	})

	// Cleanup the test file
	_ = os.Remove(filePath)
}
