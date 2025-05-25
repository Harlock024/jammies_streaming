package main

import (
	"jammies_streaming/src/db"
	"jammies_streaming/src/ws"

	"github.com/gin-gonic/gin"
)

func main() {
	db.InitDB()

	r := gin.Default()

	r.GET("/ws", ws.HandleWS)
	r.Run(":8081")
}
