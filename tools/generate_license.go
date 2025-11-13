package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/protei/monitoring/pkg/license"
)

func main() {
	// Command line flags
	customerName := flag.String("customer", "", "Customer name")
	expiryDate := flag.String("expiry", "2030-12-31", "Expiry date (YYYY-MM-DD)")
	macAddress := flag.String("mac", "", "Licensed MAC address (e.g., 00:11:22:33:44:55)")
	outputFile := flag.String("output", "license.json", "Output license file")

	// Feature flags
	enable2G := flag.Bool("2g", true, "Enable 2G support")
	enable3G := flag.Bool("3g", true, "Enable 3G support")
	enable4G := flag.Bool("4g", true, "Enable 4G support")
	enable5G := flag.Bool("5g", true, "Enable 5G support")

	// Protocol flags
	enableMAP := flag.Bool("map", true, "Enable MAP protocol")
	enableCAP := flag.Bool("cap", true, "Enable CAP protocol")
	enableINAP := flag.Bool("inap", true, "Enable INAP protocol")
	enableDiameter := flag.Bool("diameter", true, "Enable Diameter protocol")
	enableHTTP := flag.Bool("http", true, "Enable HTTP protocol")
	enableGTP := flag.Bool("gtp", true, "Enable GTP protocol")

	// Limits
	maxSubscribers := flag.Int("max-subscribers", 5000000, "Maximum subscribers")
	maxTPS := flag.Int("max-tps", 5000, "Maximum transactions per second")

	flag.Parse()

	// Validate required fields
	if *customerName == "" {
		fmt.Println("Error: customer name is required")
		flag.Usage()
		os.Exit(1)
	}

	if *macAddress == "" {
		fmt.Println("Error: MAC address is required")
		flag.Usage()
		os.Exit(1)
	}

	// Create license
	lic := &license.License{
		CustomerName:   *customerName,
		ExpiryDate:     *expiryDate,
		LicensedMAC:    *macAddress,
		Enable2G:       *enable2G,
		Enable3G:       *enable3G,
		Enable4G:       *enable4G,
		Enable5G:       *enable5G,
		EnableMAP:      *enableMAP,
		EnableCAP:      *enableCAP,
		EnableINAP:     *enableINAP,
		EnableDiameter: *enableDiameter,
		EnableHTTP:     *enableHTTP,
		EnableGTP:      *enableGTP,
		MaxSubscribers: *maxSubscribers,
		MaxTPS:         *maxTPS,
	}

	// Generate license with signature
	licenseJSON, err := license.GenerateLicense(lic, "PROTEI_MONITORING_VENDOR_KEY_2025_CHANGE_THIS_IN_PRODUCTION")
	if err != nil {
		fmt.Printf("Error generating license: %v\n", err)
		os.Exit(1)
	}

	// Write to file
	if err := os.WriteFile(*outputFile, []byte(licenseJSON), 0644); err != nil {
		fmt.Printf("Error writing license file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("License generated successfully: %s\n", *outputFile)
	fmt.Println("\nLicense Details:")
	fmt.Printf("  Customer: %s\n", *customerName)
	fmt.Printf("  Expiry: %s\n", *expiryDate)
	fmt.Printf("  MAC: %s\n", *macAddress)
	fmt.Printf("  Max Subscribers: %d\n", *maxSubscribers)
	fmt.Printf("  Max TPS: %d\n", *maxTPS)
	fmt.Println("\nEnabled Features:")
	fmt.Printf("  Generations: 2G=%v 3G=%v 4G=%v 5G=%v\n", *enable2G, *enable3G, *enable4G, *enable5G)
	fmt.Printf("  Protocols: MAP=%v CAP=%v INAP=%v Diameter=%v HTTP=%v GTP=%v\n",
		*enableMAP, *enableCAP, *enableINAP, *enableDiameter, *enableHTTP, *enableGTP)

	// Display license content
	var prettyLicense map[string]interface{}
	json.Unmarshal([]byte(licenseJSON), &prettyLicense)
	fmt.Println("\nGenerated License (JSON):")
	prettyJSON, _ := json.MarshalIndent(prettyLicense, "", "  ")
	fmt.Println(string(prettyJSON))
}
