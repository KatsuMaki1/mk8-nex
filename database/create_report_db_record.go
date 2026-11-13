package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/PretendoNetwork/nex-go/v2/types"
)

// CreateReportDBRecord stores a raw SecureConnection.SendReport payload in ./reports.
func CreateReportDBRecord(pid types.PID, reportID types.UInt32, reportData types.QBuffer) error {
	reportsDir := "reports"
	if err := os.MkdirAll(reportsDir, os.ModePerm); err != nil {
		fmt.Printf("Failed to create the reports directory: %v\n", err)
		return err
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("report_pid%d_id%d_%s.bin", pid, reportID, timestamp)
	path := filepath.Join(reportsDir, filename)

	if err := os.WriteFile(path, []byte(reportData), 0644); err != nil {
		fmt.Printf("Failed to save the report: %v\n", err)
		return err
	}

	fmt.Printf("[INFO] Saved console report: %s\n", path)
	return nil
}
