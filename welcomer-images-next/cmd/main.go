package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-images-next/service"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")

	prometheusAddress := flag.String("prometheusAddress", os.Getenv("IMAGE_PROMETHEUS_ADDRESS"), "Prometheus address")
	imageHost := flag.String("host", os.Getenv("IMAGE_HOST"), "Host to serve the image service interface from")
	proxyAddress := flag.String("proxyAddress", os.Getenv("IMAGE_PROXY_ADDRESS"), "Proxy address for requests")

	chromedpStartPort := flag.Int64("chromedpStartPort", welcomer.TryParseInt(os.Getenv("CHROMEDP_START_PORT")), "Start port for chromedp instances")
	chromedpEndPort := flag.Int64("chromedpEndPort", welcomer.TryParseInt(os.Getenv("CHROMEDP_END_PORT")), "End port for chromedp instances")
	chromedpServicePrefix := flag.String("chromedpServicePrefix", welcomer.Coalesce(os.Getenv("CHROMEDP_SERVICE_PREFIX"), "127.0.0.1"), "Hostname for chromedp services")

	chromedpServiceHost := flag.String("chromedpServiceHost", os.Getenv("CHROMEDP_SERVICE_HOST"), "Optional host to use for chromedp which will disable load balancing and directly use it instead")

	releaseMode := flag.String("ginMode", os.Getenv("GIN_MODE"), "gin mode (release/debug)")
	debug := flag.Bool("debug", false, "When enabled, images will be saved to a file.")

	flag.Parse()

	gin.SetMode(*releaseMode)

	welcomer.SetupLogger(*loggingLevel)

	imageService, err := service.NewImageService(service.ImageServiceOptions{
		Debug:             *debug,
		Host:              *imageHost,
		PrometheusAddress: *prometheusAddress,
		ProxyAddress:      *proxyAddress,
	})
	if err != nil {
		panic(err)
	}

	if *chromedpServiceHost != "" {
		imageService.URLPool = service.NewHardcodedPool(*chromedpServiceHost)
	} else {
		imageService.URLPool = discoverChromedp(*chromedpServicePrefix, *chromedpStartPort, *chromedpEndPort)
	}

	imageService.Open()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-signalCh

	if err = imageService.Close(); err != nil {
		welcomer.Logger.Warn().Err(err).Msg("Exception whilst closing image service")
	}
}

func discoverChromedp(servicePrefix string, startPort, endPort int64) service.Pool {
	urls := []string{}

	welcomer.Logger.Info().Int64("start", startPort).Int64("end", endPort).Msg("Probing chromedp ports")

	for port := startPort; port <= endPort; port++ {
		host := servicePrefix + ":" + welcomer.Itoa(port)

		ok, browser := validateChromedpInstance(host)
		if !ok {
			welcomer.Logger.Info().Str("host", host).Msg("No chromedp instance found host")

			continue
		}

		welcomer.Logger.Info().Str("host", host).Str("browser", browser).Msg("Found chromedp instance")

		urls = append(urls, "ws://"+host)
	}

	if len(urls) == 0 {
		welcomer.Logger.Panic().Msg("No chromedp instances found")
	}

	welcomer.Logger.Info().Int("count", len(urls)).Msg("Discovered chromedp instances")

	return service.NewURLPool(urls)
}

func validateChromedpInstance(host string) (bool, string) {
	req, err := http.NewRequest(http.MethodGet, "http://"+host+"/json/version", nil)
	if err != nil {
		return false, ""
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, ""
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, ""
	}

	var cv struct {
		Browser string `json:"Browser"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&cv); err != nil {
		return false, ""
	}

	if cv.Browser == "" {
		return false, ""
	}

	return true, cv.Browser
}
