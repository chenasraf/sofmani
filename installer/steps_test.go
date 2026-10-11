package installer

import (
	"testing"

	"github.com/chenasraf/sofmani/appconfig"
	"github.com/chenasraf/sofmani/logger"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func shellStep(name string) appconfig.InstallerData {
	return appconfig.InstallerData{
		Name: lo.ToPtr(name),
		Type: appconfig.InstallerTypeShell,
		Opts: &map[string]any{"command": "true"},
	}
}

func TestPrepareSteps(t *testing.T) {
	logger.InitLogger(false)
	config := &appconfig.AppConfig{}

	t.Run("keeps categories as headers", func(t *testing.T) {
		steps, problems := PrepareSteps(config, []appconfig.InstallerData{
			{Category: lo.ToPtr("Tools")},
			shellStep("tool"),
		})
		assert.Empty(t, problems)
		assert.Len(t, steps, 2)
		assert.True(t, steps[0].IsCategory())
		assert.False(t, steps[1].IsCategory())
	})

	t.Run("reports a missing name without running type validation", func(t *testing.T) {
		steps, problems := PrepareSteps(config, []appconfig.InstallerData{
			{Type: appconfig.InstallerTypeGitHubRelease},
		})
		assert.Empty(t, steps)
		assert.Len(t, problems, 1)
		assert.Equal(t, "name", problems[0].FieldName)
		assert.Equal(t, "github-release installer #1", problems[0].InstallerName)
	})

	t.Run("reports type validation errors", func(t *testing.T) {
		_, problems := PrepareSteps(config, []appconfig.InstallerData{
			{Name: lo.ToPtr("release"), Type: appconfig.InstallerTypeGitHubRelease},
		})
		assert.NotEmpty(t, problems)
		assert.Equal(t, "release", problems[0].InstallerName)
	})

	t.Run("leaves out unsupported types", func(t *testing.T) {
		steps, problems := PrepareSteps(config, []appconfig.InstallerData{
			{Name: lo.ToPtr("unknown"), Type: appconfig.InstallerType("nope")},
		})
		assert.Empty(t, steps)
		assert.Empty(t, problems)
	})
}

func TestRunSteps(t *testing.T) {
	logger.InitLogger(false)
	config := &appconfig.AppConfig{}

	t.Run("runs installers around category headers", func(t *testing.T) {
		results, err := RunSteps(config, []appconfig.InstallerData{
			{Category: lo.ToPtr("Tools")},
			shellStep("sofmani-run-steps-test"),
		})
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "sofmani-run-steps-test", results[0].Name)
	})

	t.Run("runs nothing when a step is invalid", func(t *testing.T) {
		results, err := RunSteps(config, []appconfig.InstallerData{
			shellStep("sofmani-run-steps-test"),
			{Type: appconfig.InstallerTypeShell},
		})
		assert.ErrorContains(t, err, "shell installer #2")
		assert.Empty(t, results)
	})
}

func TestGroupValidateChecksSteps(t *testing.T) {
	logger.InitLogger(false)
	group := NewGroupInstaller(&appconfig.AppConfig{}, &appconfig.InstallerData{
		Name:  lo.ToPtr("group"),
		Type:  appconfig.InstallerTypeGroup,
		Steps: &[]appconfig.InstallerData{{Category: lo.ToPtr("Tools")}, {Type: appconfig.InstallerTypeShell}},
	})
	problems := group.Validate()
	assert.Len(t, problems, 1)
	assert.Equal(t, "shell installer #2", problems[0].InstallerName)
}
