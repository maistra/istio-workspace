package model

import (
	"crypto/sha256"
	"fmt"
)

const unknownVersion = "unknown"

// GetVersion returns the version for the created resources if any. Returns unknown if not found.
func GetVersion(store LocatorStatusStore) string {
	targets := store("Deployment", "DeploymentConfig")
	for _, target := range targets {
		if target.Action != ActionDelete && target.Action != ActionRevert {
			if val, ok := target.Labels["version"]; ok {
				return val
			}
		}
	}

	return unknownVersion
}

// GetDeletedVersion returns the version for the deleted resources if any. Returns unknown if not found.
func GetDeletedVersion(store LocatorStatusStore) string {
	targets := store("Deployment", "DeploymentConfig")
	for _, target := range targets {
		if target.Action == ActionDelete || target.Action == ActionRevert {
			if val, ok := target.Labels["version"]; ok {
				return val
			}
		}
	}

	return unknownVersion
}

// GetCreatedVersion returns the new calculated version for the created resources if any. Returns unknown if not found.
func GetCreatedVersion(store LocatorStatusStore, sessionName string) string {
	targets := store("Deployment", "DeploymentConfig")
	for _, target := range targets {
		if target.Action != ActionDelete && target.Action != ActionRevert {
			if val, ok := target.Labels["version"]; ok {
				return GetSha(val) + "-" + sessionName
			}
		}
	}

	return unknownVersion
}

// GetSha computes a hash of the version and returns 8 characters substring of it.
func GetSha(version string) string {
	sum := sha256.Sum256([]byte(version))
	sha := fmt.Sprintf("%x", sum)

	return sha[:8]
}
