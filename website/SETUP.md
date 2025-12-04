# Weebcast.com Setup Guide

## Prerequisites
- Domain: weebcast.com (or similar)
- Cloudflare account (free tier works)
- Node.js installed locally
- Kubernetes cluster (for the operator)

## Step 1: Domain Setup with Cloudflare

### 1.1 Add Domain to Cloudflare
1. Go to [Cloudflare Dashboard](https://dash.cloudflare.com)
2. Click "Add a Site" → Enter `weebcast.com`
3. Select Free plan
4. Cloudflare will scan DNS records
5. Update your domain's nameservers to Cloudflare's (shown in dashboard)

### 1.2 Configure DNS
Add these DNS records:
```
Type  Name              Content
A     weebcast.com      192.0.2.1 (proxied) - placeholder, Pages will override
CNAME www               weebcast.com (proxied)
CNAME api               <your-worker>.workers.dev (proxied)
```

## Step 2: Set Up Cloudflare Workers KV

### 2.1 Install Wrangler CLI
```bash
npm install -g wrangler
wrangler login
```

### 2.2 Create KV Namespace
```bash
cd website/worker
wrangler kv:namespace create "WEEBCAST_KV"
```

Copy the output ID and update `wrangler.toml`:
```toml
[[kv_namespaces]]
binding = "WEEBCAST_KV"
id = "YOUR_ACTUAL_NAMESPACE_ID"
```

### 2.3 Deploy the Worker
```bash
wrangler deploy
```

### 2.4 Add Custom Domain (Optional)
1. Go to Workers & Pages → your worker → Settings → Triggers
2. Add Custom Domain: `api.weebcast.com`

## Step 3: Deploy Frontend to Cloudflare Pages

### 3.1 Via Dashboard
1. Go to Workers & Pages → Create → Pages
2. Connect to Git or Direct Upload
3. Upload the `website/frontend/` folder
4. Set custom domain: `weebcast.com`

### 3.2 Via CLI
```bash
cd website/frontend
wrangler pages deploy . --project-name=weebcast
```

### 3.3 Update API URL
Edit `index.html` and update `API_BASE`:
```javascript
const API_BASE = 'https://api.weebcast.com';
```

## Step 4: Configure the Operator

### 4.1 Create Cloudflare API Token
1. Go to Cloudflare Dashboard → My Profile → API Tokens
2. Create Token → Edit Cloudflare Workers (template)
3. Add permissions: Workers KV Storage (Edit)
4. Save the token securely

### 4.2 Create Kubernetes Secret
```bash
kubectl create secret generic cloudflare-credentials \
  --from-literal=account-id=YOUR_ACCOUNT_ID \
  --from-literal=kv-namespace-id=YOUR_KV_NAMESPACE_ID \
  --from-literal=api-token=YOUR_API_TOKEN \
  -n weebcast-system
```

### 4.3 Deploy the Operator
```bash
# From the operator root directory
make deploy
```

### 4.4 Create AnimeMonitors
```bash
kubectl apply -f config/samples/
```

## Step 5: Verify Everything Works

### Test the API
```bash
curl https://api.weebcast.com/api/activity
curl https://api.weebcast.com/api/trending
```

### Check the Website
Visit https://weebcast.com

### Monitor Operator Logs
```bash
kubectl logs -f -n weebcast-system deployment/weebcast-operator
```

## Cloudflare Services Summary

| Service | Purpose | Cost |
|---------|---------|------|
| **DNS** | Domain management | Free |
| **CDN** | Cache & deliver frontend | Free |
| **Pages** | Host static frontend | Free |
| **Workers** | API endpoints | Free (100k req/day) |
| **Workers KV** | Store activity data | Free (100k reads/day) |
| **DDoS Protection** | Automatic protection | Free |
| **SSL/TLS** | HTTPS certificates | Free |
| **Web Analytics** | Basic analytics | Free |

## Optional Enhancements

### Enable Cloudflare Web Analytics
1. Dashboard → Analytics → Web Analytics
2. Add site: weebcast.com
3. Auto-installed via Cloudflare proxy

### Set Up Caching Rules
1. Dashboard → Caching → Cache Rules
2. Create rule for `/api/*` with 1-minute TTL
3. Create rule for static assets with 1-day TTL

### Enable Rate Limiting (Pro)
1. Dashboard → Security → WAF → Rate Limiting
2. Limit `/api/*` to 100 req/minute per IP

## Troubleshooting

### Worker not receiving data?
Check operator logs:
```bash
kubectl logs -n weebcast-system deployment/weebcast-operator | grep cloudflare
```

### KV not updating?
Test manually:
```bash
wrangler kv:key put --namespace-id=YOUR_ID "test" "hello"
wrangler kv:key get --namespace-id=YOUR_ID "test"
```

### CORS errors?
The worker includes CORS headers. Ensure the frontend URL matches.


