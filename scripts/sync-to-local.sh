#!/bin/bash
# Sync AnimeMonitor data from Kubernetes to local API worker

API_URL="${API_URL:-http://localhost:8787}"

echo "🌀 Syncing Weeb Weather Data to $API_URL..."
echo ""

# Get all AnimeMonitors as JSON
monitors=$(kubectl get animemonitors -A -o json)

if [ $? -ne 0 ]; then
    echo "❌ Error: Failed to get AnimeMonitors from Kubernetes"
    exit 1
fi

# Process each monitor
echo "$monitors" | jq -c '.items[]' | while read -r item; do
    name=$(echo "$item" | jq -r '.metadata.name')
    activity=$(echo "$item" | jq -r '.status.activityLevel // "Unknown"')
    
    # Determine the key for storage
    anime_id=$(echo "$item" | jq -r '.spec.animeId // empty')
    anime_name=$(echo "$item" | jq -r '.spec.animeName // empty')
    
    if [ -n "$anime_id" ] && [ "$anime_id" != "null" ] && [ "$anime_id" != "0" ]; then
        key="anime-$anime_id"
    elif [[ "$name" == *"overall"* ]]; then
        key="mal-overall"
    else
        key="$name"
    fi
    
    # Build the payload - extract animeName from weebcastStatus if not in spec
    # The controller puts [AnimeName] at the start of the status message
    if [ -z "$anime_name" ] || [ "$anime_name" == "null" ]; then
        # Try to extract from status message like "[Shingeki no Kyojin] ⛈️ STORM..."
        extracted=$(echo "$item" | jq -r '.status.weebcastStatus // ""' | sed -n 's/^\[\([^]]*\)\].*/\1/p')
        if [ -n "$extracted" ]; then
            anime_name="$extracted"
        fi
    fi
    
    # Build the payload
    payload=$(echo "$item" | jq --arg key "$key" --arg animeName "$anime_name" '{
        key: $key,
        monitorName: .metadata.name,
        animeId: (.spec.animeId // null),
        animeName: (if $animeName != "" and $animeName != "null" then $animeName else null end),
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
    
    # Weather icon based on activity
    case "$activity" in
        "Critical") icon="🌀" ;;
        "High") icon="⛈️" ;;
        "Medium") icon="⛅" ;;
        "Low") icon="☀️" ;;
        *) icon="❓" ;;
    esac
    
    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        display_name="${anime_name:-$name}"
        echo "$icon $display_name → $activity"
    else
        echo "❌ Failed: $name - $response"
    fi
done

echo ""
echo "✅ Sync complete! View the forecast at http://localhost:8000"

