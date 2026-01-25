package machine

import "testing"

func TestMachines_GetShouldRunOnMachine(t *testing.T) {
	tests := []struct {
		name      string
		machines  *Machines
		machineID string
		aliases   map[string]string
		want      bool
	}{
		{
			name:      "nil machines should run on any machine",
			machines:  nil,
			machineID: "abc123",
			aliases:   nil,
			want:      true,
		},
		{
			name:      "empty machines should run on any machine",
			machines:  &Machines{},
			machineID: "abc123",
			aliases:   nil,
			want:      true,
		},
		{
			name: "only matching machine ID should run",
			machines: &Machines{
				Only: &[]string{"abc123", "def456"},
			},
			machineID: "abc123",
			aliases:   nil,
			want:      true,
		},
		{
			name: "only non-matching machine ID should not run",
			machines: &Machines{
				Only: &[]string{"abc123", "def456"},
			},
			machineID: "xyz789",
			aliases:   nil,
			want:      false,
		},
		{
			name: "except matching machine ID should not run",
			machines: &Machines{
				Except: &[]string{"abc123", "def456"},
			},
			machineID: "abc123",
			aliases:   nil,
			want:      false,
		},
		{
			name: "except non-matching machine ID should run",
			machines: &Machines{
				Except: &[]string{"abc123", "def456"},
			},
			machineID: "xyz789",
			aliases:   nil,
			want:      true,
		},
		{
			name: "only takes precedence over except",
			machines: &Machines{
				Only:   &[]string{"abc123"},
				Except: &[]string{"abc123"},
			},
			machineID: "abc123",
			aliases:   nil,
			want:      true,
		},
		{
			name: "alias matching should run",
			machines: &Machines{
				Only: &[]string{"work-laptop"},
			},
			machineID: "abc123",
			aliases:   map[string]string{"work-laptop": "abc123"},
			want:      true,
		},
		{
			name: "alias non-matching should not run",
			machines: &Machines{
				Only: &[]string{"work-laptop"},
			},
			machineID: "xyz789",
			aliases:   map[string]string{"work-laptop": "abc123"},
			want:      false,
		},
		{
			name: "except with alias matching should not run",
			machines: &Machines{
				Except: &[]string{"home-server"},
			},
			machineID: "abc123",
			aliases:   map[string]string{"home-server": "abc123"},
			want:      false,
		},
		{
			name: "except with alias non-matching should run",
			machines: &Machines{
				Except: &[]string{"home-server"},
			},
			machineID: "xyz789",
			aliases:   map[string]string{"home-server": "abc123"},
			want:      true,
		},
		{
			name: "mixed aliases and literal IDs in only",
			machines: &Machines{
				Only: &[]string{"work-laptop", "def456"},
			},
			machineID: "abc123",
			aliases:   map[string]string{"work-laptop": "abc123"},
			want:      true,
		},
		{
			name: "literal ID fallback when no alias match",
			machines: &Machines{
				Only: &[]string{"abc123"},
			},
			machineID: "abc123",
			aliases:   map[string]string{"other": "xyz789"},
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.machines.GetShouldRunOnMachine(tt.machineID, tt.aliases)
			if got != tt.want {
				t.Errorf("GetShouldRunOnMachine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMachineID(t *testing.T) {
	// Reset any cached value
	ResetMachineID()

	// Get the machine ID
	id := GetMachineID()

	// Should not be empty
	if id == "" {
		t.Error("GetMachineID() returned empty string")
	}

	// Should be 16 characters (truncated hash)
	if len(id) != 16 {
		t.Errorf("GetMachineID() returned ID with length %d, want 16", len(id))
	}

	// Should be deterministic (same value on subsequent calls)
	id2 := GetMachineID()
	if id != id2 {
		t.Errorf("GetMachineID() not deterministic: got %s then %s", id, id2)
	}
}

func TestSetMachineID(t *testing.T) {
	// Save original and restore after test
	originalID := GetMachineID()
	defer SetMachineID(originalID)

	// Set a custom ID
	customID := "testmachineid123"
	SetMachineID(customID)

	// Verify the custom ID is returned
	got := GetMachineID()
	if got != customID {
		t.Errorf("After SetMachineID(%s), GetMachineID() = %s", customID, got)
	}
}
