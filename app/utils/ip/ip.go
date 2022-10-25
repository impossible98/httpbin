package ip

import (
	// import built-in packages
	"net"
	"net/http"
	"strings"
)

// getClientIP tries to get a reasonable value for the IP address of the
// client making the request. Note that this value will likely be trivial to
// spoof, so do not rely on it for security purposes.
func GetClientIP(req *http.Request) string {
	// Special case some hosting platforms that provide the value directly.
	if clientIP := req.Header.Get("Fly-Client-IP"); clientIP != "" {
		return clientIP
	}

	// Try to pull a reasonable value from the X-Forwarded-For header, if
	// present, by taking the first entry in a comma-separated list of IPs.
	if forwardedFor := req.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		return strings.TrimSpace(strings.SplitN(forwardedFor, ",", 2)[0])
	}

	// Finally, fall back on the actual remote addr from the request.
	return req.RemoteAddr
}

func GetLocalIp() string {
	addrs, err := net.InterfaceAddrs()
	// control flow
	if err != nil {
		// return
		return ""
	}
	for _, address := range addrs {
		// control flow
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			// control flow
			if ipnet.IP.To4() != nil {
				ip := ipnet.IP.String()
				// return
				return ip
			}
		}
	}
	return ""
}
