package main

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

//go:embed Icon.png
var logo []byte

type GUI struct {
	window         fyne.Window
	app            fyne.App
	scriptsTab     *container.TabItem
	preferencesTab *container.TabItem
	tabs           *container.AppTabs
	preferences    fyne.Preferences
}

func NewGUI() *GUI {
	gui := &GUI{
		app:         app.NewWithID("dev.ibanks.bundeck-server"),
		preferences: fyne.CurrentApp().Preferences(),
	}
	gui.window = gui.app.NewWindow("BunDeck Server")
	return gui
}

func (g *GUI) Initialize() {
	g.app.Settings().SetTheme(theme.DarkTheme())
	g.window.Resize(fyne.Size{Width: 600, Height: 400})

	g.setupSystemTray()
	g.buildGUI()
	g.setupCloseHandler()
}

func (g *GUI) Run() {
	if g.preferences.Bool("minimized") {
		g.app.Run()
	} else {
		g.window.ShowAndRun()
	}
}

func (g *GUI) setupSystemTray() {
	if desk, ok := g.app.(desktop.App); ok {
		menu := fyne.NewMenu("BunDeck Server",
			fyne.NewMenuItem("Show", func() {
				g.window.Show()
			}))

		desk.SetSystemTrayMenu(menu)
		desk.SetSystemTrayIcon(fyne.NewStaticResource("logo", logo))
	}
}

func (g *GUI) setupCloseHandler() {
	g.window.SetCloseIntercept(func() {
		g.window.Hide()
	})
}

func (g *GUI) buildGUI() {
	g.scriptsTab = container.NewTabItem("Builtin Tasks", container.NewVBox())
	g.preferencesTab = container.NewTabItem("Settings", container.NewVBox())

	g.tabs = container.NewAppTabs(g.scriptsTab, g.preferencesTab)
	g.tabs.SetTabLocation(container.TabLocationLeading)

	g.tabs.OnSelected = func(tab *container.TabItem) {
		g.refreshGUI(g.tabs.SelectedIndex())
	}

	g.setupMainMenu()
	g.refreshGUI(0)
	g.window.SetContent(g.tabs)
}

func (g *GUI) setupMainMenu() {
	menu := fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("Refresh", g.buildGUI),
			fyne.NewMenuItem("New Task", g.showNewTaskDialog)))
	g.window.SetMainMenu(menu)
}

func (g *GUI) showNewTaskDialog() {
	idEntry := widget.NewEntry()
	titleEntry := widget.NewEntry()
	command := widget.NewMultiLineEntry()
	idEntry.SetText(strconv.Itoa(getMaxScriptId() + 1))

	items := []*widget.FormItem{
		widget.NewFormItem("ID", idEntry),
		widget.NewFormItem("Name", titleEntry),
		widget.NewFormItem("Script", command),
	}

	dialog.NewForm("New Task", "Confirm", "Cancel", items,
		func(confirmed bool) {
			if confirmed {
				g.handleNewTask(idEntry.Text, titleEntry.Text, command.Text)
			}
		}, g.window).Show()
}

func (g *GUI) handleNewTask(idText, title, command string) {
	id, err := strconv.Atoi(idText)
	if err != nil {
		fmt.Println("Failed to create script: ID is not a number")
		return
	}

	if err := writeScript(id, title+".ts", command); err != nil {
		fmt.Println("Failed to create script:", err.Error())
		return
	}

	g.refreshGUI(0)
}

func (g *GUI) refreshGUI(tabIndex int) {
	g.buildScriptsTab()
	g.buildPreferencesTab()
	g.tabs.SelectIndex(tabIndex)
}

func (g *GUI) buildScriptsTab() {
	scripts := getScripts()
	sort.Slice(scripts, func(i, j int) bool {
		return scripts[i].ID < scripts[j].ID
	})

	listmap := binding.NewUntypedList()
	for _, script := range scripts {
		listmap.Append(script)
	}

	list := widget.NewListWithData(listmap,
		g.createScriptListItem,
		g.updateScriptListItem)

	scroll := container.NewScroll(list)
	padded := layout.NewCustomPaddedLayout(0, 0, 16, 0)
	g.scriptsTab.Content = container.New(padded, scroll)
}

func (g *GUI) createScriptListItem() fyne.CanvasObject {
	return container.NewHBox(
		widget.NewLabel(""),
		layout.NewSpacer(),
		container.NewGridWithColumns(2,
			widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {}),
			widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {}),
		))
}

func (g *GUI) updateScriptListItem(dataItem binding.DataItem, canvasObject fyne.CanvasObject) {
	untyped, _ := dataItem.(binding.Untyped).Get()
	script := untyped.(Script)
	objects := canvasObject.(*fyne.Container).Objects
	scriptNoExt := strings.TrimSuffix(script.File, filepath.Ext(script.File))

	objects[0].(*widget.Label).SetText(scriptNoExt)

	editBtn := objects[2].(*fyne.Container).Objects[0].(*widget.Button)
	editBtn.OnTapped = func() { g.showEditTaskDialog(script) }

	deleteBtn := objects[2].(*fyne.Container).Objects[1].(*widget.Button)
	deleteBtn.OnTapped = func() {
		// TODO: Implement delete functionality
		g.refreshGUI(0)
	}
}

func (g *GUI) showEditTaskDialog(script Script) {
	taskData, err := readScript(script.File)
	if err != nil {
		return
	}

	idEntry := widget.NewEntry()
	filenameEntry := widget.NewEntry()
	commandEntry := widget.NewMultiLineEntry()

	idEntry.SetText(strconv.Itoa(script.ID))
	filenameEntry.SetText(script.File)
	commandEntry.SetText(taskData)

	items := []*widget.FormItem{
		widget.NewFormItem("ID", idEntry),
		widget.NewFormItem("Name", filenameEntry),
		widget.NewFormItem("Script", commandEntry),
	}

	dialog.NewForm("Edit Task", "Confirm", "Cancel", items,
		func(confirmed bool) {
			if confirmed {
				g.handleEditTask(script, idEntry.Text, commandEntry.Text)
			}
		}, g.window).Show()
}

func (g *GUI) handleEditTask(script Script, idText, command string) {
	id, err := strconv.Atoi(idText)
	if err != nil {
		fmt.Println("Failed to update task: ID not a number")
		return
	}

	if err := updateScript(script, id, command); err != nil {
		fmt.Println("Failed to update custom task:", err.Error())
		return
	}

	g.refreshGUI(0)
}

func (g *GUI) buildPreferencesTab() {
	minimized := g.preferences.Bool("minimized")
	port := g.preferences.StringWithFallback("port", "9212")

	minimizedBinding := binding.BindBool(&minimized)
	portBinding := binding.BindString(&port)

	minimizedCheck := widget.NewCheckWithData("", minimizedBinding)
	portInput := widget.NewEntryWithData(portBinding)

	form := widget.NewForm(
		widget.NewFormItem("Port", portInput),
		widget.NewFormItem("Start Minimized", minimizedCheck))

	form.OnSubmit = func() {
		g.preferences.SetBool("minimized", minimizedCheck.Checked)
		g.preferences.SetString("port", portInput.Text)
		StartServer()
	}

	g.preferencesTab.Content = container.NewVBox(form)
}

func main() {
	gui := NewGUI()
	StartServer()
	gui.Initialize()
	gui.Run()
}
