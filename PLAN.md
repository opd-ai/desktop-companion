# JRPG Battle System Design Plan for DDS

## ANALYSIS PHASE RESULTS

### Character Data Model Analysis

**JSON Character Card Structure:**
- **CharacterCard** (`internal/character/card.go`): Core configuration with extensible JSON schema
- **Stats System**: Map-based stats with `StatConfig` (initial, max, degradationRate, criticalThreshold)
- **Interactions**: `InteractionConfig` with effects, animations, responses, cooldowns, requirements
- **Multiplayer Config**: `MultiplayerConfig` with networking ID, bot personality, peer limits
- **Backward Compatibility**: Optional fields with `omitempty` JSON tags ensure existing cards work unchanged

**Current Stat Model:**
```go
type Stat struct {
    Current           float64
    Max               float64  
    DegradationRate   float64
    CriticalThreshold float64
}
```

### Animation System Mapping

**AnimationManager** (`internal/character/animation.go`):
- **GIF Loading**: Standard library `image/gif` decoder with frame timing
- **Frame Management**: Thread-safe with `sync.RWMutex` protection
- **Animation States**: Map-based storage `map[string]*gif.GIF`
- **Timing Control**: Uses GIF delay values for frame advancement
- **Update Pattern**: `Update()` returns bool for frame changes, called from main loop

**Key Methods:**
- `LoadAnimation(name, filepath string)`: Loads GIF from file system
- `SetCurrentAnimation(name string)`: Switches active animation
- `GetCurrentFrame()`: Returns current frame image + timing flag
- `Update()`: Advances frames based on timing, returns frame change status

### Multiplayer Communication Analysis

**NetworkManager Interface** (`internal/network/manager.go`):
- **Peer Discovery**: UDP broadcast every 5 seconds on configurable port
- **Message Protocol**: TCP with JSON payload + Ed25519 signatures
- **Message Types**: Discovery, CharacterAction, StateSync, PeerList
- **Thread Safety**: Mutex protection for concurrent access

**Message Structure:**
```go
type NetworkMessage struct {
    Type      string
    From      string 
    To        string
    Payload   []byte
    Timestamp time.Time
}

type CharacterActionPayload struct {
    Action        string
    CharacterID   string
    Position      *NetworkPosition
    Animation     string
    Response      string
    Metadata      map[string]interface{}
    InteractionID string
}
```

**MultiplayerCharacter Wrapper** (`internal/character/multiplayer.go`):
- **Embedding Pattern**: Wraps existing Character with network coordination
- **Action Broadcasting**: Overrides HandleClick/HandleRightClick for network sync
- **Handler Registration**: Message type routing with callback functions

### Inventory System Implementation

**Gift System** (`internal/character/gift_manager.go`):
- **GiftDefinition**: JSON-based items with effects, rarity, unlock requirements
- **Effect Framework**: Immediate stat effects + personality modifiers
- **Item Filtering**: Requirement-based availability (relationship level, stats)
- **Memory Integration**: Tracks gift interactions in `GiftMemory` structs

**Current Item Structure:**
```go
type GiftDefinition struct {
    ID                   string
    Name                 string
    Category             string
    Rarity               string
    Properties           GiftProperties
    GiftEffects          GiftEffects
    PersonalityModifiers map[string]map[string]float64
}

type GiftEffects struct {
    Immediate ImmediateEffects
    Memory    MemoryEffects
}

type ImmediateEffects struct {
    Stats      map[string]float64
    Animations []string
    Responses  []string  
}
```

## BATTLE SYSTEM DESIGN

### 1. Interface Mapping

**Communication Interface** (Existing):
```go
type NetworkManager interface {
    Broadcast(msg NetworkMessage) error
    RegisterHandler(msgType string, handler func(NetworkMessage, interface{}) error)
}
```

**Rendering Interface** (Existing):
```go
type AnimationManager interface {
    LoadAnimation(name, filepath string) error
    SetCurrentAnimation(name string) error
    GetCurrentFrame() (image.Image, bool)
    Update() bool
}
```

**Inventory Interface** (Existing):
```go
type GiftManager interface {
    GetAvailableGifts() []*GiftDefinition
    GiveGift(giftID, notes string) (*GiftResponse, error)
    IsGiftSystemEnabled() bool
}
```

**New Battle Interface**:
```go
type BattleManager interface {
    InitiateBattle(opponentID string) error
    PerformAction(action BattleAction, targetID string) (*BattleResult, error)
    GetBattleState() *BattleState
    GetAvailableActions() []BattleActionType
    EndBattle() error
}
```

### 2. Balance Framework

**Base Action Values** (All pets identical):
```go
const (
    BASE_ATTACK_DAMAGE  = 20.0
    BASE_DEFEND_REDUCTION = 0.5  // 50% damage reduction
    BASE_HEAL_AMOUNT    = 25.0
    BASE_STUN_DURATION  = 1      // 1 turn
    BASE_BOOST_AMOUNT   = 15.0   // +15 to attack for 3 turns
    BASE_DRAIN_RATIO    = 0.3    // 30% of damage dealt as heal
    BASE_SHIELD_ABSORPTION = 30.0
    BASE_CHARGE_MULTIPLIER = 1.5 // 50% damage boost when charged
)
```

**Item Effect Caps**:
```go
const (
    MAX_DAMAGE_MODIFIER    = 1.20  // +20% max damage
    MAX_DEFENSE_MODIFIER   = 1.15  // +15% max defense  
    MAX_SPEED_MODIFIER     = 1.10  // +10% max speed
    MAX_HEAL_MODIFIER      = 1.25  // +25% max healing
    MAX_EFFECT_STACKING    = 3     // Maximum 3 item effects
)
```

**Battle Stats** (Extended from existing system):
```go
type BattleStats struct {
    HP           float64  // Current hit points
    MaxHP        float64  // Maximum hit points
    Attack       float64  // Base attack power
    Defense      float64  // Base defense rating
    Speed        float64  // Turn order priority
    Modifiers    []BattleModifier
}

type BattleModifier struct {
    Type         ModifierType
    Value        float64
    Duration     int  // Turns remaining
    Source       string  // Item/action that created it
}
```

### 3. Battle State Design

**Core Battle State**:
```go
type BattleState struct {
    BattleID        string
    Participants    map[string]*BattleParticipant
    TurnOrder       []string
    CurrentTurn     int
    Phase           BattlePhase
    TurnTimeout     time.Duration
    Started         time.Time
    LastAction      *BattleAction
    mu              sync.RWMutex
}

type BattleParticipant struct {
    CharacterID     string
    PeerID          string
    IsLocal         bool
    Stats           BattleStats
    ActiveItems     []ActiveBattleItem
    ActionHistory   []BattleAction
    LastActionTime  time.Time
    IsReady         bool
}

type BattleAction struct {
    Type            BattleActionType
    ActorID         string
    TargetID        string
    ItemUsed        string  // Optional item enhancement
    Timestamp       time.Time
    Result          *BattleResult
}

type BattleResult struct {
    Success         bool
    Damage          float64
    Healing         float64
    StatusEffects   []StatusEffect
    Animation       string
    Response        string
    ModifiersApplied []BattleModifier
}
```

**Action Types**:
```go
type BattleActionType string

const (
    ACTION_ATTACK   BattleActionType = "attack"
    ACTION_DEFEND   BattleActionType = "defend"  
    ACTION_STUN     BattleActionType = "stun"
    ACTION_HEAL     BattleActionType = "heal"
    ACTION_BOOST    BattleActionType = "boost"
    ACTION_COUNTER  BattleActionType = "counter"
    ACTION_DRAIN    BattleActionType = "drain"
    ACTION_SHIELD   BattleActionType = "shield"
    ACTION_CHARGE   BattleActionType = "charge"
    ACTION_EVADE    BattleActionType = "evade"
    ACTION_TAUNT    BattleActionType = "taunt"
)
```

### 4. Turn Resolution Pipeline

**Turn Processing Flow**:
```go
func (bm *BattleManager) ProcessTurn(action BattleAction) (*BattleResult, error) {
    // 1. Validate action legality
    if err := bm.validateAction(action); err != nil {
        return nil, err
    }
    
    // 2. Apply item modifiers with caps
    modifiedAction := bm.applyItemModifiers(action)
    
    // 3. Calculate base effect
    baseResult := bm.calculateBaseEffect(modifiedAction)
    
    // 4. Apply fairness constraints
    cappedResult := bm.applyFairnessCaps(baseResult)
    
    // 5. Execute effect on target
    finalResult := bm.executeEffect(cappedResult)
    
    // 6. Broadcast to all participants
    bm.broadcastTurnResult(finalResult)
    
    // 7. Advance turn order
    bm.advanceTurn()
    
    return finalResult, nil
}
```

**Fairness Validation**:
```go
func (bm *BattleManager) applyFairnessCaps(result *BattleResult) *BattleResult {
    // Cap damage modifications
    if result.Damage > BASE_ATTACK_DAMAGE * MAX_DAMAGE_MODIFIER {
        result.Damage = BASE_ATTACK_DAMAGE * MAX_DAMAGE_MODIFIER
    }
    
    // Cap healing modifications  
    if result.Healing > BASE_HEAL_AMOUNT * MAX_HEAL_MODIFIER {
        result.Healing = BASE_HEAL_AMOUNT * MAX_HEAL_MODIFIER
    }
    
    // Validate modifier stacking
    if len(result.ModifiersApplied) > MAX_EFFECT_STACKING {
        result.ModifiersApplied = result.ModifiersApplied[:MAX_EFFECT_STACKING]
    }
    
    return result
}
```

### 5. AI Behavior System

**Timeout-Triggered Decisions**:
```go
type BattleAI struct {
    character       *Character
    difficulty      AIDifficulty
    availableItems  []*GiftDefinition
    lastActions     []BattleAction
    strategy        AIStrategy
}

func (ai *BattleAI) SelectAction(battleState *BattleState, timeRemaining time.Duration) BattleAction {
    // Emergency timeout decision (< 5 seconds)
    if timeRemaining < 5*time.Second {
        return ai.selectQuickAction(battleState)
    }
    
    // Analyze current situation
    threat := ai.assessThreat(battleState)
    opportunity := ai.assessOpportunity(battleState) 
    
    // Select strategy based on situation
    if threat > opportunity {
        return ai.selectDefensiveAction(battleState)
    } else {
        return ai.selectOffensiveAction(battleState)
    }
}

func (ai *BattleAI) selectQuickAction(battleState *BattleState) BattleAction {
    // Simple heuristic for timeout situations
    myHP := battleState.GetParticipant(ai.character.ID).Stats.HP
    maxHP := battleState.GetParticipant(ai.character.ID).Stats.MaxHP
    
    if myHP < maxHP * 0.3 {
        return BattleAction{Type: ACTION_HEAL, ActorID: ai.character.ID}
    }
    
    return BattleAction{Type: ACTION_ATTACK, ActorID: ai.character.ID}
}
```

**Item Integration**:
```go
func (ai *BattleAI) selectBestItem(actionType BattleActionType) string {
    availableItems := ai.giftManager.GetAvailableGifts()
    
    bestItem := ""
    bestBonus := 0.0
    
    for _, item := range availableItems {
        bonus := ai.calculateItemBonus(item, actionType)
        if bonus > bestBonus && bonus <= ai.getMaxAllowedBonus(actionType) {
            bestItem = item.ID
            bestBonus = bonus
        }
    }
    
    return bestItem
}
```

### 6. Animation Integration

**Battle Animation Mapping**:
```go
var BattleAnimationMap = map[BattleActionType]string{
    ACTION_ATTACK:  "attack",     // New animation: aggressive forward motion
    ACTION_DEFEND:  "defend",     // New animation: protective stance
    ACTION_STUN:    "stun",       // New animation: dizzying effect
    ACTION_HEAL:    "heal",       // New animation: glowing recovery
    ACTION_BOOST:   "boost",      // New animation: power-up effect
    ACTION_COUNTER: "counter",    // New animation: reactive strike
    ACTION_DRAIN:   "drain",      // New animation: energy absorption
    ACTION_SHIELD:  "shield",     // New animation: barrier formation
    ACTION_CHARGE:  "charge",     // New animation: energy building
    ACTION_EVADE:   "evade",      // New animation: quick dodge
    ACTION_TAUNT:   "taunt",      // New animation: provocative gesture
}
```

**Animation System Integration**:
```go
func (bm *BattleManager) playBattleAnimation(action BattleAction, result *BattleResult) {
    // Select primary animation for action
    primaryAnim := BattleAnimationMap[action.Type]
    
    // Add item-specific effects if item was used
    if action.ItemUsed != "" {
        if item := bm.getItemDefinition(action.ItemUsed); item != nil {
            // Overlay item effects on animation
            bm.addItemAnimationEffects(primaryAnim, item)
        }
    }
    
    // Set animation on character
    bm.setCharacterAnimation(action.ActorID, primaryAnim)
    
    // Schedule return to idle after animation completes
    bm.scheduleIdleReturn(action.ActorID, time.Second*2)
}
```

**GIF Asset Requirements**:
```
assets/characters/[archetype]/animations/battle/
├── attack.gif     # Aggressive forward motion
├── defend.gif     # Protective blocking stance  
├── stun.gif       # Dizzied/stunned state
├── heal.gif       # Glowing recovery animation
├── boost.gif      # Power-up energy effect
├── counter.gif    # Reactive counter-attack
├── drain.gif      # Energy absorption visual
├── shield.gif     # Barrier/shield formation
├── charge.gif     # Building energy/power
├── evade.gif      # Quick dodge movement
├── taunt.gif      # Provocative gesture
└── victory.gif    # Battle won celebration
```

## IMPLEMENTATION STRATEGY

### Phase 1: Core Battle Framework (Week 1)
**Files to Create:**
- `internal/battle/manager.go` - Core battle state management
- `internal/battle/actions.go` - Action processing and validation  
- `internal/battle/ai.go` - Timeout-driven AI decisions
- `internal/battle/fairness.go` - Balance constraint enforcement

**Character Card Extensions:**
```json
{
  "battleSystem": {
    "enabled": true,
    "battleStats": {
      "hp": {"base": 100, "max": 100},
      "attack": {"base": 20, "max": 25}, 
      "defense": {"base": 15, "max": 20},
      "speed": {"base": 10, "max": 15}
    },
    "aiDifficulty": "normal",
    "preferredActions": ["attack", "heal", "defend"]
  }
}
```

### Phase 2: Multiplayer Integration (Week 2)  
**Network Message Extensions:**
```go
const (
    MESSAGE_BATTLE_INVITE    = "battle_invite"
    MESSAGE_BATTLE_ACTION    = "battle_action"
    MESSAGE_BATTLE_RESULT    = "battle_result"
    MESSAGE_BATTLE_END       = "battle_end"
)
```

**MultiplayerCharacter Extensions:**
```go
func (mc *MultiplayerCharacter) InitiateBattle(targetPeerID string) error
func (mc *MultiplayerCharacter) HandleBattleInvite(invite BattleInvite) error  
func (mc *MultiplayerCharacter) PerformBattleAction(action BattleAction) error
```

### Phase 3: Animation & UI Integration (Week 3)
**Animation Loading:**
- Extend existing `LoadCard()` to load battle animations
- Add battle animation validation to character card schema
- Integrate with existing `AnimationManager.LoadAnimation()`

**UI Components:**
- Battle initiation context menu option
- Action selection interface during battle
- Turn timer display
- Battle result overlay

### Phase 4: Item System Integration (Week 4)
**Gift System Extensions:**
```go
type BattleItemEffect struct {
    ActionType      BattleActionType
    DamageModifier  float64  // Capped at MAX_DAMAGE_MODIFIER
    DefenseModifier float64  // Capped at MAX_DEFENSE_MODIFIER  
    SpeedModifier   float64  // Capped at MAX_SPEED_MODIFIER
    Duration        int      // Turns the effect lasts
}
```

**Item Integration:**
- Extend `GiftDefinition` with `BattleEffects` field
- Modify AI decision making to consider available items
- Add item usage validation in `BattleManager.validateAction()`

## FAIRNESS ENFORCEMENT

### Constraint Validation Pipeline
```go
type FairnessValidator struct {
    maxDamageModifier  float64
    maxDefenseModifier float64
    maxSpeedModifier   float64
    maxEffectStacking  int
}

func (fv *FairnessValidator) ValidateAction(action BattleAction, participant *BattleParticipant) error {
    // Check action is legal for current state
    if !fv.isActionLegal(action, participant) {
        return ErrIllegalAction
    }
    
    // Validate item effects don't exceed caps
    if action.ItemUsed != "" {
        if err := fv.validateItemEffects(action.ItemUsed); err != nil {
            return err
        }
    }
    
    // Check modifier stacking limits
    if len(participant.Stats.Modifiers) >= fv.maxEffectStacking {
        return ErrMaxModifiersReached
    }
    
    return nil
}
```

### Balance Testing Framework
```go
func TestBattleBalance(t *testing.T) {
    // Test that all base actions deal identical damage
    testCases := []struct{
        action BattleActionType
        expectedDamage float64
    }{
        {ACTION_ATTACK, BASE_ATTACK_DAMAGE},
        {ACTION_DRAIN, BASE_ATTACK_DAMAGE * BASE_DRAIN_RATIO},
    }
    
    for _, tc := range testCases {
        result := calculateBaseEffect(BattleAction{Type: tc.action})
        assert.Equal(t, tc.expectedDamage, result.Damage)
    }
}

func TestItemCapEnforcement(t *testing.T) {
    // Test that item effects are properly capped
    overpoweredItem := &GiftDefinition{
        BattleEffects: BattleItemEffect{
            DamageModifier: 2.0, // 100% boost (should be capped to 20%)
        },
    }
    
    cappedEffect := applyFairnessCaps(overpoweredItem.BattleEffects)
    assert.LessOrEqual(t, cappedEffect.DamageModifier, MAX_DAMAGE_MODIFIER)
}
```

## BACKWARD COMPATIBILITY

### Character Card Compatibility
- All existing character cards work unchanged
- Battle system is opt-in via `"battleSystem": {"enabled": true}`
- Default battle stats derived from existing game stats
- Graceful degradation when battle animations missing

### Multiplayer Compatibility  
- Battle invites only sent to peers with battle system enabled
- Non-battle-enabled characters ignore battle messages
- Existing peer discovery and communication unchanged

### Save System Compatibility
- Battle stats saved as extension to existing `GameState`
- Battle history stored in existing memory patterns
- No breaking changes to save file format

This design provides a complete, minimally invasive turn-based battle system that leverages all existing DDS infrastructure while maintaining strict fairness constraints and backward compatibility.
