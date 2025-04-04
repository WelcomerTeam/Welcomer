package service

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
)

var assets = map[string]image.Image{
	"brokeimage":    assetsBrokeImageImage,
	"defaultavatar": assetsDefaultAvatarImage,
}

var backgrounds = map[string]image.Image{
	"aesthetics":   backgroundsAestheticsImage,
	"afterwork":    backgroundsAfterworkImage,
	"airship":      backgroundsAirshipImage,
	"alone":        backgroundsAloneImage,
	"autumn":       backgroundsAutumnImage,
	"clouds":       backgroundsCloudsImage,
	"collision":    backgroundsCollisionImage,
	"cybergeek":    backgroundsCybergeekImage,
	"default":      backgroundsDefaultImage,
	"fall":         backgroundsFallImage,
	"garden":       backgroundsGardenImage,
	"glare":        backgroundsGlareImage,
	"lodge":        backgroundsLodgeImage,
	"meteorshower": backgroundsMeteorshowerImage,
	"midnightride": backgroundsMidnightrideImage,
	"mountains":    backgroundsMountainsImage,
	"neko":         backgroundsNekoImage,
	"nightview":    backgroundsNightviewImage,
	"paint":        backgroundsPaintImage,
	"peace":        backgroundsPeaceImage,
	"pika":         backgroundsPikaImage,
	"rainbow":      backgroundsRainbowImage,
	"rem":          backgroundsRemImage,
	"ribbons":      backgroundsRibbonsImage,
	"riot":         backgroundsRiotImage,
	"riversource":  backgroundsRiversourceImage,
	"sea":          backgroundsSeaImage,
	"shards":       backgroundsShardsImage,
	"solarglare":   backgroundsSolarglareImage,
	"spots":        backgroundsSpotsImage,
	"squares":      backgroundsSquaresImage,
	"stacks":       backgroundsStacksImage,
	"summer":       backgroundsSummerImage,
	"sun":          backgroundsSunImage,
	"sunrise":      backgroundsSunriseImage,
	"sunset":       backgroundsSunsetImage,
	"tanya":        backgroundsTanyaImage,
	"unova":        backgroundsUnovaImage,
	"upland":       backgroundsUplandImage,
	"utopia":       backgroundsUtopiaImage,
	"vampire":      backgroundsVampireImage,
	"vectors":      backgroundsVectorsImage,
	"wood":         backgroundsWoodImage,
}

func mustDecodeBytes(n string, src []byte) image.Image {
	res, _, err := image.Decode(bytes.NewBuffer(src))
	if err != nil {
		panic(fmt.Sprintf("image.Decode(%s): %v", n, err.Error()))
	}

	return res
}

//go:embed assets/broke_image.png
var assetsBrokeImageImageBytes []byte
var assetsBrokeImageImage = mustDecodeBytes("assetsBrokeImageImage", assetsBrokeImageImageBytes)

//go:embed assets/default_avatar.png
var assetsDefaultAvatarImageBytes []byte
var assetsDefaultAvatarImage = mustDecodeBytes("assetsDefaultAvatarImage", assetsDefaultAvatarImageBytes)

//go:embed backgrounds/aesthetics.png
var backgroundsAestheticsImageBytes []byte
var backgroundsAestheticsImage = mustDecodeBytes("backgroundsAestheticsImage", backgroundsAestheticsImageBytes)

//go:embed backgrounds/afterwork.png
var backgroundsAfterworkImageBytes []byte
var backgroundsAfterworkImage = mustDecodeBytes("backgroundsAfterworkImage", backgroundsAfterworkImageBytes)

//go:embed backgrounds/airship.png
var backgroundsAirshipImageBytes []byte
var backgroundsAirshipImage = mustDecodeBytes("backgroundsAirshipImage", backgroundsAirshipImageBytes)

//go:embed backgrounds/alone.png
var backgroundsAloneImageBytes []byte
var backgroundsAloneImage = mustDecodeBytes("backgroundsAloneImage", backgroundsAloneImageBytes)

//go:embed backgrounds/autumn.png
var backgroundsAutumnImageBytes []byte
var backgroundsAutumnImage = mustDecodeBytes("backgroundsAutumnImage", backgroundsAutumnImageBytes)

//go:embed backgrounds/clouds.png
var backgroundsCloudsImageBytes []byte
var backgroundsCloudsImage = mustDecodeBytes("backgroundsCloudsImage", backgroundsCloudsImageBytes)

//go:embed backgrounds/collision.png
var backgroundsCollisionImageBytes []byte
var backgroundsCollisionImage = mustDecodeBytes("backgroundsCollisionImage", backgroundsCollisionImageBytes)

//go:embed backgrounds/cybergeek.png
var backgroundsCybergeekImageBytes []byte
var backgroundsCybergeekImage = mustDecodeBytes("backgroundsCybergeekImage", backgroundsCybergeekImageBytes)

//go:embed backgrounds/default.png
var backgroundsDefaultImageBytes []byte
var backgroundsDefaultImage = mustDecodeBytes("backgroundsDefaultImage", backgroundsDefaultImageBytes)

//go:embed backgrounds/fall.png
var backgroundsFallImageBytes []byte
var backgroundsFallImage = mustDecodeBytes("backgroundsFallImage", backgroundsFallImageBytes)

//go:embed backgrounds/garden.png
var backgroundsGardenImageBytes []byte
var backgroundsGardenImage = mustDecodeBytes("backgroundsGardenImage", backgroundsGardenImageBytes)

//go:embed backgrounds/glare.png
var backgroundsGlareImageBytes []byte
var backgroundsGlareImage = mustDecodeBytes("backgroundsGlareImage", backgroundsGlareImageBytes)

//go:embed backgrounds/lodge.png
var backgroundsLodgeImageBytes []byte
var backgroundsLodgeImage = mustDecodeBytes("backgroundsLodgeImage", backgroundsLodgeImageBytes)

//go:embed backgrounds/meteorshower.png
var backgroundsMeteorshowerImageBytes []byte
var backgroundsMeteorshowerImage = mustDecodeBytes("backgroundsMeteorshowerImage", backgroundsMeteorshowerImageBytes)

//go:embed backgrounds/midnightride.png
var backgroundsMidnightrideImageBytes []byte
var backgroundsMidnightrideImage = mustDecodeBytes("backgroundsMidnightrideImage", backgroundsMidnightrideImageBytes)

//go:embed backgrounds/mountains.png
var backgroundsMountainsImageBytes []byte
var backgroundsMountainsImage = mustDecodeBytes("backgroundsMountainsImage", backgroundsMountainsImageBytes)

//go:embed backgrounds/neko.png
var backgroundsNekoImageBytes []byte
var backgroundsNekoImage = mustDecodeBytes("backgroundsNekoImage", backgroundsNekoImageBytes)

//go:embed backgrounds/nightview.png
var backgroundsNightviewImageBytes []byte
var backgroundsNightviewImage = mustDecodeBytes("backgroundsNightviewImage", backgroundsNightviewImageBytes)

//go:embed backgrounds/paint.png
var backgroundsPaintImageBytes []byte
var backgroundsPaintImage = mustDecodeBytes("backgroundsPaintImage", backgroundsPaintImageBytes)

//go:embed backgrounds/peace.png
var backgroundsPeaceImageBytes []byte
var backgroundsPeaceImage = mustDecodeBytes("backgroundsPeaceImage", backgroundsPeaceImageBytes)

//go:embed backgrounds/pika.png
var backgroundsPikaImageBytes []byte
var backgroundsPikaImage = mustDecodeBytes("backgroundsPikaImage", backgroundsPikaImageBytes)

//go:embed backgrounds/rainbow.png
var backgroundsRainbowImageBytes []byte
var backgroundsRainbowImage = mustDecodeBytes("backgroundsRainbowImage", backgroundsRainbowImageBytes)

//go:embed backgrounds/rem.png
var backgroundsRemImageBytes []byte
var backgroundsRemImage = mustDecodeBytes("backgroundsRemImage", backgroundsRemImageBytes)

//go:embed backgrounds/ribbons.png
var backgroundsRibbonsImageBytes []byte
var backgroundsRibbonsImage = mustDecodeBytes("backgroundsRibbonsImage", backgroundsRibbonsImageBytes)

//go:embed backgrounds/riot.png
var backgroundsRiotImageBytes []byte
var backgroundsRiotImage = mustDecodeBytes("backgroundsRiotImage", backgroundsRiotImageBytes)

//go:embed backgrounds/riversource.png
var backgroundsRiversourceImageBytes []byte
var backgroundsRiversourceImage = mustDecodeBytes("backgroundsRiversourceImage", backgroundsRiversourceImageBytes)

//go:embed backgrounds/sea.png
var backgroundsSeaImageBytes []byte
var backgroundsSeaImage = mustDecodeBytes("backgroundsSeaImage", backgroundsSeaImageBytes)

//go:embed backgrounds/shards.png
var backgroundsShardsImageBytes []byte
var backgroundsShardsImage = mustDecodeBytes("backgroundsShardsImage", backgroundsShardsImageBytes)

//go:embed backgrounds/solarglare.png
var backgroundsSolarglareImageBytes []byte
var backgroundsSolarglareImage = mustDecodeBytes("backgroundsSolarglareImage", backgroundsSolarglareImageBytes)

//go:embed backgrounds/spots.png
var backgroundsSpotsImageBytes []byte
var backgroundsSpotsImage = mustDecodeBytes("backgroundsSpotsImage", backgroundsSpotsImageBytes)

//go:embed backgrounds/squares.png
var backgroundsSquaresImageBytes []byte
var backgroundsSquaresImage = mustDecodeBytes("backgroundsSquaresImage", backgroundsSquaresImageBytes)

//go:embed backgrounds/stacks.png
var backgroundsStacksImageBytes []byte
var backgroundsStacksImage = mustDecodeBytes("backgroundsStacksImage", backgroundsStacksImageBytes)

//go:embed backgrounds/summer.png
var backgroundsSummerImageBytes []byte
var backgroundsSummerImage = mustDecodeBytes("backgroundsSummerImage", backgroundsSummerImageBytes)

//go:embed backgrounds/sun.png
var backgroundsSunImageBytes []byte
var backgroundsSunImage = mustDecodeBytes("backgroundsSunImage", backgroundsSunImageBytes)

//go:embed backgrounds/sunrise.png
var backgroundsSunriseImageBytes []byte
var backgroundsSunriseImage = mustDecodeBytes("backgroundsSunriseImage", backgroundsSunriseImageBytes)

//go:embed backgrounds/sunset.png
var backgroundsSunsetImageBytes []byte
var backgroundsSunsetImage = mustDecodeBytes("backgroundsSunsetImage", backgroundsSunsetImageBytes)

//go:embed backgrounds/tanya.png
var backgroundsTanyaImageBytes []byte
var backgroundsTanyaImage = mustDecodeBytes("backgroundsTanyaImage", backgroundsTanyaImageBytes)

//go:embed backgrounds/unova.png
var backgroundsUnovaImageBytes []byte
var backgroundsUnovaImage = mustDecodeBytes("backgroundsUnovaImage", backgroundsUnovaImageBytes)

//go:embed backgrounds/upland.png
var backgroundsUplandImageBytes []byte
var backgroundsUplandImage = mustDecodeBytes("backgroundsUplandImage", backgroundsUplandImageBytes)

//go:embed backgrounds/utopia.png
var backgroundsUtopiaImageBytes []byte
var backgroundsUtopiaImage = mustDecodeBytes("backgroundsUtopiaImage", backgroundsUtopiaImageBytes)

//go:embed backgrounds/vampire.png
var backgroundsVampireImageBytes []byte
var backgroundsVampireImage = mustDecodeBytes("backgroundsVampireImage", backgroundsVampireImageBytes)

//go:embed backgrounds/vectors.png
var backgroundsVectorsImageBytes []byte
var backgroundsVectorsImage = mustDecodeBytes("backgroundsVectorsImage", backgroundsVectorsImageBytes)

//go:embed backgrounds/wood.png
var backgroundsWoodImageBytes []byte
var backgroundsWoodImage = mustDecodeBytes("backgroundsWoodImage", backgroundsWoodImageBytes)
