package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"github.com/blan4/hexapic/wm"
	"github.com/google/go-github/github"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

const VERSION string = "\"0.4.0\""
const USERNAME string = "blan4"
const REPOSITORY string = "Hexapic"
const API_URL string = "http://hexapicserv.appspot.com"

func checkUpdate() {
	client := github.NewClient(nil)
	releases, _, err := client.Repositories.ListReleases(USERNAME, REPOSITORY, nil)
	if err != nil {
		log.Fatalf("Can't get releases info: %v", err)
	}

	if len(releases) == 0 {
		fmt.Printf("There is no releases for this program at github.com/%s/%s\n", USERNAME, REPOSITORY)
		return
	}

	latest_tag := github.Stringify(releases[0].TagName)
	latest_url := github.Stringify(releases[0].Assets[0].BrowserDownloadUrl)
	if VERSION == latest_tag {
		fmt.Println("There are no updates for you")
	} else {
		fmt.Printf("Download version %s at %s\n", latest_tag, latest_url)
	}
}

func getPicturesHome() (path string) {
	if usr, err := user.Current(); err != nil {
		log.Fatalf("Can't get user dir %s\n", err)
	} else {
		path = filepath.Join(usr.HomeDir, "Pictures", "hexapic")
	}

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0766); err != nil {
			log.Fatalf("Can't create dir %s: %v", path, err)
		}
	}

	return path
}

func getImagesFromFolder(path string) []image.Image {
	images := make([]image.Image, 0)

	walkFn := func(path string, fileInfo os.FileInfo, err error) error {
		if fileInfo != nil {
			filename := strings.ToLower(fileInfo.Name())
			if strings.HasSuffix(filename, ".jpg") {
				fmt.Printf(path)
				file, _ := os.Open(path)
				img, format, err := image.Decode(file)
				fmt.Println(format)
				if err != nil {
					log.Fatalf("Can't decode %s\n", err)
				}
				images = append(images, img)
			}
		}

		return nil
	}
	filepath.Walk(path, walkFn)
	return images
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

func loadImage(query string) image.Image {
	fmt.Println(query)
	httpClient := http.DefaultClient
	resp, err := httpClient.Get(query)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	img, format, err := image.Decode(resp.Body)
	if err != nil {
		log.Fatalf("Can't decode image of format %s: %v", format, err)
	}

	return img
}

func saveWallpaper(wallpaper image.Image) string {
	filename := filepath.Join(getPicturesHome(), randStr(20)+".jpg")

	toimg, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Can't create file %v\n", err)
	}
	defer toimg.Close()
	jpeg.Encode(toimg, wallpaper, &jpeg.Options{jpeg.DefaultQuality})

	fmt.Printf("Collage %s created\n", filename)

	return filename
}

var tag string
var userName string
var directory string
var isCheckUpdate bool
var location bool
var version bool

func init() {
	flag.StringVar(&userName, "u", "", "Make Hexapic from user's pictures. Searching by name")
	flag.StringVar(&tag, "t", "", "Make Hexapic from latest pictures by tag")
	flag.BoolVar(&location, "l", false, "Make Hexpic from nearest photos via geocoding")
	flag.BoolVar(&isCheckUpdate, "c", false, "Check for new Hexapic version")
	flag.BoolVar(&version, "v", false, "Current version")
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

	w := wm.BuildSetter()

	if location {
		wallpaper := loadImage(fmt.Sprintf("%s/%s", API_URL, "location"))
		canvasFileName := saveWallpaper(wallpaper)
		w.Set(canvasFileName)
		return
	}

	if len(tag) != 0 {
		wallpaper := loadImage(fmt.Sprintf("%s/?%s=%s", API_URL, "tag", tag))
		canvasFileName := saveWallpaper(wallpaper)
		w.Set(canvasFileName)
		return
	}

	if len(userName) != 0 {
		wallpaper := loadImage(fmt.Sprintf("%s/?%s=%s", API_URL, "user", userName))
		canvasFileName := saveWallpaper(wallpaper)
		w.Set(canvasFileName)
		return
	}

	flag.Usage()
}
