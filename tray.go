package main

import (
	"bundeck/internal/settings"

	"encoding/base64"
	"log"
	"net"
	"strconv"

	"github.com/skip2/go-qrcode"
)

func generateQRCodeBase64(content string) (string, error) {
	// Generate QR code
	qr, err := qrcode.Encode(content, qrcode.Medium, 128)

	if err != nil {
		return "", err
	}

	// Convert to base64
	return base64.StdEncoding.EncodeToString(qr), nil
}

func showQRCodeDialog() {
	title := "Scan QR Code"
	ip := GetOutboundIP().To4().String()
	port := settings.LoadSettings().Port
	url := "http://" + ip + ":" + strconv.Itoa(port)

	// Generate QR code
	qr, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		log.Printf("Error generating QR code: %v", err)
		return
	}

	// Convert QR code to image
	qrImg := qr.Image(256)

	// Show QR code dialog
	DisplayQRCode(title, qrImg)
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
