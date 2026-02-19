# BunDeck - An Open Source Stream Deck Alternative

BunDeck is a free and open-source alternative to the [Elgato Stream Deck](https://www.elgato.com/us/en/p/stream-deck-mk2-black).

Built with Go and Fyne, it aims to provide a similar experience with customizable buttons that trigger scripts and actions.

The client can be run on any of the [supported platforms](https://docs.fyne.io/started/cross-compiling.html), including but not limited to:

* Linux
* Windows
* MacOS
* Android
* iOS
* RaspberryPi/embedded linux

## Features

* **Cross-Platform Compatibility:** Works on Windows, macOS, and Linux.
* **Customizable Buttons:**  Add and arrange buttons to your liking.
* **Scripting Support:** Execute scripts using Bun.
* **Extensible:**  Easily add new features and integrations.

### Upcoming Features

* **Customizable Buttons**: Add and arrange buttons to your liking (coming soon!).
* More to come!

## Installation

### Prerequisites

* **Go:** Ensure you have Go installed. You can download it from the official website: [https://go.dev/](https://go.dev/)
* **Fyne:** Install the Fyne toolkit: [https://fyne.io/](https://docs.fyne.io/started/)

* **Bun:**  Download and install Bun from: [https://bun.sh/](https://bun.sh/)

### Build from Source

1. **Clone the repository:**

    ```bash
    git clone https://github.com/ibanks/bundeck.git
    ```

2. **Navigate to the project directory:**

    ```bash
    cd bundeck/client
    ```
    or
   ```bash
   cd bundeck/server
   ```

4. **Build the application:**

    ```bash
    go build
    ```
    
#### Alternatively you can cross compile using [fyne-cross](https://docs.fyne.io/started/cross-compiling.html)

For example:

```bash
fyne-cross android -app-id dev.ibanks.bundeck -icon Icon.png -name BunDeck
```

## Contributing

Contributions are welcome\! Feel free to open issues for bug reports or feature requests.
Pull requests are appreciated.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
