package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/rrouzbeh/CloudSonicGui/connector"
	"github.com/rrouzbeh/CloudSonicGui/fetcher"
	"github.com/rrouzbeh/CloudSonicGui/response_timer"
)

func ShowResults() {

	myApp := app.New()
	myWindow := myApp.NewWindow("Top 20 Lowest Response Times")

	content := container.NewVBox()
	progressBar := widget.NewProgressBar()
	ipList, _ := fetcher.FetchIPs()
	startButton := widget.NewButton("Start", func() {
		// clear everything
		content.Objects = nil
		progressBar.SetValue(0)

		responseTimes, err := connector.ConnectAndGetResponseTimes(ipList, progressBar)
		if err != nil {
			content.Add(widget.NewLabel(fmt.Sprintf("Error: %v", err)))
			return
		}
		top10 := response_timer.GetTopLowestResponseTimes(responseTimes, 10)

		for i, result := range top10 {
			content.Add(widget.NewLabel(fmt.Sprintf("%d. IP: %s, Response Time: %v", i+1, result.IP, result.ResponseTime)))
		}
	})

	myWindow.SetContent(container.NewVBox(startButton, progressBar, content))
	myWindow.Resize(fyne.NewSize(500, 400))
	myWindow.ShowAndRun()

}
