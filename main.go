package main

import (
	"image/color"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/sqweek/dialog"
)

func CopyFile(src string, dst string) error {
    srcFile, err := os.Open(src)
    if err != nil {
        return err
    }
    defer srcFile.Close()

    dstFile, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer dstFile.Close()

    _, err = io.Copy(dstFile, srcFile)
    return err
}

func CopyDir(src string, dst string) error {
    srcInfo, err := os.Stat(src)
    if err != nil {
        return err
    }

    // Handle directory permissions
    if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
        return err
    }

    entries, err := os.ReadDir(src)
    if err != nil {
        return err
    }

    for _, entry := range entries {
        srcPath := filepath.Join(src, entry.Name())
        dstPath := filepath.Join(dst, entry.Name())

        if entry.IsDir() {
            if err := CopyDir(srcPath, dstPath); err != nil {
                return err
            }
        } else {
            if err := CopyFile(srcPath, dstPath); err != nil {
                return err
            }
        }
    }
    return nil
}

func main() {
	user, err := user.Current()

	if err != nil {
		panic(err)
	}

	if user.Username != "root" && runtime.GOOS == "linux" {
		dialog.Message("Installing neptune requires root permissions. Please rerun neptune installer as root to continue.").Error()
		return
	}

	a := app.New()
	a.Settings().SetTheme(discordTheme{})
	w := a.NewWindow("neptune installer")
	w.SetFixedSize(true)

	selectedInstancePath := make([]string, 1)

	instances := make(map[string]string)
	var channels []string
	for _, instance := range GetChannels() {
		instances[instance.Channel] = instance.Path

		channels = append(channels, instance.Channel)
	}

	header := canvas.NewText("neptune installer", color.White)
	header.TextSize = 22
	header.Alignment = fyne.TextAlignCenter

	description := widget.NewLabel("Choose the version of TIDAL you'd like to install to, then click install.")

	installButton := widget.NewButton("Install", func() {})
	installButton.Disable()

	installedContainer := container.NewVBox()

	updateButton := widget.NewButton("Update", func() {})
	uninstallButton := widget.NewButton("Uninstall", func() {})

	installedContainer.Add(updateButton)
	installedContainer.Add(uninstallButton)

	installedContainer.Hide()

	showInstall := func() {
		installButton.Show()
		installButton.Enable()
		installedContainer.Hide()

		w.Resize(fyne.NewSize(0, 0))
	}

	showUninstall := func() {
		installButton.Hide()
		installButton.Disable()
		installedContainer.Show()

		w.Resize(fyne.NewSize(0, 0))
	}

	installButton.OnTapped = func() {
		installButton.Disable()
		neptuneZip, err := os.CreateTemp("", "neptune.zip")
		if err != nil {
			dialog.Message("Failed to create temp file!\n%s", err).Error()
			a.Quit()
		}

		tempDirectory := os.TempDir()

		resp, err := http.Get("https://github.com/uwu/neptune/archive/refs/heads/master.zip")
		if err != nil {
			dialog.Message("Failed to download neptune!\n%s", err).Error()
			a.Quit()
		}
		defer resp.Body.Close()

		if _, err := io.Copy(neptuneZip, resp.Body); err != nil {
			dialog.Message("Failed to copy neptune to temp folder!\n%s", err).Error()
			a.Quit()
		}

		err = Unzip(neptuneZip.Name(), tempDirectory)
		if err != nil {
			dialog.Message("Failed to unzip neptune!\n%s", err).Error()
			a.Quit()
		}

		// It's okay if this fails, the temp folder should ideally be cleared by the system at some point.
		os.Remove(neptuneZip.Name())

		neptuneDir := filepath.Join(tempDirectory, "neptune-master")
		injectorPath := filepath.Join(neptuneDir, "injector")

		if err := CopyDir(injectorPath, filepath.Join(selectedInstancePath[0], "app")); err != nil {
    dialog.Message("Failed to copy injector to app directory!\n%s", err).Error()
    a.Quit()
}

		// It's fine if this fails, it's in the temp dir.
		os.Remove(neptuneDir)

		appAsarPath := filepath.Join(selectedInstancePath[0], "app.asar")
		originalAsarPath := filepath.Join(selectedInstancePath[0], "original.asar")

		if _, err := os.Stat(originalAsarPath); err != nil {
			if err := os.Rename(appAsarPath, originalAsarPath); err != nil {
				dialog.Message("Failed to rename original asar!\n%s", err).Error()
				a.Quit()
			}
		}

		showUninstall()
	}

	updateButton.OnTapped = func() {
		updateButton.Disable()
		os.RemoveAll(filepath.Join(selectedInstancePath[0], "app"))
		installButton.OnTapped()
		dialog.Message("Update complete!").Title("neptune updated successfully!").Info()
		updateButton.Enable()
	}

	uninstallButton.OnTapped = func() {
		if err := os.RemoveAll(filepath.Join(selectedInstancePath[0], "app")); err != nil {
			dialog.Message("Failed to remove app directory!\n%s", err).Error()
			return
		}

		if err := os.Rename(filepath.Join(selectedInstancePath[0], "original.asar"), filepath.Join(selectedInstancePath[0], "app.asar")); err != nil {
			dialog.Message("Failed to rename original.asar to app.asar!\n%s", err).Error()
			return
		}

		showInstall()
	}

	w.SetContent(container.NewVBox(
		header,
		description,
		widget.NewSelect(channels, func(s string) {
			selectedInstancePath[0] = instances[s]

			if _, err := os.Stat(filepath.Join(selectedInstancePath[0], "original.asar")); os.IsNotExist(err) {
				showInstall()
			} else {
				showUninstall()
			}
		}),
		installButton,
		installedContainer,
	))

	w.ShowAndRun()
}
