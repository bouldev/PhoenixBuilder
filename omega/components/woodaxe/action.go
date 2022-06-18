package woodaxe

import (
	"fmt"
	"phoenixbuilder/mirror/define"
	"time"
)

type Action struct {
	Do                func()
	Undo              func()
	AffectAreas       [][2]define.CubePos
	MountedStructures []string
}

type ActionManager struct {
	commandSender        func(string)
	StructureBlockPrefix string
	UnusedStructures     []string
	CurrentStructureID   int
	ActionStack          []*Action
	ActionStackPointer   int
	Freezed              bool
	actionChan           chan func()
}

func NewActionManager(prefix string, commandSender func(string), actionChan chan func()) *ActionManager {
	am := &ActionManager{
		commandSender:        commandSender,
		StructureBlockPrefix: prefix,
		UnusedStructures:     make([]string, 0),
		CurrentStructureID:   0,
		ActionStack:          make([]*Action, 0),
		ActionStackPointer:   -1,
		Freezed:              false,
		actionChan:           actionChan,
	}
	return am
}

func (o *ActionManager) AllocateStructureName() (name string) {
	if len(o.UnusedStructures) > 0 {
		name = o.UnusedStructures[0]
		o.UnusedStructures = o.UnusedStructures[1:]
		return name
	} else {
		o.CurrentStructureID++
		return fmt.Sprintf("%v%v", o.StructureBlockPrefix, o.CurrentStructureID)
	}
}

func (o *ActionManager) RenderStructureSaveCommand(name string, AffectArea [2]define.CubePos) string {
	startPos, endPos := sortPos(AffectArea[0], AffectArea[1])
	cmd := fmt.Sprintf("structure save %v %v %v %v %v %v %v true memory ", name, startPos[0], startPos[1], startPos[2], endPos[0], endPos[1], endPos[2])
	return cmd
}

func (o *ActionManager) RenderStructureLoadCommand(name string, AffectArea [2]define.CubePos) string {
	startPos, _ := sortPos(AffectArea[0], AffectArea[1])
	cmd := fmt.Sprintf("structure load %v %v %v %v", name, startPos[0], startPos[1], startPos[2])
	return cmd
}

func (o *ActionManager) Commit(a *Action) error {
	if o.Freezed {
		hint := "Action Manager is Freezed!"
		fmt.Println(hint)
		return fmt.Errorf(hint)
	} else {
		if a.AffectAreas == nil || len(a.AffectAreas) == 0 {
			a.AffectAreas = nil
			a.MountedStructures = nil
		} else {
			if a.Undo != nil {
				return fmt.Errorf("an action can define undo function by itself or point out affect area and manage by action manager, but not both")
			}
			a.MountedStructures = make([]string, 0)
			undoCmds := []string{}
			for _, area := range a.AffectAreas {
				name := o.AllocateStructureName()
				saveCmd := o.RenderStructureSaveCommand(name, area)
				o.commandSender(saveCmd)
				undoCmds = append(undoCmds, o.RenderStructureLoadCommand(name, area))
				a.MountedStructures = append(a.MountedStructures, name)
			}
			a.Undo = func() {
				for _, cmd := range undoCmds {
					o.commandSender(cmd)
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
		o.ActionStack = append(o.ActionStack, a)
		o.actionChan <- a.Do
		o.ActionStackPointer++
		return nil
	}
}

func (o *ActionManager) Freeze() {
	o.Freezed = true
}

func (o *ActionManager) DeFreeze() {
	o.Freezed = false
}

func (o *ActionManager) Undo() error {
	if o.ActionStackPointer == -1 {
		return fmt.Errorf("Cannot Undo")
	}
	o.actionChan <- o.ActionStack[o.ActionStackPointer].Undo
	o.ActionStackPointer--
	return nil
}

func (o *ActionManager) Redo() error {
	if o.ActionStackPointer+1 == len(o.ActionStack) {
		return fmt.Errorf("Cannot Redo")
	} else {
		o.ActionStackPointer++
		o.actionChan <- o.ActionStack[o.ActionStackPointer].Do
		return nil
	}
}

func (o *ActionManager) EmptyStack() {
	for _, a := range o.ActionStack {
		for _, name := range a.MountedStructures {
			o.UnusedStructures = append(o.UnusedStructures, name)
		}
	}
	o.ActionStack = make([]*Action, 0)
	o.ActionStackPointer = -1
}

func (o *ActionManager) Trim() {
	for p := o.ActionStackPointer + 1; ; p++ {
		if p == len(o.ActionStack) {
			break
		}
		a := o.ActionStack[p]
		for _, name := range a.MountedStructures {
			o.UnusedStructures = append(o.UnusedStructures, name)
		}
	}
	o.ActionStack = o.ActionStack[:o.ActionStackPointer+1]
}
