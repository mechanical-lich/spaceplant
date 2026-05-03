# Controls

## Movement

| Key | Action |
|---|---|
| W | Move north |
| S | Move south |
| A | Move west |
| D | Move east |

## Facing (turn without moving)

Holding **Shift** while pressing a movement key rotates you to face that direction without spending a turn. Use this to pre-aim before firing.

| Key | Action |
|---|---|
| Shift+W | Face north |
| Shift+S | Face south |
| Shift+A | Face west |
| Shift+D | Face east |

## Combat

| Key | Action |
|---|---|
| F | Enter targeting mode |
| G | Enter targeting mode (burst fire) |
| H | Heal |

### Targeting Mode

Pressing **F** or **G** opens targeting mode. The nearest visible enemy in your facing direction is selected automatically and a line traces the shot path.

| Key | Action |
|---|---|
| Tab | Cycle to the next visible target |
| Enter or F | Confirm and fire |
| Shift+F | Open body part picker for the current target, then fire |
| Left click | Select the enemy at the clicked tile |
| Escape | Cancel and exit targeting mode |

## Inventory & Equipment

| Key | Action |
|---|---|
| E | Equip an item from your bag |
| P | Open nearby items — pick up or equip items within 1 tile |
| I | Open inventory |
| Shift+I | Open character overview |
| Shift+R | Open the reload modal |

## Other

| Key | Action |
|---|---|
| R | Toggle rush mode (faster movement, more energy) |
| . (Period) | Use stairs |
| Shift+C | Open class upgrade screen |
| Escape | Cancel targeting mode, close the current modal, or open the pause menu |

## Pause Menu

Press **Escape** when no modal is open to bring up the pause menu. From here you can:

| Option | What it does |
|--------|-------------|
| Save | Manually save the current game |
| Options | Adjust display and gameplay settings (see below) |
| Return to Title | Exit to the title screen |
| Close | Resume the game |

## Options

The Options screen is available from both the **title screen** and the **pause menu**. Changes take effect immediately and are written to `data/config.json`.

| Setting | Description |
|---------|-------------|
| CRT Intensity (0–1) | Strength of the CRT scanline/warp effect on the game map. `0` disables it entirely. |
| Press Delay (ticks) | How long a key must be held before it starts repeating. Lower = faster repeat. |
| Render Scale | Zoom level for the map viewport. `1.0` = pixel-perfect, `2.0` = double size. |
| NPC Turn Delay | Server ticks to pause between NPC-only turns. Higher = slower enemies. |

## Skill Hotkeys

Some skills bind additional keys when active:

| Skill | Key | Action |
|---|---|---|
| Roundhouse Kick / Martial Artist | K or R | Strike all adjacent enemies |
| Shove | V | Push the entity in front of you one tile |
| Flamethrower (item) | C | Fire a cone of flame |
| Acid Spray (skill) | C | Fire a cone of acid |
| Poisonous Bite | B | Bite the enemy directly in front of you |
| Paralyzing Vinewhip | V | Lash the enemy in front of you |
