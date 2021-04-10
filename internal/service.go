package welcomerimages

import (
	"encoding/json"
	"fmt"
	"image"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	limiter "github.com/WelcomerTeam/WelcomerImages/pkg/limiter"
	methodrouter "github.com/WelcomerTeam/WelcomerImages/pkg/methodrouter"
	"github.com/boltdb/bolt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	gotils "github.com/savsgio/gotils/strconv"
	"github.com/tevino/abool"
	"github.com/ultimate-guitar/go-imagequant"
	"github.com/valyala/fasthttp"
	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/xerrors"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v2"
)

// VERSION respects semantic versioning.
const VERSION = "0.1.2+100420210318"

const (
	ConfigurationPath         = "welcomerimages.yaml"
	ErrOnConfigurationFailure = true

	// Default TTL for images in store
	imageTTL = time.Hour * 24 * 7

	distCacheDuration = 1

	fontCacheTTL       = time.Minute * 30
	profileCacheTTL    = time.Minute * 2
	backgroundCacheTTL = time.Minute * 5

	avatarRoot = "https://cdn.discordapp.com/avatars/%[1]d/%[2]s.png?size=256"
)

var (
	attr, _    = imagequant.NewAttributes()
	bucketName = []byte("images")

	quantizationLimiter *limiter.ConcurrencyLimiter

	fontCacheSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "welcomerimages_font_cache_total",
		Help: "The total number of cached fonts",
	})

	backgroundCacheSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "welcomerimages_background_cache_total",
		Help: "The total number of cached backgrounds",
	})

	profileCacheSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "welcomerimages_profile_cache_total",
		Help: "The total number of cached profile avatars",
	})

	imagesProcessed = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "welcomerimages_create_total",
		Help: "The total number of images created",
	})

	imagesProcessTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "welcomerimages_create_seconds",
		Help:    "The total time to create a static image",
		Buckets: prometheus.LinearBuckets(0, 0.025, 20),
	})

	imagesTotalSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "welcomerimages_size_bytes_total",
		Help: "The total size of images ever made",
	})

	imagesStoreCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "welcomerimages_store_total",
		Help: "The total number of images stored",
	})

	imagesStoreSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "welcomerimages_store_bytes_total",
		Help: "The size in megabytes of reported used space",
	})

	imagesFolderCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "welcomerimages_folder_total",
		Help: "The total number of images actually stored",
	})

	imagesFolderSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "welcomerimages_folder_bytes_total",
		Help: "The size in megabytes of used space",
	})

	cdnResponseCode = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "welcomerimages_serve_code",
			Help: "The response code for HTTP requests",
		},
		[]string{
			"method",
			"code",
		},
	)

	cdnResponseTimes = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "welcomerimages_server_time",
		Help:    "The elapsed time of HTTP requests in milliseconds",
		Buckets: prometheus.ExponentialBuckets(1, 2, 15),
	})

	freedImages = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "welcomerimages_freed_total",
		Help: "Total number of images that have expired",
	})

	freedDuration = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "welcomerimages_freed_duration",
		Help: "Total time in milliseconds to check for images to free",
	})

	imageProfileResponseTimes = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "welcomerimages_profilec_seconds",
		Help:    "The elapsed time of fetching avatar HTTP requests in milliseconds",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	})

	imageProfileResponseCodes = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "welcomerimages_profilec_code",
			Help: "The response code for fetched avatar HTTP requests",
		},
		[]string{
			"code",
		},
	)
)

// WelcomerImageConfiguration represents the configuration of the service.
type WelcomerImageConfiguration struct {
	Logging struct {
		Level                 string `json:"level" yaml:"level"`
		ConsoleLoggingEnabled bool   `json:"console_logging" yaml:"console_logging"`
		FileLoggingEnabled    bool   `json:"file_logging" yaml:"file_logging"`

		EncodeAsJSON bool `json:"encode_as_json" yaml:"encode_as_json"` // Make the framework log as json

		Directory  string `json:"directory" yaml:"directory"`     // Directory to log into.
		Filename   string `json:"filename" yaml:"filename"`       // Name of logfile.
		MaxSize    int    `json:"max_size" yaml:"max_size"`       // Size in MB before a new file.
		MaxBackups int    `json:"max_backups" yaml:"max_backups"` // Number of files to keep.
		MaxAge     int    `json:"max_age" yaml:"max_age"`         // Number of days to keep a logfile.

		MinimalWebhooks bool `json:"minimal_webhooks" yaml:"minimal_webhooks"`
		// If enabled, webhooks for status changes will use one liners instead of an embed.
	} `json:"logging" yaml:"logging"`

	HTTP struct {
		Host            string `json:"host" yaml:"host"`
		BookmarkableURL string `json:"bookmarkable_url" yaml:"bookmarkable_url"`
	} `json:"http" yaml:"http"`

	Store struct {
		// Path that serves static files
		StaticPath string `json:"static_path" yaml:"static_path"`

		// Path for images to be stored in
		StorePath string `json:"store_path" yaml:"store_path"`

		// Path of backgrounds that persist throughout the lifespan of the service
		StaticBackgroundsPath string `json:"static_backgrounds_path" yaml:"static_backgrounds_path"`

		// Path for custom backgrounds to be stored in
		BackgroundsPath string `json:"backgrounds_path" yaml:"backgrounds_path"`

		// Name of StaticBackground to serve if failed to load custom one
		BackgroundFallback string `json:"background_fallback" yaml:"background_fallback"`

		// Image to serve if no image was found
		DefaultImageLocation string `json:"default_image_location" yaml:"default_image_location"`

		// Location of index folder to show on home page
		IndexLocation string `json:"index_location" yaml:"index_location"`

		// Does not require API key to use the image generation endpoints. Useful when localhost only.
		AllowAnonymousAccess bool `json:"allow_anonymous_access" yaml:"allow_anonymous_access"`

		// Location of embedded KV store.
		BoltDBLocation string `json:"bolt_db_location" yaml:"bolt_db_location"`
	} `json:"store" yaml:"store"`

	Prometheus struct {
		Enabled bool   `json:"enabled" yaml:"enabled"`
		Host    string `json:"host" yaml:"host"`
	} `json:"prometheus" yaml:"prometheus"`

	Internal struct {
		ConcurrentQuantizers int `json:"concurrent_quantizers" yaml:"concurrent_quantizers"`

		QuantizerSpeed      int `json:"quantizer_speed" yaml:"quantizer_speed"`
		QuantizerQualityMin int `json:"quantizer_quality_min" yaml:"quantizer_quality_min"`
		QuantizerQualityMax int `json:"quantizer_quality_max" yaml:"quantizer_quality_max"`
	}

	APIKeys []string `json:"api_keys" yaml:"api_keys"`
}

// WelcomerImageService stores caches and any analytical data
type WelcomerImageService struct {
	Logger zerolog.Logger `json:"-"`

	Start time.Time `json:"uptime"`

	Configuration *WelcomerImageConfiguration `json:"configuration"`

	PoolConcurrency limiter.ConcurrencyLimiter `json:"-"`
	PoolWaiter      sync.WaitGroup             `json:"-"`

	ServiceClosing abool.AtomicBool `json:"-"`

	distHandler fasthttp.RequestHandler `json:"-"`
	fs          *fasthttp.FS            `json:"-"`

	Router *methodrouter.MethodRouter `json:"-"`

	Database *bolt.DB `json:"-"`

	DefaultImage        ImageData
	DefaultImageContent []byte

	FontCacheMu sync.RWMutex
	FontCache   map[string]*FontCache

	BackgroundCacheMu sync.RWMutex
	BackgroundCache   map[string]*ImageCache

	StaticBackgroundCache map[string]*ImageCache
	// TODO: Add StaticFontCache. Loaded from autoload-fonts (configurable) + Add fonts folder (for non autoload)

	ProfileCacheMu sync.RWMutex
	ProfileCache   map[int64]*RequestCache
}

type ImageData struct {
	ID        string    `json:"id" msgpack:"i"`
	GuildID   int64     `json:"guild_id" msgpack:"g"`
	Size      int       `json:"size" msgpack:"s"`
	Path      string    `json:"path" msgpack:"p"`
	ExpiresAt time.Time `json:"expires_at" msgpack:"e"`
	CreatedAt time.Time `json:"created_at" msgpack:"c"`

	isDefault bool
}

// FontCache stores the Font, last accessed and Faces for different sizes.
type FontCache struct {
	LastAccessedMu sync.RWMutex
	LastAccessed   time.Time

	Font *sfnt.Font

	FaceCacheMu sync.RWMutex
	FaceCache   map[float64]*FaceCache
}

// LastAccess stores the last access of the structure
type LastAccess struct {
	LastAccessed   time.Time
	LastAccessedMu sync.RWMutex
}

// FaceCache stores the Face and when it was last accessed.
type FaceCache struct {
	LastAccess

	Face *font.Face
}

// FileCache stores the file body and when it was last accessed.
type FileCache struct {
	LastAccess

	Filename string
	Ext      string
	Path     string
	Body     []byte
}

// ImageCache stores the image and the extention for it.
type ImageCache struct {
	sync.RWMutex

	LastAccess

	// The image format that is represented
	Format string

	Frames []image.Image

	// Config is the global color table (palette), width and height. A nil or
	// empty-color.Palette Config.ColorModel means that each frame has its own
	// color table and there is no global color table.
	Config image.Config

	// The successive delay times, one per frame, in 100ths of a second.
	Delay []int

	// LoopCount controls the number of times an animation will be
	// restarted during display.
	LoopCount int

	// Disposal is the successive disposal methods, one per frame.
	Disposal []byte

	// BackgroundIndex is the background index in the global color table, for
	// use with the DisposalBackground disposal method.
	BackgroundIndex byte
}

// RequestCache stores the request body and when it was last accessed.
type RequestCache struct {
	LastAccess

	URL  string
	Body []byte
}

// NewService creates the Welcomer Image service and intializes it.
func NewService(logger io.Writer) (wi *WelcomerImageService, err error) {
	wi = &WelcomerImageService{
		Logger: zerolog.New(logger).With().Timestamp().Logger(),

		Configuration: &WelcomerImageConfiguration{},

		PoolWaiter: sync.WaitGroup{},

		FontCacheMu: sync.RWMutex{},
		FontCache:   make(map[string]*FontCache),

		BackgroundCacheMu: sync.RWMutex{},
		BackgroundCache:   make(map[string]*ImageCache),

		StaticBackgroundCache: make(map[string]*ImageCache),

		ProfileCacheMu: sync.RWMutex{},
		ProfileCache:   make(map[int64]*RequestCache),
	}

	configuration, err := wi.LoadConfiguration(ConfigurationPath)
	if err != nil {
		return nil, xerrors.Errorf("new service: %w", err)
	}

	wi.Configuration = configuration
	wi.DefaultImage = ImageData{
		ID:        "",
		Path:      wi.Configuration.Store.DefaultImageLocation,
		ExpiresAt: time.Date(2100, time.January, 1, 0, 0, 0, 0, time.UTC),
		CreatedAt: time.Date(2021, time.April, 9, 1, 37, 0, 0, time.UTC),

		isDefault: true,
	}

	wi.DefaultImageContent, err = ioutil.ReadFile(wi.DefaultImage.Path)
	if err != nil {
		return nil, xerrors.Errorf("new service read default: %w", err)
	}

	var writers []io.Writer

	zlLevel, err := zerolog.ParseLevel(wi.Configuration.Logging.Level)
	if err != nil {
		wi.Logger.Warn().
			Str("lvl", wi.Configuration.Logging.Level).
			Msg("Current zerolog level provided is not valid")
	} else {
		wi.Logger.Info().
			Str("lvl", wi.Configuration.Logging.Level).
			Msg("Changed logging level")
		zerolog.SetGlobalLevel(zlLevel)
	}

	if wi.Configuration.Logging.ConsoleLoggingEnabled {
		writers = append(writers, logger)
	}

	if wi.Configuration.Logging.FileLoggingEnabled {
		if err := os.MkdirAll(wi.Configuration.Logging.Directory, 0o744); err != nil {
			log.Error().Err(err).Str("path", wi.Configuration.Logging.Directory).Msg("Unable to create log directory")
		} else {
			lumber := &lumberjack.Logger{
				Filename:   path.Join(wi.Configuration.Logging.Directory, wi.Configuration.Logging.Filename),
				MaxBackups: wi.Configuration.Logging.MaxBackups,
				MaxSize:    wi.Configuration.Logging.MaxSize,
				MaxAge:     wi.Configuration.Logging.MaxAge,
			}

			if wi.Configuration.Logging.EncodeAsJSON {
				writers = append(writers, lumber)
			} else {
				writers = append(writers, zerolog.ConsoleWriter{
					Out:        lumber,
					TimeFormat: time.Stamp,
					NoColor:    true,
				})
			}
		}
	}

	mw := io.MultiWriter(writers...)
	wi.Logger = zerolog.New(mw).With().Timestamp().Logger()
	wi.Logger.Info().Msg("Logging configured")

	return wi, nil
}

// LoadConfiguration loads the service configuration.
func (wi *WelcomerImageService) LoadConfiguration(path string) (configuration *WelcomerImageConfiguration, err error) {
	wi.Logger.Debug().Msg("Loading configuration")

	defer func() {
		if err == nil {
			wi.Logger.Info().Msg("Configuration loaded")
		}
	}()

	file, err := ioutil.ReadFile(path)
	if err != nil {
		if ErrOnConfigurationFailure {
			return configuration, xerrors.Errorf("load configuration readfile: %w", err)
		}

		wi.Logger.Warn().Msg("Failed to read configuration but ErrOnConfigurationFailure is disabled")
	}

	configuration = &WelcomerImageConfiguration{}
	err = yaml.Unmarshal(file, &configuration)

	if err != nil {
		if ErrOnConfigurationFailure {
			return configuration, xerrors.Errorf("load configuration unmarshal: %w", err)
		}

		wi.Logger.Warn().Msg("Failed to unmarshal configuration but ErrOnConfigurationFailure is disabled")
	}

	return configuration, err
}

// Opens starts up the services and loads the configuration and starts up the HTTP server
func (wi *WelcomerImageService) Open() (err error) {
	wi.Start = time.Now().UTC()
	wi.Logger.Info().Msgf("Starting Welcomer Image Service")

	wi.Logger.Info().Str("staticbackgrounds", wi.Configuration.Store.StaticBackgroundsPath).Send()
	wi.Logger.Info().Str("custombackgrounds", wi.Configuration.Store.BackgroundsPath).Send()
	wi.Logger.Info().Str("store", wi.Configuration.Store.StorePath).Send()
	wi.Logger.Info().Str("bolt", wi.Configuration.Store.BoltDBLocation).Send()
	wi.Logger.Info().Str("static", wi.Configuration.Store.StaticPath).Send()

	if !fsExists(wi.Configuration.Store.StaticBackgroundsPath) {
		return xerrors.New("Static backgrounds folder does not exist")
	}

	if !fsExists(wi.Configuration.Store.BackgroundsPath) {
		return xerrors.New("Backgrounds folder does not exist")
	}

	if !fsExists(wi.Configuration.Store.StorePath) {
		return xerrors.New("Store folder does not exist")
	}

	if !fsExists(wi.Configuration.Store.StaticPath) {
		wi.Logger.Warn().Msg("Static folder does not exist")
	}

	if wi.Configuration.Store.AllowAnonymousAccess {
		wi.Logger.Warn().Msg("Anonymous access is enabled")
	}

	// Pseudo allow an infinite number of concurrency
	if wi.Configuration.Internal.ConcurrentQuantizers == 0 {
		wi.Configuration.Internal.ConcurrentQuantizers = 1024
		wi.Logger.Info().Msg("ConcurrentQuantizers was set to 0. Limiter has been set to 1024")
	}

	quantizationLimiter = limiter.NewConcurrencyLimiter(
		"",
		wi.Configuration.Internal.ConcurrentQuantizers,
	)
	attr.SetQuality(
		wi.Configuration.Internal.QuantizerQualityMin,
		wi.Configuration.Internal.QuantizerQualityMax,
	)
	attr.SetSpeed(
		wi.Configuration.Internal.QuantizerSpeed,
	)

	wi.Logger.Info().
		Msg("Releasing Bolt lock")
	db, err := bolt.Open(wi.Configuration.Store.BoltDBLocation, 0600, nil)
	if err != nil {
		return xerrors.Errorf("open service: %w", err)
	}

	wi.Logger.Info().
		Msg("Bolt unlocked")

	wi.Database = db

	tx, err := wi.Database.Begin(true)
	if err != nil {
		return xerrors.Errorf("open service begin: %w", err)
	}

	tx.CreateBucketIfNotExists([]byte("images"))

	err = tx.Commit()
	if err != nil {
		return xerrors.Errorf("open service commit: %w", err)
	}

	if wi.Configuration.Prometheus.Enabled {
		wi.Logger.Info().
			Str("path", wi.Configuration.Prometheus.Host+"/metrics").
			Msg("Starting up Prometheus handler")

		go func() {
			http.Handle("/metrics", promhttp.Handler())

			fmt.Printf("Serving prom at %s (Press CTRL+C to quit)\n", wi.Configuration.Prometheus.Host)

			err := http.ListenAndServe(wi.Configuration.Prometheus.Host, nil)
			if err != nil {
				wi.Logger.Error().
					Str("host", wi.Configuration.Prometheus.Host).Err(err).
					Msg("Failed to serve prometheus server")
			}
		}()
	}

	wi.Logger.Info().Msg("Loading static backgrounds")

	files, err := ioutil.ReadDir(wi.Configuration.Store.StaticBackgroundsPath)
	if err != nil {
		wi.Logger.Error().Err(err).
			Str("path", wi.Configuration.Store.StaticBackgroundsPath).
			Msg("Failed to list files in static backgrounds folder")
	}

	for _, f := range files {
		wi.Logger.Debug().
			Str("path", path.Join(wi.Configuration.Store.StaticBackgroundsPath, f.Name())).
			Msg("Loading static image")
		fimg, err := os.Open(path.Join(wi.Configuration.Store.StaticBackgroundsPath, f.Name()))
		if err != nil {
			wi.Logger.Error().Err(err).
				Str("file", f.Name()).
				Msg("Failed to open static image")

			continue
		}

		img, format, err := image.Decode(fimg)
		if err != nil {
			wi.Logger.Error().Err(err).
				Str("file", f.Name()).
				Msg("Failed to decode static image")

			continue
		}

		fimg.Seek(0, io.SeekStart)
		config, _, err := image.DecodeConfig(fimg)
		if err != nil {
			wi.Logger.Error().Err(err).
				Str("file", f.Name()).
				Msg("Failed to decode static image config")

			continue
		}

		name := f.Name()
		name = name[0 : len(name)-len(filepath.Ext(name))]

		wi.StaticBackgroundCache[name] = &ImageCache{
			Format: format,
			Frames: []image.Image{img},
			Config: config,
		}
	}

	wi.Logger.Info().Msgf(
		"Loaded %d/%d static backgrounds",
		len(wi.StaticBackgroundCache), len(files))

	wi.Logger.Info().Msg("Starting up HTTP server")

	if wi.Configuration.Store.StaticPath != "" {
		wi.Logger.Info().Str("path", wi.Configuration.Store.StaticPath).Msg("Serving files")
		wi.fs = &fasthttp.FS{
			Root:            wi.Configuration.Store.StaticPath,
			Compress:        true,
			CompressBrotli:  true,
			AcceptByteRange: true,
			CacheDuration:   distCacheDuration,
			PathNotFound:    fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) { ctx.WriteString("There is nothing here") }),
		}

		wi.distHandler = wi.fs.NewRequestHandler()
	}

	wi.Logger.Info().Msg("Creating endpoints")
	wi.Router = createEndpoints(wi)

	go func() {
		fmt.Printf("Serving at %s (Press CTRL+C to quit)\n", wi.Configuration.HTTP.Host)

		err = fasthttp.ListenAndServe(wi.Configuration.HTTP.Host, wi.HandleRequest)
		if err != nil {
			wi.Logger.Error().Str("host", wi.Configuration.HTTP.Host).Err(err).Msg("Failed to serve http server")
		}
	}()

	go wi.PrometheusFetcher()

	return nil
}

// PrometheusFetcher fetches extra information such as store usage
func (wi *WelcomerImageService) PrometheusFetcher() {
	var err error
	var d ImageData

	wi.Logger.Info().Msg("Started PrometheusFetcher")

	for {
		wi.Logger.Debug().Msg("Starting fetcher task")

		// Fetch imageCount and imageSize from Bolt and folder

		storeSize := 0
		storeCount := 0

		imageData := make(map[string]ImageData)

		start := time.Now()

		err = wi.Database.View(func(tx *bolt.Tx) error {
			err = tx.Bucket(bucketName).ForEach(func(k []byte, v []byte) error {
				err = json.Unmarshal(v, &d)
				if err != nil {
					wi.Logger.Error().Err(err).
						Str("key", gotils.B2S(k)).
						Msg("Invalid data received from Bolt")
					err = nil
				}

				imageData[gotils.B2S(k)] = d

				storeSize += d.Size
				storeCount += 1

				return err
			})

			return err
		})

		files, err := ioutil.ReadDir(wi.Configuration.Store.StorePath)
		if err != nil {
			wi.Logger.Error().Err(err).
				Str("path", wi.Configuration.Store.StorePath).
				Msg("Encountered error reading Store path")
		}

		folderSize := 0
		folderCount := len(files)

		for _, f := range files {
			folderSize += int(f.Size())
		}

		fi := 0

		removals := make([]string, 0)

		for i, k := range imageData {
			if k.ExpiresAt.Before(start) {
				p := path.Join(wi.Configuration.Store.StorePath, k.Path)

				if fsExists(p) {
					wi.Logger.Info().Msgf("Removing expired image at %s", p)

					err = os.Remove(p)
					if err != nil {
						wi.Logger.Error().Err(err).
							Str("path", p).
							Msg("Failed to remove image")
					}
				} else {
					wi.Logger.Warn().Str("path", p).Str("id", k.ID).
						Msg("Image does not exist however has entry present")
				}

				removals = append(removals, i)
				fi++
			}
		}

		tx, err := wi.Database.Begin(true)
		if err != nil {
			wi.Logger.Error().Err(err).Msg("Failed to begin database tx")
		}

		bucket := tx.Bucket(bucketName)
		for _, removals := range removals {
			err = bucket.Delete(gotils.S2B(removals))
			if err != nil {
				wi.Logger.Warn().Err(err).
					Str("key", removals).
					Msg("Failed to remove key from image bucket")
			}
		}

		err = tx.Commit()
		if err != nil {
			wi.Logger.Error().Err(err).Msg("Failed to commit database changes")
		}

		// free profile cache
		// free font cache

		fd := time.Since(start).Round(time.Millisecond).Milliseconds()
		wi.Logger.Debug().Int64("dur", fd).Int("freed", fi).Msg("Finished freeing images")

		freedImages.Set(float64(fi))
		freedDuration.Set(float64(fd))

		imagesStoreCount.Set(float64(storeCount))
		imagesStoreSize.Set(float64(storeSize))

		imagesFolderCount.Set(float64(folderCount))
		imagesFolderSize.Set(float64(folderSize))

		pcr := make([]int64, 0)
		fcr := make([]string, 0)
		bcr := make([]string, 0)

		// Remove expired Profile entries and count length
		wi.ProfileCacheMu.Lock()
		for k, v := range wi.ProfileCache {
			v.LastAccessedMu.RLock()
			la := v.LastAccessed
			v.LastAccessedMu.RUnlock()

			if start.After(la.Add(profileCacheTTL)) {
				pcr = append(pcr, k)
			}
		}

		for _, k := range pcr {
			delete(wi.ProfileCache, k)
		}

		profileCacheSize.Set(float64(len(wi.ProfileCache)))
		wi.ProfileCacheMu.Unlock()

		// Remove expired Background entries and count length
		wi.BackgroundCacheMu.Lock()
		for k, v := range wi.BackgroundCache {
			v.LastAccessedMu.RLock()
			la := v.LastAccessed
			v.LastAccessedMu.RUnlock()

			if start.After(la.Add(backgroundCacheTTL)) {
				bcr = append(bcr, k)
			}
		}

		for _, k := range bcr {
			delete(wi.BackgroundCache, k)
		}

		backgroundCacheSize.Set(float64(len(wi.BackgroundCache)))
		wi.BackgroundCacheMu.Unlock()

		// Remove expired Font entries and count length
		wi.FontCacheMu.Lock()
		for k, v := range wi.FontCache {
			v.LastAccessedMu.RLock()
			la := v.LastAccessed
			v.LastAccessedMu.RUnlock()

			if start.After(la.Add(profileCacheTTL)) {
				fcr = append(fcr, k)
			}
		}

		for _, k := range fcr {
			delete(wi.FontCache, k)
		}

		fontCacheSize.Set(float64(len(wi.FontCache)))
		wi.FontCacheMu.Unlock()

		time.Sleep(time.Minute)
	}

}

// Close will gracefully close the application and wait for any images being generated
func (wi *WelcomerImageService) Close() (err error) {
	wi.Logger.Info().Msg("Closing Welcomer Image Service. Waiting for any active tasks")

	wi.ServiceClosing.Set()
	wi.PoolWaiter.Wait()

	return
}

// fsExists checks if a file or folder exists and returns if it does
func fsExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// openImage returns an image, format and error
func openImage(path string) (image.Image, *gif.GIF, image.Config, string, error) {
	fi, err := os.Open(path)
	if err != nil {
		return nil, nil, image.Config{}, "", err
	}

	if strings.HasSuffix(path, ".gif") {
		g, err := gif.DecodeAll(fi)
		if err != nil {
			return nil, nil, image.Config{}, "", err
		}

		fi.Seek(0, io.SeekStart)
		cf, err := gif.DecodeConfig(fi)
		if err != nil {
			return nil, nil, image.Config{}, "", err
		}

		return nil, g, cf, "gif", nil
	} else {
		i, f, err := image.Decode(fi)
		if err != nil {
			return nil, nil, image.Config{}, f, err
		}

		fi.Seek(0, io.SeekStart)
		cf, _, err := image.DecodeConfig(fi)
		if err != nil {
			return nil, nil, image.Config{}, f, err
		}

		return i, nil, cf, f, nil
	}
}

// TODO: FetchFont

// FetchBackground fetches a background from its id. Returns the image and boolean indicating GIF
func (wi *WelcomerImageService) FetchBackground(b string, allowGifs bool) (*ImageCache, error) {
	if b == "" {
		b = "default"
	}

	wi.Logger.Debug().
		Str("name", b).
		Bool("allowGifs", allowGifs).
		Msg("Fetching background")

	c, ok := wi.StaticBackgroundCache[b]
	if ok {
		return c, nil
	}

	wi.BackgroundCacheMu.RLock()
	c, ok = wi.BackgroundCache[b]
	wi.BackgroundCacheMu.RUnlock()

	if ok {
		c.LastAccessedMu.Lock()
		c.LastAccessed = time.Now().UTC()
		c.LastAccessedMu.Unlock()

		return c, nil
	}

	p := path.Join(wi.Configuration.Store.BackgroundsPath, b)

	var lp string
	if allowGifs && fsExists(p+".gif") {
		lp = p + ".gif"
	} else {
		lp = p + ".png"
	}

	if !fsExists(lp) {
		wi.Logger.Debug().Str("path", lp).Msg("Could not find background, serving fallback")
		return wi.StaticBackgroundCache[wi.Configuration.Store.BackgroundFallback], nil
	}

	im, gi, config, format, err := openImage(lp)
	if err != nil {
		wi.Logger.Error().Err(err).
			Str("bg", b).
			Str("path", lp).
			Msg("Failed to open file")

		// TODO: Figure out how i want to handle errors in FetchBackground. At the moment
		// we use fallback and treat like there is no error.
		return wi.StaticBackgroundCache[wi.Configuration.Store.BackgroundFallback], nil
	}

	ic := &ImageCache{
		Format: format,
		Config: config,
	}

	// We store as frames reguardless of image format however
	// we should copy over the other GIF data when neccessary.
	if format == "gif" {
		ic.BackgroundIndex = gi.BackgroundIndex
		ic.Delay = gi.Delay
		ic.Disposal = gi.Disposal
		ic.LoopCount = gi.LoopCount

		ic.Frames = make([]image.Image, 0, len(gi.Image))
		for _, frame := range gi.Image {
			ic.Frames = append(ic.Frames, image.Image(frame))
		}
	} else {
		ic.Frames = make([]image.Image, 0, 1)
		ic.Frames = append(ic.Frames, im)
	}

	wi.BackgroundCacheMu.Lock()
	wi.BackgroundCache[b] = ic
	wi.BackgroundCacheMu.Unlock()

	return ic, nil
}

// FetchAvatar fetches an avatar from a user id and avatar hash
func (wi *WelcomerImageService) FetchAvatar(u int64, a string) ([]byte, error) {
	wi.Logger.Debug().
		Int64("user", u).
		Str("hash", a).
		Msg("Fetching avatar")

	wi.ProfileCacheMu.RLock()
	c, ok := wi.ProfileCache[u]
	wi.ProfileCacheMu.RUnlock()

	if ok {
		c.LastAccessedMu.Lock()
		c.LastAccessed = time.Now().UTC()
		c.LastAccessedMu.Unlock()

		return c.Body, nil
	}

	url := fmt.Sprintf(avatarRoot, u, a)

	start := time.Now().UTC()
	s, b, err := fasthttp.Get(
		nil,
		url,
	)

	ms := time.Since(start).Round(time.Millisecond).Milliseconds()
	wi.Logger.Debug().
		Str("url", url).
		Int("code", s).
		Int64("ms", ms).
		Err(err).
		Msg("Fetched avatar")

	if s < 200 || s >= 400 {
		return nil, xerrors.New(fmt.Sprintf("fetchavatar response: %d", s))
	}

	imageProfileResponseTimes.Observe(float64(ms) / 1000)
	imageProfileResponseCodes.WithLabelValues(strconv.Itoa(s)).Inc()

	if err != nil {
		wi.Logger.Error().Err(err).Msg("Failed to retrieve profile picture of user")
		return b, err
	}

	wi.ProfileCacheMu.Lock()
	wi.ProfileCache[u] = &RequestCache{
		LastAccess: LastAccess{
			LastAccessed:   time.Now().UTC(),
			LastAccessedMu: sync.RWMutex{},
		},
		URL:  url,
		Body: b,
	}
	wi.ProfileCacheMu.Unlock()

	return b, nil
}
