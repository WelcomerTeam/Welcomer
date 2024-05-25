package main

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"

	_ "github.com/joho/godotenv/autoload"
)

var fonts = []string{
	"balsamiqsans-bold",
	"balsamiqsans-regular",
	"fredokaone-regular",
	"inter-bold",
	"inter-regular",
	"luckiestguy-regular",
	"mada-bold",
	"mada-medium",
	"mina-bold",
	"mina-regular",
	"nunito-bold",
	"nunito-regular",
	"raleway-bold",
	"raleway-regular",
}

var backgrounds = []string{
	// "custom:018c1d0b-83d1-79c0-93c5-f0335df9732e",
	"aesthetics",
	"afterwork",
	"airship",
	"alone",
	"autumn",
	"blue",
	"blurple",
	"clouds",
	"collision",
	"cyan",
	"cybergeek",
	"default",
	"fall",
	"garden",
	"glare",
	"green",
	"lime",
	"lodge",
	"meteorshower",
	"midnightride",
	"neko",
	"nightview",
	"paint",
	"peace",
	"pika",
	"pink",
	"purple",
	"rainbow",
	"red",
	"rem",
	"ribbons",
	"riot",
	"riversource",
	"sea",
	"shards",
	"solarglare",
	"squares",
	"stacks",
	"summer",
	"sun",
	"sunrise",
	"sunset",
	"tanya",
	"unova",
	"upland",
	"utopia",
	"vampire",
	"vectors",
	"wood",
	"yellow",
}

var GRPC_TARGET = os.Getenv("GRPC_HOST")

func main() {
	go doLoadTest()
	go doLoadTest()
	go doLoadTest()
	go doLoadTest()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-signalCh
}

func doLoadTest() {
	for {
		start := time.Now()

		req := getImageRequest()
		jval, _ := json.Marshal(req)

		res, err := http.Post("http://127.0.0.1:4003/generate", "application/json", bytes.NewBuffer(jval))
		if err != nil {
			r, _ := json.Marshal(req)
			println(err.Error(), string(r))
		} else {
			println(res.StatusCode, time.Since(start).Milliseconds(), res.Header.Get("X-Elapsed"))
		}
	}
}

func getImageRequest() core.GenerateImageOptionsRaw {
	return core.GenerateImageOptionsRaw{
		GuildID:            341685098468343822,
		UserID:             143090142360371200,
		AllowAnimated:      randomBool(),
		AvatarURL:          "https://cdn.discordapp.com/avatars/143090142360371200/a73420b217a77a77b17fb42fa7ecfbcc.png?size=1024",
		Theme:              rand.Int31n(2),
		Background:         randomBackground(),
		Text:               "Welcome ImRock\nto the server!",
		TextFont:           randomFont(),
		TextStroke:         randomBool(),
		TextAlign:          rand.Int31n(8),
		TextColor:          0xFFFFFFFF,
		TextStrokeColor:    0xFF000000,
		ImageBorderColor:   0xFFFFFFFF,
		ImageBorderWidth:   rand.Int31n(8) + 8,
		ProfileFloat:       int32(utils.ImageAlignmentLeft),
		ProfileBorderColor: 0xFFFFFFFF,
		ProfileBorderWidth: rand.Int31n(8) + 8,
		ProfileBorderCurve: int32(utils.ImageProfileBorderTypeRounded),
	}
}

func randomBool() bool {
	return rand.Intn(2) == 1
}

func randomBackground() string {
	return backgrounds[rand.Intn(len(backgrounds))]
}

func randomFont() string {
	return fonts[rand.Intn(len(fonts))]
}
