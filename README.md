# 🌀 Weebcast Operator

A Kubernetes operator for [weebcast.com](https://weebcast.com) - your anime weather forecast! Monitors MyAnimeList (MAL) activity to predict weeb storms and traffic patterns.

## Overview

The Weebcast Operator scans the weeb-o-sphere and translates MyAnimeList activity into weather forecasts for anime traffic. When a weeb storm is brewing (new episode releases, trending shows, seasonal premieres), we'll let you know before it makes landfall!

### Features

- **🌡️ Weeb-o-sphere Scanning** - Track global anime community activity levels
- **🎯 Storm Tracking** - Monitor individual anime by MAL ID  
- **🌀 Weather Conditions** - Clear ☀️, Cloudy ⛅, Stormy ⛈️, and Typhoon 🌀 alerts
- **📡 Trending Detection** - Identify incoming storm fronts (trending shows)
- **📺 Broadcast Reports** - Automated weather forecast messages
- **⚙️ Configurable Thresholds** - Customize sensitivity levels
- **🔔 Storm Alerts** - Webhook notifications for severe weeb weather

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

## Weeb Weather Forecast Levels

| Condition | Icon | Description | Weebcast Impact |
|-----------|------|-------------|-----------------|
| **Clear Skies** | ☀️ | Peaceful day in the weeb-o-sphere | Normal traffic, perfect for backlog |
| **Cloudy** | ⛅ | Activity brewing on the horizon | Moderate traffic, stay alert |
| **Storm Warning** | ⛈️ | Heavy weeb traffic incoming! | Elevated traffic, trending hashtags |
| **Typhoon Alert** | 🌀 | MAXIMUM WEEB ENERGY DETECTED | Traffic surge, server strain expected |

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

## Local Development

### Prerequisites

- Go 1.21+
- Access to a Kubernetes cluster (local or remote)
- kubectl configured to access your cluster

### Quick Start (Full Stack)

Run the operator, API, and frontend all at once:

```bash
make dev
```

This starts:
- **Operator** on your machine (connected to your K8s cluster)
- **API Worker** at http://localhost:8787
- **Frontend** at http://localhost:8000

In another terminal, create a sample monitor:

```bash
kubectl apply -f config/samples/weebcast_v1alpha1_animemonitor.yaml
```

Press `Ctrl+C` to stop all services.

### Running Components Individually

If you prefer running components separately:

```bash
# Terminal 1: Operator
make dev-operator

# Terminal 2: API Worker  
make dev-api

# Terminal 3: Frontend
make dev-frontend
```

### Check Status

```bash
make status
# or
kubectl get animemonitors -A -o wide
kubectl describe animemonitor mal-activity
```

### Build from Source

```bash
# Build the binary
make build

# The binary is output to bin/manager
./bin/manager
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

## Website

The project includes a real-time dashboard website that displays anime activity data from the operator.

### Components

```
website/
├── frontend/          # Static HTML/CSS/JS dashboard
│   └── index.html     # Single-page app with live activity display
├── worker/            # Cloudflare Worker API
│   ├── index.js       # API endpoints for activity data
│   └── wrangler.toml  # Cloudflare Worker config
└── SETUP.md           # Detailed deployment guide
```

### Quick Local Preview

To preview the frontend locally:

```bash
cd website/frontend

# Using Python
python -m http.server 8000

# Using Node.js
npx serve .
```

Then open http://localhost:8000 in your browser.

> **Note:** The frontend will show demo data when the API is unavailable, so you can preview the UI without deploying the worker.

### Deploying the Website

The website is designed to run on Cloudflare's free tier:

1. **Deploy the API Worker:**

```bash
cd website/worker

# Install Wrangler CLI
npm install -g wrangler
wrangler login

# Create KV namespace for storing activity data
wrangler kv:namespace create "WEEBCAST_KV"
# Update wrangler.toml with the namespace ID from output

# Deploy
wrangler deploy
```

2. **Deploy the Frontend:**

```bash
cd website/frontend
wrangler pages deploy . --project-name=weebcast
```

3. **Connect the Operator:**

Create a Kubernetes secret with Cloudflare credentials:

```bash
kubectl create secret generic cloudflare-credentials \
  --from-literal=account-id=YOUR_ACCOUNT_ID \
  --from-literal=kv-namespace-id=YOUR_KV_NAMESPACE_ID \
  --from-literal=api-token=YOUR_API_TOKEN \
  -n weebcast-system
```

See [website/SETUP.md](website/SETUP.md) for the complete deployment guide.

### API Endpoints

The Cloudflare Worker exposes these endpoints:

| Endpoint | Description |
|----------|-------------|
| `GET /api/activity` | Overall MAL activity status |
| `GET /api/activity/all` | All monitored anime statuses |
| `GET /api/anime/:id` | Specific anime by MAL ID |
| `GET /api/trending` | Currently trending anime list |

Example:

```bash
curl https://api.weebcast.com/api/activity
curl https://api.weebcast.com/api/trending
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

