package welcomerimages

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
	"golang.org/x/image/font/opentype"
	"golang.org/x/xerrors"
)

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

// FetchFont fetches a font face with the specified size
func (wi *WelcomerImageService) FetchFont(f string, size float64) (*FaceCache, *FontCache, error) {
	wi.FontCacheMu.RLock()
	font, ok := wi.FontCache[f]
	wi.FontCacheMu.RUnlock()

	wi.Logger.Trace().
		Str("font", f).
		Float64("size", size).
		Msg("Fetching font")

	if !ok {
		if f != wi.Configuration.Store.DefaultFont {
			fc, foc, err := wi.FetchFont(wi.Configuration.Store.DefaultFont, size)

			if err == nil {
				return fc, foc, err
			}
		}

		return nil, nil, xerrors.Errorf("no font exists")
	}

	font.FaceCacheMu.RLock()
	face, ok := font.FaceCache[size]
	font.FaceCacheMu.RUnlock()

	if ok {
		font.LastAccessedMu.Lock()
		face.LastAccessed = time.Now().UTC()
		font.LastAccessedMu.Unlock()

		return face, font, nil
	}

	wi.Logger.Trace().
		Str("font", f).
		Float64("size", size).
		Msg("Generating new font face")

	fc, err := opentype.NewFace(font.Font, &opentype.FaceOptions{
		Size: size,
		DPI:  72,
	})

	if err != nil {
		wi.Logger.Error().Err(err).Msg("Failed to create font face")
		return nil, nil, err
	}

	fca := &FaceCache{
		LastAccess: LastAccess{sync.RWMutex{}, time.Now().UTC(), sync.RWMutex{}},
		Face:       &fc,
	}

	font.FaceCacheMu.Lock()
	font.FaceCache[size] = fca
	font.FaceCacheMu.Unlock()

	return fca, font, nil
}

// FetchBackground fetches a background from its id. Returns the image and boolean indicating GIF
func (wi *WelcomerImageService) FetchBackground(b string, allowGifs bool) (*ImageCache, error) {
	if b == "" {
		b = "default"
	}

	wi.Logger.Trace().
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

	wi.Logger.Debug().
		Str("name", b).
		Bool("allowGifs", allowGifs).
		Msg("Fetching background")

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

	c = &ImageCache{
		LastAccess: LastAccess{sync.RWMutex{}, time.Now().UTC(), sync.RWMutex{}},
		Format:     format,
		Config:     config,
	}

	// We store as frames reguardless of image format however
	// we should copy over the other GIF data when neccessary.
	if format == "gif" {
		c.BackgroundIndex = gi.BackgroundIndex
		c.Delay = gi.Delay
		c.Disposal = gi.Disposal
		c.LoopCount = gi.LoopCount
		c.Frames = make([]image.Image, len(gi.Image))

		for framenum, frame := range gi.Image {
			c.Frames[framenum] = image.Image(frame)
		}
	} else {
		c.Frames = make([]image.Image, 1)
		c.Frames[0] = im
	}

	wi.BackgroundCacheMu.Lock()
	wi.BackgroundCache[b] = c
	wi.BackgroundCacheMu.Unlock()

	return c, nil
}

// FetchAvatar fetches an avatar from a user id and avatar hash
func (wi *WelcomerImageService) FetchAvatar(u int64, a string) (*StaticImageCache, error) {
	wi.Logger.Trace().
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

		return c, nil
	}

	wi.Logger.Debug().
		Int64("user", u).
		Str("hash", a).
		Msg("Fetching avatar")

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
		if wi.UseFallbackProfile {
			return wi.FallbackProfile, nil
		} else {
			return nil, xerrors.New(fmt.Sprintf("fetchavatar response: %d", s))
		}
	}

	imageProfileResponseTimes.Observe(float64(ms) / 1000)
	imageProfileResponseCodes.WithLabelValues(strconv.Itoa(s)).Inc()

	if err != nil {
		wi.Logger.Error().Err(err).Msg("Failed to retrieve profile picture of user")

		if wi.UseFallbackProfile {
			return wi.FallbackProfile, nil
		} else {
			return nil, err
		}
	}

	im, format, err := image.Decode(bytes.NewBuffer(b))
	if err != nil {
		wi.Logger.Error().Err(err).Msg("Failed to decode profile picture of user")
	}

	sic := &StaticImageCache{
		LastAccess: LastAccess{sync.RWMutex{}, time.Now().UTC(), sync.RWMutex{}},
		Image:      im,
		Format:     format,
	}

	wi.ProfileCacheMu.Lock()
	wi.ProfileCache[u] = sic
	wi.ProfileCacheMu.Unlock()

	return sic, nil
}
