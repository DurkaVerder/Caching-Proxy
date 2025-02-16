package main

import (
	"Caching-Proxy/internal/cache"
	"Caching-Proxy/internal/service"
	"os"
	"time"
)

func main() {
	cache := cache.NewCache(30*time.Second, 1*time.Minute)
	service := service.NewServiceManager(os.Args, cache)

	service.Run()
}
