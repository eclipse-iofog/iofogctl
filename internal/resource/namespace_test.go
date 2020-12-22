package resource

import (
	"testing"
)

func TestAgents(t *testing.T) {
	ns := Namespace{}
	agents := []Agent{
		&LocalAgent{
			Name: "local",
		},
		&RemoteAgent{
			Name: "remote",
			Host: "123.123.123.123",
			SSH: SSH{
				User:    "serge",
				KeyFile: "~/.ssh/id_rsa",
				Port:    0,
			},
		},
	}
	// Add
	for idx := range agents {
		if err := ns.AddAgent(agents[idx]); err != nil {
			t.Errorf("Failed to create Agent: " + err.Error())
		}
	}
	if len(ns.GetAgents()) != 2 {
		t.Errorf("Failed to get Agents, count: %d", len(ns.GetAgents()))
	}
	// Delete
	if err := ns.DeleteAgent("local"); err != nil {
		t.Errorf("Failed to delete Agent: %s", err.Error())
	}
	if len(ns.GetAgents()) != 1 {
		t.Errorf("Failed to get Agents, count: %d", len(ns.GetAgents()))
	}
	if err := ns.DeleteAgent("remote"); err != nil {
		t.Errorf("Failed to delete Agent: %s", err.Error())
	}
	if len(ns.GetAgents()) != 0 {
		t.Errorf("Failed to get Agents, count: %d", len(ns.GetAgents()))
	}
	// Update
	for idx := 0; idx < len(agents)*2; idx++ {
		modIdx := idx % len(agents)
		if err := ns.UpdateAgent(agents[modIdx]); err != nil {
			t.Errorf("Failed to update Agent: " + err.Error())
		}
	}
	if len(ns.GetAgents()) != 2 {
		t.Errorf("Failed to get Agents, count: %d", len(ns.GetAgents()))
	}
}
