package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/stretchr/testify/assert"
)

func newTestGroupInstaller(data *appconfig.InstallerData) *GroupInstaller {
	return &GroupInstaller{
		InstallerBase: InstallerBase{
			Data: data,
		},
		Config: nil,
		Data:   data,
	}
}

func TestGroupValidation(t *testing.T) {
	logger.InitLogger(false)

	// 🟢 Valid: one sub-installer
	validStep := appconfig.InstallerData{
		Name: strPtr("child-installer"),
		Type: appconfig.InstallerTypeBrew,
	}
	validData := &appconfig.InstallerData{
		Name:  strPtr("group-valid"),
		Type:  appconfig.InstallerTypeGroup,
		Steps: &[]appconfig.InstallerData{validStep},
	}
	assertNoValidationErrors(t, newTestGroupInstaller(validData).Validate())

	// 🔴 Invalid: empty steps slice
	emptySteps := &appconfig.InstallerData{
		Name:  strPtr("group-empty"),
		Type:  appconfig.InstallerTypeGroup,
		Steps: &[]appconfig.InstallerData{},
	}
	assertValidationError(t, newTestGroupInstaller(emptySteps).Validate(), "steps")

	// 🔴 Invalid: nil steps
	nilSteps := &appconfig.InstallerData{
		Name:  strPtr("group-nil"),
		Type:  appconfig.InstallerTypeGroup,
		Steps: nil,
	}
	assertValidationError(t, newTestGroupInstaller(nilSteps).Validate(), "steps")
}

func TestGroupGetData(t *testing.T) {
	logger.InitLogger(false)

	t.Run("returns the installer data", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: strPtr("group-test"),
			Type: appconfig.InstallerTypeGroup,
		}
		installer := newTestGroupInstaller(data)
		result := installer.GetData()

		assert.Equal(t, data, result)
		assert.Equal(t, "group-test", *result.Name)
	})
}

func TestGroupGetOpts(t *testing.T) {
	logger.InitLogger(false)

	t.Run("returns empty opts", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: strPtr("group-test"),
			Type: appconfig.InstallerTypeGroup,
		}
		installer := newTestGroupInstaller(data)
		opts := installer.GetOpts()

		assert.NotNil(t, opts)
	})
}

func TestGroupGetBinName(t *testing.T) {
	logger.InitLogger(false)

	t.Run("returns name when bin_name is not set", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: strPtr("my-group"),
			Type: appconfig.InstallerTypeGroup,
		}
		installer := newTestGroupInstaller(data)
		assert.Equal(t, "my-group", installer.GetBinName())
	})

	t.Run("returns bin_name when set", func(t *testing.T) {
		binName := "custom-bin"
		data := &appconfig.InstallerData{
			Name:    strPtr("my-group"),
			Type:    appconfig.InstallerTypeGroup,
			BinName: &binName,
		}
		installer := newTestGroupInstaller(data)
		assert.Equal(t, "custom-bin", installer.GetBinName())
	})

	t.Run("returns name when bin_name is empty", func(t *testing.T) {
		binName := ""
		data := &appconfig.InstallerData{
			Name:    strPtr("my-group"),
			Type:    appconfig.InstallerTypeGroup,
			BinName: &binName,
		}
		installer := newTestGroupInstaller(data)
		assert.Equal(t, "my-group", installer.GetBinName())
	})
}

func TestGroupCheckIsInstalled(t *testing.T) {
	logger.InitLogger(false)

	t.Run("runs custom check when provided", func(t *testing.T) {
		checkCmd := "true"
		data := &appconfig.InstallerData{
			Name:           strPtr("group-test"),
			Type:           appconfig.InstallerTypeGroup,
			CheckInstalled: &checkCmd,
		}
		installer := newTestGroupInstaller(data)
		result, err := installer.CheckIsInstalled()

		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("runs custom check that fails", func(t *testing.T) {
		checkCmd := "false"
		data := &appconfig.InstallerData{
			Name:           strPtr("group-test"),
			Type:           appconfig.InstallerTypeGroup,
			CheckInstalled: &checkCmd,
		}
		installer := newTestGroupInstaller(data)
		result, err := installer.CheckIsInstalled()

		assert.NoError(t, err)
		assert.False(t, result)
	})
}

func TestGroupCheckNeedsUpdate(t *testing.T) {
	logger.InitLogger(false)

	t.Run("returns true when no custom check", func(t *testing.T) {
		data := &appconfig.InstallerData{
			Name: strPtr("group-test"),
			Type: appconfig.InstallerTypeGroup,
		}
		installer := newTestGroupInstaller(data)
		result, err := installer.CheckNeedsUpdate()

		assert.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("runs custom check when provided", func(t *testing.T) {
		checkCmd := "false" // Returns exit code 1, meaning no update
		data := &appconfig.InstallerData{
			Name:           strPtr("group-test"),
			Type:           appconfig.InstallerTypeGroup,
			CheckHasUpdate: &checkCmd,
		}
		installer := newTestGroupInstaller(data)
		result, err := installer.CheckNeedsUpdate()

		assert.NoError(t, err)
		assert.False(t, result)
	})
}

func TestNewGroupInstaller(t *testing.T) {
	logger.InitLogger(false)

	t.Run("creates installer with config and data", func(t *testing.T) {
		cfg := &appconfig.AppConfig{}
		data := &appconfig.InstallerData{
			Name: strPtr("group-test"),
			Type: appconfig.InstallerTypeGroup,
		}
		installer := NewGroupInstaller(cfg, data)

		assert.NotNil(t, installer)
		assert.Equal(t, cfg, installer.Config)
		assert.Equal(t, data, installer.Data)
	})
}

func TestGroupValidationWithMultipleSteps(t *testing.T) {
	logger.InitLogger(false)

	t.Run("valid with multiple steps", func(t *testing.T) {
		step1 := appconfig.InstallerData{
			Name: strPtr("step1"),
			Type: appconfig.InstallerTypeBrew,
		}
		step2 := appconfig.InstallerData{
			Name: strPtr("step2"),
			Type: appconfig.InstallerTypeShell,
		}
		data := &appconfig.InstallerData{
			Name:  strPtr("group-multi"),
			Type:  appconfig.InstallerTypeGroup,
			Steps: &[]appconfig.InstallerData{step1, step2},
		}
		assertNoValidationErrors(t, newTestGroupInstaller(data).Validate())
	})
}
