package main

import (
	"context"
	"flag"
	"time"

	"k8s.io/klog/v2"

	"github.com/volcano-sh/agentcube/pkg/kernelhub"
)

func main() {
	port := flag.Int("port", 8081, "Port for KernelHub server")
	workspace := flag.String("workspace", "", "Workspace directory")
	memoryLimit := flag.Int64("memory-limit", 512*1024*1024, "Memory limit per kernel")
	maxSessions := flag.Int("max-sessions", 100, "Maximum concurrent sessions")
	sessionTTL := flag.Duration("session-ttl", 24*time.Hour, "Session TTL")
	stateStore := flag.String("state-store", "memory", "State store: memory or redis")
	redisAddr := flag.String("redis-addr", "", "Redis address")
	redisPassword := flag.String("redis-password", "", "Redis password")
	cleanupInterval := flag.Duration("cleanup-interval", 5*time.Minute, "Interval for cleaning up idle sessions")

	klog.InitFlags(nil)
	flag.Parse()

	config := kernelhub.Config{
		Port:          *port,
		Workspace:     *workspace,
		MemoryLimit:   *memoryLimit,
		MaxSessions:   *maxSessions,
		SessionTTL:    *sessionTTL,
		StateStore:    *stateStore,
		RedisAddr:     *redisAddr,
		RedisPassword: *redisPassword,
	}

	server, err := kernelhub.NewServer(config)
	if err != nil {
		klog.Fatalf("Failed to create server: %v", err)
	}

	if *cleanupInterval > 0 {
		go server.StartCleanup(context.Background(), *cleanupInterval)
	}

	if err := server.Run(); err != nil {
		klog.Fatalf("Server failed: %v", err)
	}
}
