#!/bin/bash
# Sync AnimeMonitor data from Kubernetes to local API worker

API_URL="${API_URL:-http://localhost:8787}"

echo "Syncing AnimeMonitor data to $API_URL..."

# Get all AnimeMonitors as JSON
monitors=$(kubectl get animemonitors -A -o json)

if [ $? -ne 0 ]; then
    echo "Error: Failed to get AnimeMonitors from Kubernetes"
    exit 1
fi

# Process each monitor
echo "$monitors" | jq -c '.items[]' | while read -r item; do
    name=$(echo "$item" | jq -r '.metadata.name')
    
    # Determine the key for storage
    anime_id=$(echo "$item" | jq -r '.spec.animeId // empty')
    if [ -n "$anime_id" ] && [ "$anime_id" != "null" ] && [ "$anime_id" != "0" ]; then
        key="anime-$anime_id"
    elif [[ "$name" == *"overall"* ]]; then
        key="mal-overall"
    else
        key="$name"
    fi
    
    # Build the payload
    payload=$(echo "$item" | jq --arg key "$key" '{
        key: $key,
        monitorName: .metadata.name,
        animeId: (.spec.animeId // null),
        animeName: (.spec.animeName // null),
        activityLevel: .status.activityLevel,
        weebcastStatus: .status.weebcastStatus,
        metrics: {
            activeUsers: (.status.metrics.activeUsers // 0),
            watchingCount: (.status.metrics.watchingCount // 0),
            members: (.status.metrics.members // 0),
            score: (.status.metrics.score // 0),
            rank: (.status.metrics.rank // 0),
            favorites: (.status.metrics.favorites // 0)
        },
        trendingAnime: (.status.trendingAnime // []),
        lastUpdated: .status.lastChecked
    }')
    
    # Send to local API
    response=$(curl -s -X POST "$API_URL/api/sync" \
        -H "Content-Type: application/json" \
        -d "$payload")
    
    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        echo "✓ Synced: $name -> $key"
    else
        echo "✗ Failed to sync $name: $response"
    fi
done

echo ""
echo "Done! View the dashboard at http://localhost:8000"

