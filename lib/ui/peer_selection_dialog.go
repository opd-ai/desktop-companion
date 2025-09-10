package ui

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/opd-ai/desktop-companion/lib/network"
)

// PeerSelectionDialog provides a simple dialog for selecting peers for battle
// Follows existing UI patterns for consistency and minimal code changes
type PeerSelectionDialog struct {
	widget.BaseWidget
	content      *fyne.Container
	titleLabel   *widget.Label
	peerList     *widget.List
	selectButton *widget.Button
	cancelButton *widget.Button
	visible      bool
	onSelection  func(peer network.Peer)
	onCancel     func()
	peers        []network.Peer
	selectedPeer *network.Peer
	mu           sync.Mutex
}

// NewPeerSelectionDialog creates a new peer selection dialog
func NewPeerSelectionDialog() *PeerSelectionDialog {
	dialog := &PeerSelectionDialog{
		visible: false,
		peers:   []network.Peer{},
	}

	dialog.initializeComponents()
	dialog.setupLayout()

	return dialog
}

// initializeComponents creates the dialog UI components
func (psd *PeerSelectionDialog) initializeComponents() {
	psd.titleLabel = widget.NewLabel("Select Peer for Battle")

	// Create peer list widget
	psd.peerList = widget.NewList(
		func() int {
			return len(psd.peers)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Peer Name")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < len(psd.peers) {
				label := item.(*widget.Label)
				peer := psd.peers[id]
				label.SetText(peer.ID + " (" + peer.AddrStr + ")")
			}
		},
	)

	// Set selection callback
	psd.peerList.OnSelected = func(id widget.ListItemID) {
		if id < len(psd.peers) {
			psd.mu.Lock()
			psd.selectedPeer = &psd.peers[id]
			psd.mu.Unlock()
			psd.updateSelectButtonState()
		}
	}

	psd.selectButton = widget.NewButton("Select", func() {
		psd.selectPeer()
	})
	psd.selectButton.Disable() // Start disabled

	psd.cancelButton = widget.NewButton("Cancel", func() {
		psd.cancel()
	})
}

// setupLayout creates the dialog layout
func (psd *PeerSelectionDialog) setupLayout() {
	buttonContainer := container.NewHBox(
		psd.selectButton,
		psd.cancelButton,
	)

	psd.content = container.NewVBox(
		psd.titleLabel,
		psd.peerList,
		buttonContainer,
	)

	psd.content.Hide() // Start hidden
}

// Show displays the peer selection dialog with the given peers
func (psd *PeerSelectionDialog) Show(peers []network.Peer, onSelection func(peer network.Peer), onCancel func()) {
	psd.mu.Lock()
	defer psd.mu.Unlock()

	psd.peers = peers
	psd.onSelection = onSelection
	psd.onCancel = onCancel
	psd.selectedPeer = nil

	psd.peerList.UnselectAll()
	psd.peerList.Refresh()
	psd.updateSelectButtonState()

	psd.visible = true
	psd.content.Show()
	psd.Refresh()
}

// Hide hides the peer selection dialog
func (psd *PeerSelectionDialog) Hide() {
	psd.mu.Lock()
	defer psd.mu.Unlock()

	psd.visible = false
	psd.content.Hide()
	psd.Refresh()
}

// selectPeer handles peer selection
func (psd *PeerSelectionDialog) selectPeer() {
	psd.mu.Lock()
	selectedPeer := psd.selectedPeer
	callback := psd.onSelection
	psd.mu.Unlock()

	psd.Hide()

	if selectedPeer != nil && callback != nil {
		callback(*selectedPeer)
	}
}

// cancel handles dialog cancellation
func (psd *PeerSelectionDialog) cancel() {
	psd.mu.Lock()
	callback := psd.onCancel
	psd.mu.Unlock()

	psd.Hide()

	if callback != nil {
		callback()
	}
}

// updateSelectButtonState enables/disables the select button based on selection
func (psd *PeerSelectionDialog) updateSelectButtonState() {
	if psd.selectedPeer != nil {
		psd.selectButton.Enable()
	} else {
		psd.selectButton.Disable()
	}
}

// IsVisible returns whether the dialog is currently visible
func (psd *PeerSelectionDialog) IsVisible() bool {
	psd.mu.Lock()
	defer psd.mu.Unlock()
	return psd.visible
}

// GetContainer returns the container for embedding in the window
func (psd *PeerSelectionDialog) GetContainer() *fyne.Container {
	return psd.content
}

// CreateRenderer creates the Fyne renderer for the peer selection dialog
func (psd *PeerSelectionDialog) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(psd.content)
}
