package docu

import (
	"fmt"
	"strings"
)

func Supported(sqlserver bool, sqldatabase bool, synapseserver bool, synapsedatabase bool) string {

	var supported []string
	var notSupported []string

	if sqlserver {
		supported = append(supported, "`SQL Server`")
	} else {
		notSupported = append(notSupported, "`Azure SQL Server`")
	}

	if sqldatabase {
		supported = append(supported, "`SQL Database`")
	} else {
		notSupported = append(notSupported, "`Azure SQL Database`")
	}

	if synapseserver {
		supported = append(supported, "`Synapse server`")
	} else {
		notSupported = append(notSupported, "`Synapse server`")
	}

	if synapseserver {
		supported = append(supported, "`Synapse serverless database`")
	} else {
		notSupported = append(notSupported, "`Synapse serverless pool database`")
	}

	notSupported = append(notSupported, "`Synapse dedicated database`")

	return fmt.Sprintf(`

**Supported**: %s \
**Not supported**: %s 
	`, strings.Join(supported, ", "), strings.Join(notSupported, ", "))
}
