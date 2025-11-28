package acceptance

import (
	"os"
	"strings"
	"sync"
	"testing"
)

type BackendKind string

const (
	BackendSQLServer         BackendKind = "sqlserver"
	BackendSynapseServerless BackendKind = "synapse-serverless"
	BackendSynapseDedicated  BackendKind = "synapse-dedicated"
	BackendFabric            BackendKind = "fabric"
)

type Backend struct {
	Kind        BackendKind
	Name        string
	requiredEnv []string
	serverConn  func(TestData) string
	databaseConn func(TestData) string
}

func (b Backend) ServerConn(data TestData) string {
	if b.serverConn == nil {
		return ""
	}
	return b.serverConn(data)
}

func (b Backend) DatabaseConn(data TestData) string {
	if b.databaseConn == nil {
		return ""
	}
	return b.databaseConn(data)
}

func (b Backend) ProviderConfig() string {
	return ProviderConfig
}

var backendRegistry = map[BackendKind]Backend{
	BackendSQLServer: {
		Kind:        BackendSQLServer,
		Name:        "sqlserver",
		requiredEnv: []string{"AZURE_SQL_SERVER", "AZURE_SQL_SERVER_PORT", "AZURE_SQL_DATABASE"},
		serverConn: func(data TestData) string {
			return data.SQLServer_connection
		},
		databaseConn: func(data TestData) string {
			return data.SQLDatabase_connection
		},
	},
	BackendSynapseServerless: {
		Kind:        BackendSynapseServerless,
		Name:        "synapse-serverless",
		requiredEnv: []string{"AZURE_SYNAPSE_SERVER", "AZURE_SYNAPSE_SERVER_PORT", "AZURE_SYNAPSE_DATABASE"},
		serverConn: func(data TestData) string {
			return data.SynapseServer_connection
		},
		databaseConn: func(data TestData) string {
			return data.SynapseDatabase_connection
		},
	},
	BackendSynapseDedicated: {
		Kind:        BackendSynapseDedicated,
		Name:        "synapse-dedicated",
		requiredEnv: []string{"AZURE_SYNAPSE_DEDICATED_SERVER", "AZURE_SYNAPSE_DEDICATED_PORT", "AZURE_SYNAPSE_DEDICATED_DATABASE"},
		serverConn: func(data TestData) string {
			return data.SynapseDedicatedServer_connection
		},
		databaseConn: func(data TestData) string {
			return data.SynapseDedicatedDatabase_connection
		},
	},
	BackendFabric: {
		Kind:        BackendFabric,
		Name:        "fabric",
		requiredEnv: []string{"AZURE_FABRIC_SERVER", "AZURE_FABRIC_DATABASE"},
		serverConn: func(data TestData) string {
			return data.FabricServer_connection
		},
		databaseConn: func(data TestData) string {
			return data.FabricDatabase_connection
		},
	},
}

var backendAliases = map[string]BackendKind{
	"sql":                  BackendSQLServer,
	"sqlserver":            BackendSQLServer,
	"synapse":              BackendSynapseServerless,
	"synapse-serverless":   BackendSynapseServerless,
	"synapseserverless":    BackendSynapseServerless,
	"synapse-dedicated":    BackendSynapseDedicated,
	"synapsededicated":     BackendSynapseDedicated,
	"synapse_dedicated":    BackendSynapseDedicated,
	"fabric":               BackendFabric,
}

var (
	targetsOnce sync.Once
	targets     map[BackendKind]struct{}
)

func Backends(t *testing.T, kinds ...BackendKind) []Backend {
	enabled := backendTargets()
	requested := kinds
	if len(requested) == 0 {
		requested = allBackendKinds()
	}

	var selected []Backend
	for _, kind := range requested {
		backend, ok := backendRegistry[kind]
		if !ok {
			t.Fatalf("Unknown backend %q requested", kind)
		}
		if _, ok := enabled[kind]; !ok {
			t.Logf("Skipping backend %s; not enabled in AZURESQL_TEST_BACKENDS", backend.Name)
			continue
		}
		selected = append(selected, backend)
	}

	if len(selected) == 0 {
		var names []string
		for _, kind := range requested {
			if backend, ok := backendRegistry[kind]; ok {
				names = append(names, backend.Name)
			} else {
				names = append(names, string(kind))
			}
		}
		t.Skipf("No enabled backends for %s", strings.Join(names, ", "))
	}

	return selected
}

func backendTargets() map[BackendKind]struct{} {
	targetsOnce.Do(func() {
		targets = map[BackendKind]struct{}{}
		raw := os.Getenv("AZURESQL_TEST_BACKENDS")
		if raw == "" {
			for kind := range backendRegistry {
				targets[kind] = struct{}{}
			}
			return
		}

		for _, entry := range strings.Split(raw, ",") {
			entry = strings.ToLower(strings.TrimSpace(entry))
			if entry == "" {
				continue
			}
			kind, ok := backendAliases[entry]
			if !ok {
				if backend, ok := backendByName(entry); ok {
					kind = backend.Kind
				} else {
					continue
				}
			}
			targets[kind] = struct{}{}
		}

		if len(targets) == 0 {
			for kind := range backendRegistry {
				targets[kind] = struct{}{}
			}
		}
	})

	return targets
}

func backendByName(name string) (Backend, bool) {
	for _, backend := range backendRegistry {
		if backend.Name == name {
			return backend, true
		}
	}
	return Backend{}, false
}

func allBackendKinds() []BackendKind {
	kinds := make([]BackendKind, 0, len(backendRegistry))
	for kind := range backendRegistry {
		kinds = append(kinds, kind)
	}
	return kinds
}

func requireEnv(t *testing.T, vars ...string) {
	for _, variable := range vars {
		value := os.Getenv(variable)
		if value == "" {
			t.Skipf("%q must be set for acceptance tests", variable)
		}
	}
}

func DatabaseConnection(t *testing.T, backend Backend, data TestData) string {
	connection := backend.DatabaseConn(data)
	if connection == "" {
		t.Skipf("backend %s does not have a database connection configured", backend.Name)
	}
	return connection
}

func ServerConnection(t *testing.T, backend Backend, data TestData) string {
	connection := backend.ServerConn(data)
	if connection == "" {
		t.Skipf("backend %s does not have a server connection configured", backend.Name)
	}
	return connection
}

func ForEachBackend(t *testing.T, kinds []BackendKind, fn func(t *testing.T, backend Backend, data TestData)) {
	for _, backend := range Backends(t, kinds...) {
		backend := backend
		t.Run(backend.Name, func(t *testing.T) {
			PreCheckBackend(t, backend)
			data := BuildTestData(t)
			fn(t, backend, data)
		})
	}
}
