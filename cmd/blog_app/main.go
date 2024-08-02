package main

import (
	"github.com/yux77yux/blog-backend/cmd/server"
	"github.com/yux77yux/blog-backend/utils/database_utils"
)

func main() {
	databaseutils.CreateTables()
	server.Server()
}
