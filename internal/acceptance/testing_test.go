package acceptance

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"terraform-provider-azuresql/internal/logging"
	"testing"

	_ "github.com/microsoft/go-mssqldb/azuread"
)

func TestAzureCLI(t *testing.T) {
	commandLine := "az --version"

	cliCmd := exec.CommandContext(logging.GetTestContext(), "/bin/sh", "-c", commandLine)

	var stderr bytes.Buffer
	cliCmd.Stderr = &stderr

	_, err := cliCmd.Output()
	if err != nil {
		msg := stderr.String()
		var exErr *exec.ExitError
		t.Error(msg)
		if errors.As(err, &exErr) && exErr.ExitCode() == 127 || strings.HasPrefix(msg, "'az' is not recognized") {
			t.Error("Azure CLI not found on path")

		}
		if msg == "" {
			t.Error(err.Error())
		}
	}
}

func TestExecuteSQL(t *testing.T) {
	PreCheck(t)
	data := BuildTestData(t)

	ExecuteSQL(data.SQLDatabase_connection, fmt.Sprintf("create table dbo.tftable_%s (col1 int)", data.RandomString))
	defer ExecuteSQL(data.SQLDatabase_connection, fmt.Sprintf("DROP table dbo.tftable_%s", data.RandomString))
}
