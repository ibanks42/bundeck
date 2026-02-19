package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type BuiltinTask struct {
	Title    string `json:"title"`
	Pathname string `json:"pathname"`
	ID       int    `json:"id"`
	Enabled  bool   `json:"enabled"`
}

type CustomTask struct {
	Title string `json:"title"`
	ID    int    `json:"id"`
}

var (
	tabs        *container.AppTabs
	window      fyne.Window
	fyne_app    fyne.App
	preferences fyne.Preferences
)

func main() {
	fyne_app = app.NewWithID("dev.ibanks.bundeck")
	window = fyne_app.NewWindow("BunDeck Client")
	fyne_app.Settings().SetTheme(theme.DarkTheme())

	window.Resize(fyne.NewSize(600, 420))

	initializePreferences()

	buildGui()

	window.ShowAndRun()
}

func initializePreferences() {
	preferences = fyne_app.Preferences()
	if hostname := preferences.String("hostname"); len(hostname) == 0 {
		fyne_app.Preferences().SetString("hostname", "localhost")
	}
	if port := preferences.String("port"); len(port) == 0 {
		fyne_app.Preferences().SetString("port", "9212")
	}
	window.SetFullScreen(preferences.Bool("fullscreen"))

	fyne_app.Preferences().AddChangeListener(func() {
		buildScriptsTab()
		window.SetFullScreen(preferences.Bool("fullscreen"))
	})
}

func buildGui() {
	tabs = container.NewAppTabs()
	scripts_tab := container.NewTabItem("Tasks", container.NewVBox())
	settings_tab := container.NewTabItem("Settings", container.NewVBox())
	tabs.Items = append(tabs.Items, scripts_tab, settings_tab)
	tabs.SetTabLocation(container.TabLocationLeading)

	buildScriptsTab()
	buildSettingsTab()

	tabs.SelectIndex(0)

	window.SetContent(tabs)
}

func buildScriptsTab() {
	hostname := preferences.String("hostname")
	port := preferences.String("port")
	url := "http://" + hostname + ":" + port

	scripts, err := getScripts(hostname, port)
	if err != nil {
		fmt.Println(err)
		setFallbackContainer(0, "Failed to load tasks. Try again?")
		return
	}

	// create bottom widget bar
	connection_lbl := widget.NewLabel("Connected to: " + url)
	refresh_btn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		buildScriptsTab()
	})
	close_btn := widget.NewButtonWithIcon("", theme.WindowCloseIcon(), func() { os.Exit(0) })
	btn_box := container.New(
		layout.NewCustomPaddedLayout(4, 4, 4, 4),
		container.NewHBox(
			close_btn,
			layout.NewSpacer(),
			connection_lbl,
			layout.NewSpacer(),
			refresh_btn,
		),
	)

	// create task list
	var containers []fyne.CanvasObject
	for _, s := range scripts {
		title := strings.TrimSuffix(s, filepath.Ext(s))
		button := widget.NewButton(title, func() {
			go func() {
				resp, err := http.Get(url + "/scripts/" + s)
				if err != nil {
					connection_lbl.SetText(err.Error())
					return
				}

				response, err := io.ReadAll(resp.Body)
				if err != nil {
					connection_lbl.SetText(err.Error())
					return
				}

				connection_lbl.SetText(title + ": " + string(response))
			}()
		})
		layout := layout.NewCustomPaddedLayout(12, 12, 12, 12)
		containers = append(containers, container.New(layout, button))
	}

	grid := container.NewGridWrap(fyne.NewSize(256, 192), containers...)
	scroll := container.NewVScroll(grid)

	tabs.Items[0].Content = container.NewBorder(btn_box, nil, nil, nil, scroll)
}

func buildSettingsTab() {
	host := preferences.String("hostname")
	port := preferences.String("port")
	full := preferences.Bool("fullscreen")

	host_bnd := binding.BindString(&host)
	port_bnd := binding.BindString(&port)
	full_bnd := binding.BindBool(&full)

	host_ent := widget.NewEntryWithData(host_bnd)
	host_ent.Validator = nil
	port_ent := widget.NewEntryWithData(port_bnd)
	port_ent.Validator = nil
	full_chk := widget.NewCheckWithData("", full_bnd)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Hostname", Widget: host_ent},
			{Text: "Port", Widget: port_ent},
			{Text: "Fullscreen", Widget: full_chk},
		},
		SubmitText: "Save",
		OnSubmit: func() {
			preferences.SetString("hostname", host_ent.Text)
			preferences.SetString("port", port_ent.Text)
			preferences.SetBool("fullscreen", full_chk.Checked)
		},
	}

	tabs.Items[1].Content = container.NewVBox(container.NewPadded(form))
}

func setFallbackContainer(index int, text string) {
	label := widget.NewLabel(text)
	button := widget.NewButtonWithIcon("Reload", theme.ViewRefreshIcon(), func() {
		buildScriptsTab()
	})

	vbox := container.NewVBox(label, button)
	center := container.NewCenter(vbox)

	tabs.Items[index].Content = center
}
