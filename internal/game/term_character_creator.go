package game

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rltermgui"
	"github.com/mechanical-lich/spaceplant/internal/background"
	"github.com/mechanical-lich/spaceplant/internal/class"
	"github.com/mechanical-lich/spaceplant/internal/skill"
)

// termCCStep is the current step in the terminal character creation flow.
type termCCStep int

const (
	termCCName       termCCStep = iota // Enter name
	termCCStats                        // Distribute stat points
	termCCClass                        // Choose class
	termCCBackground                   // Choose background
	termCCConfirm                      // Review and confirm
)

const termStatPoints = 10 // extra points to distribute beyond defaults

// TermCharacterCreator is a step-by-step terminal character creation flow.
// While active, it consumes all key input. On completion it fires OnComplete.
type TermCharacterCreator struct {
	active     bool
	step       termCCStep
	OnComplete func(data CharacterData)

	// Name step
	name []rune

	// Stats step — same order as GUI: PH, AG, MA, CL, LD, HTCS
	stats      [6]int
	statCursor int

	// Class step
	classes    []*class.ClassDef
	classCursor int

	// Background step
	backgrounds []*background.BackgroundDef
	bgCursor    int
}

// NewTermCharacterCreator creates and activates a terminal character creator.
func NewTermCharacterCreator() *TermCharacterCreator {
	cc := &TermCharacterCreator{
		active: true,
		step:   termCCName,
		stats:  [6]int{10, 10, 10, 10, 8, 30},
	}
	cc.classes = class.All()
	cc.backgrounds = background.All()
	return cc
}

func (cc *TermCharacterCreator) Active() bool  { return cc.active }
func (cc *TermCharacterCreator) Visible() bool { return cc.active }

func (cc *TermCharacterCreator) remainingPoints() int {
	defaults := [6]int{10, 10, 10, 10, 8, 30}
	used := 0
	for i, v := range cc.stats {
		d := v - defaults[i]
		if d > 0 {
			used += d
		}
	}
	return termStatPoints - used
}

// HandleKey processes a key event. Returns true if consumed.
func (cc *TermCharacterCreator) HandleKey(ev *tcell.EventKey) bool {
	if !cc.active {
		return false
	}
	switch cc.step {
	case termCCName:
		cc.handleName(ev)
	case termCCStats:
		cc.handleStats(ev)
	case termCCClass:
		cc.handleList(ev, len(cc.classes), &cc.classCursor)
	case termCCBackground:
		cc.handleList(ev, len(cc.backgrounds), &cc.bgCursor)
	case termCCConfirm:
		cc.handleConfirm(ev)
	}
	return true
}

func (cc *TermCharacterCreator) handleName(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEnter:
		cc.step = termCCStats
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(cc.name) > 0 {
			cc.name = cc.name[:len(cc.name)-1]
		}
	default:
		r := ev.Rune()
		if r >= 32 && r < 127 && len(cc.name) < 24 {
			cc.name = append(cc.name, r)
		}
	}
}

func (cc *TermCharacterCreator) handleStats(ev *tcell.EventKey) {
	defaults := [6]int{10, 10, 10, 10, 8, 30}
	switch ev.Key() {
	case tcell.KeyUp:
		if cc.statCursor > 0 {
			cc.statCursor--
		}
	case tcell.KeyDown:
		if cc.statCursor < 5 {
			cc.statCursor++
		}
	case tcell.KeyRight, tcell.KeyRune:
		r := ev.Rune()
		if ev.Key() == tcell.KeyRight || r == '+' {
			if cc.remainingPoints() > 0 {
				cc.stats[cc.statCursor]++
			}
		} else if r == '-' {
			if cc.stats[cc.statCursor] > defaults[cc.statCursor] {
				cc.stats[cc.statCursor]--
			}
		}
	case tcell.KeyLeft:
		if cc.stats[cc.statCursor] > defaults[cc.statCursor] {
			cc.stats[cc.statCursor]--
		}
	case tcell.KeyEnter:
		cc.step = termCCClass
	case tcell.KeyEscape:
		cc.step = termCCName
	}
}

func (cc *TermCharacterCreator) handleList(ev *tcell.EventKey, length int, cursor *int) {
	switch ev.Key() {
	case tcell.KeyUp:
		if *cursor > 0 {
			*cursor--
		}
	case tcell.KeyDown:
		if *cursor < length-1 {
			*cursor++
		}
	case tcell.KeyEnter:
		cc.step++
	case tcell.KeyEscape:
		cc.step--
	}
}

func (cc *TermCharacterCreator) handleConfirm(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEnter:
		cc.submit()
	case tcell.KeyEscape:
		cc.step = termCCBackground
	}
}

func (cc *TermCharacterCreator) submit() {
	name := strings.TrimSpace(string(cc.name))
	if name == "" {
		name = "Unnamed"
	}
	classID := ""
	if cc.classCursor >= 0 && cc.classCursor < len(cc.classes) {
		classID = cc.classes[cc.classCursor].ID
	}
	bgID := ""
	if cc.bgCursor >= 0 && cc.bgCursor < len(cc.backgrounds) {
		bgID = cc.backgrounds[cc.bgCursor].ID
	}
	cc.active = false
	if cc.OnComplete != nil {
		cc.OnComplete(CharacterData{
			Name:         name,
			PH:           cc.stats[statPH],
			AG:           cc.stats[statAG],
			MA:           cc.stats[statMA],
			CL:           cc.stats[statCL],
			LD:           cc.stats[statLD],
			HTCS:         cc.stats[statHTCS],
			ClassID:      classID,
			BackgroundID: bgID,
			BodyType:     "mid",
			BodyIndex:    0,
			HairIndex:    0,
		})
	}
}

// Draw renders the character creator to the terminal screen.
func (cc *TermCharacterCreator) Draw(s tcell.Screen) {
	if !cc.active {
		return
	}
	s.Clear()
	w, h := s.Size()

	title := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack)
	normal := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	selected := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
	dim := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack)
	hint := tcell.StyleDefault.Foreground(tcell.ColorTeal).Background(tcell.ColorBlack)

	steps := []string{"Name", "Stats", "Class", "Background", "Confirm"}
	breadcrumb := ""
	for i, s2 := range steps {
		if termCCStep(i) == cc.step {
			breadcrumb += "[" + s2 + "] "
		} else {
			breadcrumb += s2 + " "
		}
	}
	rltermgui.DrawText(s, 0, 0, "CHARACTER CREATION — "+breadcrumb, title)
	rltermgui.DrawText(s, 0, 1, strings.Repeat("─", w), dim)

	switch cc.step {
	case termCCName:
		rltermgui.DrawText(s, 2, 3, "Enter your name:", normal)
		nameStr := string(cc.name) + "_"
		rltermgui.DrawText(s, 2, 4, nameStr, selected)
		rltermgui.DrawText(s, 2, h-2, "[Enter] continue", hint)

	case termCCStats:
		rltermgui.DrawText(s, 2, 3, fmt.Sprintf("Distribute stat points  (remaining: %d)", cc.remainingPoints()), normal)
		statLabels := [6]string{"PH  Physique", "AG  Agility", "MA  Mental", "CL  Cool", "LD  Leadership", "HTCS Hand-to-Hand"}
		for i, label := range statLabels {
			st := normal
			if i == cc.statCursor {
				st = selected
			}
			rltermgui.DrawText(s, 2, 5+i, fmt.Sprintf("  %-22s %3d", label, cc.stats[i]), st)
		}
		rltermgui.DrawText(s, 2, h-2, "[↑↓] select  [←/−] decrease  [→/+] increase  [Enter] next  [Esc] back", hint)

	case termCCClass:
		rltermgui.DrawText(s, 2, 3, "Choose a class:", normal)
		descX := w/2 + 1
		for i, cl := range cc.classes {
			st := normal
			if i == cc.classCursor {
				st = selected
			}
			rltermgui.DrawText(s, 2, 5+i, fmt.Sprintf("  %s", cl.Name), st)
		}
		if cc.classCursor >= 0 && cc.classCursor < len(cc.classes) {
			cl := cc.classes[cc.classCursor]
			rltermgui.DrawText(s, descX, 3, cl.Name, title)
			rltermgui.DrawText(s, descX, 4, cl.Description, normal)
			if len(cl.StartingSkills) > 0 {
				rltermgui.DrawText(s, descX, 6, "Starting skills:", dim)
				for j, sk := range cl.StartingSkills {
					def := skill.Get(sk)
					name := sk
					if def != nil {
						name = def.Name
					}
					rltermgui.DrawText(s, descX+2, 7+j, name, normal)
				}
			}
		}
		rltermgui.DrawText(s, 2, h-2, "[↑↓] select  [Enter] next  [Esc] back", hint)

	case termCCBackground:
		rltermgui.DrawText(s, 2, 3, "Choose a background:", normal)
		descX := w/2 + 1
		for i, bg := range cc.backgrounds {
			st := normal
			if i == cc.bgCursor {
				st = selected
			}
			rltermgui.DrawText(s, 2, 5+i, fmt.Sprintf("  %s", bg.Name), st)
		}
		if cc.bgCursor >= 0 && cc.bgCursor < len(cc.backgrounds) {
			bg := cc.backgrounds[cc.bgCursor]
			rltermgui.DrawText(s, descX, 3, bg.Name, title)
			rltermgui.DrawText(s, descX, 4, bg.Description, normal)
			if len(bg.Skills) > 0 {
				rltermgui.DrawText(s, descX, 6, "Bonus skills:", dim)
				for j, sk := range bg.Skills {
					def := skill.Get(sk)
					name := sk
					if def != nil {
						name = def.Name
					}
					rltermgui.DrawText(s, descX+2, 7+j, name, normal)
				}
			}
		}
		rltermgui.DrawText(s, 2, h-2, "[↑↓] select  [Enter] next  [Esc] back", hint)

	case termCCConfirm:
		rltermgui.DrawText(s, 2, 3, "Review your character:", normal)
		name := strings.TrimSpace(string(cc.name))
		if name == "" {
			name = "Unnamed"
		}
		y := 5
		rltermgui.DrawText(s, 4, y, fmt.Sprintf("Name:  %s", name), normal)
		y++
		statLabels := [6]string{"PH", "AG", "MA", "CL", "LD", "HTCS"}
		statLine := ""
		for i, label := range statLabels {
			statLine += fmt.Sprintf("%s:%d  ", label, cc.stats[i])
		}
		rltermgui.DrawText(s, 4, y, "Stats: "+statLine, normal)
		y++
		className := "(none)"
		if cc.classCursor >= 0 && cc.classCursor < len(cc.classes) {
			className = cc.classes[cc.classCursor].Name
		}
		rltermgui.DrawText(s, 4, y, fmt.Sprintf("Class: %s", className), normal)
		y++
		bgName := "(none)"
		if cc.bgCursor >= 0 && cc.bgCursor < len(cc.backgrounds) {
			bgName = cc.backgrounds[cc.bgCursor].Name
		}
		rltermgui.DrawText(s, 4, y, fmt.Sprintf("Background: %s", bgName), normal)
		rltermgui.DrawText(s, 2, h-2, "[Enter] start game  [Esc] back", hint)
	}

	s.Show()
}
