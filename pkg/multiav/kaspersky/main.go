// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package kaspersky

import (
	"strings"

	"github.com/saferwall/saferwall/pkg/utils"
)

// Our consts
const (
	kav4fs = "/opt/kaspersky/kav4fs/bin/kav4fs-control"
)

// Result represents detection results
type Result struct {
	Infected bool   `json:"infected"`
	Output   string `json:"output"`
}

// Version represents database components' versions
type Version struct {
	CurrentAVDatabasesDate    string `json:"current_av_db_ate"`
	LastAVDatabasesUpdateDate string `json:"last_av_db_update_date"`
	CurrentAVDatabasesState   string `json:"current_av_db_state"`
	CurrentAVDatabasesRecords string `json:"current_av_db_records"`
}

// GetProgramVersion returns Kaspersky Anti-Virus for Linux File Server version
func GetProgramVersion() (string, error) {

	// Run kav4s to grab the version
	out, err := utils.ExecCommand(kav4fs, "-S", "--app-info")
	if err != nil {
		return "", err
	}

	version := ""
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Version:") {
			version = strings.TrimSpace(strings.TrimPrefix(line, "Version:"))
			break
		}
	}

	return version, nil
}

// GetDatabaseVersion returns AV database update version
func GetDatabaseVersion() (Version, error) {

	// Run kav4s to grab the database update version
	databaseOut, err := utils.ExecCommand(kav4fs, "--get-stat", "Update")

	ver := Version{}
	if err != nil {
		return ver, nil
	}

	lines := strings.Split(databaseOut, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Current AV databases date") {
			ver.CurrentAVDatabasesDate = strings.TrimSpace(strings.TrimPrefix(line, "Current AV databases date:"))
		} else if strings.Contains(line, "Last AV databases update date") {
			ver.LastAVDatabasesUpdateDate = strings.TrimSpace(strings.TrimPrefix(line, "Last AV databases update date:"))
		} else if strings.Contains(line, "Current AV databases state") {
			ver.CurrentAVDatabasesState = strings.TrimSpace(strings.TrimPrefix(line, "Current AV databases state:"))
		} else if strings.Contains(line, "Current AV databases records") {
			ver.CurrentAVDatabasesRecords = strings.TrimSpace(strings.TrimPrefix(line, "Current AV databases records:"))
		}
	}
	return ver, nil
}

// ScanFile a file with Kaspersky scanner
func ScanFile(filePath string) (Result, error) {

	// Run now
	out, err := utils.ExecCommand(kav4fs, "--scan-file", filePath)
	// /opt/kaspersky/kav4fs/bin/kav4fs-control --scan-file locky
	// Objects scanned:     1
	// Threats found:       1
	// Riskware found:      0
	// Infected:            1
	// Suspicious:          0
	// Cured:               0
	// Moved to quarantine: 0
	// Removed:             0
	// Not cured:           0
	// Scan errors:         0
	// Password protected:  0

	res := Result{}
	if err != nil {
		return res, err
	}

	// Check if file is infected
	if !strings.Contains(out, "Threats found:       1") {
		return res, nil
	}

	// Clean the states
	_, stateError := utils.ExecCommand(kav4fs, "--clean-stat")
	if err != nil {
		return res, stateError
	}

	// Grab detection name with a separate cmd
	kavOut, err := utils.ExecCommand(kav4fs, "--top-viruses", "1")
	// Viruses found: 1
	// Virus name:       Trojan-Ransom.Win32.Locky.d
	// Infected objects: 1
	if err != nil {
		return res, err
	}

	lines := strings.Split(kavOut, "\n")
	if len(lines) > 0 {
		res.Output = strings.TrimSpace(strings.Split(lines[1], ":")[1])
		res.Infected = true
	}

	return res, nil
}