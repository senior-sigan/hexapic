package main

import (
	"crypto/rand"
	"fmt"
	"github.com/carbocation/go-instagram/instagram"
	"image"
	"image/draw"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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
	log.Printf("Set desktop wallpaper %s", path)
	err := exec.Command("gsettings", "set", "org.gnome.desktop.background", "picture-uri", "file:///", path).Start()
	if err != nil {
		log.Printf("Can't set desktop wallpaper %s: %v", path, err)
	}
}

func (m *MateWallpaperSetter) Set(path string) {
	log.Printf("Set desktop wallpaper %s", path)
	err := exec.Command("gsettings", "set", "org.mate.background", "picture-filename", path).Start()
	if err != nil {
		log.Printf("Can't set desktop wallpaper %s: %v", path, err)
	}
}

func (x *XFCE4WallpaperSetter) Set(path string) {
	log.Printf("Set desktop wallpaper %s", path)
	exec.Command("xfconf-query", "-c", "xfce4-desktop", "-p", "/backdrop/screen0/monitorLVDS1/workspace0/last-image", "-s", path)
	exec.Command("xfconf-query", "-c", "xfce4-desktop", "-p", "/backdrop/screen0/monitorLVDS1/workspace1/last-image", "-s", path)
}

func (c *CinnamonWallpaperSetter) Set(path string) {
	log.Fatal("Not implemented for Cinnamon desktop manager")
}

func (k *KDE4WallpaperSetter) Set(path string) {
	log.Fatal("Not implemented for KDE desktop manager")
}

func BuildSetter() (w WallpaperSetter) {
	w = WM[GetWMName()]
	return
}

var WM = map[string]WallpaperSetter{
	"Metacity (Marco)": new(MateWallpaperSetter),
	"Xfwm4":            new(XFCE4WallpaperSetter),
	"Gnome3":           new(Gnome3WallpaperSetter),
	"Gala":             new(Gnome3WallpaperSetter),
	"Kwin":             new(KDE4WallpaperSetter),
	"Mutter (Muffin)":  new(CinnamonWallpaperSetter),
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

	log.Printf("Founded %s windows manager", wm)
	return wm
}

func randStr(str_size int) string {
	alphanum := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, str_size)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

func main() {
	client := instagram.NewClient(nil)
	client.ClientID = "417c3ee8c9544530b83aa1c24de2abb3"
	media, _, err := client.Tags.RecentMedia("cat", nil)
	if err != nil {
		log.Fatalf("Can't load data from instagram: %v", err)
	}

	canvas_filename := filepath.Join("/home/ilya/Pictures/go", randStr(20)+".jpg")
	canvas_image := image.NewRGBA(image.Rect(0, 0, 1920, 1280))

	for index, m := range media[0:6] {
		fmt.Printf("ID: %v, Type: %v, Url: %v\n", m.ID, m.Type, m.Images.StandardResolution.URL)
		resp, err := http.Get(m.Images.StandardResolution.URL)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		img, _, err := image.Decode(resp.Body)
		if err != nil {
			log.Fatalf("Can't decode image %s: %v", m.Images.StandardResolution.URL, err)
		} else {
			x := 640 * (index % 3)
			y := 640 * (index % 2)
			fmt.Printf("%v %v\n", x, y)
			draw.Draw(canvas_image, img.Bounds().Add(image.Pt(x, y)), img, image.ZP, draw.Src)
		}
	}
	toimg, err := os.Create(canvas_filename)
	if err != nil {
		log.Fatalf("Can't create file %v", err)
	}
	defer toimg.Close()
	jpeg.Encode(toimg, canvas_image, &jpeg.Options{jpeg.DefaultQuality})

	log.Printf("Collage %s created", canvas_filename)
	w := BuildSetter()
	w.Set(canvas_filename)
}
