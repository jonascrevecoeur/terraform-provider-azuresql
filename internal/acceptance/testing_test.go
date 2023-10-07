package acceptance

import (
	"fmt"
	"testing"
)

func TestExecuteSQL(t *testing.T) {
	PreCheck(t)
	data := BuildTestData(t)

	ExecuteSQL(data.SQLDatabase_connection, fmt.Sprintf("create table dbo.tftable_%s (col1 int)", data.RandomString))
	defer ExecuteSQL(data.SQLDatabase_connection, fmt.Sprintf("DROP table dbo.tftable_%s", data.RandomString))
}
