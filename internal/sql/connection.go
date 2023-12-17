// Package sql implements the database operations to perform the
// create, read, update and delete triggered by Terraform
package sql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"terraform-provider-azuresql/internal/logging"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/synapse/armsynapse"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/kofalt/go-memoize"

	_ "github.com/microsoft/go-mssqldb/azuread"
)

// The azuresql provider instantiates a single ConnectionCache
// every server/database connection required for provisioning the
// resources is added to this cache. The cache ensures maximal
// reusability of these connections.
type ConnectionCache struct {
	Cache               *memoize.Memoizer
	SubscriptionId      string
	CheckDatabaseExists bool
	CheckServerExists   bool
}

type ConnectionResourceStatus int

const (
	ConnectionResourceStatusUndefined ConnectionResourceStatus = iota
	ConnectionResourceStatusExists
	ConnectionResourceStatusNotFound
	// differs from undefined, in that this state signals that the existence of the resource
	// could not be determined because of the provider settings
	ConnectionResourceStatusUnknown
)

// A single connection stored in the connection cache.
type Connection struct {
	Connection               *sql.DB
	ConnectionId             string
	ConnectionString         string
	Provider                 string
	Server                   string
	Database                 string
	IsServerConnection       bool
	ConnectionResourceStatus ConnectionResourceStatus
}

// Create a new cache. This function is called when starting
// a new azuresql provider
func NewCache(subscriptionId string, check_server_exists bool, check_database_exists bool) ConnectionCache {
	return ConnectionCache{
		Cache:               memoize.NewMemoizer(2*time.Hour, time.Hour),
		SubscriptionId:      subscriptionId,
		CheckServerExists:   check_server_exists,
		CheckDatabaseExists: check_database_exists,
	}
}

func (cache ConnectionCache) synapseServerExists(ctx context.Context, connection Connection) (status ConnectionResourceStatus) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		logging.AddError(ctx, "Failed to obtain a credential", err)
		return ConnectionResourceStatusUnknown
	}

	clientFactory, err := armsynapse.NewClientFactory(cache.SubscriptionId, cred, nil)
	if err != nil {
		logging.AddError(ctx,
			fmt.Sprintf("Failed to create a synapse client for subscription %s",
				cache.SubscriptionId), err)
		return ConnectionResourceStatusUnknown
	}

	pager := clientFactory.NewWorkspacesClient().NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			logging.AddError(ctx,
				fmt.Sprintf("Failed to read all Synapse workspaces in subscription %s",
					cache.SubscriptionId), err)
			return ConnectionResourceStatusUnknown
		}
		for _, workspace := range page.Value {
			// You could use page here. We use blank identifier for just demo purposes.
			if *workspace.Name == connection.Server {
				return ConnectionResourceStatusExists
			}
		}
	}

	return ConnectionResourceStatusNotFound
}

func (cache ConnectionCache) sqlServerExists(ctx context.Context, connection Connection) (status ConnectionResourceStatus) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		logging.AddError(ctx, "Failed to obtain a credential", err)
		return ConnectionResourceStatusUnknown
	}

	policy := policy.TokenRequestOptions{Scopes: []string{"https://management.azure.com/"}}
	token, err := cred.GetToken(ctx, policy)
	if err != nil {
		logging.AddError(ctx,
			fmt.Sprintf(fmt.Sprintf("Failed to request a token for subscriptions/%s/providers/Microsoft.Sql", cache.SubscriptionId)), err)
		return ConnectionResourceStatusUnknown
	}

	url := fmt.Sprintf("https://management.azure.com/subscriptions/%s/providers/Microsoft.Sql/servers?api-version=2021-11-01", cache.SubscriptionId)

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+token.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logging.AddError(ctx, "Failed to list all sql servers in subscription", err)
		return ConnectionResourceStatusUnknown
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logging.AddError(ctx, "Error while reading sql servers", err)
		return ConnectionResourceStatusUnknown
	}

	if resp.StatusCode != 200 {
		logging.AddError(ctx, "Failed to list all sql servers in subscription", string(body))
		return ConnectionResourceStatusUnknown
	}

	var result map[string][]map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		logging.AddError(ctx, "Error while reading sql servers", err)
		return ConnectionResourceStatusUnknown
	}

	for _, server := range result["value"] {
		if server["name"] == connection.Server {
			return ConnectionResourceStatusExists
		}
	}

	return ConnectionResourceStatusNotFound
}

func (cache ConnectionCache) ServerExists(ctx context.Context, connection Connection) (status ConnectionResourceStatus) {
	// Checking the existence of a server connection is only possible when a
	// subscription id is provided
	if cache.SubscriptionId == "" {
		return ConnectionResourceStatusUnknown
	}

	if connection.Provider == "synapse" {
		return cache.synapseServerExists(ctx, connection)
	} else if connection.Provider == "sqlserver" {
		return cache.sqlServerExists(ctx, connection)
	} else {
		logging.AddError(ctx, "Existence check not implemented", fmt.Sprintf("Checking existence for provider %s is not implemnted", connection.Provider))
	}
	return ConnectionResourceStatusUnknown
}

func (cache ConnectionCache) DatabaseExists(ctx context.Context, connection Connection) (status ConnectionResourceStatus) {

	serverConnectionId := strings.TrimSuffix(connection.ConnectionId, ":"+connection.Database)
	serverConnection := cache.Connect(ctx, serverConnectionId, true, false)

	if logging.HasError(ctx) {
		return ConnectionResourceStatusUnknown
	}

	if serverConnection.ConnectionResourceStatus == ConnectionResourceStatusNotFound {
		return ConnectionResourceStatusNotFound
	}

	var response int64

	query := "select database_id from sys.databases where name = @name"
	err := serverConnection.Connection.QueryRowContext(ctx, query, sql.Named("name", connection.Database)).Scan(&response)

	switch {
	case err == sql.ErrNoRows:
		return ConnectionResourceStatusNotFound
	case err != nil:
		logging.AddError(ctx, "Checking database existence failed", err)
		return
	}
	return ConnectionResourceStatusExists
}

func retrySynapsePoolWarmup(ctx context.Context, connection *sql.DB) (err error) {
	// Try connecting again after 2, 15, 60, 180 seconds
	// delay contains the diff of these delays
	var delay = []int{
		3, 12, 45, 120,
	}

	for _, wait := range delay {
		tflog.Info(ctx, fmt.Sprintf("Waiting %d seconds for Synapse to prepare the SQL pools.", wait))
		time.Sleep(time.Duration(wait) * time.Second)

		err = connection.PingContext(ctx)

		error_sql_pool := regexp.MustCompile("The SQL pool is warming up.")
		if err == nil || !error_sql_pool.MatchString(err.Error()) {
			// the pool is no longer warning up, return the new error code
			return err
		}
	}
	return err
}

// Convert a connectionId into an actual SQL connection
// The connectionId is a required parameter of each azuresql terraform resource
func (cache ConnectionCache) Connect(ctx context.Context, connectionId string, server bool, requiresExist bool) Connection {

	tflog.Info(ctx, fmt.Sprintf("Fetching connection to %s", connectionId))

	connection, err, cached := cache.Cache.Memoize(
		connectionId,
		func() (interface{}, error) {
			connection := ParseConnectionId(ctx, connectionId)

			if logging.HasError(ctx) {
				tflog.Debug(ctx, fmt.Sprintf("Parsing of connectionId %s failed", connectionId))
				return connection, nil
			}

			tflog.Debug(ctx, fmt.Sprintf("IsServerConnection: %t", connection.IsServerConnection))
			if cache.CheckServerExists && connection.IsServerConnection {
				tflog.Debug(ctx, fmt.Sprintf("Check server exists"))
				connection.ConnectionResourceStatus = cache.ServerExists(ctx, connection)
			}

			if cache.CheckDatabaseExists && !connection.IsServerConnection {
				tflog.Debug(ctx, fmt.Sprintf("Check database exists"))
				connection.ConnectionResourceStatus = cache.DatabaseExists(ctx, connection)
			}

			if logging.HasError(ctx) || connection.ConnectionResourceStatus == ConnectionResourceStatusNotFound {
				tflog.Debug(ctx, fmt.Sprintf("Connection %s not found", connectionId))
				return connection, nil
			}

			con, err := sql.Open("azuresql", connection.ConnectionString)
			connection.Connection = con
			if err == nil {
				tflog.Debug(ctx, "Pinging database")
				err = connection.Connection.PingContext(ctx)

				error_sql_pool := regexp.MustCompile("The SQL pool is warming up.")
				if err != nil && error_sql_pool.MatchString(err.Error()) {
					err = retrySynapsePoolWarmup(ctx, connection.Connection)
				}
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

	if requiresExist && conn.ConnectionResourceStatus == ConnectionResourceStatusNotFound {
		logging.AddError(ctx, fmt.Sprintf("Database or server not found"), fmt.Sprintf("Connection %s doesn't exist", connectionId))
		return conn
	}

	if conn.IsServerConnection != server {
		if conn.IsServerConnection {
			logging.AddError(ctx, "Attribute error", "Expecting a database connection, but received a server connection.")
		} else {
			logging.AddError(ctx, "Attribute error", "Expecting a server connection, but received a database connection.")
		}
	}

	return conn
}

func (cache ConnectionCache) Connect_server_or_database(ctx context.Context, server string, database string, requiresExist bool) (connection Connection) {

	if server != "" && database != "" {
		logging.AddError(ctx, "Connection failed", "Server and database cannot be both specified when making an SQL connection")
		return
	}

	if server == "" {
		return cache.Connect(ctx, database, false, requiresExist)
	} else {
		return cache.Connect(ctx, server, true, requiresExist)
	}
}

// Convert a connection id into a valid connection string
// ConnectionId format: {provider}::{servername}:{port}:{database}
func ParseConnectionId(ctx context.Context, connectionId string) (connection Connection) {
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
