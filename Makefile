.PHONY: build clean build-macos-intel build-macos-apple build-macos

# Default build target
build:
	go build -o bundeck

# Clean build artifacts
clean:
	rm -f bundeck
	rm -f BunDeck-*
	rm -f *.zip
	rm -rf dist

# Build macOS app bundle for Intel/x86_64
build-macos-intel:
	chmod +x ./build-macos-app.sh
	./build-macos-app.sh intel

# Build macOS app bundle for Apple Silicon
build-macos-apple:
	chmod +x ./build-macos-app.sh
	./build-macos-app.sh apple

# Build macOS app bundles for both architectures
build-macos: build-macos-intel build-macos-apple
