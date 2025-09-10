package ui

import (
	"github.com/opd-ai/desktop-companion/lib/character"
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// AchievementNotification displays floating achievement unlock notifications
// Follows the "lazy programmer" principle using existing Fyne widgets with simple animations
type AchievementNotification struct {
	widget.BaseWidget
	container    *fyne.Container
	background   *canvas.Rectangle
	titleLabel   *widget.RichText
	descLabel    *widget.RichText
	visible      bool
	fadeTimer    *time.Timer
	hideCallback func()
}

// NewAchievementNotification creates a new achievement notification widget
func NewAchievementNotification() *AchievementNotification {
	an := &AchievementNotification{
		visible: false,
	}

	// Create golden background for achievement feel
	an.background = canvas.NewRectangle(color.RGBA{R: 255, G: 215, B: 0, A: 200}) // Gold with transparency
	an.background.StrokeColor = color.RGBA{R: 218, G: 165, B: 32, A: 255}         // Dark gold border
	an.background.StrokeWidth = 2

	// Create achievement title with larger, bold text
	an.titleLabel = widget.NewRichTextFromMarkdown("**🏆 Achievement Unlocked!**")
	an.titleLabel.Wrapping = fyne.TextWrapWord

	// Create description label
	an.descLabel = widget.NewRichTextFromMarkdown("*Achievement description*")
	an.descLabel.Wrapping = fyne.TextWrapWord

	// Create container with padding
	content := container.NewVBox(
		an.titleLabel,
		an.descLabel,
	)

	an.container = container.NewWithoutLayout(
		an.background,
		container.NewPadded(content),
	)

	an.ExtendBaseWidget(an)
	return an
}

// ShowAchievement displays an achievement notification with auto-hide
func (an *AchievementNotification) ShowAchievement(details character.AchievementDetails) {
	// Update content with achievement details
	an.titleLabel.ParseMarkdown(fmt.Sprintf("**🏆 %s**", details.Name))
	an.descLabel.ParseMarkdown(fmt.Sprintf("*%s*", details.Description))

	// Add reward information if available
	if details.Reward != nil {
		rewardText := an.formatRewardText(details.Reward)
		if rewardText != "" {
			an.descLabel.ParseMarkdown(fmt.Sprintf("*%s*\n\n%s", details.Description, rewardText))
		}
	}

	an.Show()

	// Auto-hide after 4 seconds
	if an.fadeTimer != nil {
		an.fadeTimer.Stop()
	}
	an.fadeTimer = time.AfterFunc(4*time.Second, func() {
		an.Hide()
	})
}

// formatRewardText creates a user-friendly description of achievement rewards
func (an *AchievementNotification) formatRewardText(reward *character.AchievementReward) string {
	if reward == nil {
		return ""
	}

	var parts []string

	// Describe stat boosts
	if len(reward.StatBoosts) > 0 {
		parts = append(parts, "**Stat Boosts:**")
		for stat, boost := range reward.StatBoosts {
			parts = append(parts, fmt.Sprintf("• %s +%.1f", stat, boost))
		}
	}

	// Describe animation unlocks
	if len(reward.Animations) > 0 {
		parts = append(parts, "**New Animations Unlocked:**")
		for name := range reward.Animations {
			parts = append(parts, fmt.Sprintf("• %s", name))
		}
	}

	// Describe size changes
	if reward.Size > 0 {
		parts = append(parts, fmt.Sprintf("**Size Change:** %d pixels", reward.Size))
	}

	if len(parts) == 0 {
		return ""
	}

	result := ""
	for i, part := range parts {
		if i > 0 {
			result += "\n"
		}
		result += part
	}
	return result
}

// Show makes the notification visible
func (an *AchievementNotification) Show() {
	an.visible = true
	an.Refresh()
}

// Hide makes the notification invisible
func (an *AchievementNotification) Hide() {
	an.visible = false
	an.Refresh()
	if an.hideCallback != nil {
		an.hideCallback()
	}
}

// IsVisible returns whether the notification is currently visible
func (an *AchievementNotification) IsVisible() bool {
	return an.visible
}

// SetHideCallback sets a callback to be called when the notification is hidden
func (an *AchievementNotification) SetHideCallback(callback func()) {
	an.hideCallback = callback
}

// CreateRenderer returns the renderer for this widget
func (an *AchievementNotification) CreateRenderer() fyne.WidgetRenderer {
	return &achievementNotificationRenderer{
		notification: an,
		container:    an.container,
	}
}

// achievementNotificationRenderer handles rendering the achievement notification
type achievementNotificationRenderer struct {
	notification *AchievementNotification
	container    *fyne.Container
}

// Layout arranges the notification components
func (r *achievementNotificationRenderer) Layout(size fyne.Size) {
	if !r.notification.visible {
		return
	}

	// Center the notification
	notificationSize := fyne.NewSize(300, 120)
	x := (size.Width - notificationSize.Width) / 2
	y := size.Height / 4 // Position in upper quarter of screen

	r.container.Resize(notificationSize)
	r.container.Move(fyne.NewPos(x, y))

	// Layout background to fill container
	r.notification.background.Resize(notificationSize)
	r.notification.background.Move(fyne.NewPos(0, 0))
}

// MinSize returns the minimum size for the notification
func (r *achievementNotificationRenderer) MinSize() fyne.Size {
	if !r.notification.visible {
		return fyne.NewSize(0, 0)
	}
	return fyne.NewSize(300, 120)
}

// Refresh updates the visual state
func (r *achievementNotificationRenderer) Refresh() {
	r.container.Refresh()
}

// Objects returns the rendered objects
func (r *achievementNotificationRenderer) Objects() []fyne.CanvasObject {
	if !r.notification.visible {
		return []fyne.CanvasObject{}
	}
	return []fyne.CanvasObject{r.container}
}

// Destroy cleans up resources
func (r *achievementNotificationRenderer) Destroy() {
	if r.notification.fadeTimer != nil {
		r.notification.fadeTimer.Stop()
	}
}
