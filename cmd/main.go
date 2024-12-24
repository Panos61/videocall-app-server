package main

import (
	"server/internal"
	"server/internal/rdb"
)

func main() {
	rdb.InitRedisClient()
	internal.Run()
}
