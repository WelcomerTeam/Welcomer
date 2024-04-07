package welcomer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"

	"github.com/rs/zerolog"
)

const (
	IPIntelEndpoint = "https://check.getipintel.net/check.php"
)

type IPIntelResponse struct {
	Success      string  `json:"success"`
	ResultString string  `json:"result"`
	Result       float64 `json:"-"`
	Country      string  `json:"Country"`
}

var errorCodes = map[float64]string{
	-1: "Invalid no input",
	-2: "Invalid IP address",
	-3: "Unroutable address / private address",
	-4: "Unable to reach database",
	-5: "Your connecting IP has been banned from the system or you do not have permission to access a particular service.",
	-6: "You did not provide any contact information with your query or the contact information is invalid.",
}

type IPIntelFlags string
type IPIntelOFlags string

const (
	IPIntelFlagDefaultLookup               IPIntelFlags = ""
	IPIntelFlagDynamicBanList              IPIntelFlags = "m"
	IPIntelFlagDynamicBanListDynamicChecks IPIntelFlags = "b"
	IPIntelFlagForceFullLookup             IPIntelFlags = "f"

	IPIntelOFlagOnlyBadIP   IPIntelOFlags = "b"
	IPIntelOFlagShowCountry IPIntelOFlags = "c"
	IPIntelOFlagShowVPN     IPIntelOFlags = "i"
	IPIntelOFlgagShowASN    IPIntelOFlags = "a"
)

type IPChecker interface {
	CheckIP(ipaddress string, flags IPIntelFlags, oflags IPIntelOFlags) (response IPIntelResponse, err error)
}

type BasicIPChecker struct {
	logger zerolog.Logger
}

// NewBasicIPChecker creates a new basic IP checker.
func NewBasicIPChecker(logger zerolog.Logger) *BasicIPChecker {
	return &BasicIPChecker{
		logger: logger,
	}
}

func (c *BasicIPChecker) CheckIP(ipaddress string, flags IPIntelFlags, oflags IPIntelOFlags) (response IPIntelResponse, err error) {
	return checkIPIntel(c.logger, ipaddress, flags, oflags)
}

type LRUIPChecker struct {
	logger      zerolog.Logger
	maxSize     int
	cache       map[string]IPIntelResponse
	accessOrder []string
	mutex       sync.RWMutex
}

// NewLRUIPChecker creates a new LRU IP checker with the specified maximum cache size.
func NewLRUIPChecker(logger zerolog.Logger, maxSize int) *LRUIPChecker {
	return &LRUIPChecker{
		logger:      logger,
		maxSize:     maxSize,
		cache:       make(map[string]IPIntelResponse),
		accessOrder: make([]string, 0),
		mutex:       sync.RWMutex{},
	}
}

func (c *LRUIPChecker) CheckIP(ipaddress string, flags IPIntelFlags, oflags IPIntelOFlags) (response IPIntelResponse, err error) {
	// Check if the IP address is already in the cache
	c.mutex.RLock()
	cachedResponse, ok := c.cache[ipaddress]
	c.mutex.RUnlock()

	if ok {
		// Move the IP address to the front of the access order
		c.mutex.Lock()
		c.moveToFront(ipaddress)
		c.mutex.Unlock()

		return cachedResponse, nil
	}

	// Perform the IP check using the basic IP checker
	response, err = checkIPIntel(c.logger, ipaddress, flags, oflags)
	if err != nil {
		return response, err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Add the IP address and response to the cache
	c.cache[ipaddress] = response
	c.accessOrder = append(c.accessOrder, ipaddress)

	// If the cache size exceeds the maximum size, remove the least recently used IP address
	if len(c.cache) > c.maxSize {
		oldestIP := c.accessOrder[0]
		delete(c.cache, oldestIP)
		c.accessOrder = c.accessOrder[1:]
	}

	return response, nil
}

func (c *LRUIPChecker) moveToFront(ipaddress string) {
	// Find the index of the IP address in the access order
	index := -1
	for i, addr := range c.accessOrder {
		if addr == ipaddress {
			index = i
			break
		}
	}

	// If the IP address is already at the front, no need to move
	if index == 0 {
		return
	}

	// Move the IP address to the front by swapping it with the previous addresses
	for i := index; i > 0; i-- {
		c.accessOrder[i] = c.accessOrder[i-1]
	}
	c.accessOrder[0] = ipaddress
}

func checkIPIntel(logger zerolog.Logger, ipaddress string, flags IPIntelFlags, oflags IPIntelOFlags) (response IPIntelResponse, err error) {
	reqParams := url.Values{}
	reqParams.Set("ip", ipaddress)
	reqParams.Set("contact", os.Getenv("IPINTEL_CONTACT"))
	reqParams.Set("format", "json")
	reqParams.Set("flags", string(flags))

	if oflags != "" {
		reqParams.Set("oflags", string(oflags))
	}

	req, err := http.NewRequest(http.MethodGet, IPIntelEndpoint+"?"+reqParams.Encode(), nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create IPIntel request")

		return
	}

	req.Header.Set("User-Agent", UserAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error().Err(err).Int("status_code", resp.StatusCode).Msg("Failed to send IPIntel request")

		return
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to decode IPIntel response")

		return
	}

	response.Result, err = strconv.ParseFloat(response.ResultString, 64)
	if err != nil {
		logger.Error().Err(err).Str("response", response.ResultString).Msg("Failed to parse IPIntel response result")

		return
	}

	if response.Result < 0 {
		logger.Error().Float64("result", response.Result).Msg("IPIntel returned an error")

		if message, ok := errorCodes[response.Result]; ok {
			return response, fmt.Errorf(fmt.Sprintf("ipintel failed with code %f: %s", response.Result, message))
		}

		return response, fmt.Errorf(fmt.Sprintf("ipintel failed with code %f: unknown error", response.Result))
	}

	return response, nil
}
