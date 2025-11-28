package docu

import "fmt"

func FieldDatabase(name string, allowServer bool, isResource bool) string {
	if isResource && allowServer {
		return fmt.Sprintf("The ID of the database in which the %s should be created. Either `database` or `server` is required.", name)
	}
	if isResource && !allowServer {
		return fmt.Sprintf("The ID of the database in which the %s should be created.", name)
	}
	if !isResource && allowServer {
		return fmt.Sprintf("The ID of the database in which the %s exists. Either `database` or `server` is required.", name)
	}
	if !isResource && !allowServer {
		return fmt.Sprintf("The ID of the database in which the %s exists.", name)
	}
	return ""
}
