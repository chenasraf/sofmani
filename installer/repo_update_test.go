package installer

import (
	"errors"
	"os"
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestRunRepoUpdateOnce(t *testing.T) {
	t.Run("runs function only once per key", func(t *testing.T) {
		ResetRepoUpdateTracker()
		callCount := 0
		fn := func() error {
			callCount++
			return nil
		}

		err := RunRepoUpdateOnce("test", fn)
		assert.NoError(t, err)
		assert.Equal(t, 1, callCount)

		err = RunRepoUpdateOnce("test", fn)
		assert.NoError(t, err)
		assert.Equal(t, 1, callCount)

		err = RunRepoUpdateOnce("test", fn)
		assert.NoError(t, err)
		assert.Equal(t, 1, callCount)
	})

	t.Run("different keys run independently", func(t *testing.T) {
		ResetRepoUpdateTracker()
		countA := 0
		countB := 0

		_ = RunRepoUpdateOnce("a", func() error { countA++; return nil })
		_ = RunRepoUpdateOnce("b", func() error { countB++; return nil })
		_ = RunRepoUpdateOnce("a", func() error { countA++; return nil })
		_ = RunRepoUpdateOnce("b", func() error { countB++; return nil })

		assert.Equal(t, 1, countA)
		assert.Equal(t, 1, countB)
	})

	t.Run("caches and returns error on subsequent calls", func(t *testing.T) {
		ResetRepoUpdateTracker()
		expectedErr := errors.New("update failed")
		callCount := 0

		err := RunRepoUpdateOnce("fail", func() error {
			callCount++
			return expectedErr
		})
		assert.ErrorIs(t, err, expectedErr)
		assert.Equal(t, 1, callCount)

		err = RunRepoUpdateOnce("fail", func() error {
			callCount++
			return nil
		})
		assert.ErrorIs(t, err, expectedErr)
		assert.Equal(t, 1, callCount)
	})
}

func TestMarkRepoUpdated(t *testing.T) {
	ResetRepoUpdateTracker()

	assert.False(t, IsRepoUpdated("key"))
	MarkRepoUpdated("key")
	assert.True(t, IsRepoUpdated("key"))
	assert.False(t, IsRepoUpdated("other"))
}

func TestResetRepoUpdateTracker(t *testing.T) {
	ResetRepoUpdateTracker()

	MarkRepoUpdated("key")
	assert.True(t, IsRepoUpdated("key"))

	ResetRepoUpdateTracker()
	assert.False(t, IsRepoUpdated("key"))
}

func newAptInstallerWithMode(mode appconfig.RepoUpdateMode) *AptInstaller {
	repoUpdate := map[appconfig.InstallerType]appconfig.RepoUpdateMode{
		appconfig.InstallerTypeApt: mode,
	}
	return &AptInstaller{
		InstallerBase:  InstallerBase{Data: &appconfig.InstallerData{Name: lo.ToPtr("test-pkg"), Type: appconfig.InstallerTypeApt}},
		Config:         &appconfig.AppConfig{RepoUpdate: &repoUpdate},
		Info:           &appconfig.InstallerData{Name: lo.ToPtr("test-pkg"), Type: appconfig.InstallerTypeApt},
		PackageManager: AptPackageManager("true"), // "true" command always succeeds
	}
}

func TestAptRepoUpdateMode(t *testing.T) {
	logger.InitLogger(false)

	t.Run("never mode skips repo update", func(t *testing.T) {
		ResetRepoUpdateTracker()
		inst := newAptInstallerWithMode(appconfig.RepoUpdateNever)

		err := inst.runRepoUpdate()
		assert.NoError(t, err)
		assert.False(t, IsRepoUpdated("true-update"), "tracker should not be set in never mode")
	})

	t.Run("once mode runs repo update only once", func(t *testing.T) {
		ResetRepoUpdateTracker()
		callCount := 0
		origFn := RunRepoUpdateOnce

		inst := newAptInstallerWithMode(appconfig.RepoUpdateOnce)

		// First call should run
		err := inst.runRepoUpdate()
		assert.NoError(t, err)
		assert.True(t, IsRepoUpdated("true-update"), "tracker should be set after first call")

		// Second call should be skipped by RunRepoUpdateOnce
		// We verify by checking the tracker was already set
		_ = RunRepoUpdateOnce("true-update", func() error {
			callCount++
			return nil
		})
		assert.Equal(t, 0, callCount, "function should not run again for same key")
		_ = origFn
	})

	t.Run("always mode runs repo update every time", func(t *testing.T) {
		ResetRepoUpdateTracker()
		inst := newAptInstallerWithMode(appconfig.RepoUpdateAlways)

		// Should succeed and NOT use the tracker
		err := inst.runRepoUpdate()
		assert.NoError(t, err)
		assert.False(t, IsRepoUpdated("true-update"), "tracker should not be set in always mode")

		// Second call should also succeed (not blocked by tracker)
		err = inst.runRepoUpdate()
		assert.NoError(t, err)
	})
}

func newBrewInstallerWithMode(mode appconfig.RepoUpdateMode) *BrewInstaller {
	repoUpdate := map[appconfig.InstallerType]appconfig.RepoUpdateMode{
		appconfig.InstallerTypeBrew: mode,
	}
	return &BrewInstaller{
		InstallerBase: InstallerBase{Data: &appconfig.InstallerData{Name: lo.ToPtr("test-pkg"), Type: appconfig.InstallerTypeBrew}},
		Config:        &appconfig.AppConfig{RepoUpdate: &repoUpdate},
		Info:          &appconfig.InstallerData{Name: lo.ToPtr("test-pkg"), Type: appconfig.InstallerTypeBrew},
	}
}

func TestBrewRepoUpdateMode(t *testing.T) {
	logger.InitLogger(false)

	t.Run("never mode always suppresses auto-update", func(t *testing.T) {
		ResetRepoUpdateTracker()
		_ = os.Unsetenv("HOMEBREW_NO_AUTO_UPDATE")

		inst := newBrewInstallerWithMode(appconfig.RepoUpdateNever)
		inst.handleBrewRepoUpdate()

		assert.Equal(t, "1", os.Getenv("HOMEBREW_NO_AUTO_UPDATE"))

		// Cleanup
		_ = os.Unsetenv("HOMEBREW_NO_AUTO_UPDATE")
	})

	t.Run("always mode never suppresses auto-update", func(t *testing.T) {
		ResetRepoUpdateTracker()
		_ = os.Unsetenv("HOMEBREW_NO_AUTO_UPDATE")

		inst := newBrewInstallerWithMode(appconfig.RepoUpdateAlways)

		// Even after marking as updated, always mode should not suppress
		MarkRepoUpdated("brew")
		inst.handleBrewRepoUpdate()

		assert.Empty(t, os.Getenv("HOMEBREW_NO_AUTO_UPDATE"))
	})

	t.Run("once mode lets first through then suppresses", func(t *testing.T) {
		ResetRepoUpdateTracker()
		_ = os.Unsetenv("HOMEBREW_NO_AUTO_UPDATE")

		inst := newBrewInstallerWithMode(appconfig.RepoUpdateOnce)

		// First call: brew not yet updated, should NOT suppress
		inst.handleBrewRepoUpdate()
		assert.Empty(t, os.Getenv("HOMEBREW_NO_AUTO_UPDATE"), "first call should not suppress")

		// Simulate first brew command completing
		inst.markBrewRepoUpdated()
		assert.True(t, IsRepoUpdated("brew"), "tracker should be set after mark")

		// Second call: brew already updated, should suppress
		inst.handleBrewRepoUpdate()
		assert.Equal(t, "1", os.Getenv("HOMEBREW_NO_AUTO_UPDATE"), "second call should suppress")

		// Cleanup
		_ = os.Unsetenv("HOMEBREW_NO_AUTO_UPDATE")
	})

	t.Run("markBrewRepoUpdated only marks in once mode", func(t *testing.T) {
		ResetRepoUpdateTracker()

		instAlways := newBrewInstallerWithMode(appconfig.RepoUpdateAlways)
		instAlways.markBrewRepoUpdated()
		assert.False(t, IsRepoUpdated("brew"), "should not mark in always mode")

		instNever := newBrewInstallerWithMode(appconfig.RepoUpdateNever)
		instNever.markBrewRepoUpdated()
		assert.False(t, IsRepoUpdated("brew"), "should not mark in never mode")

		instOnce := newBrewInstallerWithMode(appconfig.RepoUpdateOnce)
		instOnce.markBrewRepoUpdated()
		assert.True(t, IsRepoUpdated("brew"), "should mark in once mode")
	})
}
