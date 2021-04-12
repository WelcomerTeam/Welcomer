package welcomerimages

import (
	"encoding/json"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
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
	"golang.org/x/image/font/opentype"
	"golang.org/x/xerrors"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v2"
)

// VERSION respects semantic versioning.
const VERSION = "0.3+110420211952"

const (
	ConfigurationPath         = "welcomerimages.yaml"
	ErrOnConfigurationFailure = true

	// Default time to keep images in store.
	imageTTL = time.Hour * 24 * 7

	// Default time to cache static files.
	distCacheDuration = time.Hour

	faceCacheTTL       = time.Minute * 15
	profileCacheTTL    = time.Minute * 5
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
	} `json:"logging" yaml:"logging"`

	HTTP struct {
		Host            string `json:"host" yaml:"host"`
		BookmarkableURL string `json:"bookmarkable_url" yaml:"bookmarkable_url"`
	} `json:"http" yaml:"http"`

	Store struct {
		// List of paths for fonts to be stored in
		FontPath []string `json:"font_path" yaml:"font_path"`

		// Path for custom backgrounds to be stored in
		BackgroundsPath string `json:"backgrounds_path" yaml:"backgrounds_path"`

		// Path that serves static files
		StaticPath string `json:"static_path" yaml:"static_path"`

		// Path for images to be stored in
		StorePath string `json:"store_path" yaml:"store_path"`

		// Path of backgrounds that persist throughout the lifespan of the service
		StaticBackgroundsPath string `json:"static_backgrounds_path" yaml:"static_backgrounds_path"`

		// Name of StaticBackground to serve if failed to load custom one
		BackgroundFallback string `json:"background_fallback" yaml:"background_fallback"`

		// Font to serve if the one specified was not found
		DefaultFont string `json:"default_font" yaml:"default_font"`

		// Image to serve if no image was found
		DefaultImageLocation string `json:"default_image_location" yaml:"default_image_location"`

		// Image to serve when backgrounds fail. If empty will return a 500 rather than continue on
		FallbackProfileLocation string `json:"fallback_profile_location" yaml:"fallback_profile_location"`

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

	FallbackFonts []string `json:"fallback_fonts" yaml:"fallback_fonts"`
}

// WelcomerImageService stores caches and any analytical data.
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

	UseFallbackProfile bool
	FallbackProfile    *StaticImageCache

	FallbackFonts []string

	FontCacheMu sync.RWMutex
	FontCache   map[string]*FontCache

	BackgroundCacheMu sync.RWMutex
	BackgroundCache   map[string]*ImageCache

	StaticBackgroundCache map[string]*ImageCache

	ProfileCacheMu sync.RWMutex
	ProfileCache   map[int64]*StaticImageCache
}

// NewService creates the Welcomer Image service and intializes it.
func NewService(logger io.Writer) (wi *WelcomerImageService, err error) {
	wi = &WelcomerImageService{
		Logger: zerolog.New(logger).With().Timestamp().Logger(),

		Configuration: &WelcomerImageConfiguration{},

		PoolWaiter: sync.WaitGroup{},

		FontCacheMu: sync.RWMutex{},
		FontCache:   make(map[string]*FontCache),

		FallbackFonts: make([]string, 0),

		BackgroundCacheMu: sync.RWMutex{},
		BackgroundCache:   make(map[string]*ImageCache),

		StaticBackgroundCache: make(map[string]*ImageCache),

		ProfileCacheMu: sync.RWMutex{},
		ProfileCache:   make(map[int64]*StaticImageCache),
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

	if wi.Configuration.Store.FallbackProfileLocation != "" {
		f, err := os.Open(wi.Configuration.Store.FallbackProfileLocation)
		if err != nil {
			return nil, xerrors.Errorf("new service read fallback: %w", err)
		}

		im, format, err := image.Decode(f)
		if err != nil {
			return nil, xerrors.Errorf("new service decode fallback avatar: %w", err)
		}

		wi.FallbackProfile = &StaticImageCache{
			Format: format,
			Image:  im,
		}

		wi.UseFallbackProfile = true
	}

	wi.DefaultImageContent, err = ioutil.ReadFile(wi.DefaultImage.Path)
	if err != nil {
		return nil, xerrors.Errorf("new service read default bg: %w", err)
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

// Opens starts up the services and loads the configuration and starts up the HTTP server.
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
		wi.Logger.Debug().Msg("ConcurrentQuantizers was set to 0. Limiter has been set to 1024")
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

	wi.Logger.Debug().
		Msg("Releasing Bolt lock")

	db, err := bolt.Open(wi.Configuration.Store.BoltDBLocation, 0o600, nil)
	if err != nil {
		return xerrors.Errorf("open service: %w", err)
	}

	wi.Logger.Debug().
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

	wi.Logger.Debug().Msg("Loading fonts")

	for _, folder := range wi.Configuration.Store.FontPath {
		files, err := ioutil.ReadDir(folder)
		if err != nil {
			wi.Logger.Error().Err(err).
				Str("path", folder).
				Msg("Failed to list files in fonts folder")

			continue
		}

		for _, f := range files {
			name := f.Name()
			path := path.Join(folder, name)
			nametrim := name[0 : len(name)-len(filepath.Ext(name))]

			wi.Logger.Trace().
				Str("path", path).
				Msg("Loading font")

			if _, ok := wi.FontCache[nametrim]; ok {
				wi.Logger.Error().Err(err).
					Str("name", nametrim).
					Msg("Font with name already exists")

				continue
			}

			fb, err := ioutil.ReadFile(path)
			if err != nil {
				wi.Logger.Error().Err(err).
					Str("path", path).
					Msg("Failed to open font")

				continue
			}

			fnt, err := opentype.Parse(fb)
			if err != nil {
				wi.Logger.Error().Err(err).
					Str("path", path).
					Msg("Failed to parse font")

				continue
			}

			wi.FontCache[nametrim] = &FontCache{
				Font:        fnt,
				FaceCacheMu: sync.RWMutex{},
				FaceCache:   make(map[float64]*FaceCache),
			}
		}
	}

	wi.Logger.Info().Msgf("Loaded %d fonts", len(wi.FontCache))

	for _, fallback := range wi.Configuration.FallbackFonts {
		if _, ok := wi.FontCache[fallback]; ok {
			wi.FallbackFonts = append(wi.FallbackFonts, fallback)
		} else {
			wi.Logger.Warn().Str("font", fallback).Msg("Referenced invalid fallback font")
		}
	}

	wi.Logger.Info().
		Msgf(
			"Discovered %d/%d fallback fonts",
			len(wi.FallbackFonts),
			len(wi.Configuration.FallbackFonts),
		)

	wi.Logger.Debug().Msg("Loading static backgrounds")

	files, err := ioutil.ReadDir(wi.Configuration.Store.StaticBackgroundsPath)
	if err != nil {
		wi.Logger.Error().Err(err).
			Str("path", wi.Configuration.Store.StaticBackgroundsPath).
			Msg("Failed to list files in static backgrounds folder")
	}

	for _, f := range files {
		name := f.Name()
		path := path.Join(wi.Configuration.Store.StaticBackgroundsPath, name)
		nametrim := name[0 : len(name)-len(filepath.Ext(name))]

		wi.Logger.Trace().
			Str("path", path).
			Msg("Loading static image")

		if _, ok := wi.StaticBackgroundCache[nametrim]; ok {
			wi.Logger.Error().Err(err).
				Str("name", nametrim).
				Msg("Static background with name already exists")

			continue
		}

		fimg, err := os.Open(path)
		if err != nil {
			wi.Logger.Error().Err(err).
				Str("path", path).
				Msg("Failed to open static image")

			continue
		}

		img, format, err := image.Decode(fimg)
		if err != nil {
			wi.Logger.Error().Err(err).
				Str("path", path).
				Msg("Failed to decode static image")

			continue
		}

		fimg.Seek(0, io.SeekStart)

		config, _, err := image.DecodeConfig(fimg)
		if err != nil {
			wi.Logger.Error().Err(err).
				Str("path", path).
				Msg("Failed to decode static image config")

			continue
		}

		wi.StaticBackgroundCache[nametrim] = &ImageCache{
			Format: format,
			Frames: []image.Image{img},
			Config: config,
		}
	}

	wi.Logger.Info().Msgf(
		"Loaded %d/%d static backgrounds",
		len(wi.StaticBackgroundCache), len(files))

	if wi.Configuration.Prometheus.Enabled {
		wi.Logger.Info().
			Str("path", wi.Configuration.Prometheus.Host+"/metrics").
			Msg("Starting up Prometheus handler")

		go func() {
			http.Handle("/metrics", promhttp.Handler())

			wi.Logger.Info().
				Str("host", wi.Configuration.Prometheus.Host).
				Msg("Serving prom")

			err := http.ListenAndServe(wi.Configuration.Prometheus.Host, nil)
			if err != nil {
				wi.Logger.Error().
					Str("host", wi.Configuration.Prometheus.Host).Err(err).
					Msg("Failed to serve prometheus server")
			}
		}()
	}

	wi.Logger.Info().Msg("Starting up HTTP server")

	if wi.Configuration.Store.StaticPath != "" {
		wi.Logger.Info().Str("path", wi.Configuration.Store.StaticPath).Msg("Serving files")
		wi.fs = &fasthttp.FS{
			Root:            wi.Configuration.Store.StaticPath,
			Compress:        true,
			CompressBrotli:  true,
			AcceptByteRange: true,
			CacheDuration:   distCacheDuration,
			PathNotFound: fasthttp.RequestHandler(
				func(ctx *fasthttp.RequestCtx) {
					ctx.WriteString("There is nothing here")
				},
			),
		}

		wi.distHandler = wi.fs.NewRequestHandler()
	}

	wi.Logger.Info().Msg("Creating endpoints")
	wi.Router = createEndpoints(wi)

	go func() {
		wi.Logger.Info().
			Str("host", wi.Configuration.HTTP.Host).
			Msg("Serving HTTP")

		err = fasthttp.ListenAndServe(wi.Configuration.HTTP.Host, wi.HandleRequest)
		if err != nil {
			wi.Logger.Error().Str("host", wi.Configuration.HTTP.Host).Err(err).Msg("Failed to serve http server")
		}
	}()

	go wi.PrometheusFetcher()

	println("Service running (Press CTRL+C to quit)")

	return nil
}

// PrometheusFetcher fetches extra information such as store usage.
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
			err = tx.Bucket(bucketName).ForEach(func(k, v []byte) error {
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

		pcr := make([]int64, 0)
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
			// We attempt to lock as a last resort to ensure nobody is still using it.
			// Problem is long running tasks may block this for a long time causing
			// other things to slow down.
			wi.ProfileCache[k].Lock()
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
			wi.BackgroundCache[k].Lock()
			delete(wi.BackgroundCache, k)
		}

		backgroundCacheSize.Set(float64(len(wi.BackgroundCache)))
		wi.BackgroundCacheMu.Unlock()

		// We do not want to remove actual fonts,
		// only faces.
		totalFaces := 0
		freedFaces := 0

		wi.FontCacheMu.Lock()
		for _, v := range wi.FontCache {
			fcr := make([]float64, 0)

			v.FaceCacheMu.Lock()
			for fk, fv := range v.FaceCache {
				if start.After(fv.LastAccessed.Add(faceCacheTTL)) {
					fcr = append(fcr, fk)
				}
			}

			for _, k := range fcr {
				v.FaceCache[k].Lock()
				delete(v.FaceCache, k)
			}

			totalFaces += len(v.FaceCache)
			freedFaces += len(fcr)
			v.FaceCacheMu.Unlock()
		}

		fontCacheSize.Set(float64(totalFaces))
		wi.FontCacheMu.Unlock()

		fd := time.Since(start).Round(time.Millisecond).Milliseconds()
		wi.Logger.Debug().
			Int64("dur", fd).
			Int("freed_images", fi).
			Int("freed_faces", freedFaces).
			Int("freed_profiles", len(pcr)).
			Int("freed_backgrounds", len(bcr)).
			Msg("Finished freeing")

		freedImages.Set(float64(fi))
		freedDuration.Set(float64(fd))

		imagesStoreCount.Set(float64(storeCount))
		imagesStoreSize.Set(float64(storeSize))

		imagesFolderCount.Set(float64(folderCount))
		imagesFolderSize.Set(float64(folderSize))

		time.Sleep(time.Minute)
	}
}

// Close will gracefully close the application and wait for any images being generated.
func (wi *WelcomerImageService) Close() (err error) {
	wi.Logger.Info().Msg("Closing Welcomer Image Service. Waiting for any active tasks")

	wi.ServiceClosing.Set()
	wi.PoolWaiter.Wait()

	return
}
