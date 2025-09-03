package ui

import (
	"github.com/opd-ai/desktop-companion/internal/character"
	"testing"
	"time"
)

func TestNewAchievementNotification(t *testing.T) {
	notification := NewAchievementNotification()

	if notification == nil {
		t.Fatal("NewAchievementNotification returned nil")
	}

	if notification.IsVisible() {
		t.Error("New notification should start hidden")
	}

	if notification.background == nil {
		t.Error("Background should be initialized")
	}

	if notification.titleLabel == nil {
		t.Error("Title label should be initialized")
	}

	if notification.descLabel == nil {
		t.Error("Description label should be initialized")
	}

	if notification.container == nil {
		t.Error("Container should be initialized")
	}
}

func TestAchievementNotificationShowAchievement(t *testing.T) {
	notification := NewAchievementNotification()

	achievementDetails := character.AchievementDetails{
		Name:        "Test Achievement",
		Description: "This is a test achievement",
		Timestamp:   time.Now(),
	}

	notification.ShowAchievement(achievementDetails)

	if !notification.IsVisible() {
		t.Error("Notification should be visible after ShowAchievement")
	}

	// Check if content was updated - we can't easily verify the exact text due to markdown parsing,
	// but we can verify the notification is showing
	if notification.titleLabel == nil {
		t.Error("Title label should exist")
	}

	if notification.descLabel == nil {
		t.Error("Description label should exist")
	}
}

func TestAchievementNotificationShowHide(t *testing.T) {
	notification := NewAchievementNotification()

	// Initially hidden
	if notification.IsVisible() {
		t.Error("Notification should start hidden")
	}

	// Show notification
	notification.Show()
	if !notification.IsVisible() {
		t.Error("Notification should be visible after Show()")
	}

	// Hide notification
	notification.Hide()
	if notification.IsVisible() {
		t.Error("Notification should be hidden after Hide()")
	}
}

func TestAchievementNotificationAutoHide(t *testing.T) {
	notification := NewAchievementNotification()

	achievementDetails := character.AchievementDetails{
		Name:        "Auto Hide Test",
		Description: "This should auto-hide",
		Timestamp:   time.Now(),
	}

	// Show achievement - this should auto-hide after 4 seconds
	notification.ShowAchievement(achievementDetails)

	if !notification.IsVisible() {
		t.Error("Notification should be visible immediately after ShowAchievement")
	}

	// Note: We don't wait 4 seconds in the test as that would slow down testing
	// The auto-hide functionality is verified by checking that the timer is set
	if notification.fadeTimer == nil {
		t.Error("Fade timer should be set after ShowAchievement")
	}
}

func TestAchievementNotificationHideCallback(t *testing.T) {
	notification := NewAchievementNotification()

	callbackCalled := false
	notification.SetHideCallback(func() {
		callbackCalled = true
	})

	notification.Show()
	notification.Hide()

	if !callbackCalled {
		t.Error("Hide callback should be called when notification is hidden")
	}
}

func TestAchievementNotificationFormatRewardText(t *testing.T) {
	notification := NewAchievementNotification()

	tests := []struct {
		name     string
		reward   *character.AchievementReward
		expected bool // true if we expect non-empty result
	}{
		{
			name:     "nil reward",
			reward:   nil,
			expected: false,
		},
		{
			name:     "empty reward",
			reward:   &character.AchievementReward{},
			expected: false,
		},
		{
			name: "stat boosts only",
			reward: &character.AchievementReward{
				StatBoosts: map[string]float64{
					"health":    10.0,
					"happiness": 5.0,
				},
			},
			expected: true,
		},
		{
			name: "animations only",
			reward: &character.AchievementReward{
				Animations: map[string]string{
					"special": "special_animation.gif",
				},
			},
			expected: true,
		},
		{
			name: "size change only",
			reward: &character.AchievementReward{
				Size: 150,
			},
			expected: true,
		},
		{
			name: "multiple rewards",
			reward: &character.AchievementReward{
				StatBoosts: map[string]float64{"health": 5.0},
				Animations: map[string]string{"dance": "dance.gif"},
				Size:       200,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := notification.formatRewardText(tt.reward)
			hasContent := len(result) > 0

			if hasContent != tt.expected {
				t.Errorf("formatRewardText() expected content=%v, got content=%v (result: %q)",
					tt.expected, hasContent, result)
			}
		})
	}
}

func TestAchievementNotificationRenderer(t *testing.T) {
	notification := NewAchievementNotification()
	renderer := notification.CreateRenderer()

	if renderer == nil {
		t.Fatal("CreateRenderer should return non-nil renderer")
	}

	// Test renderer when notification is hidden
	minSize := renderer.MinSize()
	if minSize.Width != 0 || minSize.Height != 0 {
		t.Error("Hidden notification should have zero minimum size")
	}

	objects := renderer.Objects()
	if len(objects) != 0 {
		t.Error("Hidden notification should have no rendered objects")
	}

	// Test renderer when notification is visible
	notification.Show()
	minSize = renderer.MinSize()
	if minSize.Width == 0 || minSize.Height == 0 {
		t.Error("Visible notification should have non-zero minimum size")
	}

	objects = renderer.Objects()
	if len(objects) == 0 {
		t.Error("Visible notification should have rendered objects")
	}

	// Test layout
	renderer.Layout(minSize)

	// Test refresh
	renderer.Refresh() // Should not panic

	// Test destroy
	renderer.Destroy() // Should not panic and should stop timer
}

func TestAchievementNotificationIntegration(t *testing.T) {
	// Test showing multiple achievements
	notification := NewAchievementNotification()

	achievement1 := character.AchievementDetails{
		Name:        "First Achievement",
		Description: "First test achievement",
		Timestamp:   time.Now(),
	}

	achievement2 := character.AchievementDetails{
		Name:        "Second Achievement",
		Description: "Second test achievement",
		Timestamp:   time.Now(),
		Reward: &character.AchievementReward{
			StatBoosts: map[string]float64{"happiness": 10.0},
		},
	}

	// Show first achievement
	notification.ShowAchievement(achievement1)
	if !notification.IsVisible() {
		t.Error("First achievement should be visible")
	}

	// Show second achievement (should replace first)
	notification.ShowAchievement(achievement2)
	if !notification.IsVisible() {
		t.Error("Second achievement should be visible")
	}

	// Verify that the timer gets reset when showing a new achievement
	if notification.fadeTimer == nil {
		t.Error("Timer should be active after showing achievement")
	}
}

func TestAchievementNotificationWithVariousRewards(t *testing.T) {
	notification := NewAchievementNotification()

	// Test with complex reward structure
	complexReward := &character.AchievementReward{
		StatBoosts: map[string]float64{
			"health":    15.5,
			"happiness": 20.0,
			"energy":    -5.0, // Negative values should also work
		},
		Animations: map[string]string{
			"celebration": "celebrate.gif",
			"special":     "special_move.gif",
		},
		Size: 256,
	}

	achievement := character.AchievementDetails{
		Name:        "Master Achievement",
		Description: "You've mastered the game!",
		Timestamp:   time.Now(),
		Reward:      complexReward,
	}

	notification.ShowAchievement(achievement)

	if !notification.IsVisible() {
		t.Error("Complex achievement should be visible")
	}

	// Verify that the reward formatting doesn't crash with complex data
	rewardText := notification.formatRewardText(complexReward)
	if len(rewardText) == 0 {
		t.Error("Complex reward should generate non-empty text")
	}
}
