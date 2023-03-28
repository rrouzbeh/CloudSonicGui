package main

import (
	"github.com/rrouzbeh/CloudSonicGui/gui"
)

func main() {
	// Fetch IP addresses from the URL
	// ipList, err := fetcher.FetchIPs()
	// if err != nil {
	// 	fmt.Println("Error fetching IP addresses:", err)
	// 	return
	// }

	// Create custom TLS connections and calculate response times
	// responseTimes, err := connector.ConnectAndGetResponseTimes(ipList)
	// if err != nil {
	// 	fmt.Println("Error connecting to IPs and getting response times:", err)
	// 	return
	// }

	// Get the top 20 lowest response times
	// top20 := response_timer.GetTopLowestResponseTimes(responseTimes, 20)

	// Display the results in a Fyne GUI
	gui.ShowResults()
}
