package acceptance

import (
	"fmt"
	"strings"
)

func TerraformConnectionId(connection string) string {
	if len(strings.Split(connection, ":")) == 4 {
		return fmt.Sprintf("server = \"%s\"", connection)
	} else {
		return fmt.Sprintf("database = \"%s\"", connection)
	}
}
