package helper

import "github.com/spf13/viper"

// TableNameWithSchema returns a fully qualified table name when a schema is configured.
func TableNameWithSchema(table string) string {
	schema := viper.GetString(POSTGRES_SCHEMA)
	if schema == "" {
		return table
	}
	return schema + "." + table
}
