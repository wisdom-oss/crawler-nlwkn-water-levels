package globals

import "github.com/jackc/pgx/v5/pgxpool"

// This file contains all globally shared connections (e.g., Databases)

// Db contains the globally available connection to the database
var Db *pgxpool.Pool
