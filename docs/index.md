# Gatus Monitor

Welcome to Gatus Monitor documentation!

## Overview

Gatus Monitor is a cross-platform system tray application that monitors multiple [Gatus](https://github.com/TwiN/gatus) health check endpoints and provides visual feedback on their status through a color-coded icon.

## Features

- **Multi-endpoint monitoring**: Monitor multiple Gatus instances simultaneously
- **Visual status indicators**: Color-coded system tray icon (green/orange/red)
- **Configurable intervals**: Set custom query intervals (default: 60s)
- **Staggered queries**: Intelligent query distribution to avoid server load spikes
- **Cross-platform**: Works on Linux, Windows, and macOS
- **Lightweight**: Minimal resource usage (<50MB RAM, <1% CPU when idle)
- **Easy configuration**: Simple settings panel for managing endpoints

## Quick Start

1. Download the latest release for your platform
2. Launch the application - it will appear in your system tray
3. Right-click the tray icon and select "Settings"
4. Add your Gatus URLs and configure the query interval
5. Save and enjoy automatic monitoring!

## Status Indicators

| Icon Color | Meaning | Condition |
|-----------|---------|-----------|
| Green | All systems operational | 0 errors across all endpoints |
| Orange | Minor issues detected | 1-2 errors on any endpoint |
| Red | Critical issues detected | 3+ errors on any endpoint |

## Support

- [User Guide](user-guide/getting-started.md) - Learn how to use Gatus Monitor
- [Administrator Guide](admin-guide/deployment.md) - Deploy and manage the application
- [Developer Guide](developer-guide/architecture.md) - Contribute to the project
- [GitHub Issues](https://github.com/kartoza/gatus-monitor/issues) - Report bugs or request features

## Funding

This project is open source and freely available. If you find it useful, please consider supporting its development:

- [GitHub Sponsors](https://github.com/sponsors/kartoza)
- [Ko-fi](https://ko-fi.com/kartoza)
- [Kartoza](https://kartoza.com/en/sponsorship/)

---

Made with 💗 by [Kartoza](https://kartoza.com)
