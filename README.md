# Gatus Monitor

A cross-platform system tray application for monitoring [Gatus](https://github.com/TwiN/gatus) health check endpoints with visual status indicators.

## Features

- **Multi-endpoint monitoring**: Monitor multiple Gatus instances simultaneously
- **Visual status indicators**: Color-coded system tray icon (green/orange/red)
- **Configurable intervals**: Set custom query intervals (default: 60s)
- **Staggered queries**: Intelligent query distribution to avoid server load spikes
- **Cross-platform**: Works on Linux, Windows, and macOS
- **Lightweight**: Minimal resource usage (<50MB RAM, <1% CPU when idle)
- **Easy configuration**: Simple settings panel for managing endpoints

## Status Indicators

| Icon Color | Meaning | Condition |
|-----------|---------|-----------|
| Green | All systems operational | 0 errors across all endpoints |
| Orange | Minor issues detected | 1-2 errors on any endpoint |
| Red | Critical issues detected | 3+ errors on any endpoint |

## Installation

### Linux

#### Debian/Ubuntu (.deb)
```bash
wget https://github.com/kartoza/gatus-monitor/releases/latest/download/gatus-monitor_linux_amd64.deb
sudo dpkg -i gatus-monitor_linux_amd64.deb
```

#### Red Hat/Fedora (.rpm)
```bash
wget https://github.com/kartoza/gatus-monitor/releases/latest/download/gatus-monitor_linux_amd64.rpm
sudo rpm -i gatus-monitor_linux_amd64.rpm
```

#### AppImage
```bash
wget https://github.com/kartoza/gatus-monitor/releases/latest/download/gatus-monitor_linux_amd64.AppImage
chmod +x gatus-monitor_linux_amd64.AppImage
./gatus-monitor_linux_amd64.AppImage
```

### macOS

#### Direct Download
Download the `.dmg` file from the [releases page](https://github.com/kartoza/gatus-monitor/releases) and drag to Applications.

### Windows

#### MSI Installer
Download and run the `.msi` installer from the [releases page](https://github.com/kartoza/gatus-monitor/releases).

### From Source

```bash
git clone https://github.com/kartoza/gatus-monitor.git
cd gatus-monitor
nix develop  # Enter development environment
go build -o gatus-monitor ./cmd/gatus-monitor
```

## Quick Start

1. **Launch the application** - it will appear in your system tray
2. **Right-click the tray icon** and select "Settings"
3. **Add your Gatus URLs** (e.g., `https://status.example.com`)
4. **Set your preferred query interval** (default: 60 seconds)
5. **Save and close** - monitoring will start automatically

## Configuration

Configuration is stored in a platform-specific location:

- **Linux**: `~/.config/gatus-monitor/config.json`
- **macOS**: `~/Library/Application Support/gatus-monitor/config.json`
- **Windows**: `%APPDATA%\gatus-monitor\config.json`

Example configuration:
```json
{
  "query_interval": 60,
  "gatus_urls": [
    "https://status.example.com",
    "https://status2.example.com"
  ]
}
```

## Documentation

Full documentation is available at [https://kartoza.github.io/gatus-monitor](https://kartoza.github.io/gatus-monitor)

- [User Guide](https://kartoza.github.io/gatus-monitor/user-guide/)
- [Administrator Guide](https://kartoza.github.io/gatus-monitor/admin-guide/)
- [Developer Guide](https://kartoza.github.io/gatus-monitor/developer-guide/)
- [API Reference](https://kartoza.github.io/gatus-monitor/api/)

## Development

### Prerequisites

- [Nix](https://nixos.org/download.html) with flakes enabled
- Go 1.21+ (provided by Nix)

### Development Environment

```bash
# Enter development environment
nix develop

# Run the application
go run ./cmd/gatus-monitor

# Run tests
go test ./...

# Run linters
golangci-lint run

# Build for all platforms
nix run .#build-all
```

### Project Structure

```
gatus-monitor/
├── cmd/
│   └── gatus-monitor/    # Main application entry point
├── internal/
│   ├── config/           # Configuration management
│   ├── gatus/            # Gatus API client
│   ├── icons/            # Embedded icon resources
│   ├── monitor/          # Core monitoring logic
│   ├── scheduler/        # Query scheduling
│   ├── storage/          # Persistent storage
│   ├── systray/          # System tray integration
│   └── ui/               # Settings UI
├── docs/                 # MkDocs documentation
├── .github/
│   └── workflows/        # CI/CD workflows
└── tests/                # Integration tests
```

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`go test ./...`)
5. Commit your changes (commits will be signed by pre-commit hooks)
6. Push to your branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## Support

- **Documentation**: https://kartoza.github.io/gatus-monitor
- **Issue Tracker**: https://github.com/kartoza/gatus-monitor/issues
- **Discussions**: https://github.com/kartoza/gatus-monitor/discussions

## Funding

This project is open source and freely available. If you find it useful, please consider supporting its development:

- [GitHub Sponsors](https://github.com/sponsors/kartoza)
- [Ko-fi](https://ko-fi.com/kartoza)
- [Kartoza](https://kartoza.com/en/sponsorship/)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Credits

Made with 💗 by [Kartoza](https://kartoza.com)

### Built With

- [Go](https://golang.org/) - The Go Programming Language
- [Fyne](https://fyne.io/) - Cross-platform GUI toolkit
- [Systray](https://github.com/fyne-io/systray) - System tray integration
- [Gatus](https://github.com/TwiN/gatus) - Health check and monitoring

## Acknowledgments

- The Gatus team for creating an excellent monitoring solution
- The Fyne community for their excellent GUI framework
- All our contributors and supporters
