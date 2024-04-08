package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/monobearotaku/online-chat-api/internal/di"
)

func main() {
	ctx := context.Background()

	diContainer := di.NewDiContainer(ctx)
	defer diContainer.Stop()

	diContainer.Run()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	<-exit
}
