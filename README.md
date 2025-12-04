# 🎌 Weebcast Operator

A Kubernetes operator for [weebcast.com](https://weebcast.com) that monitors MyAnimeList (MAL) activity to predict and track anime-related traffic patterns.

## Overview

The Weebcast Operator watches MyAnimeList activity levels and translates them into actionable insights for weebcast.com. When anime activity is high on MAL (new episode releases, trending shows, seasonal premieres), this typically correlates with increased traffic on anime-related sites.

### Features

- **📊 Overall MAL Activity Monitoring** - Track global anime community engagement
- **🎯 Specific Anime Monitoring** - Monitor individual anime by MAL ID
- **📈 Activity Level Tracking** - Low, Medium, High, and Critical activity levels
- **🔥 Trending Anime Detection** - Identify currently trending shows
- **⚡ Weebcast Status Derivation** - Automatic status messages for weebcast.com
- **🔔 Configurable Thresholds** - Customize activity level boundaries
- **📡 Webhook Notifications** - Get notified when activity is high

## Installation

### Prerequisites

- Kubernetes cluster (v1.25+)
- kubectl configured to access your cluster
- (Optional) Docker for building custom images

### Quick Start

1. **Install the CRD and operator:**

```bash
# Clone the repository
git clone https://github.com/weebcast/weebcast-operator.git
cd weebcast-operator

# Deploy to your cluster
make deploy
```

2. **Create an AnimeMonitor resource:**

```yaml
apiVersion: weebcast.com/v1alpha1
kind: AnimeMonitor
metadata:
  name: mal-activity
spec:
  pollingIntervalSeconds: 300
  highActivityThreshold: 1000
  mediumActivityThreshold: 500
```

3. **Check the status:**

```bash
kubectl get animemonitors
```

## Custom Resource Definition

### AnimeMonitor Spec

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `animeId` | int | - | MAL ID of specific anime to monitor (optional) |
| `animeName` | string | - | Display name for the anime |
| `pollingIntervalSeconds` | int | 300 | How often to check MAL (min: 60) |
| `highActivityThreshold` | int | 1000 | Threshold for "High" activity |
| `mediumActivityThreshold` | int | 500 | Threshold for "Medium" activity |
| `notifyOnHighActivity` | bool | false | Enable webhook notifications |
| `webhookUrl` | string | - | URL for activity notifications |

### AnimeMonitor Status

| Field | Description |
|-------|-------------|
| `phase` | Current phase (Initializing, Monitoring, Error) |
| `activityLevel` | Low, Medium, High, or Critical |
| `weebcastStatus` | Human-readable status for weebcast.com |
| `metrics` | Detailed activity metrics |
| `trendingAnime` | List of currently trending anime |
| `lastChecked` | Timestamp of last MAL check |
| `lastActivityChange` | When activity level changed |

## Examples

### Monitor Overall MAL Activity

```yaml
apiVersion: weebcast.com/v1alpha1
kind: AnimeMonitor
metadata:
  name: mal-overall
spec:
  pollingIntervalSeconds: 300
  highActivityThreshold: 1000
```

### Monitor a Specific Anime

```yaml
apiVersion: weebcast.com/v1alpha1
kind: AnimeMonitor
metadata:
  name: attack-on-titan
spec:
  animeId: 16498
  animeName: "Shingeki no Kyojin"
  pollingIntervalSeconds: 300
  highActivityThreshold: 5000
```

### Monitor Currently Airing Anime

```yaml
apiVersion: weebcast.com/v1alpha1
kind: AnimeMonitor
metadata:
  name: solo-leveling
spec:
  animeId: 52299
  animeName: "Solo Leveling"
  pollingIntervalSeconds: 180
  notifyOnHighActivity: true
```

## Activity Levels

| Level | Emoji | Description | Weebcast Impact |
|-------|-------|-------------|-----------------|
| **Low** | 💤 | Baseline activity | Normal traffic |
| **Medium** | 📈 | Moderate interest | Steady traffic |
| **High** | ⚡ | Popular anime trending | Elevated traffic |
| **Critical** | 🔥 | Major anime event | Traffic surge |

## Usage

### View All Monitors

```bash
kubectl get animemonitors -A
```

### Detailed Status

```bash
kubectl describe animemonitor mal-overall
```

### Watch Activity Changes

```bash
kubectl get animemonitors -w
```

### View Operator Logs

```bash
make logs
# or
kubectl logs -f -n weebcast-system deployment/weebcast-operator
```

## Development

### Build Locally

```bash
# Build the binary
make build

# Run locally (requires kubeconfig)
make run
```

### Build Docker Image

```bash
# Build image
make docker-build IMG=your-registry/weebcast-operator:tag

# Push image
make docker-push IMG=your-registry/weebcast-operator:tag
```

### Run Tests

```bash
make test
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                        │
│  ┌────────────────────────────────────────────────────────┐ │
│  │                  Weebcast Operator                      │ │
│  │  ┌──────────────┐    ┌──────────────┐                  │ │
│  │  │  Controller  │───▶│  MAL Client  │───▶ Jikan API    │ │
│  │  └──────────────┘    └──────────────┘                  │ │
│  │         │                                               │ │
│  │         ▼                                               │ │
│  │  ┌──────────────┐                                      │ │
│  │  │AnimeMonitor  │ ◀── Custom Resource                  │ │
│  │  │   Status     │                                      │ │
│  │  └──────────────┘                                      │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## API Reference

The operator uses the [Jikan API](https://jikan.moe/) (unofficial MyAnimeList API) to fetch:
- Anime details and statistics
- Top airing anime
- Current season anime
- User engagement metrics

**Note:** Jikan has rate limits. The operator respects these limits and will retry on rate limit errors.

## Troubleshooting

### Common Issues

**Rate Limited:**
```
Error: rate limited by MAL API, retry later
```
Solution: Increase `pollingIntervalSeconds` to reduce API calls.

**Anime Not Found:**
```
Error: unexpected status code: 404
```
Solution: Verify the `animeId` is correct on MyAnimeList.

**No Activity Data:**
Check that the operator has network access to `api.jikan.moe`.

## License

MIT License - See [LICENSE](LICENSE) for details.

## Contributing

Contributions welcome! Please read our contributing guidelines before submitting PRs.

---

Made with ❤️ for the anime community by [weebcast.com](https://weebcast.com)

