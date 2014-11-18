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
	Set()
}

type MateWallpaperSetter struct{}
type XFCE4WallpaperSetter struct{}
type Gnome3WallpaperSetter struct{}

func (g *Gnome3WallpaperSetter) Set() {
	fmt.Println("Hello form Gnome3")
}

func (m *MateWallpaperSetter) Set() {
	fmt.Println("Hello from Mate")
}

func (x *XFCE4WallpaperSetter) Set() {
	fmt.Println("Hello from XFCE4")
}

func Builder() (w WallpaperSetter) {
	w = WM[GetWMName()]
	return
}

var WM = map[string]WallpaperSetter{
	"Metacity (Marco)": new(MateWallpaperSetter),
	"Xfwm4":            new(XFCE4WallpaperSetter),
	"Gnome3":           new(Gnome3WallpaperSetter),
	"Gala":             new(Gnome3WallpaperSetter),
}

func GetWMName() string {
	var (
		out   []byte
		err   error
		index int
	)
	out, err = exec.Command("xprop", "-root", "_NET_SUPPORTING_WM_CHECK").Output()
	if err != nil {
		fmt.Println("Unable to open X session")
		log.Fatal(err)
	}
	index = strings.LastIndex(string(out), "#")
	window_id := string(out)[index+2:]

	out, err = exec.Command("xprop", "-id", window_id, "8s", "_NET_WM_NAME").Output()
	if err != nil {
		log.Fatal(err)
	}
	index = strings.LastIndex(string(out), "=")
	wm := string(out)[index+3 : len(out)-2]

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
	w := Builder()
	w.Set()
	client := instagram.NewClient(nil)
	client.ClientID = "417c3ee8c9544530b83aa1c24de2abb3"
	media, _, _ := client.Tags.RecentMedia("cat", nil)

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
}
