package service

import (
	"context"
	"fmt"
	"image/color"
	"os"
	"runtime/debug"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	pb "github.com/WelcomerTeam/Welcomer/welcomer-images/protobuf"
)

type routeImageGenerationServiceServer struct {
	pb.ImageGenerationServiceServer

	is *ImageService
}

func (is *ImageService) newImageGenerationServiceServer() *routeImageGenerationServiceServer {
	return &routeImageGenerationServiceServer{
		is: is,
	}
}

func onGRPCRequest() {
	grpcImgenRequests.Inc()
}

func (grpc *routeImageGenerationServiceServer) GenerateImage(ctx context.Context, req *pb.GenerateImageRequest) (response *pb.GenerateImageResponse, err error) {
	defer func() {
		if r := recover(); r != nil {
			errorMessage, ok := r.(error)

			if ok {
				response.BaseResponse.Error = errorMessage.Error()

				grpc.is.Logger.Error().
					Err(errorMessage).
					Interface("request", req).
					Msg("Recovered panic in GenerateImage")
			} else {
				grpc.is.Logger.Error().
					Str("err", "[unknown]").
					Interface("request", req).
					Msg("Recovered panic in GenerateImage")
			}

			fmt.Println(string(debug.Stack()))
		}
	}()

	onGRPCRequest()

	response = &pb.GenerateImageResponse{}
	response.BaseResponse = &pb.BaseResponse{}

	file, format, err := grpc.is.GenerateImage(generateImageRequestToOptions(req))
	if err != nil {
		response.BaseResponse.Error = err.Error()

		return
	}

	if grpc.is.Options.Debug {
		os.WriteFile("output.png", file, 0o644)
	}

	response.File = file
	response.Filetype = format.String()
	response.BaseResponse.Ok = true

	return
}

func generateImageRequestToOptions(req *pb.GenerateImageRequest) GenerateImageOptions {
	return GenerateImageOptions{
		GuildID:            discord.Snowflake(req.GuildID),
		UserID:             discord.Snowflake(req.UserID),
		AllowAnimated:      req.AllowAnimated,
		AvatarURL:          req.AvatarURL,
		Theme:              welcomer.ImageTheme(req.Theme),
		Background:         req.Background,
		Text:               req.Text,
		TextFont:           req.TextFont,
		TextStroke:         formatTextStroke(req.TextStroke),
		TextAlign:          welcomer.ImageAlignment(req.TextAlign),
		TextColor:          convertToRGBA(req.TextColor),
		TextStrokeColor:    convertToRGBA(req.TextStrokeColor),
		ImageBorderColor:   convertToRGBA(req.ImageBorderColor),
		ImageBorderWidth:   int(req.ImageBorderWidth),
		ProfileFloat:       welcomer.ImageAlignment(req.ProfileFloat),
		ProfileBorderColor: convertToRGBA(req.ProfileBorderColor),
		ProfileBorderWidth: int(req.ProfileBorderWidth),
		ProfileBorderCurve: welcomer.ImageProfileBorderType(req.ProfileBorderCurve),
	}
}

func formatTextStroke(v bool) int {
	if v {
		return 4
	}

	return 0
}

func convertToRGBA(int32Color int64) color.RGBA {
	alpha := uint8(int32Color >> 24 & 0xFF)
	red := uint8(int32Color >> 16 & 0xFF)
	green := uint8(int32Color >> 8 & 0xFF)
	blue := uint8(int32Color & 0xFF)

	return color.RGBA{R: red, G: green, B: blue, A: alpha}
}
