// Package battle - Tournament and ranking system for competitive play
//
// This file implements a basic tournament system with brackets, ranking,
// and competitive match management. Tournaments support multiple formats
// and maintain player statistics for ranking calculations.
//
// Design principles:
// - Standard library only implementation
// - Fair tournament bracket generation
// - ELO-based ranking system for skill assessment
// - Comprehensive match history and statistics
package battle

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"time"
)

// Tournament system errors
var (
	ErrTournamentFull         = errors.New("tournament is full")
	ErrTournamentNotStarted   = errors.New("tournament has not started")
	ErrTournamentInProgress   = errors.New("tournament is in progress")
	ErrInvalidTournamentSize  = errors.New("invalid tournament size")
	ErrPlayerNotFound         = errors.New("player not found")
	ErrMatchNotFound          = errors.New("match not found")
	ErrInvalidMatchResult     = errors.New("invalid match result")
)

// TournamentFormat defines different tournament structures
type TournamentFormat string

const (
	FORMAT_SINGLE_ELIMINATION TournamentFormat = "single_elimination"
	FORMAT_DOUBLE_ELIMINATION TournamentFormat = "double_elimination"
	FORMAT_ROUND_ROBIN        TournamentFormat = "round_robin"
	FORMAT_SWISS_SYSTEM       TournamentFormat = "swiss_system"
)

// TournamentStatus represents the current state of a tournament
type TournamentStatus string

const (
	STATUS_REGISTRATION TournamentStatus = "registration"
	STATUS_IN_PROGRESS  TournamentStatus = "in_progress"
	STATUS_COMPLETED    TournamentStatus = "completed"
	STATUS_CANCELLED    TournamentStatus = "cancelled"
)

// MatchResult represents the outcome of a tournament match
type MatchResult string

const (
	RESULT_PLAYER1_WIN MatchResult = "player1_win"
	RESULT_PLAYER2_WIN MatchResult = "player2_win"
	RESULT_DRAW        MatchResult = "draw"
	RESULT_FORFEIT     MatchResult = "forfeit"
)

// TournamentPlayer represents a participant in tournaments
type TournamentPlayer struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Rating       int       `json:"rating"`       // ELO rating
	MatchesWon   int       `json:"matchesWon"`
	MatchesLost  int       `json:"matchesLost"`
	MatchesDrawn int       `json:"matchesDrawn"`
	Tournaments  int       `json:"tournaments"`  // Tournaments participated
	Victories    int       `json:"victories"`    // Tournament wins
	LastActive   time.Time `json:"lastActive"`
	Personality  string    `json:"personality"`  // AI personality for bot players
}

// TournamentMatch represents a single match in a tournament
type TournamentMatch struct {
	ID          string       `json:"id"`
	Player1ID   string       `json:"player1Id"`
	Player2ID   string       `json:"player2Id"`
	Result      MatchResult  `json:"result"`
	WinnerID    string       `json:"winnerId"`
	Round       int          `json:"round"`
	BracketPos  int          `json:"bracketPos"`
	ScheduledAt time.Time    `json:"scheduledAt"`
	CompletedAt time.Time    `json:"completedAt"`
	BattleLog   *BattleState `json:"battleLog,omitempty"` // Optional detailed battle data
}

// Tournament represents a competitive tournament
type Tournament struct {
	ID              string             `json:"id"`
	Name            string             `json:"name"`
	Format          TournamentFormat   `json:"format"`
	Status          TournamentStatus   `json:"status"`
	MaxPlayers      int                `json:"maxPlayers"`
	Players         []*TournamentPlayer `json:"players"`
	Matches         []*TournamentMatch  `json:"matches"`
	CurrentRound    int                `json:"currentRound"`
	WinnerID        string             `json:"winnerId"`
	CreatedAt       time.Time          `json:"createdAt"`
	StartedAt       time.Time          `json:"startedAt"`
	CompletedAt     time.Time          `json:"completedAt"`
	PrizeStructure  map[int]string     `json:"prizeStructure"` // Position -> Prize description
}

// TournamentManager handles tournament creation and management
type TournamentManager struct {
	tournaments map[string]*Tournament     `json:"tournaments"`
	players     map[string]*TournamentPlayer `json:"players"`
	nextMatchID int                        `json:"nextMatchId"`
	nextTournID int                        `json:"nextTournId"`
}

// ELO rating constants for fair skill assessment
const (
	INITIAL_RATING     = 1200 // Starting ELO rating for new players
	K_FACTOR          = 32   // Rating change factor
	MIN_RATING        = 100  // Minimum possible rating
	MAX_RATING        = 3000 // Maximum possible rating
	RATING_DIFFERENCE = 400  // Rating difference for 90% win probability
)

// NewTournamentManager creates a new tournament management system
func NewTournamentManager() *TournamentManager {
	return &TournamentManager{
		tournaments: make(map[string]*Tournament),
		players:     make(map[string]*TournamentPlayer),
		nextMatchID: 1,
		nextTournID: 1,
	}
}

// RegisterPlayer adds a new player to the tournament system
func (tm *TournamentManager) RegisterPlayer(id, name, personality string) *TournamentPlayer {
	if player, exists := tm.players[id]; exists {
		// Update existing player
		player.Name = name
		player.Personality = personality
		player.LastActive = time.Now()
		return player
	}

	// Create new player
	player := &TournamentPlayer{
		ID:          id,
		Name:        name,
		Rating:      INITIAL_RATING,
		Personality: personality,
		LastActive:  time.Now(),
	}

	tm.players[id] = player
	return player
}

// CreateTournament creates a new tournament with the specified format
func (tm *TournamentManager) CreateTournament(name string, format TournamentFormat, maxPlayers int) (*Tournament, error) {
	// Validate tournament size
	if !tm.isValidTournamentSize(format, maxPlayers) {
		return nil, ErrInvalidTournamentSize
	}

	tournament := &Tournament{
		ID:             fmt.Sprintf("tournament_%d", tm.nextTournID),
		Name:           name,
		Format:         format,
		Status:         STATUS_REGISTRATION,
		MaxPlayers:     maxPlayers,
		Players:        make([]*TournamentPlayer, 0, maxPlayers),
		Matches:        make([]*TournamentMatch, 0),
		CurrentRound:   0,
		CreatedAt:      time.Now(),
		PrizeStructure: tm.getDefaultPrizeStructure(maxPlayers),
	}

	tm.nextTournID++
	tm.tournaments[tournament.ID] = tournament
	return tournament, nil
}

// isValidTournamentSize checks if the tournament size is valid for the format
func (tm *TournamentManager) isValidTournamentSize(format TournamentFormat, maxPlayers int) bool {
	if maxPlayers < 2 || maxPlayers > 64 {
		return false
	}

	switch format {
	case FORMAT_SINGLE_ELIMINATION, FORMAT_DOUBLE_ELIMINATION:
		// Must be power of 2 for elimination formats
		return maxPlayers > 0 && (maxPlayers&(maxPlayers-1)) == 0
	case FORMAT_ROUND_ROBIN:
		// Any size is valid for round robin
		return maxPlayers <= 16 // Reasonable limit for round robin
	case FORMAT_SWISS_SYSTEM:
		// Any size is valid for Swiss system
		return maxPlayers >= 4
	default:
		return false
	}
}

// getDefaultPrizeStructure returns a default prize structure based on tournament size
func (tm *TournamentManager) getDefaultPrizeStructure(maxPlayers int) map[int]string {
	prizes := make(map[int]string)
	
	if maxPlayers >= 4 {
		prizes[1] = "Tournament Champion"
		prizes[2] = "Runner-up"
	}
	
	if maxPlayers >= 8 {
		prizes[3] = "Semi-finalist"
		prizes[4] = "Semi-finalist"
	}
	
	if maxPlayers >= 16 {
		prizes[5] = "Quarter-finalist"
		prizes[6] = "Quarter-finalist" 
		prizes[7] = "Quarter-finalist"
		prizes[8] = "Quarter-finalist"
	}
	
	return prizes
}

// JoinTournament adds a player to a tournament
func (tm *TournamentManager) JoinTournament(tournamentID, playerID string) error {
	tournament, exists := tm.tournaments[tournamentID]
	if !exists {
		return errors.New("tournament not found")
	}

	if tournament.Status != STATUS_REGISTRATION {
		return errors.New("tournament registration is closed")
	}

	if len(tournament.Players) >= tournament.MaxPlayers {
		return ErrTournamentFull
	}

	player, exists := tm.players[playerID]
	if !exists {
		return ErrPlayerNotFound
	}

	// Check if player already joined
	for _, p := range tournament.Players {
		if p.ID == playerID {
			return errors.New("player already joined")
		}
	}

	tournament.Players = append(tournament.Players, player)
	player.Tournaments++
	return nil
}

// StartTournament begins a tournament and generates the initial bracket
func (tm *TournamentManager) StartTournament(tournamentID string) error {
	tournament, exists := tm.tournaments[tournamentID]
	if !exists {
		return errors.New("tournament not found")
	}

	if tournament.Status != STATUS_REGISTRATION {
		return errors.New("tournament is not in registration phase")
	}

	if len(tournament.Players) < 2 {
		return errors.New("insufficient players to start tournament")
	}

	// Generate bracket based on format
	var err error
	switch tournament.Format {
	case FORMAT_SINGLE_ELIMINATION:
		err = tm.generateSingleEliminationBracket(tournament)
	case FORMAT_ROUND_ROBIN:
		err = tm.generateRoundRobinMatches(tournament)
	case FORMAT_SWISS_SYSTEM:
		err = tm.generateSwissRound(tournament)
	default:
		return fmt.Errorf("unsupported tournament format: %v", tournament.Format)
	}

	if err != nil {
		return fmt.Errorf("failed to generate bracket: %w", err)
	}

	tournament.Status = STATUS_IN_PROGRESS
	tournament.StartedAt = time.Now()
	tournament.CurrentRound = 1

	return nil
}

// generateSingleEliminationBracket creates matches for single elimination format
func (tm *TournamentManager) generateSingleEliminationBracket(tournament *Tournament) error {
	players := make([]*TournamentPlayer, len(tournament.Players))
	copy(players, tournament.Players)

	// Seed players by rating (highest rated gets best position)
	sort.Slice(players, func(i, j int) bool {
		return players[i].Rating > players[j].Rating
	})

	// Pair players for first round
	round := 1
	position := 0

	for i := 0; i < len(players); i += 2 {
		if i+1 < len(players) {
			match := &TournamentMatch{
				ID:         fmt.Sprintf("match_%d", tm.nextMatchID),
				Player1ID:  players[i].ID,
				Player2ID:  players[i+1].ID,
				Round:      round,
				BracketPos: position,
			}
			tm.nextMatchID++
			position++
			tournament.Matches = append(tournament.Matches, match)
		}
	}

	return nil
}

// generateRoundRobinMatches creates all matches for round robin format
func (tm *TournamentManager) generateRoundRobinMatches(tournament *Tournament) error {
	players := tournament.Players
	
	round := 1
	position := 0

	// Generate all possible pairings
	for i := 0; i < len(players); i++ {
		for j := i + 1; j < len(players); j++ {
			match := &TournamentMatch{
				ID:         fmt.Sprintf("match_%d", tm.nextMatchID),
				Player1ID:  players[i].ID,
				Player2ID:  players[j].ID,
				Round:      round,
				BracketPos: position,
			}
			tm.nextMatchID++
			position++
			tournament.Matches = append(tournament.Matches, match)
		}
	}

	return nil
}

// generateSwissRound creates matches for the current Swiss system round
func (tm *TournamentManager) generateSwissRound(tournament *Tournament) error {
	// For now, implement simple Swiss pairing (pair players with similar records)
	// This is a simplified version - full Swiss system would be more complex
	
	players := make([]*TournamentPlayer, len(tournament.Players))
	copy(players, tournament.Players)

	// Sort by tournament performance (wins first, then rating)
	sort.Slice(players, func(i, j int) bool {
		iWins := tm.getTournamentWins(tournament.ID, players[i].ID)
		jWins := tm.getTournamentWins(tournament.ID, players[j].ID)
		
		if iWins == jWins {
			return players[i].Rating > players[j].Rating
		}
		return iWins > jWins
	})

	// Pair adjacent players
	position := 0
	for i := 0; i < len(players); i += 2 {
		if i+1 < len(players) {
			match := &TournamentMatch{
				ID:         fmt.Sprintf("match_%d", tm.nextMatchID),
				Player1ID:  players[i].ID,
				Player2ID:  players[i+1].ID,
				Round:      tournament.CurrentRound,
				BracketPos: position,
			}
			tm.nextMatchID++
			position++
			tournament.Matches = append(tournament.Matches, match)
		}
	}

	return nil
}

// getTournamentWins counts wins for a player in a specific tournament
func (tm *TournamentManager) getTournamentWins(tournamentID, playerID string) int {
	tournament := tm.tournaments[tournamentID]
	if tournament == nil {
		return 0
	}

	wins := 0
	for _, match := range tournament.Matches {
		if match.WinnerID == playerID {
			wins++
		}
	}
	return wins
}

// ReportMatchResult records the result of a tournament match
func (tm *TournamentManager) ReportMatchResult(tournamentID, matchID string, result MatchResult, winnerID string) error {
	tournament, exists := tm.tournaments[tournamentID]
	if !exists {
		return errors.New("tournament not found")
	}

	if tournament.Status != STATUS_IN_PROGRESS {
		return ErrTournamentNotStarted
	}

	// Find the match
	var match *TournamentMatch
	for _, m := range tournament.Matches {
		if m.ID == matchID {
			match = m
			break
		}
	}

	if match == nil {
		return ErrMatchNotFound
	}

	// Validate result
	if result == RESULT_PLAYER1_WIN && winnerID != match.Player1ID {
		return ErrInvalidMatchResult
	}
	if result == RESULT_PLAYER2_WIN && winnerID != match.Player2ID {
		return ErrInvalidMatchResult
	}

	// Record match result
	match.Result = result
	match.WinnerID = winnerID
	match.CompletedAt = time.Now()

	// Update player statistics
	player1 := tm.players[match.Player1ID]
	player2 := tm.players[match.Player2ID]

	if player1 != nil && player2 != nil {
		switch result {
		case RESULT_PLAYER1_WIN:
			player1.MatchesWon++
			player2.MatchesLost++
			tm.updateELORating(player1, player2, 1.0) // Player 1 wins
		case RESULT_PLAYER2_WIN:
			player1.MatchesLost++
			player2.MatchesWon++
			tm.updateELORating(player1, player2, 0.0) // Player 2 wins
		case RESULT_DRAW:
			player1.MatchesDrawn++
			player2.MatchesDrawn++
			tm.updateELORating(player1, player2, 0.5) // Draw
		}
	}

	// Check if tournament round/tournament is complete
	tm.checkTournamentProgress(tournament)

	return nil
}

// updateELORating updates player ratings based on match result
func (tm *TournamentManager) updateELORating(player1, player2 *TournamentPlayer, score float64) {
	// Calculate expected scores
	expectedScore1 := tm.calculateExpectedScore(player1.Rating, player2.Rating)
	expectedScore2 := 1.0 - expectedScore1

	// Calculate new ratings
	newRating1 := player1.Rating + int(K_FACTOR*(score-expectedScore1))
	newRating2 := player2.Rating + int(K_FACTOR*((1.0-score)-expectedScore2))

	// Apply rating bounds
	player1.Rating = tm.clampRating(newRating1)
	player2.Rating = tm.clampRating(newRating2)
}

// calculateExpectedScore calculates expected score for ELO rating
func (tm *TournamentManager) calculateExpectedScore(rating1, rating2 int) float64 {
	return 1.0 / (1.0 + math.Pow(10.0, float64(rating2-rating1)/RATING_DIFFERENCE))
}

// clampRating ensures rating stays within valid bounds
func (tm *TournamentManager) clampRating(rating int) int {
	if rating < MIN_RATING {
		return MIN_RATING
	}
	if rating > MAX_RATING {
		return MAX_RATING
	}
	return rating
}

// checkTournamentProgress checks if tournament round or tournament is complete
func (tm *TournamentManager) checkTournamentProgress(tournament *Tournament) {
	switch tournament.Format {
	case FORMAT_SINGLE_ELIMINATION:
		tm.checkSingleEliminationProgress(tournament)
	case FORMAT_ROUND_ROBIN:
		tm.checkRoundRobinProgress(tournament)
	case FORMAT_SWISS_SYSTEM:
		tm.checkSwissProgress(tournament)
	}
}

// checkSingleEliminationProgress checks if elimination tournament needs to advance
func (tm *TournamentManager) checkSingleEliminationProgress(tournament *Tournament) {
	// Check if current round is complete
	currentRoundMatches := 0
	completedMatches := 0

	for _, match := range tournament.Matches {
		if match.Round == tournament.CurrentRound {
			currentRoundMatches++
			if !match.CompletedAt.IsZero() {
				completedMatches++
			}
		}
	}

	if currentRoundMatches > 0 && completedMatches == currentRoundMatches {
		// Current round is complete
		winners := tm.getRoundWinners(tournament, tournament.CurrentRound)
		
		if len(winners) == 1 {
			// Tournament is complete
			tournament.Status = STATUS_COMPLETED
			tournament.WinnerID = winners[0]
			tournament.CompletedAt = time.Now()
			
			// Update winner statistics
			if winner := tm.players[winners[0]]; winner != nil {
				winner.Victories++
			}
		} else if len(winners) > 1 {
			// Generate next round
			tm.generateNextEliminationRound(tournament, winners)
			tournament.CurrentRound++
		}
	}
}

// getRoundWinners gets all winners from a specific round
func (tm *TournamentManager) getRoundWinners(tournament *Tournament, round int) []string {
	var winners []string
	
	for _, match := range tournament.Matches {
		if match.Round == round && match.WinnerID != "" {
			winners = append(winners, match.WinnerID)
		}
	}
	
	return winners
}

// generateNextEliminationRound creates matches for the next elimination round
func (tm *TournamentManager) generateNextEliminationRound(tournament *Tournament, winners []string) {
	nextRound := tournament.CurrentRound + 1
	position := 0

	for i := 0; i < len(winners); i += 2 {
		if i+1 < len(winners) {
			match := &TournamentMatch{
				ID:         fmt.Sprintf("match_%d", tm.nextMatchID),
				Player1ID:  winners[i],
				Player2ID:  winners[i+1],
				Round:      nextRound,
				BracketPos: position,
			}
			tm.nextMatchID++
			position++
			tournament.Matches = append(tournament.Matches, match)
		}
	}
}

// checkRoundRobinProgress checks if round robin tournament is complete
func (tm *TournamentManager) checkRoundRobinProgress(tournament *Tournament) {
	// Check if all matches are complete
	totalMatches := len(tournament.Matches)
	completedMatches := 0

	for _, match := range tournament.Matches {
		if !match.CompletedAt.IsZero() {
			completedMatches++
		}
	}

	if totalMatches > 0 && completedMatches == totalMatches {
		// Tournament is complete - determine winner by points
		tournament.Status = STATUS_COMPLETED
		tournament.WinnerID = tm.getRoundRobinWinner(tournament)
		tournament.CompletedAt = time.Now()
		
		// Update winner statistics
		if winner := tm.players[tournament.WinnerID]; winner != nil {
			winner.Victories++
		}
	}
}

// getRoundRobinWinner determines the winner of a round robin tournament
func (tm *TournamentManager) getRoundRobinWinner(tournament *Tournament) string {
	// Calculate points for each player (3 for win, 1 for draw, 0 for loss)
	points := make(map[string]int)
	
	for _, player := range tournament.Players {
		points[player.ID] = 0
	}

	for _, match := range tournament.Matches {
		switch match.Result {
		case RESULT_PLAYER1_WIN:
			points[match.Player1ID] += 3
		case RESULT_PLAYER2_WIN:
			points[match.Player2ID] += 3
		case RESULT_DRAW:
			points[match.Player1ID] += 1
			points[match.Player2ID] += 1
		}
	}

	// Find player with most points
	var winnerID string
	maxPoints := -1
	
	for playerID, playerPoints := range points {
		if playerPoints > maxPoints {
			maxPoints = playerPoints
			winnerID = playerID
		}
	}

	return winnerID
}

// checkSwissProgress checks if Swiss tournament needs another round or is complete
func (tm *TournamentManager) checkSwissProgress(tournament *Tournament) {
	// Simple Swiss implementation - run 5 rounds or until clear winner
	const maxSwissRounds = 5

	// Check if current round is complete
	currentRoundMatches := 0
	completedMatches := 0

	for _, match := range tournament.Matches {
		if match.Round == tournament.CurrentRound {
			currentRoundMatches++
			if !match.CompletedAt.IsZero() {
				completedMatches++
			}
		}
	}

	if currentRoundMatches > 0 && completedMatches == currentRoundMatches {
		if tournament.CurrentRound >= maxSwissRounds {
			// Tournament is complete
			tournament.Status = STATUS_COMPLETED
			tournament.WinnerID = tm.getSwissWinner(tournament)
			tournament.CompletedAt = time.Now()
			
			// Update winner statistics
			if winner := tm.players[tournament.WinnerID]; winner != nil {
				winner.Victories++
			}
		} else {
			// Generate next round
			tournament.CurrentRound++
			tm.generateSwissRound(tournament)
		}
	}
}

// getSwissWinner determines the winner of a Swiss tournament
func (tm *TournamentManager) getSwissWinner(tournament *Tournament) string {
	// Winner is player with most wins, then highest rating as tiebreaker
	winCounts := make(map[string]int)
	
	for _, player := range tournament.Players {
		winCounts[player.ID] = tm.getTournamentWins(tournament.ID, player.ID)
	}

	// Find player with most wins
	var winnerID string
	maxWins := -1
	maxRating := -1
	
	for _, player := range tournament.Players {
		wins := winCounts[player.ID]
		if wins > maxWins || (wins == maxWins && player.Rating > maxRating) {
			maxWins = wins
			maxRating = player.Rating
			winnerID = player.ID
		}
	}

	return winnerID
}

// GetTournament retrieves a tournament by ID
func (tm *TournamentManager) GetTournament(tournamentID string) (*Tournament, error) {
	tournament, exists := tm.tournaments[tournamentID]
	if !exists {
		return nil, errors.New("tournament not found")
	}
	return tournament, nil
}

// GetPlayer retrieves a player by ID
func (tm *TournamentManager) GetPlayer(playerID string) (*TournamentPlayer, error) {
	player, exists := tm.players[playerID]
	if !exists {
		return nil, ErrPlayerNotFound
	}
	return player, nil
}

// GetLeaderboard returns top players sorted by rating
func (tm *TournamentManager) GetLeaderboard(limit int) []*TournamentPlayer {
	players := make([]*TournamentPlayer, 0, len(tm.players))
	
	for _, player := range tm.players {
		players = append(players, player)
	}

	// Sort by rating (highest first)
	sort.Slice(players, func(i, j int) bool {
		return players[i].Rating > players[j].Rating
	})

	if limit > 0 && limit < len(players) {
		return players[:limit]
	}
	
	return players
}

// GetActiveTournaments returns all tournaments that are accepting registrations or in progress
func (tm *TournamentManager) GetActiveTournaments() []*Tournament {
	var active []*Tournament
	
	for _, tournament := range tm.tournaments {
		if tournament.Status == STATUS_REGISTRATION || tournament.Status == STATUS_IN_PROGRESS {
			active = append(active, tournament)
		}
	}
	
	return active
}

// CalculateWinRate calculates a player's win rate percentage
func (tm *TournamentManager) CalculateWinRate(playerID string) float64 {
	player, exists := tm.players[playerID]
	if !exists {
		return 0.0
	}

	totalMatches := player.MatchesWon + player.MatchesLost + player.MatchesDrawn
	if totalMatches == 0 {
		return 0.0
	}

	return float64(player.MatchesWon) / float64(totalMatches) * 100.0
}