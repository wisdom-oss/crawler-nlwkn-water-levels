package globals

import (
	"github.com/qustavo/dotsql"
)

// This file contains globally shared variables (e.g., service name, sql queries)

// ServiceName contains the global identifier for the service
const ServiceName = "nlwkn-water-level-crawler"

// SqlQueries contains the prepared sql queries from the resources folder
var SqlQueries *dotsql.DotSql

// Environment contains a mapping between the environment variables and the values
// they were set to. However, this variable only contains the configured environment
// variables
var Environment map[string]string = make(map[string]string)
