module github.com/opd-ai/desktop-companion/cmd/easy-embedded

go 1.21

require (
	fyne.io/fyne/v2 v2.4.5
	github.com/opd-ai/desktop-companion v0.0.0-00010101000000-000000000000
	github.com/jdkato/prose/v2 v2.0.0
	github.com/mmcdole/gofeed v1.3.0
)

// Single replace directive - much simpler than the old approach
replace github.com/opd-ai/desktop-companion => /home/user/go/src/github.com/opd-ai/desktop-companion
