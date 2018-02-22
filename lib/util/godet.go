package util

import "github.com/raff/godet"

// SetDeviceMetricsOverride is a wrapper for `Emulation.setDeviceMetricsOverride` call
// (see https://chromedevtools.github.io/devtools-protocol/tot/Emulation#method-setDeviceMetricsOverride).
func SetDeviceMetricsOverride(
	remote *godet.RemoteDebugger, width int, height int,
	deviceScaleFactor float64, mobile bool, fitWindow bool,
) (err error) {
	_, err = remote.SendRequest(
		"Emulation.setDeviceMetricsOverride",
		godet.Params{
			"width":             int(width),
			"height":            int(height),
			"deviceScaleFactor": float64(deviceScaleFactor),
			"mobile":            bool(mobile),
			"fitWindow":         bool(fitWindow),
		},
	)
	return
}
