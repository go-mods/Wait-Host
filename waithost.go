package waithost

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// waithost data
type WaitHost struct {
	Scheme         string
	Host           string
	Port           uint
	Timeout        time.Duration
	ConnectTimeout time.Duration

	logger         logger
	waitMessage    string
	retryMessage   string
	successMessage string
	timeoutMessage string
}

var (
	// Default host to use if none is present
	// in the address
	DefaultHost = "localhost"

	// Default scheme to use if none is present
	// in the address
	DefaultScheme = "tcp"

	// Default Timeout.
	// If 0, WaitHost while not return until
	// the connection is established
	DefaultTimeout = time.Duration(0)

	DefaultRetryMessage = "Waiting for connection on {host}:{port}"
)

var (
	// List of accepted schemes
	acceptedScheme = map[string]bool{
		"tcp":   true,
		"http":  true,
		"https": true,
	}

	// Port mapping
	portMap = map[string]uint{
		"http":  80,
		"https": 443,
	}

	// Check function to use depending on scheme
	checkFunc = map[string]func(wp *WaitHost) error{
		"tcp":   checkTcp,
		"http":  checkHttp,
		"https": checkHttps,
		"": func(wp *WaitHost) error {
			return &WaitHostError{BAD_SCHEME}
		},
	}
)

func New(target string) (*WaitHost, error) {
	wh, err := extractTarget(target)
	return wh, err
}

// Wait for the host port to be open on the target
func Wait(target string) error {
	if wh, err := extractTarget(target); err != nil {
		return err
	} else {
		if err := wh.Wait(); err != nil {
			return err
		} else {
			return nil
		}
	}
}

// Wait for the host port to be open on the target
func (wh *WaitHost) Wait() error {
	// Validate
	if err := validateTarget(wh); err != nil {
		return err
	}
	// Check
	return checkFunc[wh.Scheme](wh)
}

// SetLogger replace default logger
func (wh *WaitHost) SetLogger(log logger) {
	wh.logger = log
}

func (wh *WaitHost) SetWaitMessage(message string) {
	wh.waitMessage = message
}

func (wh *WaitHost) SetRetryMessage(message string) {
	wh.retryMessage = message
}

func (wh *WaitHost) SetSuccessMessage(message string) {
	wh.successMessage = message
}

func (wh *WaitHost) SetTimeoutMessage(message string) {
	wh.timeoutMessage = message
}

func extractTarget(target string) (*WaitHost, error) {
	// default host
	if strings.Index(target, ":") == 0 && strings.Index(target, "://") == -1 {
		target = DefaultHost + target
	}
	// default scheme
	if strings.Index(target, "://") == -1 {
		target = DefaultScheme + "://" + target
	}
	// Parse url
	if u, err := url.Parse(target); err != nil {
		return nil, err
	} else {
		// Create WaitHost data
		wh := &WaitHost{
			Scheme:         u.Scheme,
			Host:           u.Hostname(),
			Timeout:        DefaultTimeout,
			ConnectTimeout: time.Duration(1) * time.Second,
			logger:         defaultLogger,
			waitMessage:    "",
			retryMessage:   DefaultRetryMessage,
			successMessage: "",
			timeoutMessage: "",
		}
		// Port
		if port, err := strconv.Atoi(u.Port()); err != nil {
			wh.Port = portMap[u.Scheme]
		} else {
			wh.Port = uint(port)
		}
		return wh, nil
	}
}

func validateTarget(wh *WaitHost) error {
	target := fmt.Sprintf("%s://%s:%d", wh.Scheme, wh.Host, wh.Port)
	if _, err := url.Parse(target); err != nil {
		return err
	}
	if len(wh.Scheme) == 0 || acceptedScheme[wh.Scheme] == false {
		return &WaitHostError{BAD_SCHEME}
	}
	if len(wh.Host) == 0 {
		return &WaitHostError{BAD_HOST}
	}
	if wh.Port == 0 {
		return &WaitHostError{BAD_PORT}
	}
	if wh.ConnectTimeout == 0 {
		wh.ConnectTimeout = time.Duration(1) * time.Second
	}
	return nil
}

func check(wh *WaitHost, network, address string) error {
	// Start time
	start := time.Now()

	// wait message
	wh.printMessage(wh.waitMessage)

	for {
		// Check if the timeout is reached
		if wh.Timeout > 0 && time.Now().Sub(start) >= wh.Timeout {
			// timeout message
			wh.printMessage(wh.timeoutMessage)
			// error
			return &WaitHostError{TIMEOUT}
		}
		// Current time
		td := time.Now()
		// try to connect
		if conn, err := net.Dial(network, address); err == nil {
			// success message
			wh.printMessage(wh.successMessage)
			// success
			return nil
		} else if conn != nil {
			if err := conn.Close(); err != nil {
				return err
			} else {
				return nil
			}
		}
		// Wait
		if tw := time.Now().Sub(td); tw < wh.ConnectTimeout {
			time.Sleep(wh.ConnectTimeout - tw)
		}
		// retry message
		wh.printMessage(wh.retryMessage)
	}
}

func checkTcp(wh *WaitHost) error {
	return check(wh, "tcp", fmt.Sprintf("%s:%d", wh.Host, wh.Port))
}

func checkHttp(wh *WaitHost) error {
	return check(wh, "tcp", fmt.Sprintf("%s:%s", wh.Host, wh.Scheme))
}

func checkHttps(wh *WaitHost) error {
	return check(wh, "tcp", fmt.Sprintf("%s:%s", wh.Host, wh.Scheme))
}

func (wh *WaitHost) printMessage(message string) {
	if wh.logger == nil {
		return
	}
	if len(message) == 0 {
		return
	}

	message = strings.ReplaceAll(message, "{scheme}", wh.Scheme)
	message = strings.ReplaceAll(message, "{host}", wh.Host)
	message = strings.ReplaceAll(message, "{port}", fmt.Sprintf("%d", wh.Port))

	wh.logger.Print(message)
}
