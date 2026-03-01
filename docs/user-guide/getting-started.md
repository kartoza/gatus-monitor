# Getting Started

This guide will help you get started with Gatus Monitor.

## Installation

See the [Installation Guide](installation.md) for detailed installation instructions for your platform.

## First Run

1. **Launch the application**

   After installation, launch Gatus Monitor. The application will start minimized to your system tray.

2. **Locate the tray icon**

   Look for the Gatus Monitor icon in your system tray:
   - **Windows**: Bottom-right corner (notification area)
   - **macOS**: Top-right corner (menu bar)
   - **Linux**: Varies by desktop environment (typically top or bottom panel)

3. **Initial status**

   The icon will be green initially, indicating no endpoints are configured.

## Configuring Endpoints

1. **Open Settings**

   Right-click the tray icon and select "Settings" from the menu.

2. **Add Gatus URLs**

   In the settings window:
   - Click "Add URL"
   - Enter your Gatus instance URL (e.g., `https://status.example.com`)
   - Click "Add"
   - Repeat for additional endpoints

3. **Configure Query Interval**

   Set how often Gatus Monitor should check your endpoints:
   - Default: 60 seconds
   - Minimum: 10 seconds
   - Maximum: 3600 seconds (1 hour)

4. **Save Configuration**

   Click "Save" to apply your changes. Monitoring will start automatically.

## Understanding the Status

The tray icon color indicates the overall health status:

- **Green**: All endpoints are healthy (no errors)
- **Orange**: Some endpoints have 1-2 errors
- **Red**: One or more endpoints have 3+ errors or are unreachable

### Tooltip Information

Hover over the tray icon to see detailed status information:

- Overall status
- Number of endpoints
- Healthy, unreachable, and error counts

## Manual Refresh

To force an immediate refresh of all endpoints:

1. Right-click the tray icon
2. Select "Refresh Now"

This will query all endpoints immediately, regardless of the configured interval.

## Exiting the Application

To exit Gatus Monitor:

1. Right-click the tray icon
2. Select "Quit"

Your configuration will be saved automatically.

## Next Steps

- Learn about [Configuration Options](configuration.md)
- Troubleshoot issues in the [Troubleshooting Guide](troubleshooting.md)
- Check the [FAQ](faq.md) for common questions
