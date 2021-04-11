package welcomerimages

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"math"
	"mime"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/WelcomerTeam/WelcomerImages/pkg/methodrouter"
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"github.com/kalafut/imohash"
	"github.com/rs/zerolog"
	gotils "github.com/savsgio/gotils/strconv"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// HandleRequest handles the HTTP requests.
func (wi *WelcomerImageService) HandleRequest(ctx *fasthttp.RequestCtx) {
	var processingMS int64

	start := time.Now().UTC()
	path := gotils.B2S(ctx.Path())

	defer func() {
		var zlog *zerolog.Event

		statusCode := ctx.Response.StatusCode()

		switch {
		case (statusCode >= 400 && statusCode <= 499):
			zlog = wi.Logger.Warn()
		case (statusCode >= 500 && statusCode <= 599):
			zlog = wi.Logger.Error()
		default:
			zlog = wi.Logger.Info()
		}

		zlog.Msgf("%s %s %s %d %d %dms",
			ctx.RemoteAddr(),
			ctx.Request.Header.Method(),
			ctx.Request.URI().PathOriginal(),
			statusCode,
			len(ctx.Response.Body()),
			processingMS,
		)

		cdnResponseCode.WithLabelValues(
			gotils.B2S(ctx.Request.Header.Method()),
			strconv.Itoa(statusCode),
		).Add(1)
		cdnResponseTimes.Observe(float64(processingMS) / 1000)
	}()

	fasthttp.CompressHandlerBrotliLevel(func(ctx *fasthttp.RequestCtx) {
		fasthttpadaptor.NewFastHTTPHandler(wi.Router)(ctx)
		if path == "/" {
			ctx.Response.Header.Set("Content-Type", "text/html")
			ctx.SendFile(wi.Configuration.Store.IndexLocation)
			return
		}

		// If there is no URL in router then try serving from the dist
		// folder.
		if ctx.Response.StatusCode() == http.StatusNotFound &&
			path != "/" && wi.Configuration.Store.StaticPath != "" {
			ctx.Response.Reset()
			wi.distHandler(ctx)
		}
	}, fasthttp.CompressBrotliDefaultCompression, fasthttp.CompressDefaultCompression)(ctx)

	processingMS = int64(math.Max(float64(time.Since(start).Round(time.Millisecond).Milliseconds()), 1))
	ctx.Response.Header.Set("X-Elapsed", strconv.FormatInt(processingMS, 10))
}

// GET /images/{guid}
// Retrieves an image based off of its guid. This does not require Authorization.

// Input headers:
// Output headers:

// ImagesGet handles retrieving images
func ImagesGet(wi *WelcomerImageService) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var err error

		vars := mux.Vars(r)

		guid, ok := vars["guid"]
		if !ok {
			http.Error(rw, "{'message': 'Missing \"guid\" argument'}", http.StatusBadRequest)
			return
		}

		var d ImageData
		var v []byte

		// removes .png, .gif etc.
		guid = guid[0 : len(guid)-len(filepath.Ext(guid))]

		wi.Database.View(func(tx *bolt.Tx) error {
			v = tx.Bucket(bucketName).Get(gotils.S2B(guid))
			return nil
		})

		if v != nil {
			err = json.Unmarshal(v, &d)
			if err != nil {
				wi.Logger.Error().Err(err).Str("key", guid).Msg("Invalid data received from Bolt")
				d = wi.DefaultImage
			}

			d.Path = path.Join(wi.Configuration.Store.StorePath, d.Path)
		} else {
			d = wi.DefaultImage
		}

		if !d.isDefault && fsExists(d.Path) {
			f, err := ioutil.ReadFile(d.Path)
			if err == nil {
				rw.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(d.Path)))
				rw.WriteHeader(200)
				rw.Write(f)

				return
			} else {
				wi.Logger.Error().Err(err).Msg("Failed to read file")
			}
		}

		rw.Header().Set("Content-Type", "image/png")
		rw.WriteHeader(200)
		rw.Write(wi.DefaultImageContent)
	}
}

// GET /images/{guid}
// Creates a new image and returns either a hotlink or the image.
// This does not require Authorization if AllowAnonymousAccess is enabled
// else the Authorization header must match an APIKey in the configuration.

// Input headers:
//		Authorization: APIKey
// Body: Refer to ImageCreateArguments structure
// Output headers:
//      X-Gen-Elapsed: Time taken to generate image

// ImagesCreate handles creating images
func ImagesCreate(wi *WelcomerImageService) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		// Ensures we finish any active ImagesCreate requests
		// before closing down.
		wi.PoolWaiter.Add(1)
		defer wi.PoolWaiter.Done()

		var rd ImageCreateArguments

		err := json.NewDecoder(r.Body).Decode(&rd)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		apiKey := r.Header.Get("Authorization")

		if !wi.Configuration.Store.AllowAnonymousAccess {
			for _, v := range wi.Configuration.APIKeys {
				if v == apiKey {
					goto postAuth
				}
			}

			rw.WriteHeader(http.StatusUnauthorized)
			return
		}

	postAuth:
		var b bytes.Buffer

		start := time.Now().UTC()
		ext, err := wi.GenerateImage(&b, rd.Options)
		now := time.Now().UTC()

		ms := now.Sub(start).Round(time.Millisecond).Milliseconds()

		rw.Header().Set("X-Gen-Elapsed", strconv.FormatInt(ms, 10))

		if err != nil {
			wi.Logger.Error().Err(err).
				Msg("Failed to generate image")

			resp := ImageCreateResponse{
				Success: false,
				Message: err.Error(),
			}

			res, err := json.Marshal(resp)
			if err != nil {
				wi.Logger.Error().Err(err).Msg("Caught exception marshalling response")
			}

			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write(res)
			return
		}

		imagesProcessed.Inc()
		imagesProcessTime.Observe(float64(ms) / 1000)
		imagesTotalSize.Add(float64(b.Len()))

		// u := md5.New()
		// k := strings.ReplaceAll(base64.URLEncoding.EncodeToString(u.Sum(gotils.S2B(
		// 	strconv.FormatInt(rd.Options.GuildId, 10) +
		// 		strconv.FormatInt(rd.Options.UserId, 10) +
		// 		strconv.FormatInt(time.Now().Unix(), 10),
		// ))[:]), "=", "")

		u := imohash.New()
		by := u.Sum(b.Bytes())
		k := strings.TrimRight(base64.URLEncoding.EncodeToString(by[:]), "=")

		storeName := k + "." + ext
		storePath := path.Join(wi.Configuration.Store.StorePath, storeName)
		storeImage := rd.ForceCache || b.Len() >= rd.FilesizeLimit

		wi.Logger.Info().
			Bool("store", storeImage).
			Str("path", storePath).
			Int64("ms", ms).
			Msg("Generated image")

		if !storeImage {
			rw.WriteHeader(http.StatusOK)
			_, err = b.WriteTo(rw)
			if err != nil {
				wi.Logger.Error().Err(err).Msg("Failed to write image to body")
			}

			return
		}

		err = ioutil.WriteFile(storePath, b.Bytes(), 0644)
		if err != nil {
			wi.Logger.Error().Err(err).
				Str("path", storePath).
				Msg("Failed to write file")

			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		wi.Logger.Debug().
			Str("path", storePath).
			Msg("Successfuly written file")

		d := ImageData{
			ID:        k,
			GuildID:   rd.Options.GuildId,
			Size:      b.Len(),
			Path:      storeName,
			CreatedAt: now,
			ExpiresAt: now.Add(imageTTL),
		}

		v, err := json.Marshal(d)
		if err != nil {
			wi.Logger.Error().Err(err).Msg("Caught exception marshaling ImageData")
		}

		err = wi.Database.Update(func(tx *bolt.Tx) error {
			return tx.Bucket(bucketName).Put(gotils.S2B(k), v)
		})
		if err != nil {
			wi.Logger.Error().Err(err).Msg("Caught exception inserting into Bolt")
		}

		resp := ImageCreateResponse{
			Success:      true,
			Bookmarkable: wi.Configuration.HTTP.BookmarkableURL + "/images/" + d.ID + "." + ext,
			ImageData:    &d,
		}

		res, err := json.Marshal(resp)
		if err != nil {
			wi.Logger.Error().Err(err).Msg("Caught exception marshalling response")
		}

		rw.WriteHeader(http.StatusOK)
		rw.Write(res)
	}
}

type ImageCreateArguments struct {
	// cache = force_cache or filesize >= filesize_limit
	ForceCache    bool      `json:"force_cache"`
	FilesizeLimit int       `json:"filesize_limit"`
	Options       ImageOpts `json:"options"`
}

type ImageCreateResponse struct {
	Success      bool       `json:"success"`
	Message      string     `json:"message,omitempty"`
	Bookmarkable string     `json:"bookmarkable,omitempty"`
	ImageData    *ImageData `json:"image_data,omitempty"`
}

func createEndpoints(wi *WelcomerImageService) *methodrouter.MethodRouter {
	router := methodrouter.NewMethodRouter()

	router.HandleFunc("/images/{guid}", ImagesGet(wi), "GET") // Retrieves image
	router.HandleFunc("/images", ImagesCreate(wi), "POST")    // Creates image

	return router
}
