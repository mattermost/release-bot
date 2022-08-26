package version

import (
	"time"

	log "github.com/sirupsen/logrus"
)

const appName = "Release-Bot"
const dev = "dev"

// Provisioned by ldflags
var (
	name       string
	version    string
	commitHash string
	buildDate  string
)

type Info struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Hash    string `json:"hash"`
	Date    string `json:"date"`
}

func init() {
	if name == "" {
		name = appName
	}
	if version == "" {
		version = dev
	}
	if commitHash == "" {
		commitHash = dev
	}
	if buildDate == "" {
		buildDate = time.Now().Format(time.RFC3339)
	}
}

func Full() *Info {
	return &Info{
		Name:    name,
		Version: version,
		Hash:    commitHash,
		Date:    buildDate,
	}
}

func Log() {
	log.WithFields(log.Fields{
		"name":      name,
		"version":   version,
		"hash":      commitHash,
		"buildDate": buildDate,
	}).Info("Version")
}
