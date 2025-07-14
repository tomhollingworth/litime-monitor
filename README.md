# LiTime Monitor

A Go application that connects to LiTime MPPT Solar Charge Controllers via Bluetooth Low Energy (BLE) to collect real-time metrics and send them to InfluxDB for monitoring and analysis.

## Description

This application monitors solar charge controller data from LiTime MPPT controllers by establishing a Bluetooth connection and continuously collecting metrics such as:

- Battery voltage, current, and power
- Controller temperature
- Load voltage, current, and power
- Solar panel voltage
- Daily energy statistics
- Total energy and running days

The collected data is automatically sent to an InfluxDB instance for storage, enabling you to create dashboards and monitor your solar power system's performance over time.

## Features

- Real-time data collection via Bluetooth Low Energy
- Automatic InfluxDB integration
- Configurable via command line arguments, environment variables, or YAML config file
- Robust error handling for Bluetooth communication
- Periodic logging of collected metrics

## Installation

Make sure you have Go installed, then build the application:

```bash
go build -o litime-monitor
```

## Usage

```bash
./litime-monitor [flags]
```

### Command Line Arguments

| Flag | Default | Description |
|------|---------|-------------|
| `--influxdb.url` | `http://127.0.0.1:8086` | InfluxDB server URL |
| `--influxdb.token` | `my-token` | InfluxDB authentication token |
| `--influxdb.org` | `my-org` | InfluxDB organization name |
| `--influxdb.bucket` | `solar_charge_controller` | InfluxDB bucket name for data storage |
| `--device.address` | `00:00:00:00:00:00` | Bluetooth MAC address of your LiTime controller |
| `--device.service` | `0000ffe0-0000-1000-8000-00805f9b34fb` | Bluetooth service UUID |
| `--device.characteristic` | `0000ffe1-0000-1000-8000-00805f9b34fb` | Bluetooth characteristic UUID |
| `-h, --help` | | Show help information |

### Configuration

The application supports multiple configuration methods (in order of precedence):

1. **Command line flags** (highest priority)
2. **Environment variables** with `LITIME_MONITOR_` prefix
3. **Configuration files** in YAML format

#### Environment Variables

Prefix all configuration keys with `LITIME_MONITOR_` and use uppercase:

```bash
export LITIME_MONITOR_INFLUXDB_URL="http://your-influxdb:8086"
export LITIME_MONITOR_INFLUXDB_TOKEN="your-token"
export LITIME_MONITOR_DEVICE_ADDRESS="AA:BB:CC:DD:EE:FF"
```

#### Configuration File

Create a `config.yaml` file in one of these locations:
- `/etc/litime-monitor/config.yaml`
- `$HOME/.litime-monitor/config.yaml`
- `./config.yaml` (current directory)

Example configuration:

```yaml
influxdb:
  url: "http://your-influxdb:8086"
  token: "your-influxdb-token"
  org: "your-org"
  bucket: "solar_data"

device:
  address: "AA:BB:CC:DD:EE:FF"
  service: "0000ffe0-0000-1000-8000-00805f9b34fb"
  characteristic: "0000ffe1-0000-1000-8000-00805f9b34fb"
```

## Setup

1. **Find your LiTime controller's Bluetooth address**: Use your system's Bluetooth settings or a BLE scanner app to identify your devices bluetooth address
2. **Set up InfluxDB**: Ensure you have an InfluxDB instance running and create a bucket for the data
3. **Configure the application**: Set the device address and InfluxDB connection details
4. **Run the application**: The monitor will automatically connect and start collecting data

## Data Schema

Data is stored in InfluxDB with the measurement name `solar_charge_controller` and the following fields:

- `battery_voltage` (float)
- `battery_current` (float) 
- `battery_power` (int)
- `controller_temperature` (float)
- `load_voltage` (float)
- `load_current` (float)
- `load_power` (float)
- `panel_voltage` (float)
- `max_charge_power` (int)
- `energy_today` (int)
- `running_days` (int)
- `total_energy` (int)

## Contributing

Contributions are welcome! I'm particularly interested in:

- **Additional data sinks**: Support for other time-series databases (Prometheus, TimescaleDB, etc.)
- **Improved response data handling**: Better parsing and validation of controller responses
- **Additional device support**: Compatibility with other MPPT controller brands

Please feel free to open issues or submit pull requests.

## Changelog

### v1.0.0 (2025-07-14)
- Initial release
- Bluetooth Low Energy connectivity for LiTime MPPT controllers
- InfluxDB data sink integration
- Configurable via CLI args, environment variables, and YAML files
- Real-time monitoring of battery, load, and solar panel metrics
- Robust error handling and logging

## License

This project is open source. Please check the license file for details.
