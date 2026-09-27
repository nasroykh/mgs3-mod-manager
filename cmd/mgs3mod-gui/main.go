// Command mgs3mod-gui is the window app for the MGS3 mod manager. It only
// binds internal/gui to a Wails window; the logic and every safety check live
// in internal/gui and internal/manager.
package main

import (
	"context"
	"embed"
	"io/fs"
	"mgs3mod/internal/gui"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Only the app's own files: wails build also writes frontend/wailsjs.
//
//go:embed frontend/index.html frontend/app.css frontend/app.js frontend/icons.svg
var frontend embed.FS

// App holds the methods the window calls.
type App struct {
	ctx  context.Context
	svc  *gui.Service
	live atomic.Pointer[context.Context] // for events from action goroutines
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.live.Store(&ctx)
}

// progress forwards the service's events to the window as "progress".
func (a *App) progress(e gui.Event) {
	if ctx := a.live.Load(); ctx != nil {
		runtime.EventsEmit(*ctx, "progress", e)
	}
}

// beforeClose refuses to close while a manager command runs: each command is
// a transaction and should finish. The window explains it.
func (a *App) beforeClose(ctx context.Context) bool {
	if a.svc.Busy() {
		runtime.EventsEmit(ctx, "close-refused")
		return true
	}
	return false
}

func (a *App) Version() string                                 { return gui.Version }
func (a *App) Detect() gui.Detection                           { return a.svc.Detect() }
func (a *App) Doctor(root string) gui.Check                    { return a.svc.Doctor(root) }
func (a *App) Status(root string) gui.Overview                 { return a.svc.Status(root) }
func (a *App) Install(root string) gui.Outcome                 { return a.svc.Install(root) }
func (a *App) Uninstall(root string) gui.Outcome               { return a.svc.Uninstall(root) }
func (a *App) Verify(root string) gui.Outcome                  { return a.svc.Verify(root) }
func (a *App) PlanUninstall(root string) gui.Preview           { return a.svc.PlanUninstall(root) }
func (a *App) LaunchInfo(root string) gui.LaunchInfo           { return a.svc.LaunchInfo(root) }
func (a *App) Launch(root, profileID string) gui.Outcome       { return a.svc.Launch(root, profileID) }
func (a *App) LaunchOptions() gui.LaunchOptions                { return a.svc.Options() }
func (a *App) DeleteProfile(root, id string) gui.ProfileResult { return a.svc.DeleteProfile(root, id) }
func (a *App) SetDefaultProfile(root, id string) gui.ProfileResult {
	return a.svc.SetDefaultProfile(root, id)
}

func (a *App) SaveProfile(root, original, name, region, language, controller string, makeDefault bool) gui.ProfileResult {
	return a.svc.SaveProfile(root, original, name, region, language, controller, makeDefault)
}

func (a *App) SetEnabled(root, id string, on bool) gui.Outcome {
	return a.svc.SetEnabled(root, id, on)
}

func (a *App) AddPackage(root, path string) gui.Outcome { return a.svc.AddPackage(root, path) }

// Remember stores the chosen folder; it returns an error text or "".
func (a *App) Remember(root string) string {
	if err := a.svc.Remember(root); err != nil {
		return err.Error()
	}
	return ""
}

// ChooseFolder opens a folder picker; "" when cancelled.
func (a *App) ChooseFolder() string {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: "Choose the MGS3 folder (the one with METAL GEAR SOLID3.exe)"})
	if err != nil {
		return ""
	}
	return dir
}

// ChoosePackage opens a file picker for a mod package; "" when cancelled.
func (a *App) ChoosePackage() string {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "Choose a mod package",
		Filters: []runtime.FileFilter{{DisplayName: "Mod packages (*.zip)", Pattern: "*.zip"}},
	})
	if err != nil {
		return ""
	}
	return path
}

// OpenFolder shows the game folder in Explorer; it returns an error text or "".
func (a *App) OpenFolder(root string) string {
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return "The game folder cannot be opened."
	}
	cmd := exec.Command("explorer.exe", root)
	if err := cmd.Start(); err != nil {
		return err.Error()
	}
	go cmd.Wait()
	return ""
}

func main() {
	settings, _ := gui.DefaultSettingsPath()
	app := &App{svc: gui.Production(gui.ExeDir(), settings)}
	app.svc.SetProgress(app.progress)
	// Keep WebView2's own data beside gui.json in %LOCALAPPDATA%\mgs3mod
	// instead of Wails' default, a folder named after the exe in %APPDATA%.
	webviewData := ""
	if settings != "" {
		webviewData = filepath.Join(filepath.Dir(settings), "webview")
	}
	assets, err := fs.Sub(frontend, "frontend")
	if err != nil {
		println("Error:", err.Error())
		os.Exit(1)
	}
	err = wails.Run(&options.App{
		Title:         "MGS3 Mod Manager",
		Width:         980,
		Height:        720,
		MinWidth:      640,
		MinHeight:     520,
		AssetServer:   &assetserver.Options{Assets: assets},
		OnStartup:     app.startup,
		OnBeforeClose: app.beforeClose,
		Bind:          []interface{}{app},
		DragAndDrop:   &options.DragAndDrop{DisableWebViewDrop: true},
		Windows:       &windows.Options{Theme: windows.SystemDefault, WebviewUserDataPath: webviewData},
	})
	if err != nil {
		println("Error:", err.Error())
		os.Exit(1)
	}
}
