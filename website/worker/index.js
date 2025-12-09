// Cloudflare Worker - API for weebcast.com
// Deploy with: wrangler deploy

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    
    // CORS headers
    const corsHeaders = {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type',
    };

    if (request.method === 'OPTIONS') {
      return new Response(null, { headers: corsHeaders });
    }

    // API Routes
    if (url.pathname === '/api/activity') {
      return handleGetActivity(env, corsHeaders);
    }
    
    if (url.pathname === '/api/activity/all') {
      return handleGetAllActivity(env, corsHeaders);
    }

    if (url.pathname.startsWith('/api/anime/')) {
      const animeId = url.pathname.split('/')[3];
      return handleGetAnime(env, animeId, corsHeaders);
    }

    if (url.pathname === '/api/trending') {
      return handleGetTrending(env, corsHeaders);
    }

    if (url.pathname === '/api/seasonal') {
      return handleGetSeasonal(env, corsHeaders);
    }

    // POST endpoint for syncing data (local development / operator webhook)
    if (url.pathname === '/api/sync' && request.method === 'POST') {
      return handleSync(request, env, corsHeaders);
    }

    return new Response(JSON.stringify({ 
      error: 'Not found',
      endpoints: [
        '/api/activity - Overall MAL activity',
        '/api/activity/all - All monitors',
        '/api/anime/:id - Specific anime activity',
        '/api/trending - Trending anime list',
        '/api/seasonal - Current season anime',
        'POST /api/sync - Sync data from operator (dev)'
      ]
    }), {
      status: 404,
      headers: { ...corsHeaders, 'Content-Type': 'application/json' }
    });
  },
};

async function handleGetActivity(env, corsHeaders) {
  try {
    const data = await env.WEEBCAST_KV.get('mal-overall', 'json');
    
    if (!data) {
      return new Response(JSON.stringify({
        activityLevel: 'Unknown',
        weebcastStatus: 'No data available yet',
        lastUpdated: null
      }), {
        headers: { ...corsHeaders, 'Content-Type': 'application/json' }
      });
    }

    return new Response(JSON.stringify(data), {
      headers: { ...corsHeaders, 'Content-Type': 'application/json' }
    });
  } catch (error) {
    return new Response(JSON.stringify({ error: error.message }), {
      status: 500,
      headers: { ...corsHeaders, 'Content-Type': 'application/json' }
    });
  }
}

async function handleGetAllActivity(env, corsHeaders) {
  try {
    const keys = await env.WEEBCAST_KV.list();
    const results = [];

    for (const key of keys.keys) {
      const data = await env.WEEBCAST_KV.get(key.name, 'json');
      if (data) {
        results.push(data);
      }
    }

    return new Response(JSON.stringify({ monitors: results }), {
      headers: { ...corsHeaders, 'Content-Type': 'application/json' }
    });
  } catch (error) {
    return new Response(JSON.stringify({ error: error.message }), {
      status: 500,
      headers: { ...corsHeaders, 'Content-Type': 'application/json' }
    });
  }
}

async function handleGetAnime(env, animeId, corsHeaders) {
  try {
    const data = await env.WEEBCAST_KV.get(`anime-${animeId}`, 'json');
    
    if (!data) {
      return new Response(JSON.stringify({ error: 'Anime not being monitored' }), {
        status: 404,
        headers: { ...corsHeaders, 'Content-Type': 'application/json' }
      });
    }

    return new Response(JSON.stringify(data), {
      headers: { ...corsHeaders, 'Content-Type': 'application/json' }
    });
  } catch (error) {
    return new Response(JSON.stringify({ error: error.message }), {
      status: 500,
      headers: { ...corsHeaders, 'Content-Type': 'application/json' }
    });
  }
}

async function handleGetTrending(env, corsHeaders) {
  try {
    const data = await env.WEEBCAST_KV.get('mal-overall', 'json');
    
    if (!data || !data.trendingAnime) {
      return new Response(JSON.stringify({ trending: [] }), {
        headers: { ...corsHeaders, 'Content-Type': 'application/json' }
      });
    }

    return new Response(JSON.stringify({ trending: data.trendingAnime }), {
      headers: { ...corsHeaders, 'Content-Type': 'application/json' }
    });
  } catch (error) {
    return new Response(JSON.stringify({ error: error.message }), {
      status: 500,
      headers: { ...corsHeaders, 'Content-Type': 'application/json' }
    });
  }
}

async function handleGetSeasonal(env, corsHeaders) {
  try {
    const data = await env.WEEBCAST_KV.get('mal-overall', 'json');
    
    if (!data) {
      return new Response(JSON.stringify({ seasonal: [], season: getCurrentSeason() }), {
        headers: { ...corsHeaders, 'Content-Type': 'application/json' }
      });
    }

    // Use seasonalAnime if available, fall back to trendingAnime for backwards compatibility
    const seasonalList = data.seasonalAnime || data.trendingAnime || [];

    return new Response(JSON.stringify({ 
      seasonal: seasonalList,
      season: data.currentSeason || getCurrentSeason()
    }), {
      headers: { ...corsHeaders, 'Content-Type': 'application/json' }
    });
  } catch (error) {
    return new Response(JSON.stringify({ error: error.message }), {
      status: 500,
      headers: { ...corsHeaders, 'Content-Type': 'application/json' }
    });
  }
}

function getCurrentSeason() {
  const now = new Date();
  const month = now.getMonth();
  const year = now.getFullYear();
  
  let season;
  if (month >= 0 && month <= 2) season = 'Winter';
  else if (month >= 3 && month <= 5) season = 'Spring';
  else if (month >= 6 && month <= 8) season = 'Summer';
  else season = 'Fall';
  
  return `${season} ${year}`;
}

// Handle sync from operator (for local development)
async function handleSync(request, env, corsHeaders) {
  try {
    const payload = await request.json();
    const key = payload.key || 'mal-overall';
    
    // Store the activity data
    await env.WEEBCAST_KV.put(key, JSON.stringify({
      monitorName: payload.monitorName,
      animeId: payload.animeId,
      animeName: payload.animeName,
      activityLevel: payload.activityLevel,
      weebcastStatus: payload.weebcastStatus,
      metrics: payload.metrics,
      trendingAnime: payload.trendingAnime,
      seasonalAnime: payload.seasonalAnime,
      currentSeason: payload.currentSeason,
      lastUpdated: payload.lastUpdated || new Date().toISOString()
    }));

    return new Response(JSON.stringify({ success: true, key }), {
      headers: { ...corsHeaders, 'Content-Type': 'application/json' }
    });
  } catch (error) {
    return new Response(JSON.stringify({ error: error.message }), {
      status: 500,
      headers: { ...corsHeaders, 'Content-Type': 'application/json' }
    });
  }
}


