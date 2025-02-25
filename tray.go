package main

import (
	"bundeck/internal/settings"
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strings"

	"log"
	"net"

	"fyne.io/systray"
)

func initTray(settings *settings.Settings) {
	if runtime.GOOS == "darwin" {
		systray.SetIcon(macLogo)
	} else if runtime.GOOS == "linux" {
		systray.SetIcon(linuxLogo)
		systray.SetTitle("BunDeck")
	} else {
		systray.SetIcon(winLogo)
		systray.SetTitle("BunDeck")
	}

	browser := systray.AddMenuItem("Open App", "Open App")
	qr := systray.AddMenuItem("Show QR Code", "Show QR Code")
	quit := systray.AddMenuItem("Exit", "Exit")

	go func() {
		<-quit.ClickedCh
		systray.Quit()
	}()

	go func() {
		<-qr.ClickedCh
		ip := GetOutboundIP().To4().String()
		qrUrl := fmt.Sprintf("http://%s:%d", ip, settings.Port)
		fullUrl := fmt.Sprintf("http://localhost:%d/qr/%s", settings.Port, url.PathEscape(qrUrl))
		openURL(fullUrl)
	}()

	go func() {
		<-browser.ClickedCh
		openURL(fmt.Sprintf("http://localhost:%d", settings.Port))
	}()
}

// Get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

// https://stackoverflow.com/questions/39320371/how-start-web-server-to-open-page-in-browser-in-golang
// openURL opens the specified URL in the default browser of the user.
func openURL(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default: // "linux", "freebsd", "openbsd", "netbsd"
		// Check if running under WSL
		if isWSL() {
			// Use 'cmd.exe /c start' to open the URL in the default Windows browser
			cmd = "cmd.exe"
			args = []string{"/c", "start", url}
		} else {
			// Use xdg-open on native Linux environments
			cmd = "xdg-open"
			args = []string{url}
		}
	}
	if len(args) > 1 {
		// args[0] is used for 'start' command argument, to prevent issues with URLs starting with a quote
		args = append(args[:1], append([]string{""}, args[1:]...)...)
	}
	return exec.Command(cmd, args...).Start()
}

// isWSL checks if the Go program is running inside Windows Subsystem for Linux
func isWSL() bool {
	releaseData, err := exec.Command("uname", "-r").Output()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(releaseData)), "microsoft")
}
