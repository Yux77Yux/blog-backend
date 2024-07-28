package blogapp

import (
	"github.com/yux77yux/blog-backend/cmd/server"
	"github.com/yux77yux/blog-backend/utils"
)

func main() {
	utils.CreateTables()
	server.Server()
}
