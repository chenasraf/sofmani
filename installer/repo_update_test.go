package installer

import (
	"errors"
	"testing"

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
