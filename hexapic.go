package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"github.com/carbocation/go-instagram/instagram"
	"github.com/google/go-github/github"
	"image"
	"image/draw"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

const VERSION string = "\"0.0.1\""
const USERNAME string = "blan4"
const REPOSITORY string = "HexapicRB"
const CLIENT_ID string = "417c3ee8c9544530b83aa1c24de2abb3"

func checkUpdate() {
	client := github.NewClient(nil)
	releases, _, err := client.Repositories.ListReleases(USERNAME, REPOSITORY, nil)
	if err != nil {
		log.Fatalf("Can't get releases info: %v.", err)
	}

	if len(releases) == 0 {
		log.Println("There is no releases for this program at github.com/%s/%s.", USERNAME, REPOSITORY)
		return
	}

	latest_tag := github.Stringify(releases[0].TagName)
	latest_url := github.Stringify(releases[0].Assets[0].BrowserDownloadUrl)
	if VERSION == latest_tag {
		log.Println("There are no updates for you.")
	} else {
		log.Printf("Download version %s at %s.", latest_tag, latest_url)
	}
}

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
		log.Printf("Can't set desktop wallpaper %s: %v.", path, err)
	}
}

func (m *MateWallpaperSetter) Set(path string) {
	log.Printf("Set desktop wallpaper %s", path)
	err := exec.Command("gsettings", "set", "org.mate.background", "picture-filename", path).Start()
	if err != nil {
		log.Printf("Can't set desktop wallpaper %s: %v.", path, err)
	}
}

func (x *XFCE4WallpaperSetter) Set(path string) {
	log.Printf("Set desktop wallpaper %s", path)
	exec.Command("xfconf-query", "-c", "xfce4-desktop", "-p", "/backdrop/screen0/monitorLVDS1/workspace0/last-image", "-s", path)
	exec.Command("xfconf-query", "-c", "xfce4-desktop", "-p", "/backdrop/screen0/monitorLVDS1/workspace1/last-image", "-s", path)
}

func (c *CinnamonWallpaperSetter) Set(path string) {
	log.Fatal("Not implemented for Cinnamon desktop manager.")
}

func (k *KDE4WallpaperSetter) Set(path string) {
	log.Fatal("Not implemented for KDE desktop manager.")
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
		log.Fatalf("Unable to open X session: %v.", err)
	}
	index = strings.LastIndex(string(out), "#")
	window_id := string(out)[index+2:]

	out, err = exec.Command("xprop", "-id", window_id, "8s", "_NET_WM_NAME").Output()
	if err != nil {
		log.Fatal(err)
	}
	index = strings.LastIndex(string(out), "=")
	wm := string(out)[index+3 : len(out)-2]

	log.Printf("Founded %s windows manager.", wm)
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

func getImages(media []instagram.Media) []image.Image {
	images := make([]image.Image, 0)
	for _, m := range media {
		fmt.Printf("ID: %v, Type: %v, Url: %v\n", m.ID, m.Type, m.Images.StandardResolution.URL)
		resp, err := http.Get(m.Images.StandardResolution.URL)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		img, _, err := image.Decode(resp.Body)
		images = append(images, img)
		if err != nil {
			log.Fatalf("Can't decode image %s: %v", m.Images.StandardResolution.URL, err)
		}
	}

	return images
}

func getPicturesHome() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Can't get user dir %s.", err)
	}
	//TODO: create dir if not exists
	return filepath.Join(usr.HomeDir, "Pictures", "hexapic")
}

func generateWallpaper(images []image.Image) string {
	if len(images) < 6 {
		log.Fatalf("Need 6 pixs, founded %d", len(images))
	}

	canvas_filename := filepath.Join(getPicturesHome(), randStr(20)+".jpg")
	canvas_image := image.NewRGBA(image.Rect(0, 0, 1920, 1280))
	log.Printf("Found %d pics", len(images))
	for index, img := range images[0:6] {
		x := 640 * (index % 3)
		y := 640 * (index % 2)
		fmt.Printf("%v %v\n", x, y)
		draw.Draw(canvas_image, img.Bounds().Add(image.Pt(x, y)), img, image.ZP, draw.Src)
	}

	toimg, err := os.Create(canvas_filename)
	if err != nil {
		log.Fatalf("Can't create file %v", err)
	}
	defer toimg.Close()
	jpeg.Encode(toimg, canvas_image, &jpeg.Options{jpeg.DefaultQuality})

	log.Printf("Collage %s created", canvas_filename)

	return canvas_filename
}

func searchByName(userName string) []instagram.Media {
	log.Printf("Searching by username %s", userName)
	client := instagram.NewClient(nil)
	client.ClientID = CLIENT_ID
	users, _, err := client.Users.Search(userName, nil)
	if err != nil {
		log.Fatalf("Can't find user with name %s", userName)
	}
	log.Printf("Found user %s", users[0].Username)
	media, _, err := client.Users.RecentMedia(users[0].ID, nil)
	if err != nil {
		log.Fatalf("Can't load data from instagram: %v.", err)
	}

	return media[0:6]
}

func searchByTag(tag string) []instagram.Media {
	log.Printf("Searching by tag %s", tag)
	client := instagram.NewClient(nil)
	client.ClientID = CLIENT_ID
	media, _, err := client.Tags.RecentMedia(tag, nil)
	if err != nil {
		log.Fatalf("Can't load data from instagram: %v.", err)
	}

	return media[0:6]
}

func getImagesFromFolder(path string) []image.Image {
	images := make([]image.Image, 0)

	walkFn := func(path string, fileInfo os.FileInfo, err error) error {
		if fileInfo != nil {
			filename := strings.ToLower(fileInfo.Name())
			if strings.HasSuffix(filename, ".jpg") {
				log.Printf(path)
				file, _ := os.Open(path)
				img, format, err := image.Decode(file)
				log.Println(format)
				if err != nil {
					log.Fatalf("Can't decode %s", err)
				}
				images = append(images, img)
			}
		}

		return nil
	}
	filepath.Walk(path, walkFn)
	return images
}

var tag string
var userName string
var directory string
var isCheckUpdate bool
var version bool

func init() {
	flag.StringVar(&userName, "u", "", "Make Hexapic from user's pictures. Searching by name")
	flag.StringVar(&tag, "t", "", "Make Hexapic from latest pictures by tag")
	flag.BoolVar(&isCheckUpdate, "c", false, "Check for new Hexapic version")
	flag.BoolVar(&version, "v", false, "Current version")
	flag.StringVar(&directory, "d", "", "Debug mode. Path to folder with images.")
}

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Parse()
	if len(flag.Args()) != 0 {
		flag.Usage()
		os.Exit(1)
	}

	if isCheckUpdate {
		checkUpdate()
		return
	}

	if version {
		fmt.Println(VERSION)
		return
	}

	w := BuildSetter()

	if len(directory) != 0 {
		canvasFileName := generateWallpaper(getImagesFromFolder(directory))
		w.Set(canvasFileName)
		return
	}

	if len(tag) != 0 {
		canvasFileName := generateWallpaper(getImages(searchByTag(tag)))
		w.Set(canvasFileName)
		return
	}

	if len(userName) != 0 {
		canvasFileName := generateWallpaper(getImages(searchByName(userName)))
		w.Set(canvasFileName)
		return
	}

	flag.Usage()
}
