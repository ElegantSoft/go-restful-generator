package db

func AddUUIDExtension() {
	DB.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
}
