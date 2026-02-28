# Installation

Gatus Monitor is available for Linux, Windows, and macOS.

## Linux

### Ubuntu/Debian (.deb)

```bash
wget https://github.com/kartoza/gatus-monitor/releases/latest/download/gatus-monitor_linux_amd64.deb
sudo dpkg -i gatus-monitor_linux_amd64.deb
```

### Red Hat/Fedora (.rpm)

```bash
wget https://github.com/kartoza/gatus-monitor/releases/latest/download/gatus-monitor_linux_amd64.rpm
sudo rpm -i gatus-monitor_linux_amd64.rpm
```

### Binary Installation

```bash
wget https://github.com/kartoza/gatus-monitor/releases/latest/download/gatus-monitor_linux_amd64
chmod +x gatus-monitor_linux_amd64
sudo mv gatus-monitor_linux_amd64 /usr/local/bin/gatus-monitor
```

### Running

```bash
gatus-monitor
```

## macOS

### Direct Download

1. Download the appropriate version for your Mac:
   - Apple Silicon (M1/M2): `gatus-monitor_darwin_arm64`
   - Intel: `gatus-monitor_darwin_amd64`

2. Make it executable:

```bash
chmod +x gatus-monitor_darwin_arm64
```

3. Run the application:

```bash
./gatus-monitor_darwin_arm64
```

Note: On first run, you may need to allow the application in System Preferences → Security & Privacy.

## Windows

### MSI Installer

1. Download `gatus-monitor_windows_amd64.msi`
2. Run the installer
3. Follow the installation wizard
4. Launch from Start Menu or Desktop shortcut

### Binary

1. Download `gatus-monitor_windows_amd64.exe`
2. Place it in a folder of your choice
3. Double-click to run

## From Source

### Prerequisites

- [Go 1.21 or later](https://golang.org/dl/)
- [Nix with flakes enabled](https://nixos.org/download.html) (recommended)

### Using Nix

```bash
git clone https://github.com/kartoza/gatus-monitor.git
cd gatus-monitor
nix develop
go build -o gatus-monitor ./cmd/gatus-monitor
./gatus-monitor
```

### Without Nix

```bash
git clone https://github.com/kartoza/gatus-monitor.git
cd gatus-monitor
go mod download
go build -o gatus-monitor ./cmd/gatus-monitor
./gatus-monitor
```

Note: You'll need to install system dependencies manually (see [Building](../developer-guide/building.md)).

## Verification

After installation, verify the installation by running:

```bash
gatus-monitor --version
```

You should see output similar to:
```
Gatus Monitor v0.1.0
```

## Next Steps

Now that you have Gatus Monitor installed, proceed to the [Getting Started Guide](getting-started.md) to configure your endpoints.
