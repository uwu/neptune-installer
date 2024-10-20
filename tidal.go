package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

type TIDALInstance struct {
	// Path to "resources" directory
	Path    string
	Channel string
}

func GetInstance(channel string) (TIDALInstance, error) {
	channelString := "TIDAL"

	instance := TIDALInstance{
		Path:    "",
		Channel: channel,
	}

	// Generate channel strings (e.g TIDAL, TIDAL-Beta, TIDAL Beta)
	if channel != "Stable" {
		switch os := runtime.GOOS; os {
		case "darwin":
			channelString = channelString + " " + channel
		case "windows":
			channelString = channelString + "-" + channel // This is the only confirmed TIDAL Beta string formatting, could be the same on  macOS.
		default: // Linux and BSD are basically the same thing
			return instance, errors.New("instance doesn't exist") // non-stable channels do not exist on Linux
		}
	}

	switch OS := runtime.GOOS; OS {
	case "darwin":
		instance.Path = filepath.Join("/Applications", channelString+".app", "Contents", "Resources")
	case "windows":
		starterPath := filepath.Join(os.Getenv("localappdata"), channelString, "/")
		currentParsedVersion := 0

		filepath.Walk(starterPath, func(path string, _ fs.FileInfo, _ error) error {
			fileName := filepath.Base(path)
			if strings.HasPrefix(fileName, "app-") {
				if parsedVersion, err := strconv.Atoi(strings.Replace(fileName[4:], ".", "", -1)); err == nil {
					if parsedVersion > currentParsedVersion {
						currentParsedVersion = parsedVersion
						instance.Path = filepath.Join(path, "resources")
					}
				}
			}

			return nil
		})
	default: // Linux and BSD are *still* basically the same thing
		path := os.Getenv("PATH")
		channel := "tidal-hifi"

		for _, pathItem := range strings.Split(path, ":") {
			joinedPath := filepath.Join(pathItem, channel)
			if _, err := os.Stat(joinedPath); err == nil {
				possiblepath, _ := filepath.EvalSymlinks(joinedPath)
				if possiblepath != joinedPath {
					instance.Path = filepath.Join(possiblepath, "..", "resources")
				}
			}
		}
	}

	if _, err := os.Stat(instance.Path); err == nil {
		return instance, nil
	} else {
		return instance, errors.New("instance doesn't exist")
	}
}

func GetChannels() []TIDALInstance {
	possible := []string{"Stable", "Beta"}
	var channels []TIDALInstance

	for _, channel := range possible {
		c, err := GetInstance(channel)
		if err == nil {
			channels = append(channels, c)
		}
	}

	return channels
}

func NewTIDALInstance(path string) (*TIDALInstance, error) {
	instance := TIDALInstance{
		Path:    path,
		Channel: "Unknown",
	}

	if _, err := os.Stat(filepath.Join(instance.Path, "app.asar")); err == nil {
		return &instance, nil
	} else {
		return nil, errors.New("instance doesn't exist")
	}
}
