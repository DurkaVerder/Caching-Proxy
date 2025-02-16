package main

import (
	"Caching-Proxy/internal/cache"
	"Caching-Proxy/internal/service"
	"os"
	"time"
)

func main() {
	cache := cache.NewCache(1*time.Second, 2*time.Minute, 3000)
	service := service.NewServiceManager(os.Args, cache)

	service.Run()
}
