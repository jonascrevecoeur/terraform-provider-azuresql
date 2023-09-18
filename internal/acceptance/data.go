package acceptance

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
)

const (
	// charSetAlphaNum is the alphanumeric character set for use with randStringFromCharSet
	charSetAlphaNum = "abcdefghijklmnopqrstuvwxyz012346789"
)

type TestData struct {
	SQLServer_connection       string
	SQLDatabase_connection     string
	SynapseServer_connection   string
	SynapseDatabase_connection string

	// RandomInteger is a random integer which is unique to this test case
	RandomInteger int

	// RandomString is a random 5 character string is unique to this test case
	RandomString string
}

func BuildTestData(t *testing.T) TestData {
	testData := TestData{
		SQLServer_connection:       fmt.Sprintf("sqlserver::%s:%s", os.Getenv("AZURE_SQL_SERVER"), os.Getenv("AZURE_SQL_SERVER_PORT")),
		SQLDatabase_connection:     fmt.Sprintf("sqlserver::%s:%s:%s", os.Getenv("AZURE_SQL_SERVER"), os.Getenv("AZURE_SQL_SERVER_PORT"), os.Getenv("AZURE_SQL_DATABASE")),
		SynapseServer_connection:   fmt.Sprintf("synapse::%s:%s", os.Getenv("AZURE_SYNAPSE_SERVER"), os.Getenv("AZURE_SYNAPSE_SERVER_PORT")),
		SynapseDatabase_connection: fmt.Sprintf("synapse::%s:%s:%s", os.Getenv("AZURE_SYNAPSE_SERVER"), os.Getenv("AZURE_SYNAPSE_SERVER_PORT"), os.Getenv("AZURE_SYNAPSE_DATABASE")),
		RandomInteger:              RandTimeInt(),
		RandomString:               randString(5),
	}
	return testData
}

// randString generates a random alphanumeric string of the length specified
func randString(strlen int) string {
	return randStringFromCharSet(strlen, charSetAlphaNum)
}

// randStringFromCharSet generates a random string by selecting characters from
// the charset provided
func randStringFromCharSet(strlen int, charSet string) string {
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = charSet[rand.Intn(len(charSet))]
	}
	return string(result)
}

func RandTimeInt() int {
	// acctest.RantInt() returns a value of size:
	// 000000000000000000
	// YYMMddHHmmsshhRRRR

	// go format: 2006-01-02 15:04:05.00

	timeStr := strings.Replace(time.Now().Local().Format("060102150405.00"), ".", "", 1) // no way to not have a .?
	postfix := acctest.RandStringFromCharSet(4, "0123456789")

	i, err := strconv.Atoi(timeStr + postfix)
	if err != nil {
		panic(err)
	}

	return i
}
