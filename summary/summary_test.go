package summary

import (
	"testing"

	"github.com/chenasraf/sofmani/logger"
	"github.com/stretchr/testify/assert"
)

func TestNewSummary(t *testing.T) {
	s := NewSummary()
	assert.NotNil(t, s)
	assert.Empty(t, s.results)
}

func TestSummaryAdd(t *testing.T) {
	s := NewSummary()

	result := InstallResult{
		Name:   "test-package",
		Type:   "brew",
		Action: ActionInstalled,
	}

	s.Add(result)
	assert.Len(t, s.results, 1)
	assert.Equal(t, "test-package", s.results[0].Name)

	// Add another
	s.Add(InstallResult{Name: "another", Type: "npm", Action: ActionUpgraded})
	assert.Len(t, s.results, 2)
}

func TestSummaryPrint(t *testing.T) {
	// Initialize logger for tests
	logger.InitLogger(false)

	t.Run("No results prints nothing message", func(t *testing.T) {
		s := NewSummary()
		// This test verifies no panic; actual output goes to logger
		s.Print()
	})

	t.Run("Only skipped results prints nothing message", func(t *testing.T) {
		s := NewSummary()
		s.Add(InstallResult{Name: "skipped-pkg", Type: "brew", Action: ActionSkipped})
		s.Add(InstallResult{Name: "uptodate-pkg", Type: "npm", Action: ActionUpToDate})
		s.Print()
	})

	t.Run("Installed results", func(t *testing.T) {
		s := NewSummary()
		s.Add(InstallResult{Name: "new-pkg", Type: "brew", Action: ActionInstalled})
		s.Print()
	})

	t.Run("Upgraded results", func(t *testing.T) {
		s := NewSummary()
		s.Add(InstallResult{Name: "updated-pkg", Type: "npm", Action: ActionUpgraded})
		s.Print()
	})

	t.Run("Both installed and upgraded", func(t *testing.T) {
		s := NewSummary()
		s.Add(InstallResult{Name: "new-pkg", Type: "brew", Action: ActionInstalled})
		s.Add(InstallResult{Name: "updated-pkg", Type: "npm", Action: ActionUpgraded})
		s.Print()
	})
}

func TestCollectByAction(t *testing.T) {
	logger.InitLogger(false)

	t.Run("Empty summary returns empty slice", func(t *testing.T) {
		s := NewSummary()
		results := s.collectByAction(ActionInstalled)
		assert.Empty(t, results)
	})

	t.Run("Collects matching action", func(t *testing.T) {
		s := NewSummary()
		s.Add(InstallResult{Name: "pkg1", Type: "brew", Action: ActionInstalled})
		s.Add(InstallResult{Name: "pkg2", Type: "npm", Action: ActionUpgraded})
		s.Add(InstallResult{Name: "pkg3", Type: "apt", Action: ActionInstalled})

		installed := s.collectByAction(ActionInstalled)
		assert.Len(t, installed, 2)
		assert.Equal(t, "pkg1", installed[0].Name)
		assert.Equal(t, "pkg3", installed[1].Name)

		upgraded := s.collectByAction(ActionUpgraded)
		assert.Len(t, upgraded, 1)
		assert.Equal(t, "pkg2", upgraded[0].Name)
	})

	t.Run("Collects from nested children", func(t *testing.T) {
		s := NewSummary()
		s.Add(InstallResult{
			Name:   "my-group",
			Type:   "group",
			Action: ActionInstalled,
			Children: []InstallResult{
				{Name: "child1", Type: "brew", Action: ActionInstalled},
				{Name: "child2", Type: "npm", Action: ActionUpgraded},
			},
		})

		installed := s.collectByAction(ActionInstalled)
		assert.Len(t, installed, 1)
		assert.Equal(t, "my-group", installed[0].Name)
		assert.Len(t, installed[0].Children, 1)
		assert.Equal(t, "child1", installed[0].Children[0].Name)
	})
}

func TestCollectResultsByAction(t *testing.T) {
	t.Run("Result matches action", func(t *testing.T) {
		r := InstallResult{Name: "pkg", Type: "brew", Action: ActionInstalled}
		results := collectResultsByAction(r, ActionInstalled)
		assert.Len(t, results, 1)
		assert.Equal(t, "pkg", results[0].Name)
	})

	t.Run("Result does not match action", func(t *testing.T) {
		r := InstallResult{Name: "pkg", Type: "brew", Action: ActionSkipped}
		results := collectResultsByAction(r, ActionInstalled)
		assert.Empty(t, results)
	})

	t.Run("Parent matches, filters children", func(t *testing.T) {
		r := InstallResult{
			Name:   "group",
			Type:   "group",
			Action: ActionInstalled,
			Children: []InstallResult{
				{Name: "child1", Type: "brew", Action: ActionInstalled},
				{Name: "child2", Type: "npm", Action: ActionUpgraded},
				{Name: "child3", Type: "apt", Action: ActionInstalled},
			},
		}

		results := collectResultsByAction(r, ActionInstalled)
		assert.Len(t, results, 1)
		assert.Equal(t, "group", results[0].Name)
		assert.Len(t, results[0].Children, 2)
		assert.Equal(t, "child1", results[0].Children[0].Name)
		assert.Equal(t, "child3", results[0].Children[1].Name)
	})

	t.Run("Parent does not match but children do", func(t *testing.T) {
		r := InstallResult{
			Name:   "group",
			Type:   "group",
			Action: ActionUpToDate,
			Children: []InstallResult{
				{Name: "child1", Type: "brew", Action: ActionInstalled},
				{Name: "child2", Type: "npm", Action: ActionSkipped},
			},
		}

		results := collectResultsByAction(r, ActionInstalled)
		assert.Len(t, results, 1)
		assert.Equal(t, "group", results[0].Name)
		assert.Equal(t, ActionUpToDate, results[0].Action) // Parent keeps its action
		assert.Len(t, results[0].Children, 1)
		assert.Equal(t, "child1", results[0].Children[0].Name)
	})

	t.Run("Neither parent nor children match", func(t *testing.T) {
		r := InstallResult{
			Name:   "group",
			Type:   "group",
			Action: ActionUpToDate,
			Children: []InstallResult{
				{Name: "child1", Type: "brew", Action: ActionSkipped},
				{Name: "child2", Type: "npm", Action: ActionUpToDate},
			},
		}

		results := collectResultsByAction(r, ActionInstalled)
		assert.Empty(t, results)
	})
}

func TestFilterChildrenByAction(t *testing.T) {
	t.Run("Empty children returns empty slice", func(t *testing.T) {
		filtered := filterChildrenByAction([]InstallResult{}, ActionInstalled)
		assert.Empty(t, filtered)
	})

	t.Run("Filters matching children", func(t *testing.T) {
		children := []InstallResult{
			{Name: "child1", Type: "brew", Action: ActionInstalled},
			{Name: "child2", Type: "npm", Action: ActionUpgraded},
			{Name: "child3", Type: "apt", Action: ActionInstalled},
		}

		filtered := filterChildrenByAction(children, ActionInstalled)
		assert.Len(t, filtered, 2)
		assert.Equal(t, "child1", filtered[0].Name)
		assert.Equal(t, "child3", filtered[1].Name)
	})

	t.Run("Includes container with matching grandchildren", func(t *testing.T) {
		children := []InstallResult{
			{
				Name:   "nested-group",
				Type:   "group",
				Action: ActionUpToDate,
				Children: []InstallResult{
					{Name: "grandchild", Type: "brew", Action: ActionInstalled},
				},
			},
		}

		filtered := filterChildrenByAction(children, ActionInstalled)
		assert.Len(t, filtered, 1)
		assert.Equal(t, "nested-group", filtered[0].Name)
		assert.Len(t, filtered[0].Children, 1)
		assert.Equal(t, "grandchild", filtered[0].Children[0].Name)
	})

	t.Run("Recursively filters deeply nested children", func(t *testing.T) {
		children := []InstallResult{
			{
				Name:   "level1",
				Type:   "group",
				Action: ActionUpToDate,
				Children: []InstallResult{
					{
						Name:   "level2",
						Type:   "group",
						Action: ActionUpToDate,
						Children: []InstallResult{
							{Name: "level3-installed", Type: "brew", Action: ActionInstalled},
							{Name: "level3-skipped", Type: "npm", Action: ActionSkipped},
						},
					},
				},
			},
		}

		filtered := filterChildrenByAction(children, ActionInstalled)
		assert.Len(t, filtered, 1)
		assert.Equal(t, "level1", filtered[0].Name)
		assert.Len(t, filtered[0].Children, 1)
		assert.Equal(t, "level2", filtered[0].Children[0].Name)
		assert.Len(t, filtered[0].Children[0].Children, 1)
		assert.Equal(t, "level3-installed", filtered[0].Children[0].Children[0].Name)
	})
}

func TestHasChildrenWithAction(t *testing.T) {
	t.Run("Empty children returns false", func(t *testing.T) {
		result := hasChildrenWithAction([]InstallResult{}, ActionInstalled)
		assert.False(t, result)
	})

	t.Run("Direct child matches", func(t *testing.T) {
		children := []InstallResult{
			{Name: "child1", Type: "brew", Action: ActionInstalled},
		}
		result := hasChildrenWithAction(children, ActionInstalled)
		assert.True(t, result)
	})

	t.Run("No children match", func(t *testing.T) {
		children := []InstallResult{
			{Name: "child1", Type: "brew", Action: ActionSkipped},
			{Name: "child2", Type: "npm", Action: ActionUpToDate},
		}
		result := hasChildrenWithAction(children, ActionInstalled)
		assert.False(t, result)
	})

	t.Run("Grandchild matches", func(t *testing.T) {
		children := []InstallResult{
			{
				Name:   "group",
				Type:   "group",
				Action: ActionUpToDate,
				Children: []InstallResult{
					{Name: "grandchild", Type: "brew", Action: ActionInstalled},
				},
			},
		}
		result := hasChildrenWithAction(children, ActionInstalled)
		assert.True(t, result)
	})

	t.Run("Deep nested match", func(t *testing.T) {
		children := []InstallResult{
			{
				Name:   "level1",
				Type:   "group",
				Action: ActionUpToDate,
				Children: []InstallResult{
					{
						Name:   "level2",
						Type:   "group",
						Action: ActionSkipped,
						Children: []InstallResult{
							{Name: "level3", Type: "brew", Action: ActionUpgraded},
						},
					},
				},
			},
		}
		result := hasChildrenWithAction(children, ActionUpgraded)
		assert.True(t, result)
	})
}

func TestGroupWithNoMatchingChildren(t *testing.T) {
	logger.InitLogger(false)

	t.Run("Group with no matching children is excluded", func(t *testing.T) {
		s := NewSummary()
		// Group is marked as Installed, but all children are UpToDate
		s.Add(InstallResult{
			Name:   "my-group",
			Type:   "group",
			Action: ActionInstalled,
			Children: []InstallResult{
				{Name: "child1", Type: "brew", Action: ActionUpToDate},
				{Name: "child2", Type: "npm", Action: ActionUpToDate},
			},
		})

		installed := s.collectByAction(ActionInstalled)
		assert.Empty(t, installed, "Group with no installed children should not appear in installed list")

		upgraded := s.collectByAction(ActionUpgraded)
		assert.Empty(t, upgraded, "Group with no upgraded children should not appear in upgraded list")
	})

	t.Run("Group appears only in sections where it has matching children", func(t *testing.T) {
		s := NewSummary()
		s.Add(InstallResult{
			Name:   "my-group",
			Type:   "group",
			Action: ActionInstalled,
			Children: []InstallResult{
				{Name: "child1", Type: "brew", Action: ActionInstalled},
				{Name: "child2", Type: "npm", Action: ActionUpToDate},
			},
		})

		installed := s.collectByAction(ActionInstalled)
		assert.Len(t, installed, 1)
		assert.Equal(t, "my-group", installed[0].Name)
		assert.Len(t, installed[0].Children, 1)
		assert.Equal(t, "child1", installed[0].Children[0].Name)

		upgraded := s.collectByAction(ActionUpgraded)
		assert.Empty(t, upgraded, "Group should not appear in upgraded since no children were upgraded")
	})

	t.Run("Empty group is excluded", func(t *testing.T) {
		s := NewSummary()
		s.Add(InstallResult{
			Name:     "empty-group",
			Type:     "group",
			Action:   ActionInstalled,
			Children: []InstallResult{},
		})

		installed := s.collectByAction(ActionInstalled)
		assert.Empty(t, installed, "Empty group should not appear")
	})
}

func TestActionConstants(t *testing.T) {
	// Verify action constants have expected values
	assert.Equal(t, Action(0), ActionSkipped)
	assert.Equal(t, Action(1), ActionUpToDate)
	assert.Equal(t, Action(2), ActionInstalled)
	assert.Equal(t, Action(3), ActionUpgraded)
}

func TestInstallResultStructure(t *testing.T) {
	r := InstallResult{
		Name:   "test",
		Type:   "brew",
		Action: ActionInstalled,
		Children: []InstallResult{
			{Name: "child", Type: "npm", Action: ActionUpgraded},
		},
	}

	assert.Equal(t, "test", r.Name)
	assert.Equal(t, "brew", r.Type)
	assert.Equal(t, ActionInstalled, r.Action)
	assert.Len(t, r.Children, 1)
	assert.Equal(t, "child", r.Children[0].Name)
}

func TestComplexHierarchy(t *testing.T) {
	logger.InitLogger(false)

	// Simulate a complex real-world scenario with mixed results
	s := NewSummary()

	// Top-level installed package
	s.Add(InstallResult{Name: "standalone-pkg", Type: "brew", Action: ActionInstalled})

	// Group with mixed results
	s.Add(InstallResult{
		Name:   "dev-tools",
		Type:   "group",
		Action: ActionInstalled,
		Children: []InstallResult{
			{Name: "git", Type: "brew", Action: ActionUpToDate},
			{Name: "node", Type: "brew", Action: ActionInstalled},
			{Name: "yarn", Type: "npm", Action: ActionUpgraded},
			{
				Name:   "linters",
				Type:   "group",
				Action: ActionInstalled,
				Children: []InstallResult{
					{Name: "eslint", Type: "npm", Action: ActionInstalled},
					{Name: "prettier", Type: "npm", Action: ActionSkipped},
				},
			},
		},
	})

	// Top-level upgraded package
	s.Add(InstallResult{Name: "updated-pkg", Type: "apt", Action: ActionUpgraded})

	// Verify installed collection
	installed := s.collectByAction(ActionInstalled)
	assert.Len(t, installed, 2) // standalone-pkg and dev-tools

	// Verify the group has filtered children
	devTools := installed[1]
	assert.Equal(t, "dev-tools", devTools.Name)
	assert.Len(t, devTools.Children, 2) // node and linters (git is up-to-date, yarn is upgraded)

	// Verify nested group
	linters := devTools.Children[1]
	assert.Equal(t, "linters", linters.Name)
	assert.Len(t, linters.Children, 1) // only eslint (prettier is skipped)

	// Verify upgraded collection
	upgraded := s.collectByAction(ActionUpgraded)
	assert.Len(t, upgraded, 2) // dev-tools (because of yarn) and updated-pkg

	// Print should not panic
	s.Print()
}
