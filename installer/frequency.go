package installer

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/chenasraf/sofmani/logger"
	"github.com/chenasraf/sofmani/utils"
)

// frequencyCacheFileName returns the cache file name for a given installer name,
// escaping characters that are not safe for file names.
func frequencyCacheFileName(name string) string {
	replacer := strings.NewReplacer(
		"/", "__",
		"\\", "__",
		":", "__",
		" ", "_",
	)
	return "freq_" + replacer.Replace(name)
}

// checkFrequency checks whether enough time has passed since the last successful run
// for an installer with the given name and frequency string.
// Returns true if the installer should run (frequency elapsed or no previous run).
func checkFrequency(name string, frequency string) (bool, error) {
	dur, err := utils.ParsePrettyDuration(frequency)
	if err != nil {
		return false, err
	}

	cacheDir, err := utils.GetCacheDir()
	if err != nil {
		return true, nil // if we can't get cache dir, just run
	}

	cacheFile := filepath.Join(cacheDir, frequencyCacheFileName(name))
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		// No previous run recorded
		return true, nil
	}

	ts, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		// Corrupt cache file, just run
		logger.Debug("Invalid frequency cache for %s, will run", logger.H(name))
		return true, nil
	}

	lastRun := time.Unix(ts, 0)
	if time.Since(lastRun) < dur {
		return false, nil
	}

	return true, nil
}

// writeFrequencyTimestamp writes the current timestamp to the frequency cache file
// for the given installer name.
func writeFrequencyTimestamp(name string) error {
	cacheDir, err := utils.GetCacheDir()
	if err != nil {
		return err
	}

	cacheFile := filepath.Join(cacheDir, frequencyCacheFileName(name))
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	return os.WriteFile(cacheFile, []byte(ts), 0644)
}
