package controller

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	weebcastv1alpha1 "github.com/weebcast/weebcast-operator/api/v1alpha1"
	"github.com/weebcast/weebcast-operator/pkg/mal"
)

// AnimeMonitorReconciler reconciles an AnimeMonitor object
type AnimeMonitorReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	MALClient *mal.Client
}

// +kubebuilder:rbac:groups=weebcast.com,resources=animemonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=weebcast.com,resources=animemonitors/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=weebcast.com,resources=animemonitors/finalizers,verbs=update

// Reconcile handles the reconciliation loop for AnimeMonitor resources
func (r *AnimeMonitorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling AnimeMonitor", "name", req.Name, "namespace", req.Namespace)

	// Fetch the AnimeMonitor instance
	monitor := &weebcastv1alpha1.AnimeMonitor{}
	if err := r.Get(ctx, req.NamespacedName, monitor); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Keep a copy of the original for patching
	monitorPatch := client.MergeFrom(monitor.DeepCopy())

	// Update status phase
	monitor.Status.Phase = "Monitoring"

	// Determine what to monitor
	if monitor.Spec.AnimeID > 0 {
		// Monitor specific anime
		if err := r.reconcileSpecificAnime(ctx, monitor); err != nil {
			logger.Error(err, "Failed to reconcile specific anime")
			r.setErrorCondition(monitor, err)
			if updateErr := r.Status().Patch(ctx, monitor, monitorPatch); updateErr != nil {
				logger.Error(updateErr, "Failed to update status")
			}
			return ctrl.Result{RequeueAfter: time.Minute}, nil
		}
	} else {
		// Monitor overall MAL activity
		if err := r.reconcileOverallActivity(ctx, monitor); err != nil {
			logger.Error(err, "Failed to reconcile overall activity")
			r.setErrorCondition(monitor, err)
			if updateErr := r.Status().Patch(ctx, monitor, monitorPatch); updateErr != nil {
				logger.Error(updateErr, "Failed to update status")
			}
			return ctrl.Result{RequeueAfter: time.Minute}, nil
		}
	}

	// Set success condition
	r.setReadyCondition(monitor)

	// Patch the status (more resilient to concurrent modifications than Update)
	if err := r.Status().Patch(ctx, monitor, monitorPatch); err != nil {
		logger.Error(err, "Failed to update AnimeMonitor status")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully reconciled AnimeMonitor",
		"activityLevel", monitor.Status.ActivityLevel,
		"weebcastStatus", monitor.Status.WeebcastStatus)

	// Calculate requeue interval
	requeueAfter := time.Duration(monitor.Spec.PollingIntervalSeconds) * time.Second
	if requeueAfter == 0 {
		requeueAfter = 5 * time.Minute // Default 5 minutes
	}

	return ctrl.Result{RequeueAfter: requeueAfter}, nil
}

// reconcileSpecificAnime handles monitoring of a specific anime
func (r *AnimeMonitorReconciler) reconcileSpecificAnime(ctx context.Context, monitor *weebcastv1alpha1.AnimeMonitor) error {
	logger := log.FromContext(ctx)
	logger.Info("Fetching specific anime data", "animeID", monitor.Spec.AnimeID)

	// Fetch anime details
	anime, err := r.MALClient.GetAnime(ctx, monitor.Spec.AnimeID)
	if err != nil {
		return fmt.Errorf("fetching anime %d: %w", monitor.Spec.AnimeID, err)
	}

	// Update anime name if not set
	if monitor.Spec.AnimeName == "" && anime.Title != "" {
		monitor.Spec.AnimeName = anime.Title
	}

	// Fetch statistics
	stats, err := r.MALClient.GetAnimeStatistics(ctx, monitor.Spec.AnimeID)
	if err != nil {
		logger.Info("Could not fetch statistics, using basic data", "error", err)
	}

	// Update metrics
	monitor.Status.Metrics = weebcastv1alpha1.AnimeActivityMetrics{
		Score:         anime.Score,
		ScoredByCount: anime.ScoredBy,
		Rank:          anime.Rank,
		Popularity:    anime.Popularity,
		Members:       anime.Members,
		Favorites:     anime.Favorites,
	}

	if stats != nil {
		monitor.Status.Metrics.WatchingCount = stats.Watching
		monitor.Status.Metrics.CompletedCount = stats.Completed
		monitor.Status.Metrics.DroppedCount = stats.Dropped
		monitor.Status.Metrics.PlanToWatchCount = stats.PlanToWatch
		monitor.Status.Metrics.ActiveUsers = stats.Watching + stats.Completed/10 // Rough estimate
	}

	// Calculate activity level based on engagement
	activityScore := calculateActivityScore(monitor.Status.Metrics)
	previousLevel := monitor.Status.ActivityLevel
	monitor.Status.ActivityLevel = r.determineActivityLevel(activityScore, monitor.Spec)

	if previousLevel != monitor.Status.ActivityLevel {
		monitor.Status.LastActivityChange = metav1.Now()
	}

	// Set Weebcast status based on activity
	monitor.Status.WeebcastStatus = r.deriveWeebcastStatus(monitor.Status.ActivityLevel, anime.Title)
	monitor.Status.LastChecked = metav1.Now()
	monitor.Status.Message = fmt.Sprintf("Monitoring '%s' - %d members, %.2f score",
		anime.Title, anime.Members, anime.Score)

	return nil
}

// reconcileOverallActivity handles monitoring of overall MAL activity
func (r *AnimeMonitorReconciler) reconcileOverallActivity(ctx context.Context, monitor *weebcastv1alpha1.AnimeMonitor) error {
	logger := log.FromContext(ctx)
	logger.Info("Fetching overall MAL activity")

	// Get overall activity metrics
	metrics, err := r.MALClient.GetOverallActivity(ctx)
	if err != nil {
		return fmt.Errorf("fetching overall activity: %w", err)
	}

	// Get trending anime (top airing by popularity)
	topAiring, err := r.MALClient.GetTopAiring(ctx, 10)
	if err != nil {
		logger.Info("Could not fetch top airing", "error", err)
	}

	// Get seasonal anime (current season releases)
	seasonalAnime, err := r.MALClient.GetSeasonNow(ctx, 10)
	if err != nil {
		logger.Info("Could not fetch seasonal anime", "error", err)
	}

	// Update metrics
	monitor.Status.Metrics = weebcastv1alpha1.AnimeActivityMetrics{
		ActiveUsers: metrics.TotalActiveUsers,
		Members:     metrics.TotalMembers,
		Score:       metrics.AverageScore,
	}

	// Build trending anime list (top airing)
	monitor.Status.TrendingAnime = make([]weebcastv1alpha1.TrendingAnime, 0, len(topAiring))
	for _, anime := range topAiring {
		monitor.Status.TrendingAnime = append(monitor.Status.TrendingAnime, r.buildTrendingEntry(anime))
	}

	// Build seasonal anime list
	monitor.Status.SeasonalAnime = make([]weebcastv1alpha1.TrendingAnime, 0, len(seasonalAnime))
	for _, anime := range seasonalAnime {
		monitor.Status.SeasonalAnime = append(monitor.Status.SeasonalAnime, r.buildTrendingEntry(anime))
	}

	// Set current season
	monitor.Status.CurrentSeason = getCurrentSeason()

	// Calculate overall activity level
	activityScore := metrics.TotalActiveUsers + (metrics.TotalMembers / 1000)
	previousLevel := monitor.Status.ActivityLevel
	monitor.Status.ActivityLevel = r.determineActivityLevel(activityScore, monitor.Spec)

	if previousLevel != monitor.Status.ActivityLevel {
		monitor.Status.LastActivityChange = metav1.Now()
	}

	// Set Weebcast status
	monitor.Status.WeebcastStatus = r.deriveWeebcastStatus(monitor.Status.ActivityLevel, "")
	monitor.Status.LastChecked = metav1.Now()
	monitor.Status.Message = fmt.Sprintf("Overall MAL Activity: %d active users across %d members, tracking %d trending + %d seasonal anime",
		metrics.TotalActiveUsers, metrics.TotalMembers, len(topAiring), len(seasonalAnime))

	return nil
}

// calculateActivityScore computes a normalized activity score from metrics
func calculateActivityScore(metrics weebcastv1alpha1.AnimeActivityMetrics) int {
	// Weight different factors to determine activity
	score := 0
	score += metrics.ActiveUsers * 10
	score += metrics.WatchingCount * 5
	score += metrics.Members / 1000
	score += metrics.Favorites / 100
	if metrics.Score > 8.0 {
		score += 500 // High-rated anime tend to have more engagement
	}
	return score
}

// determineActivityLevel maps an activity score to an activity level
func (r *AnimeMonitorReconciler) determineActivityLevel(score int, spec weebcastv1alpha1.AnimeMonitorSpec) weebcastv1alpha1.ActivityLevel {
	highThreshold := spec.HighActivityThreshold
	mediumThreshold := spec.MediumActivityThreshold

	if highThreshold == 0 {
		highThreshold = 1000
	}
	if mediumThreshold == 0 {
		mediumThreshold = 500
	}

	criticalThreshold := highThreshold * 2

	switch {
	case score >= criticalThreshold:
		return weebcastv1alpha1.ActivityLevelCritical
	case score >= highThreshold:
		return weebcastv1alpha1.ActivityLevelHigh
	case score >= mediumThreshold:
		return weebcastv1alpha1.ActivityLevelMedium
	default:
		return weebcastv1alpha1.ActivityLevelLow
	}
}

// deriveWeebcastStatus generates a Weebcast weather forecast status message
func (r *AnimeMonitorReconciler) deriveWeebcastStatus(level weebcastv1alpha1.ActivityLevel, animeName string) string {
	var statusPrefix string
	if animeName != "" {
		statusPrefix = fmt.Sprintf("[%s] ", animeName)
	}

	switch level {
	case weebcastv1alpha1.ActivityLevelCritical:
		return statusPrefix + "🌀 TYPHOON ALERT! A category 5 weeb storm is making landfall! Extreme anime energy detected across all sectors. This is not a drill - expect maximum hype levels, server strain, and spontaneous waifu debates. All weebs advised to secure their watchlists!"
	case weebcastv1alpha1.ActivityLevelHigh:
		return statusPrefix + "⛈️ STORM WARNING! A massive weeb front is moving in! Heavy anime discussions expected with a high chance of trending hashtags. Take shelter in your favorite streaming site - it's going to be a wild one!"
	case weebcastv1alpha1.ActivityLevelMedium:
		return statusPrefix + "⛅ Partly cloudy conditions in the weeb-o-sphere. Moderate anime activity detected with occasional bursts of excitement. Good conditions for casual binge-watching. Keep an umbrella ready for surprise episode drops!"
	default:
		return statusPrefix + "☀️ Clear skies across the anime landscape! A peaceful day in the weeb-o-sphere. Perfect weather for catching up on your backlog or discovering hidden gems. Enjoy the calm before the next seasonal storm!"
	}
}

// setReadyCondition sets the Ready condition to True
func (r *AnimeMonitorReconciler) setReadyCondition(monitor *weebcastv1alpha1.AnimeMonitor) {
	meta.SetStatusCondition(&monitor.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             "MonitoringActive",
		Message:            "Successfully fetched MAL activity data",
		LastTransitionTime: metav1.Now(),
	})
}

// setErrorCondition sets the Ready condition to False with error details
func (r *AnimeMonitorReconciler) setErrorCondition(monitor *weebcastv1alpha1.AnimeMonitor, err error) {
	monitor.Status.Phase = "Error"
	monitor.Status.Message = fmt.Sprintf("Error: %v", err)
	meta.SetStatusCondition(&monitor.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionFalse,
		Reason:             "MonitoringFailed",
		Message:            err.Error(),
		LastTransitionTime: metav1.Now(),
	})
}

// buildTrendingEntry creates a TrendingAnime entry from MAL anime data
func (r *AnimeMonitorReconciler) buildTrendingEntry(anime mal.AnimeData) weebcastv1alpha1.TrendingAnime {
	activityLevel := weebcastv1alpha1.ActivityLevelLow
	if anime.Members > 1000000 {
		activityLevel = weebcastv1alpha1.ActivityLevelHigh
	} else if anime.Members > 500000 {
		activityLevel = weebcastv1alpha1.ActivityLevelMedium
	}

	imageURL := ""
	if anime.Images.JPG.ImageURL != "" {
		imageURL = anime.Images.JPG.ImageURL
	}

	return weebcastv1alpha1.TrendingAnime{
		ID:            anime.MalID,
		Title:         anime.Title,
		Score:         anime.Score,
		Members:       anime.Members,
		ActivityLevel: activityLevel,
		ImageURL:      imageURL,
	}
}

// getCurrentSeason returns the current anime season (e.g., "Winter 2025")
func getCurrentSeason() string {
	now := time.Now()
	month := now.Month()
	year := now.Year()

	var season string
	switch {
	case month >= 1 && month <= 3:
		season = "Winter"
	case month >= 4 && month <= 6:
		season = "Spring"
	case month >= 7 && month <= 9:
		season = "Summer"
	default:
		season = "Fall"
	}

	return fmt.Sprintf("%s %d", season, year)
}

// SetupWithManager sets up the controller with the Manager
func (r *AnimeMonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&weebcastv1alpha1.AnimeMonitor{}).
		Complete(r)
}
