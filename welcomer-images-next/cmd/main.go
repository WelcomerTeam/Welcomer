package main

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"time"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-images-next/service"
)

var pool *service.URLPool

func call() {
	var customWelcomerImage welcomer.CustomWelcomerImage

	_ = json.Unmarshal([]byte(`{"fill": "rainbow", "layers": [{"fill": "ref:019b0a95-a5c0-767b-b7f8-9d928ca4a615", "type": 2, "value": "", "stroke": {"color": "#ffffff00", "width": 0}, "position": [222, 315], "rotation": 0, "dimensions": [759, 206], "inverted_x": false, "inverted_y": false, "border_radius": ["0", "0", "0", "0"]}, {"fill": "#ffffff", "type": 0, "value": "Hello {{User.Name}} ", "stroke": {"color": "#000000", "width": 6}, "position": [241, 20], "rotation": 0, "dimensions": [696, 225], "inverted_x": false, "inverted_y": false, "typography": {"font_size": 42, "font_family": "Balsamiq Sans", "font_weight": "regular", "line_height": 1.2, "letter_spacing": 0, "vertical_alignment": "center", "horizontal_alignment": "left"}, "border_radius": ["0", "0", "0", "0"]}, {"fill": "#ffffff", "type": 1, "value": "{{User.Avatar}}", "stroke": {"color": "#FFFFFF", "width": 16}, "position": [49, 55], "rotation": 0, "dimensions": [150, 150], "inverted_x": false, "inverted_y": false, "border_radius": ["100%", "100%", "100%", "100%"]}], "stroke": {"color": "#FFFFFF", "width": 16}, "dimensions": [1000, 300]}`), &customWelcomerImage)

	builder := service.GenerateCanvas(customWelcomerImage)
	html := builder.String()

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, time.Second*5)

	resp, err := service.ScreenshotFromHTML(ctx, pool, html)
	if err != nil {
		panic(err)
	}

	file, _ := os.OpenFile("out.png", os.O_CREATE|os.O_WRONLY, 0o644)

	defer file.Close()

	file.Write(resp)
}

func discoverChromedp(startPort, endPort int64) {
	urls := []string{}

	println("Probing ports")

	for port := startPort; port <= endPort; port++ {
		conn, err := net.Dial("tcp", "127.0.0.1:"+welcomer.Itoa(port))
		if err != nil {
			println("Cannot connect to", port)

			continue
		}

		conn.Close()

		urls = append(urls, "ws://127.0.0.1:"+welcomer.Itoa(port))
	}

	println("Done")

	pool = service.NewURLPool(urls)
}

func main() {
	discoverChromedp(15000, 15005)
	call()
}
