package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// Embed resource files.
//
//go:embed resources/edited-localization-string-tables-english(en)_assets_all.bundle
var patchResource []byte

//go:embed resources/edited-localization-string-tables-english(en)_assets_all.bundle
var patchResourceSwitch []byte

//go:embed resources/edited-localization-string-tables-english(en)_assets_all-switch.bundle
var restoreResourcePc []byte

//go:embed resources/original-localization-string-tables-english(en)_assets_all-switch.bundle
var restoreResourceSwitch []byte

func main() {
	a := app.New()
	w := a.NewWindow("Eiyuden Patcher Tool")

	// Set window size
	w.Resize(fyne.NewSize(800, 500))

	// Apply custom theme with reduced font size
	a.Settings().SetTheme(&CustomTheme{})

	// Entry field with increased width
	folderEntry := widget.NewEntry()
	folderEntry.SetPlaceHolder("No folder selected...")
	folderEntry.Resize(fyne.NewSize(650, 50))

	// Function to open folder dialog
	// Function to open folder dialog
	openFolder := func() {
		selectedFolder := openFolderWithExplorer()
		if selectedFolder != "" {
			folderEntry.SetText(selectedFolder)
		} else {
			dialog.ShowInformation("Error", "No folder selected.", w)
		}
	}

	// Radio button group for selecting platform
	platform := widget.NewRadioGroup([]string{"PC", "Switch"}, func(selected string) {
		fmt.Println("Selected Platform:", selected)
	})
	platform.SetSelected("PC") // Default selection
	platform.Horizontal = true // Ensure buttons are in a single row

	// Layout: "Platform" label + radio buttons
	platformContainer := container.NewCenter(
		container.NewHBox(
			widget.NewLabel("Platform:"),
			widget.NewLabel("   "), // Small gap
			platform,
		),
	)

	// Function to patch a file (overwrite only if it exists)
	patchFile := func() {
		folder := folderEntry.Text
		if folder == "" {
			dialog.ShowInformation("Error", "Please select a folder first.", w)
			return
		}

		// Determine output path and resource based on selected platform
		var outputPath string
		var resource []byte

		if platform.Selected == "Switch" {
			if !strings.Contains(folder, "atmosphere") {
				dialog.ShowError(fmt.Errorf("Selected folder is not from 'atmosphere'"), w)
				return
			}
			resource = patchResourceSwitch
			outputPath = filepath.Join(
				folder, "contents", "0100ED9018F3E000", "romfs", "Data", "StreamingAssets", "aa", "Switch", "localization-string-tables-english(en)_assets_all.bundle",
			)
		} else {
			resource = patchResource
			outputPath = filepath.Join(folder, "EiyudenChronicle_Data", "StreamingAssets", "aa", "StandaloneWindows64", "localization-string-tables-english(en)_assets_all.bundle")
		}

		if platform.Selected == "Switch" {
			if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
				dialog.ShowError(fmt.Errorf("Failed to create directories: %v", err), w)
				return
			}
		}

		// Ensure file exists before overwriting
		if _, err := os.Stat(outputPath); os.IsNotExist(err) && !(platform.Selected == "Switch") {
			dialog.ShowError(fmt.Errorf("File does not exist!"), w)
			return
		}

		// Write the patch file
		err := ioutil.WriteFile(outputPath, resource, 0644)
		if err != nil {
			dialog.ShowError(err, w)
		} else {
			dialog.ShowInformation("Success", fmt.Sprintf("File patched successfully!"), w)
		}
	}

	// Function to restore a file (always creates a new one)
	restoreFile := func() {
		folder := folderEntry.Text
		if folder == "" {
			dialog.ShowInformation("Error", "Please select a folder first.", w)
			return
		}

		// Determine output path and resource based on selected platform
		var outputPath string
		var resource []byte

		if platform.Selected == "Switch" {
			if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
				dialog.ShowError(fmt.Errorf("Failed to create directories: %v", err), w)
				return
			}

			resource = restoreResourceSwitch
			outputPath = filepath.Join(
				folder, "contents", "0100ED9018F3E000", "romfs", "Data", "StreamingAssets", "aa", "Switch", "localization-string-tables-english(en)_assets_all.bundle",
			)
		} else {
			resource = restoreResourcePc
			outputPath = filepath.Join(folder, "EiyudenChronicle_Data", "StreamingAssets", "aa", "StandaloneWindows64", "localization-string-tables-english(en)_assets_all.bundle")
		}

		if platform.Selected == "Switch" {
			if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
				dialog.ShowError(fmt.Errorf("Failed to create directories: %v", err), w)
				return
			}
		}

		// Ensure file exists before overwriting
		if _, err := os.Stat(outputPath); os.IsNotExist(err) && !(platform.Selected == "Switch") {
			fmt.Println("File does not exist: %s", outputPath)
			dialog.ShowError(fmt.Errorf("File does not exist!"), w)
			return
		}

		// Write the patch file
		err := ioutil.WriteFile(outputPath, resource, 0644)
		if err != nil {
			println("Error:", err)
			dialog.ShowError(err, w)
		} else {
			dialog.ShowInformation("Success", fmt.Sprintf("File patched successfully!"), w)
		}
	}

	// Buttons with better sizing
	browseBtn := widget.NewButtonWithIcon(" Browse", theme.FolderOpenIcon(), openFolder)
	patchBtn := widget.NewButtonWithIcon(" Patch File", theme.DocumentSaveIcon(), patchFile)
	restoreBtn := widget.NewButtonWithIcon(" Restore File", theme.HistoryIcon(), restoreFile)

	// Layout for buttons
	buttons := container.New(layout.NewGridLayout(2), patchBtn, restoreBtn)

	// UI Layout
	title := widget.NewLabelWithStyle(
		"Patch & Restore Tool",
		fyne.TextAlignCenter,       // Centered
		fyne.TextStyle{Bold: true}, // Bold
	)

	subtitle := widget.NewLabelWithStyle(
		"Select a folder and patch/restore files",
		fyne.TextAlignCenter, // Centered
		fyne.TextStyle{},     // Regular text
	)

	// Wrap title & subtitle in a centered container
	titleContainer := container.NewVBox(title, subtitle)

	// Create a card with centered content
	card := widget.NewCard(
		"", "", // Empty title/subtitle (we use our own labels)
		container.NewVBox(
			titleContainer, // Centered title & subtitle
			widget.NewSeparator(),
			container.NewPadded(),
			platformContainer, // Radio buttons now side by side
			widget.NewSeparator(),
			container.NewPadded(),
			folderEntry,
			widget.NewSeparator(),
			container.NewPadded(),
			browseBtn,
			container.NewPadded(),
			buttons,
		),
	)

	// Wrap the card with margin (100px top & bottom)
	w.SetContent(container.NewVBox(
		layout.NewSpacer(), layout.NewSpacer(), // 100px margin at the top
		container.NewPadded(card),              // Card with internal padding
		layout.NewSpacer(), layout.NewSpacer(), // 100px margin at the bottom
	))

	// Dark mode toggle
	darkModeEnabled := false
	darkModeToggle := func() {
		if darkModeEnabled {
			a.Settings().SetTheme(theme.LightTheme())
		} else {
			a.Settings().SetTheme(theme.DarkTheme())
		}
		darkModeEnabled = !darkModeEnabled
	}

	//Open about window
	showAbout := func() {
		showLicenseWindow(a)
	}

	// Open "Lapor Bug" link
	openBugReport := func() {
		bugURL, _ := url.Parse("https://docs.google.com/forms/d/e/1FAIpQLSelhsVi5gZqmXmLEIp01-wfBSc4XcdUG0XpqdZpy1RUAlTh8g/viewform")
		_ = a.OpenURL(bugURL)
	}

	// Open "Lapor Bug" link
	openUpdateFile := func() {
		bugURL, _ := url.Parse("https://your-bug-report-url.com")
		_ = a.OpenURL(bugURL)
	}

	// Create menu items
	aboutMenu := fyne.NewMenuItem("About", showAbout)
	darkModeMenu := fyne.NewMenuItem("Dark Mode", darkModeToggle)
	laporBugMenu := fyne.NewMenuItem("Lapor Bug", openBugReport)
	updateMenu := fyne.NewMenuItem("Check For Update", openUpdateFile)

	// Create menu bar
	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("Options", aboutMenu, darkModeMenu, laporBugMenu, updateMenu),
	)

	// Assign menu to window
	w.SetMainMenu(mainMenu)

	w.ShowAndRun()
}

// CustomTheme reduces font size globally (scaled-down by 25%)
type CustomTheme struct{}

func (c *CustomTheme) Font(s fyne.TextStyle) fyne.Resource {
	return theme.DefaultTextFont()
}

func (c *CustomTheme) Size(n fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(n) // Use default size without scaling issues
}

func (c *CustomTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(n, v)
}

func (c *CustomTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

// Function to show the license window
func showLicenseWindow(a fyne.App) {

	aboutText := `
Eiyuden Patch Tool v1.0.0

Presented: Koda Translation

Credit :
- Baslipro (Lead, Translator, Playtester)
- Sazanka (Translator, Playtester)
- Mr. C (Translator, Playtester, Programmer)
- tedo (Translator)
- Budidoom (Translator)
- Pangrangga (Translator)

Patch ini bebas biaya dan tidak dipungut biaya sama sekali. Pastikan untuk
mendapatkannya langsung dari sumbernya Koda Translation.

LICENSE:
This software is provided free of charge for personal and non-commercial use only.
Users are not allowed to modify, distribute, or sell this software or any derivative works.

The software is provided 'AS IS', without warranty of any kind, express or implied, 
including but not limited to the warranties of merchantability, fitness for a particular purpose, 
and noninfringement. In no event shall the author be liable for any claim, damages, 
or other liability, whether in an action of contract, tort, or otherwise, arising from, 
out of, or in connection with the software or the use or other dealings in the software.

By using this software, you agree to these terms.`
	licenseWin := a.NewWindow("License Agreement")
	licenseWin.Resize(fyne.NewSize(800, 600)) // Make the window big enough

	// Label with fixed width, monospaced font for better text formatting
	title := widget.NewLabelWithStyle("About", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	label := widget.NewLabelWithStyle(aboutText, fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})

	// Add close button
	closeBtn := widget.NewButton("Close", func() {
		licenseWin.Close()
	})

	// Layout
	licenseWin.SetContent(container.NewVBox(
		title,
		container.NewPadded(label), // Reduces padding
		closeBtn,
	))

	licenseWin.Show()
}

func listConnectedDevices() ([]string, error) {
	cmd := exec.Command("powershell", "-Command", "Get-PnpDevice | Where-Object { $_.InstanceId -like '*USB*' -or $_.InstanceId -like '*MTP*' } | Select-Object -ExpandProperty FriendlyName")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	devices := strings.Split(out.String(), "\r\n")
	return devices, nil
}

func listMTPDevices(callback func([]string)) {
	go func() {
		cmd := exec.Command("mtp-detect")
		output, err := cmd.CombinedOutput()
		if err != nil {
			callback([]string{"Error listing devices"})
			return
		}

		lines := strings.Split(string(output), "\n")
		var devices []string
		for _, line := range lines {
			if strings.Contains(line, "Device") {
				devices = append(devices, strings.TrimSpace(line))
			}
		}

		if len(devices) == 0 {
			devices = append(devices, "No devices found")
		}
		callback(devices)
	}()
}

func openFolderWithExplorer() string {
	cmd := exec.Command("powershell", "-NoProfile", "-Command", `
		Add-Type -AssemblyName System.Windows.Forms; 
		$folderBrowser = New-Object System.Windows.Forms.FolderBrowserDialog;
		$folderBrowser.Description = 'Select Folder';
		if ($folderBrowser.ShowDialog() -eq 'OK') { $folderBrowser.SelectedPath } else { exit 1 }
	`)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
