// Package sql implements the database operations to perform the
// create, read, update and delete triggered by Terraform
package sql

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-azuresql/internal/logging"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/kofalt/go-memoize"

	_ "github.com/microsoft/go-mssqldb/azuread"
)

// The azuresql provider instantiates a single ConnectionCache
// every server/database connection required for provisioning the
// resources is added to this cache. The cache ensures maximal
// reusability of these connections.
type ConnectionCache struct {
	Cache *memoize.Memoizer
}

// A single connection stored in the connection cache.
type Connection struct {
	Connection         *sql.DB
	ConnectionId       string
	ConnectionString   string
	Provider           string
	Server             string
	Database           string
	IsServerConnection bool
}

// Create a new cache. This function is called when starting
// a new azuresql provider
func NewCache() ConnectionCache {
	return ConnectionCache{
		Cache: memoize.NewMemoizer(2*time.Hour, time.Hour),
	}
}

// Convert a connectionId into an actual SQL connection
// The connectionId is a required parameter of each azuresql terraform resource
func (cache ConnectionCache) Connect(ctx context.Context, connectionId string, server bool) Connection {

	tflog.Info(ctx, fmt.Sprintf("Fetching connection to %s", connectionId))

	connection, err, cached := cache.Cache.Memoize(
		connectionId,
		func() (interface{}, error) {
			connection := parseConnectionId(ctx, connectionId)
			con, err := sql.Open("azuresql", connection.ConnectionString)
			connection.Connection = con
			if err == nil {
				err = connection.Connection.PingContext(ctx)
			}
			return connection, err
		},
	)

	if logging.HasError(ctx) {
		return Connection{}
	}

	if err != nil {
		logging.AddError(ctx, "Failed to establish an SQL connection", err)
		return Connection{
			ConnectionString: connectionId,
		}
	}

	if cached {
		tflog.Info(ctx, "Succesfully retrieved an existing connection.")
	} else {
		tflog.Info(ctx, "Successfully opened a new connection.")
	}

	conn := connection.(Connection)

	if conn.IsServerConnection != server {
		if conn.IsServerConnection {
			logging.AddError(ctx, "Attribute error", "Expecting a database connection, but received a server connection.")
		} else {
			logging.AddError(ctx, "Attribute error", "Expecting a server connection, but received a database connection.")
		}
	}

	return conn
}

func (cache ConnectionCache) Connect_server_or_database(ctx context.Context, server string, database string) (connection Connection) {

	if server != "" && database != "" {
		logging.AddError(ctx, "Connection failed", "Server and database cannot be both specified when making an SQL connection")
		return
	}

	if server == "" {
		return cache.Connect(ctx, database, false)
	} else {
		return cache.Connect(ctx, server, true)
	}
}

// Convert a connection id into a valid connection string
// ConnectionId format: {provider}::{servername}:{port}:{database}
func parseConnectionId(ctx context.Context, connectionId string) (connection Connection) {
	parts := strings.Split(connectionId, ":")

	if len(parts) < 4 || len(parts) > 5 || parts[1] != "" {
		logging.AddError(ctx, "Invalid connection id", fmt.Sprintf("connection id %s is invalid", connectionId))
		return
	}

	provider := parts[0]
	if provider != "sqlserver" && provider != "synapse" {
		logging.AddError(ctx, "Invalid SQL provider in connection id", fmt.Sprintf("SQL provider %s is invalid. Only sqlserver and synapse are currently supported.", provider))
		return
	}

	server := parts[2]
	port, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		logging.AddError(ctx, "Invalid port in connection id", fmt.Sprintf("Port %s is invalid", parts[3]))
		return
	}

	if provider == "sqlserver" {
		if len(parts) == 5 {
			return Connection{
				ConnectionId:       connectionId,
				ConnectionString:   fmt.Sprintf("%s://%s.database.windows.net:%d?database=%s&fedauth=ActiveDirectoryDefault", provider, server, port, parts[4]),
				IsServerConnection: false,
				Provider:           provider,
				Server:             server,
				Database:           parts[4],
			}
		} else {
			return Connection{
				ConnectionId:       connectionId,
				ConnectionString:   fmt.Sprintf("%s://%s.database.windows.net:%d?fedauth=ActiveDirectoryDefault", provider, server, port),
				IsServerConnection: true,
				Provider:           provider,
				Server:             server,
			}
		}
	} else {
		if len(parts) == 5 {
			return Connection{
				ConnectionId:       connectionId,
				ConnectionString:   fmt.Sprintf("sqlserver://%s-ondemand.sql.azuresynapse.net:%d?database=%s&fedauth=ActiveDirectoryDefault", server, port, parts[4]),
				IsServerConnection: false,
				Provider:           provider,
				Server:             server,
				Database:           parts[4],
			}
		} else {
			return Connection{
				ConnectionId:       connectionId,
				ConnectionString:   fmt.Sprintf("sqlserver://%s-ondemand.sql.azuresynapse.net:%d?fedauth=ActiveDirectoryDefault", server, port),
				IsServerConnection: true,
				Provider:           provider,
				Server:             server,
			}
		}
	}

}
