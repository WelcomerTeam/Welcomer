package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-images-proxy/service"
)

func main() {
	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")

	proxyHost := flag.String("host", os.Getenv("IMAGE_PROXY_HOST"), "Host to serve the image proxy service interface from")

	debug := flag.Bool("debug", false, "When enabled, requests will be logged")

	flag.Parse()

	var err error

	welcomer.SetupLogger(*loggingLevel)

	// Proxy Service initialization
	var proxyService *service.ProxyService
	if proxyService, err = service.NewProxyService(service.ProxyServiceOptions{
		Debug: *debug,
		Host:  *proxyHost,
	}); err != nil {
		welcomer.Logger.Panic().Err(err).Msg("Cannot create image proxy service")
	}

	proxyService.Open()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-signalCh

	proxyService.Close()
}
