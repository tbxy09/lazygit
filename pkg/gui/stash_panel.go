package gui

import (
	"fmt"

	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/commands"
)

func (gui *Gui) refreshStashEntries(g *gocui.Gui) error {
	g.Update(func(g *gocui.Gui) error {
		v, err := g.View("stash")
		if err != nil {
			panic(err)
		}
		gui.State.StashEntries = gui.GitCommand.GetStashEntries()
		v.Clear()
		for _, stashEntry := range gui.State.StashEntries {
			fmt.Fprintln(v, stashEntry.DisplayString)
		}
		return gui.resetOrigin(v)
	})
	return nil
}

func (gui *Gui) getSelectedStashEntry(v *gocui.View) *commands.StashEntry {
	if len(gui.State.StashEntries) == 0 {
		return nil
	}
	lineNumber := gui.getItemPosition(v)
	return &gui.State.StashEntries[lineNumber]
}

func (gui *Gui) renderStashOptions(g *gocui.Gui) error {
	return gui.renderOptionsMap(g, map[string]string{
		"space":   gui.Tr.SLocalize("apply"),
		"g":       gui.Tr.SLocalize("pop"),
		"d":       gui.Tr.SLocalize("drop"),
		"← → ↑ ↓": gui.Tr.SLocalize("navigate"),
	})
}

func (gui *Gui) handleStashEntrySelect(g *gocui.Gui, v *gocui.View) error {
	if err := gui.renderStashOptions(g); err != nil {
		return err
	}
	go func() {
		stashEntry := gui.getSelectedStashEntry(v)
		if stashEntry == nil {
			gui.renderString(g, "main", gui.Tr.SLocalize("NoStashEntries"))
			return
		}
		diff, _ := gui.GitCommand.GetStashEntryDiff(stashEntry.Index)
		gui.renderString(g, "main", diff)
	}()
	return nil
}

func (gui *Gui) handleStashApply(g *gocui.Gui, v *gocui.View) error {
	return gui.stashDo(g, v, "apply")
}

func (gui *Gui) handleStashPop(g *gocui.Gui, v *gocui.View) error {
	return gui.stashDo(g, v, "pop")
}

func (gui *Gui) handleStashDrop(g *gocui.Gui, v *gocui.View) error {
	title := gui.Tr.SLocalize("StashDrop")
	message := gui.Tr.SLocalize("SureDropStashEntry")
	return gui.createConfirmationPanel(g, v, title, message, func(g *gocui.Gui, v *gocui.View) error {
		return gui.stashDo(g, v, "drop")
	}, nil)
}

func (gui *Gui) stashDo(g *gocui.Gui, v *gocui.View, method string) error {
	stashEntry := gui.getSelectedStashEntry(v)
	if stashEntry == nil {
		errorMessage := gui.Tr.TemplateLocalize(
			"NoStashTo",
			Teml{
				"method": method,
			},
		)
		return gui.createErrorPanel(g, errorMessage)
	}
	if err := gui.GitCommand.StashDo(stashEntry.Index, method); err != nil {
		gui.createErrorPanel(g, err.Error())
	}
	gui.refreshStashEntries(g)
	return gui.refreshFiles(g)
}

func (gui *Gui) handleStashSave(g *gocui.Gui, filesView *gocui.View) error {
	if len(gui.trackedFiles()) == 0 && len(gui.stagedFiles()) == 0 {
		return gui.createErrorPanel(g, gui.Tr.SLocalize("NoTrackedStagedFilesStash"))
	}
	gui.createPromptPanel(g, filesView, gui.Tr.SLocalize("StashChanges"), func(g *gocui.Gui, v *gocui.View) error {
		if err := gui.GitCommand.StashSave(gui.trimmedContent(v)); err != nil {
			gui.createErrorPanel(g, err.Error())
		}
		gui.refreshStashEntries(g)
		return gui.refreshFiles(g)
	})
	return nil
}
