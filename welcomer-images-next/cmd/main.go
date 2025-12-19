package main

import (
	"context"
	"encoding/json"
	"net"
	"os"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-images-next/service"
)

var pool *service.URLPool

func call() {
	var customWelcomerImage welcomer.CustomWelcomerImage

	// _ = json.Unmarshal([]byte(`{"fill": "solid:profile", "layers": [{"fill": "#ffffff", "type": 0, "value": "Welcome {{User.Name}} to {{Guild.Name}}\nyou are the {{Ordinal(Guild.Members)}} member!", "stroke": {"color": "#000000FF", "width": 3}, "position": [264, 32], "rotation": 0, "dimensions": [704, 236], "inverted_x": false, "inverted_y": false, "typography": {"font_size": 32, "font_family": "Balsamiq Sans", "font_weight": "bold", "line_height": 1.2, "letter_spacing": 0, "vertical_alignment": "center", "horizontal_alignment": "left"}, "border_radius": ["0", "0", "0", "0"]}, {"fill": "#FFFFFFFF", "type": 1, "value": "{{User.Avatar}}", "stroke": {"color": "#FFFFFFFF", "width": 8}, "position": [32, 50], "rotation": 0, "dimensions": [200, 200], "inverted_x": false, "inverted_y": false, "border_radius": ["100%", "100%", "100%", "100%"]}], "stroke": {"color": "#00000000", "width": 0}, "dimensions": [1000, 300]}`), &customWelcomerImage)
	_ = json.Unmarshal([]byte(`{"fill": "rainbow", "layers": [{"fill": "ref:019b0a95-a5c0-767b-b7f8-9d928ca4a615", "type": 2, "value": "", "stroke": {"color": "#ffffff00", "width": 0}, "position": [222, 315], "rotation": 0, "dimensions": [759, 206], "inverted_x": false, "inverted_y": false, "border_radius": ["0", "0", "0", "0"]}, {"fill": "#ffffff", "type": 0, "value": "Hello {{User.Name}} ", "stroke": {"color": "#000000", "width": 6}, "position": [241, 20], "rotation": 0, "dimensions": [696, 225], "inverted_x": false, "inverted_y": false, "typography": {"font_size": 42, "font_family": "Luckiest Guy", "font_weight": "regular", "line_height": 1.2, "letter_spacing": 0, "vertical_alignment": "center", "horizontal_alignment": "left"}, "border_radius": ["0", "0", "0", "0"]}, {"fill": "#ffffff", "type": 1, "value": "{{User.Avatar}}", "stroke": {"color": "#FFFFFF", "width": 16}, "position": [49, 55], "rotation": 0, "dimensions": [150, 150], "inverted_x": false, "inverted_y": false, "border_radius": ["100%", "100%", "100%", "100%"]}], "stroke": {"color": "#FFFFFF", "width": 16}, "dimensions": [1000, 300]}`), &customWelcomerImage)

	builder := service.GenerateCanvas(customWelcomerImage)
	html := builder.String()

	println(html)

	ctx := context.Background()
	// ctx, _ = context.WithTimeout(ctx, time.Second*5)

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
	for port := startPort; port <= endPort; port++ {
		println("Probing port", port)

		conn, err := net.Dial("tcp", "127.0.0.1:"+welcomer.Itoa(port))
		if err != nil {
			continue
		}

		conn.Close()

		urls = append(urls, "ws://127.0.0.1:"+welcomer.Itoa(port))
	}

	pool = service.NewURLPool(urls)
}

func main() {
	discoverChromedp(15000, 15005)
	call()
}
