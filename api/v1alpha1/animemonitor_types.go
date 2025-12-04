package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AnimeMonitorSpec defines the desired state of AnimeMonitor
type AnimeMonitorSpec struct {
	// AnimeID is the MyAnimeList ID of a specific anime to monitor (optional)
	// If not set, monitors overall MAL activity
	// +optional
	AnimeID int `json:"animeId,omitempty"`

	// AnimeName is the name of the anime being monitored (for display purposes)
	// +optional
	AnimeName string `json:"animeName,omitempty"`

	// PollingIntervalSeconds defines how often to check MAL activity
	// +kubebuilder:default=300
	// +kubebuilder:validation:Minimum=60
	PollingIntervalSeconds int `json:"pollingIntervalSeconds,omitempty"`

	// HighActivityThreshold defines the threshold for "high" activity level
	// Represents number of active users/interactions per minute
	// +kubebuilder:default=1000
	HighActivityThreshold int `json:"highActivityThreshold,omitempty"`

	// MediumActivityThreshold defines the threshold for "medium" activity level
	// +kubebuilder:default=500
	MediumActivityThreshold int `json:"mediumActivityThreshold,omitempty"`

	// NotifyOnHighActivity enables notifications when activity is high
	// +optional
	NotifyOnHighActivity bool `json:"notifyOnHighActivity,omitempty"`

	// WebhookURL for sending activity notifications
	// +optional
	WebhookURL string `json:"webhookUrl,omitempty"`
}

// ActivityLevel represents the current activity state
// +kubebuilder:validation:Enum=Low;Medium;High;Critical
type ActivityLevel string

const (
	ActivityLevelLow      ActivityLevel = "Low"
	ActivityLevelMedium   ActivityLevel = "Medium"
	ActivityLevelHigh     ActivityLevel = "High"
	ActivityLevelCritical ActivityLevel = "Critical"
)

// AnimeActivityMetrics contains detailed activity metrics
type AnimeActivityMetrics struct {
	// ActiveUsers is the estimated number of active users
	ActiveUsers int `json:"activeUsers,omitempty"`

	// WatchingCount is the number of users currently watching
	WatchingCount int `json:"watchingCount,omitempty"`

	// CompletedCount is the number of users who completed
	CompletedCount int `json:"completedCount,omitempty"`

	// DroppedCount is the number of users who dropped
	DroppedCount int `json:"droppedCount,omitempty"`

	// PlanToWatchCount is the number of users planning to watch
	PlanToWatchCount int `json:"planToWatchCount,omitempty"`

	// Score is the current MAL score
	Score float64 `json:"score,omitempty"`

	// ScoredByCount is the number of users who scored
	ScoredByCount int `json:"scoredByCount,omitempty"`

	// Rank is the current popularity rank
	Rank int `json:"rank,omitempty"`

	// Popularity is the popularity score
	Popularity int `json:"popularity,omitempty"`

	// Members is the total member count
	Members int `json:"members,omitempty"`

	// Favorites is the favorites count
	Favorites int `json:"favorites,omitempty"`
}

// TrendingAnime represents a trending anime entry
type TrendingAnime struct {
	// ID is the MAL anime ID
	ID int `json:"id"`

	// Title is the anime title
	Title string `json:"title"`

	// Score is the current score
	Score float64 `json:"score,omitempty"`

	// Members is the member count
	Members int `json:"members,omitempty"`

	// ActivityLevel for this specific anime
	ActivityLevel ActivityLevel `json:"activityLevel,omitempty"`

	// ImageURL is the cover image URL
	ImageURL string `json:"imageUrl,omitempty"`
}

// AnimeMonitorStatus defines the observed state of AnimeMonitor
type AnimeMonitorStatus struct {
	// Phase represents the current phase of the monitor
	// +kubebuilder:default=Initializing
	Phase string `json:"phase,omitempty"`

	// ActivityLevel indicates the current activity level
	ActivityLevel ActivityLevel `json:"activityLevel,omitempty"`

	// WeebcastStatus indicates the derived status for weebcast.com
	// High MAL activity = High Weebcast engagement expected
	WeebcastStatus string `json:"weebcastStatus,omitempty"`

	// Metrics contains detailed activity metrics
	Metrics AnimeActivityMetrics `json:"metrics,omitempty"`

	// TrendingAnime lists currently trending anime on MAL
	// +optional
	TrendingAnime []TrendingAnime `json:"trendingAnime,omitempty"`

	// LastChecked is the timestamp of the last activity check
	LastChecked metav1.Time `json:"lastChecked,omitempty"`

	// LastActivityChange is when the activity level last changed
	LastActivityChange metav1.Time `json:"lastActivityChange,omitempty"`

	// Message provides additional context about the current status
	Message string `json:"message,omitempty"`

	// Conditions represent the latest available observations
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Activity",type="string",JSONPath=".status.activityLevel",description="Current activity level"
// +kubebuilder:printcolumn:name="Weebcast",type="string",JSONPath=".status.weebcastStatus",description="Weebcast status"
// +kubebuilder:printcolumn:name="Score",type="number",JSONPath=".status.metrics.score",description="MAL Score"
// +kubebuilder:printcolumn:name="Members",type="integer",JSONPath=".status.metrics.members",description="Member count"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// AnimeMonitor is the Schema for the animemonitors API
type AnimeMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AnimeMonitorSpec   `json:"spec,omitempty"`
	Status AnimeMonitorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AnimeMonitorList contains a list of AnimeMonitor
type AnimeMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AnimeMonitor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AnimeMonitor{}, &AnimeMonitorList{})
}

