package connector

import (
	"crypto/tls"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2/widget"
	"github.com/miekg/dns"
)

type IPResponseTime struct {
	IP           string
	ResponseTime time.Duration
	StatusCode   int
}

type IPDownloadSpeed struct {
	IP            string
	DownloadSpeed int
}

func getGoogleResponseTime() (time.Duration, error) {
	start := time.Now()
	resp, err := http.Get("https://www.google.com")
	if err != nil {
		return 0, fmt.Errorf("error connecting to Google: %w", err)
	}
	defer resp.Body.Close()

	elapsed := time.Since(start)
	return elapsed, nil
}

// ConnectAndGetResponseTimes connects to the IPs using custom TLS settings,
// calculates response time, sends an HTTP request, and returns a slice of IPResponseTime structs.
// It also updates the progress bar during the process.
func ConnectAndGetResponseTimes(ipList []string, progressBar *widget.ProgressBar) ([]IPResponseTime, error) {
	var responseTimes []IPResponseTime
	progress := 0.0
	progressStep := 1.0 / float64(20)

	googleResponseTime, err := getGoogleResponseTime()
	if err != nil {
		return nil, err
	}

	threshold := time.Duration(float64(googleResponseTime) * 1.5)

	for _, ip := range ipList {
		// Replace the last octet with a random number between 1 and 254
		ip = strings.Split(ip, ".")[0] + "." + strings.Split(ip, ".")[1] + "." + strings.Split(ip, ".")[2] + "." + strconv.Itoa(rand.Intn(254-0)+1)
		start := time.Now()

		conn, statusCode, err := createCustomTLSConnection(ip, "/cdn-cgi/trace")
		if err != nil {
			return nil, fmt.Errorf("error creating custom TLS connection: %w", err)
		}
		defer conn.Close()

		elapsed := time.Since(start)
		if elapsed <= threshold {
			responseTimes = append(responseTimes, IPResponseTime{
				IP:           ip,
				ResponseTime: elapsed,
				StatusCode:   statusCode,
			})
			progress += progressStep
			progressBar.SetValue(progress)
		}

		// Update the progress bar

		// Stop the scan if responseTimes has 20 or more elements
		if len(responseTimes) >= 20 {
			break
		}
	}

	return responseTimes, nil
}

// createCustomTLSConnection creates a custom TLS connection to the specified IP
// with SNI set to "speed.cloudflare.com", sends an HTTP request, and returns the response code.
func createCustomTLSConnection(ip string, url string) (*tls.Conn, int, error) {
	echconf, err := FetchECH()
	if err != nil {
		log.Fatalf("FetchECH: %v", err)
	}
	echPEMKey := fmt.Sprintf("-----BEGIN ECH CONFIGS-----\n%s\n-----END ECH CONFIGS-----", echconf)

	block, rest := pem.Decode([]byte(echPEMKey))
	if block == nil || block.Type != "ECH CONFIGS" || len(rest) > 0 {
		fmt.Println("failed to PEM-decode the ECH configs")
	}

	echconfig, err := tls.UnmarshalECHConfigs(block.Bytes)
	if err != nil {
		log.Fatalf("UnmarshalECHConfigs: %v", err)
	}
	conf := &tls.Config{
		ServerName:       "speed.cloudflare.com",
		ECHEnabled:       true,
		ClientECHConfigs: echconfig,
	}

	addr := fmt.Sprintf("%s:443", ip)
	conn, err := tls.Dial("tcp", addr, conf)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequest("GET", "https://"+ip+url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("error creating HTTP request: %w", err)
	}
	req.Host = "speed.cloudflare.com"
	req.Header.Set("User-Agent", "Mozilla/5.0")

	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				ServerName:       "speed.cloudflare.com",
				ECHEnabled:       true,
				ClientECHConfigs: echconfig,
			},
		},
	}

	resp, err := client.Do(req)
	fmt.Println(resp.StatusCode)
	// print the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("error reading response body: %w", err)
	}
	fmt.Println(string(body))
	if err != nil {
		return nil, 0, fmt.Errorf("error sending HTTP request: %w", err)
	}

	return conn, resp.StatusCode, nil
}

func FetchECH() (string, error) {
	dc := dns.Client{Timeout: 10 * time.Second}

	d := dns.Fqdn("crypto.cloudflare.com")
	q := dns.Question{
		Name:   d,
		Qtype:  dns.TypeHTTPS,
		Qclass: dns.ClassINET,
	}

	dnsAddr := "1.1.1.1:53"

	r, _, err := dc.Exchange(&dns.Msg{
		MsgHdr: dns.MsgHdr{
			Id:               dns.Id(),
			RecursionDesired: true,
		},
		Question: []dns.Question{q},
	}, dnsAddr)
	if err != nil {
		return "", err
	}

	for _, v := range r.Answer {
		if vv, ok := v.(*dns.HTTPS); ok {
			for _, vvv := range vv.SVCB.Value {
				if vvv.Key().String() == "ech" {
					return vvv.String(), nil
				}
			}
		}
	}

	return "", errors.New("failed to find ech in response")
}
