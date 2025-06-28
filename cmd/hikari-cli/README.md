# Hikari

**Hikari** (光) is a fast, cross-platform TUI application for discovering and controlling [LIFX](https://www.lifx.com/) smart lights on your local network (LAN).

> Built in Go. No cloud required.

---

## ✨ Features

- 🧭 Automatically discovers LIFX lights on your LAN
- 💡 Control power, brightness, and color
- 🔍 View device info and statuses
- ⚡️ Blazing fast — all local, no internet needed
- 🖥️ Works on macOS, Linux, and Windows

---

## 📦 Installation

Download a binary for your OS from the [Releases page](https://github.com/yourusername/hikari/releases), or build from source.

### Prebuilt binaries:

- `hikari-darwin-arm64.zip`
- `hikari-darwin-amd64.zip`
- `hikari-linux-amd64.zip`
- `hikari-windows-amd64.zip`

Each zip contains:

- The `hikari` binary
- A copy of this README
- A `VERSION` file
- A `LICENSE` file

> No installer needed — just unzip and run!

---

## 🚀 Usage

Once installed, just run:

```bash
./hikari
```

Or on Windows:

```bash
hikari.exe
```

Inside the TUI:

- Press i to inspect a device
- Press enter/e to select a device/command/parameter

* Press enter/e to send a simple command (e.g, on/off)

- Press a after editing parameters to apply and send the command
- Press esc/b to go back

* Press / to filter a device by name, group, location
* Press q to quit

---

🔧 Build From Source

```bash
git clone https://github.com/yourusername/hikari.git
cd hikari
go build ./cmd/main.go
```

---

📜 License

This project is licensed under the MIT License. See LICENSE for details.

---

🙏 Acknowledgements

Built with ❤️ using:

- Go
- Bubble Tea
- LIFX Public Lan Protocol
- LIFX Public Products registry
