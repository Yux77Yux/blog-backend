package main

import (
	"github.com/yux77yux/blog-backend/cmd/server"
	"github.com/yux77yux/blog-backend/utils/database_utils"
	"github.com/yux77yux/blog-backend/utils/log_utils"
)

func main() {
	defer log_utils.CloseLogFile()

	database_utils.CreateTables()
	server.Server()
}
