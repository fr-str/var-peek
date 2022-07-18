package util

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"syslabit.com/git/syslabit/log"
)

// ReadBodyFromRequest ...
func ReadBodyFromRequest(r *http.Request) string {
	body, _ := ioutil.ReadAll(r.Body)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return string(body)
}

// ReadBodyFromResponse ...
func ReadBodyFromResponse(r *http.Response) string {
	body, _ := ioutil.ReadAll(r.Body)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return string(body)
}

func GetFile(url string) ([]byte, *log.RecordS) {
	res, err := http.Get(url)
	if err != nil {
		return nil, log.Error(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, log.Error("Failed to get file: {{status}}", log.Vars{"status": res.Status})
	}

	file, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, log.Error(err)
	}

	return file, nil
}

// Get preferred outbound ip of this machine
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return ""
	}
	return localAddr.IP.String()
}

func RequestIP(r *http.Request) string {
	if prior := r.Header.Get("X-Forwarded-For"); prior != "" {
		proxies := strings.Split(prior, ",")
		if len(proxies) > 0 {
			return strings.Trim(proxies[0], " ")
		}
	}

	// X-Real-Ip is less supported, but worth checking in the
	// absence of X-Forwarded-For
	if realIP := r.Header.Get("X-Real-Ip"); realIP != "" {
		return realIP
	}

	return strings.Split(r.RemoteAddr, ":")[0]
}
