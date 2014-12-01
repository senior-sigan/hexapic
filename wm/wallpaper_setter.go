package wm

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type WallpaperSetter interface {
	Set(path string)
}

type MateWallpaperSetter struct{}
type XFCE4WallpaperSetter struct{}
type Gnome3WallpaperSetter struct{}
type CinnamonWallpaperSetter struct{}
type KDE4WallpaperSetter struct{}

func (g *Gnome3WallpaperSetter) Set(path string) {
	fmt.Printf("Set desktop wallpaper %s\n", path)
	err := exec.Command("gsettings", "set", "org.gnome.desktop.background", "picture-uri", fmt.Sprintf("file:%s", path)).Run()
	if err != nil {
		fmt.Printf("Can't set desktop wallpaper %s: %s\n", path, err)
	}
}

func (m *MateWallpaperSetter) Set(path string) {
	fmt.Printf("Set desktop wallpaper %s\n", path)
	err := exec.Command("gsettings", "set", "org.mate.background", "picture-filename", path).Run()
	if err != nil {
		fmt.Printf("Can't set desktop wallpaper %s: %v", path, err)
	}
}

func (x *XFCE4WallpaperSetter) Set(path string) {
	fmt.Printf("Set desktop wallpaper %s\n", path)
	displays := GetDisplayNames()
	for _, display := range displays {
		err := exec.Command("xfconf-query", "-c", "xfce4-desktop", "-p", fmt.Sprintf("/backdrop/screen0/monitor%s/workspace0/last-image", display), "-s", path).Run()
		if err != nil {
			log.Fatalf("Can't set wallpaper on display %s: %s", display, err)
		}
	}
}

func (c *CinnamonWallpaperSetter) Set(path string) {
	log.Fatal("Not implemented for Cinnamon desktop manager")
}

func (k *KDE4WallpaperSetter) Set(path string) {
	log.Fatal("Not implemented for KDE desktop manager")
}

func BuildSetter() (w WallpaperSetter) {
	w = WM[GetWMName()]
	if w == nil {
		log.Fatalf("%s not supported. Ask maitainer\n", GetWMName())
	}
	return
}

var WM = map[string]WallpaperSetter{
	"Metacity (Marco)": new(MateWallpaperSetter),
	"Xfwm4":            new(XFCE4WallpaperSetter),
	"Gnome3":           new(Gnome3WallpaperSetter),
	"Gala":             new(Gnome3WallpaperSetter),
	"Kwin":             new(KDE4WallpaperSetter),
	"Mutter (Muffin)":  new(CinnamonWallpaperSetter),
	"Compiz":           new(Gnome3WallpaperSetter),
}

func GetWMName() string {
	var (
		out   []byte
		err   error
		index int
	)
	out, err = exec.Command("xprop", "-root", "_NET_SUPPORTING_WM_CHECK").Output()
	if err != nil {
		log.Fatalf("Unable to open X session: %v", err)
	}
	index = strings.LastIndex(string(out), "#")
	window_id := string(out)[index+2:]

	out, err = exec.Command("xprop", "-id", window_id, "8s", "_NET_WM_NAME").Output()
	if err != nil {
		log.Fatal(err)
	}
	index = strings.LastIndex(string(out), "=")
	wm := string(out)[index+3 : len(out)-2]

	fmt.Printf("Founded %s windows manager\n", wm)
	return wm
}

func GetDisplayNames() []string {
	displays := make([]string, 0)
	out, err := exec.Command("xrandr").Output()
	if err != nil {
		log.Fatalf("%s", err)
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		i := strings.LastIndex(line, " connected")
		if i != -1 {
			displays = append(displays, line[0:i])
		}
	}

	return displays
}
