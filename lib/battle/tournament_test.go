package battle

import (
	"fmt"
	"testing"
)

func TestNewTournamentManager(t *testing.T) {
	tm := NewTournamentManager()

	if tm == nil {
		t.Fatal("NewTournamentManager returned nil")
	}

	if tm.tournaments == nil {
		t.Error("tournaments map not initialized")
	}

	if tm.players == nil {
		t.Error("players map not initialized")
	}

	if tm.nextMatchID != 1 {
		t.Errorf("Expected nextMatchID to be 1, got %d", tm.nextMatchID)
	}

	if tm.nextTournID != 1 {
		t.Errorf("Expected nextTournID to be 1, got %d", tm.nextTournID)
	}
}

func TestRegisterPlayer(t *testing.T) {
	tm := NewTournamentManager()

	player := tm.RegisterPlayer("player1", "Test Player", string(PERSONALITY_AGGRESSIVE))

	if player == nil {
		t.Fatal("RegisterPlayer returned nil")
	}

	if player.ID != "player1" {
		t.Errorf("Expected player ID to be 'player1', got %s", player.ID)
	}

	if player.Name != "Test Player" {
		t.Errorf("Expected player name to be 'Test Player', got %s", player.Name)
	}

	if player.Rating != INITIAL_RATING {
		t.Errorf("Expected initial rating to be %d, got %d", INITIAL_RATING, player.Rating)
	}

	if player.Personality != string(PERSONALITY_AGGRESSIVE) {
		t.Errorf("Expected personality to be %s, got %s", PERSONALITY_AGGRESSIVE, player.Personality)
	}

	// Test updating existing player
	updatedPlayer := tm.RegisterPlayer("player1", "Updated Player", string(PERSONALITY_DEFENSIVE))
	if updatedPlayer.Name != "Updated Player" {
		t.Error("Player name should be updated")
	}

	if updatedPlayer.Personality != string(PERSONALITY_DEFENSIVE) {
		t.Error("Player personality should be updated")
	}

	if updatedPlayer.Rating != INITIAL_RATING {
		t.Error("Player rating should remain unchanged when updating")
	}
}

func TestCreateTournament(t *testing.T) {
	tm := NewTournamentManager()

	tests := []struct {
		name        string
		format      TournamentFormat
		maxPlayers  int
		shouldError bool
	}{
		{"Valid Single Elimination", FORMAT_SINGLE_ELIMINATION, 8, false},
		{"Valid Round Robin", FORMAT_ROUND_ROBIN, 6, false},
		{"Valid Swiss", FORMAT_SWISS_SYSTEM, 12, false},
		{"Invalid Size - Too Small", FORMAT_SINGLE_ELIMINATION, 1, true},
		{"Invalid Size - Too Large", FORMAT_ROUND_ROBIN, 100, true},
		{"Invalid Size - Not Power of 2", FORMAT_SINGLE_ELIMINATION, 6, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tournament, err := tm.CreateTournament(tt.name, tt.format, tt.maxPlayers)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("CreateTournament failed: %v", err)
			}

			if tournament == nil {
				t.Fatal("Tournament is nil")
			}

			if tournament.Name != tt.name {
				t.Errorf("Expected tournament name to be %s, got %s", tt.name, tournament.Name)
			}

			if tournament.Format != tt.format {
				t.Errorf("Expected format to be %v, got %v", tt.format, tournament.Format)
			}

			if tournament.MaxPlayers != tt.maxPlayers {
				t.Errorf("Expected max players to be %d, got %d", tt.maxPlayers, tournament.MaxPlayers)
			}

			if tournament.Status != STATUS_REGISTRATION {
				t.Errorf("Expected status to be %v, got %v", STATUS_REGISTRATION, tournament.Status)
			}

			if len(tournament.PrizeStructure) == 0 {
				t.Error("Tournament should have a prize structure")
			}
		})
	}
}

func TestTournament_SingleElimination(t *testing.T) {
	tm := NewTournamentManager()

	// Create tournament
	tournament, err := tm.CreateTournament("test", FORMAT_SINGLE_ELIMINATION, 4)
	if err != nil {
		t.Fatalf("Failed to create tournament: %v", err)
	}

	// Register players
	player1 := tm.RegisterPlayer("p1", "Player 1", string(PERSONALITY_BALANCED))
	player2 := tm.RegisterPlayer("p2", "Player 2", string(PERSONALITY_BALANCED))
	player3 := tm.RegisterPlayer("p3", "Player 3", string(PERSONALITY_BALANCED))
	player4 := tm.RegisterPlayer("p4", "Player 4", string(PERSONALITY_BALANCED))

	// Join tournament
	err = tm.JoinTournament(tournament.ID, player1.ID)
	if err != nil {
		t.Errorf("Failed to join tournament: %v", err)
	}

	err = tm.JoinTournament(tournament.ID, player2.ID)
	if err != nil {
		t.Errorf("Failed to join tournament: %v", err)
	}

	err = tm.JoinTournament(tournament.ID, player3.ID)
	if err != nil {
		t.Errorf("Failed to join tournament: %v", err)
	}

	err = tm.JoinTournament(tournament.ID, player4.ID)
	if err != nil {
		t.Errorf("Failed to join tournament: %v", err)
	}

	// Start tournament
	err = tm.StartTournament(tournament.ID)
	if err != nil {
		t.Fatalf("Failed to start tournament: %v", err)
	}

	// Check tournament status
	if tournament.Status != STATUS_IN_PROGRESS {
		t.Errorf("Expected tournament status to be in_progress, got %v", tournament.Status)
	}

	// Check that matches were created
	if len(tournament.Matches) == 0 {
		t.Error("Expected matches to be created")
	}
}

func TestStartTournament(t *testing.T) {
	tm := NewTournamentManager()

	tests := []struct {
		name        string
		format      TournamentFormat
		playerCount int
		shouldError bool
	}{
		{"Single Elimination 4 players", FORMAT_SINGLE_ELIMINATION, 4, false},
		{"Round Robin 4 players", FORMAT_ROUND_ROBIN, 4, false},
		{"Swiss 6 players", FORMAT_SWISS_SYSTEM, 6, false},
		{"Insufficient players", FORMAT_SINGLE_ELIMINATION, 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tournament, err := tm.CreateTournament(tt.name, tt.format, 8)
			if err != nil {
				t.Fatalf("CreateTournament failed: %v", err)
			}

			// Add players
			for i := 0; i < tt.playerCount; i++ {
				player := tm.RegisterPlayer(fmt.Sprintf("p%d", i), fmt.Sprintf("Player %d", i), string(PERSONALITY_BALANCED))
				tm.JoinTournament(tournament.ID, player.ID)
			}

			err = tm.StartTournament(tournament.ID)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("StartTournament failed: %v", err)
			}

			if tournament.Status != STATUS_IN_PROGRESS {
				t.Errorf("Expected status to be %v, got %v", STATUS_IN_PROGRESS, tournament.Status)
			}

			if tournament.CurrentRound != 1 {
				t.Errorf("Expected current round to be 1, got %d", tournament.CurrentRound)
			}

			if len(tournament.Matches) == 0 {
				t.Error("Tournament should have matches after starting")
			}

			// Verify match structure based on format
			switch tt.format {
			case FORMAT_SINGLE_ELIMINATION:
				expectedMatches := tt.playerCount / 2
				if len(tournament.Matches) != expectedMatches {
					t.Errorf("Expected %d matches for single elimination, got %d", expectedMatches, len(tournament.Matches))
				}
			case FORMAT_ROUND_ROBIN:
				expectedMatches := tt.playerCount * (tt.playerCount - 1) / 2
				if len(tournament.Matches) != expectedMatches {
					t.Errorf("Expected %d matches for round robin, got %d", expectedMatches, len(tournament.Matches))
				}
			case FORMAT_SWISS_SYSTEM:
				expectedMatches := tt.playerCount / 2
				if len(tournament.Matches) != expectedMatches {
					t.Errorf("Expected %d matches for first Swiss round, got %d", expectedMatches, len(tournament.Matches))
				}
			}
		})
	}
}

func TestReportMatchResult(t *testing.T) {
	tm := NewTournamentManager()

	// Create and start tournament
	tournament, err := tm.CreateTournament("Test Tournament", FORMAT_SINGLE_ELIMINATION, 4)
	if err != nil {
		t.Fatalf("CreateTournament failed: %v", err)
	}

	player1 := tm.RegisterPlayer("p1", "Player 1", string(PERSONALITY_BALANCED))
	player2 := tm.RegisterPlayer("p2", "Player 2", string(PERSONALITY_AGGRESSIVE))
	player3 := tm.RegisterPlayer("p3", "Player 3", string(PERSONALITY_DEFENSIVE))
	player4 := tm.RegisterPlayer("p4", "Player 4", string(PERSONALITY_TACTICAL))

	tm.JoinTournament(tournament.ID, player1.ID)
	tm.JoinTournament(tournament.ID, player2.ID)
	tm.JoinTournament(tournament.ID, player3.ID)
	tm.JoinTournament(tournament.ID, player4.ID)

	err = tm.StartTournament(tournament.ID)
	if err != nil {
		t.Fatalf("StartTournament failed: %v", err)
	}

	if len(tournament.Matches) != 2 {
		t.Fatalf("Expected 2 matches, got %d", len(tournament.Matches))
	}

	// Test valid match result
	match1 := tournament.Matches[0]
	initialRating1 := tm.players[match1.Player1ID].Rating
	initialRating2 := tm.players[match1.Player2ID].Rating

	err = tm.ReportMatchResult(tournament.ID, match1.ID, RESULT_PLAYER1_WIN, match1.Player1ID)
	if err != nil {
		t.Fatalf("ReportMatchResult failed: %v", err)
	}

	if match1.Result != RESULT_PLAYER1_WIN {
		t.Errorf("Expected match result to be %v, got %v", RESULT_PLAYER1_WIN, match1.Result)
	}

	if match1.WinnerID != match1.Player1ID {
		t.Errorf("Expected winner to be %s, got %s", match1.Player1ID, match1.WinnerID)
	}

	if match1.CompletedAt.IsZero() {
		t.Error("Match completion time should be set")
	}

	// Check that ratings were updated
	newRating1 := tm.players[match1.Player1ID].Rating
	newRating2 := tm.players[match1.Player2ID].Rating

	if newRating1 <= initialRating1 {
		t.Error("Winner's rating should increase")
	}

	if newRating2 >= initialRating2 {
		t.Error("Loser's rating should decrease")
	}

	// Check that match statistics were updated
	winner := tm.players[match1.Player1ID]
	loser := tm.players[match1.Player2ID]

	if winner.MatchesWon == 0 {
		t.Error("Winner should have match win recorded")
	}

	if loser.MatchesLost == 0 {
		t.Error("Loser should have match loss recorded")
	}

	// Test invalid result reporting
	err = tm.ReportMatchResult(tournament.ID, match1.ID, RESULT_PLAYER2_WIN, match1.Player1ID)
	if err != ErrInvalidMatchResult {
		t.Errorf("Expected ErrInvalidMatchResult, got %v", err)
	}
}

func TestTournamentCompletion(t *testing.T) {
	tm := NewTournamentManager()

	// Test single elimination completion
	tournament, err := tm.CreateTournament("Test Tournament", FORMAT_SINGLE_ELIMINATION, 4)
	if err != nil {
		t.Fatalf("CreateTournament failed: %v", err)
	}

	// Add players
	players := make([]*TournamentPlayer, 4)
	for i := 0; i < 4; i++ {
		players[i] = tm.RegisterPlayer(fmt.Sprintf("p%d", i), fmt.Sprintf("Player %d", i), string(PERSONALITY_BALANCED))
		tm.JoinTournament(tournament.ID, players[i].ID)
	}

	err = tm.StartTournament(tournament.ID)
	if err != nil {
		t.Fatalf("StartTournament failed: %v", err)
	}

	// Complete first round
	for _, match := range tournament.Matches {
		if match.Round == 1 {
			tm.ReportMatchResult(tournament.ID, match.ID, RESULT_PLAYER1_WIN, match.Player1ID)
		}
	}

	// Check that second round was created
	secondRoundMatches := 0
	for _, match := range tournament.Matches {
		if match.Round == 2 {
			secondRoundMatches++
		}
	}

	if secondRoundMatches != 1 {
		t.Errorf("Expected 1 second round match, got %d", secondRoundMatches)
	}

	if tournament.CurrentRound != 2 {
		t.Errorf("Expected current round to be 2, got %d", tournament.CurrentRound)
	}

	// Complete final match
	var finalMatch *TournamentMatch
	for _, match := range tournament.Matches {
		if match.Round == 2 {
			finalMatch = match
			break
		}
	}

	if finalMatch == nil {
		t.Fatal("Final match not found")
	}

	initialVictories := tm.players[finalMatch.Player1ID].Victories

	tm.ReportMatchResult(tournament.ID, finalMatch.ID, RESULT_PLAYER1_WIN, finalMatch.Player1ID)

	// Check tournament completion
	if tournament.Status != STATUS_COMPLETED {
		t.Errorf("Expected tournament status to be %v, got %v", STATUS_COMPLETED, tournament.Status)
	}

	if tournament.WinnerID != finalMatch.Player1ID {
		t.Errorf("Expected winner to be %s, got %s", finalMatch.Player1ID, tournament.WinnerID)
	}

	if tournament.CompletedAt.IsZero() {
		t.Error("Tournament completion time should be set")
	}

	// Check that winner's victory count was incremented
	finalVictories := tm.players[finalMatch.Player1ID].Victories
	if finalVictories != initialVictories+1 {
		t.Errorf("Expected winner victories to increase by 1, got %d -> %d", initialVictories, finalVictories)
	}
}

func TestELORatingUpdate(t *testing.T) {
	tm := NewTournamentManager()

	player1 := tm.RegisterPlayer("p1", "Player 1", string(PERSONALITY_BALANCED))
	player2 := tm.RegisterPlayer("p2", "Player 2", string(PERSONALITY_BALANCED))

	// Set specific ratings for predictable test
	player1.Rating = 1200
	player2.Rating = 1200

	// Test win/loss rating changes
	tm.updateELORating(player1, player2, 1.0) // Player 1 wins

	if player1.Rating <= 1200 {
		t.Error("Winner's rating should increase")
	}

	if player2.Rating >= 1200 {
		t.Error("Loser's rating should decrease")
	}

	// Reset and test draw
	player1.Rating = 1200
	player2.Rating = 1200

	tm.updateELORating(player1, player2, 0.5) // Draw

	// For equal ratings, draw should not change ratings significantly
	if abs(player1.Rating-1200) > 1 {
		t.Errorf("Player 1 rating changed too much in draw: %d", player1.Rating)
	}

	if abs(player2.Rating-1200) > 1 {
		t.Errorf("Player 2 rating changed too much in draw: %d", player2.Rating)
	}

	// Test rating bounds
	player1.Rating = MIN_RATING
	player2.Rating = MAX_RATING

	tm.updateELORating(player1, player2, 0.0) // Player 1 loses to much higher rated player

	if player1.Rating < MIN_RATING {
		t.Errorf("Rating should not go below minimum: %d", player1.Rating)
	}

	if player2.Rating > MAX_RATING {
		t.Errorf("Rating should not go above maximum: %d", player2.Rating)
	}
}

func TestGetLeaderboard(t *testing.T) {
	tm := NewTournamentManager()

	// Create players with different ratings
	players := []*TournamentPlayer{
		tm.RegisterPlayer("p1", "Player 1", string(PERSONALITY_BALANCED)),
		tm.RegisterPlayer("p2", "Player 2", string(PERSONALITY_AGGRESSIVE)),
		tm.RegisterPlayer("p3", "Player 3", string(PERSONALITY_DEFENSIVE)),
		tm.RegisterPlayer("p4", "Player 4", string(PERSONALITY_TACTICAL)),
	}

	// Set different ratings
	players[0].Rating = 1500
	players[1].Rating = 1300
	players[2].Rating = 1600
	players[3].Rating = 1400

	// Test full leaderboard
	leaderboard := tm.GetLeaderboard(0)
	if len(leaderboard) != 4 {
		t.Errorf("Expected 4 players in leaderboard, got %d", len(leaderboard))
	}

	// Check sorting (highest rating first)
	if leaderboard[0].Rating != 1600 {
		t.Errorf("Expected highest rated player (1600) first, got %d", leaderboard[0].Rating)
	}

	if leaderboard[3].Rating != 1300 {
		t.Errorf("Expected lowest rated player (1300) last, got %d", leaderboard[3].Rating)
	}

	// Test limited leaderboard
	topTwo := tm.GetLeaderboard(2)
	if len(topTwo) != 2 {
		t.Errorf("Expected 2 players in limited leaderboard, got %d", len(topTwo))
	}

	if topTwo[0].Rating != 1600 || topTwo[1].Rating != 1500 {
		t.Error("Top 2 players not correctly ordered")
	}
}

func TestWinRateCalculation(t *testing.T) {
	tm := NewTournamentManager()

	player := tm.RegisterPlayer("p1", "Player 1", string(PERSONALITY_BALANCED))

	// Test zero matches
	winRate := tm.CalculateWinRate(player.ID)
	if winRate != 0.0 {
		t.Errorf("Expected 0%% win rate for no matches, got %f", winRate)
	}

	// Set match statistics
	player.MatchesWon = 7
	player.MatchesLost = 2
	player.MatchesDrawn = 1

	winRate = tm.CalculateWinRate(player.ID)
	expectedWinRate := 70.0 // 7 wins out of 10 matches

	if abs(int(winRate-expectedWinRate)) > 0 {
		t.Errorf("Expected win rate %f, got %f", expectedWinRate, winRate)
	}

	// Test non-existent player
	winRate = tm.CalculateWinRate("nonexistent")
	if winRate != 0.0 {
		t.Errorf("Expected 0%% win rate for non-existent player, got %f", winRate)
	}
}

func TestRoundRobinTournament(t *testing.T) {
	tm := NewTournamentManager()

	tournament, err := tm.CreateTournament("Round Robin Test", FORMAT_ROUND_ROBIN, 4)
	if err != nil {
		t.Fatalf("CreateTournament failed: %v", err)
	}

	// Add 4 players
	for i := 0; i < 4; i++ {
		player := tm.RegisterPlayer(fmt.Sprintf("p%d", i), fmt.Sprintf("Player %d", i), string(PERSONALITY_BALANCED))
		tm.JoinTournament(tournament.ID, player.ID)
	}

	err = tm.StartTournament(tournament.ID)
	if err != nil {
		t.Fatalf("StartTournament failed: %v", err)
	}

	// Should have 6 matches (4 choose 2)
	if len(tournament.Matches) != 6 {
		t.Errorf("Expected 6 matches in round robin with 4 players, got %d", len(tournament.Matches))
	}

	// Complete all matches with predictable results
	for i, match := range tournament.Matches {
		result := RESULT_PLAYER1_WIN
		winner := match.Player1ID

		// Make player 0 win most matches
		if i%2 == 0 {
			result = RESULT_PLAYER1_WIN
			winner = match.Player1ID
		} else {
			result = RESULT_PLAYER2_WIN
			winner = match.Player2ID
		}

		tm.ReportMatchResult(tournament.ID, match.ID, result, winner)
	}

	// Tournament should be complete
	if tournament.Status != STATUS_COMPLETED {
		t.Errorf("Expected tournament to be completed, got status %v", tournament.Status)
	}

	if tournament.WinnerID == "" {
		t.Error("Tournament should have a winner")
	}
}

// Helper function for absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Benchmark tests for tournament performance
func BenchmarkCreateTournament(b *testing.B) {
	tm := NewTournamentManager()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := tm.CreateTournament(fmt.Sprintf("Tournament %d", i), FORMAT_SINGLE_ELIMINATION, 8)
		if err != nil {
			b.Fatalf("CreateTournament failed: %v", err)
		}
	}
}

func BenchmarkJoinTournament(b *testing.B) {
	tm := NewTournamentManager()

	tournament, err := tm.CreateTournament("Benchmark Tournament", FORMAT_SINGLE_ELIMINATION, 64)
	if err != nil {
		b.Fatalf("CreateTournament failed: %v", err)
	}

	// Pre-register players
	for i := 0; i < 64; i++ {
		tm.RegisterPlayer(fmt.Sprintf("p%d", i), fmt.Sprintf("Player %d", i), string(PERSONALITY_BALANCED))
	}

	b.ResetTimer()

	for i := 0; i < b.N && i < 64; i++ {
		err := tm.JoinTournament(tournament.ID, fmt.Sprintf("p%d", i))
		if err != nil {
			b.Fatalf("JoinTournament failed: %v", err)
		}
	}
}

func BenchmarkELORatingUpdate(b *testing.B) {
	tm := NewTournamentManager()

	player1 := tm.RegisterPlayer("p1", "Player 1", string(PERSONALITY_BALANCED))
	player2 := tm.RegisterPlayer("p2", "Player 2", string(PERSONALITY_BALANCED))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Alternate wins to maintain reasonable ratings
		score := 1.0
		if i%2 == 0 {
			score = 0.0
		}
		tm.updateELORating(player1, player2, score)
	}
}
