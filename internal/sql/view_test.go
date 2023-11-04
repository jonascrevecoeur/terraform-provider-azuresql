package sql

import (
	"fmt"
	"terraform-provider-azuresql/internal/logging"
	"testing"
)

func TestExtractViewDefintion(t *testing.T) {
	ctx := logging.GetTestContext()

	definition := "select * from table"
	statement := fmt.Sprintf(`
		create view %s.%s as (
			%s
		)
	`, "schema", "view", definition)

	parsed := extractViewDefintion(ctx, statement)
	if parsed != definition {
		t.Errorf("Parsed view definition does not match original definition, %s, %s", definition, parsed)
	}
}
