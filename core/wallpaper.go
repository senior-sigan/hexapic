package core

import (
	"errors"
	"fmt"
	"github.com/blan4/hexapic/instagramFix"
	"github.com/carbocation/go-instagram/instagram"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const CLIENT_ID string = "417c3ee8c9544530b83aa1c24de2abb3"

func randMedia(media []instagram.Media) []instagram.Media {
	if len(media) < 6 {
		log.Fatalf("Not enough media")
	}
	if len(media) == 6 {
		return media
	}

	res := make([]instagram.Media, 6)
	rand.Seed(time.Now().UTC().UnixNano())
	list := rand.Perm(len(media))[0:6]
	for i, n := range list {
		res[i] = media[n]
	}

	return res
}

func getImages(media []instagram.Media, httpClient *http.Client) []image.Image {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	images := make([]image.Image, 6)
	var wg sync.WaitGroup
	for i, m := range media[0:6] {
		wg.Add(1)
		go func(i int, m instagram.Media) {
			defer wg.Done()
			log.Printf("ID: %v, Type: %v, Url: %v\n", m.ID, m.Type, m.Images.StandardResolution.URL)
			resp, err := httpClient.Get(m.Images.StandardResolution.URL)
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			img, format, err := image.Decode(resp.Body)
			if err != nil {
				log.Fatalf("Can't decode image %s of format %s: %v", m.Images.StandardResolution.URL, format, err)
			}

			images[i] = img
			fmt.Print(".")
		}(i, m)
	}
	wg.Wait()
	fmt.Println()

	return images
}

func GenerateWallpaper(images []image.Image) image.Image {
	if len(images) < 6 {
		log.Fatalf("Need 6 pixs, founded %d", len(images))
	}

	canvas_image := image.NewRGBA(image.Rect(0, 0, 1920, 1280))
	fmt.Printf("Found %d pics", len(images))
	for index, img := range images[0:6] {
		x := 640 * (index % 3)
		y := 640 * (index % 2)
		draw.Draw(canvas_image, img.Bounds().Add(image.Pt(x, y)), img, image.ZP, draw.Src)
	}

	return image.Image(canvas_image)
}

func searchByName(userName string, httpClient *http.Client) []instagram.Media {
	fmt.Printf("Searching by username %s\n", userName)
	client := instagram.NewClient(httpClient)
	client.ClientID = CLIENT_ID
	users, _, err := client.Users.Search(userName, nil)
	if err != nil {
		log.Fatalf("Can't find user with name %s\n", userName)
	}
	fmt.Printf("Found user %s", users[0].Username)
	params := &instagram.Parameters{Count: 100}
	media, _, err := client.Users.RecentMedia(users[0].ID, params)
	if err != nil {
		log.Fatalf("Can't load data from instagram: %v\n", err)
	}

	return randMedia(media)
}

func searchByTag(tag string, httpClient *http.Client) []instagram.Media {
	fmt.Printf("Searching by tag %s\n", tag)
	c := instagram.NewClient(httpClient)
	c.ClientID = CLIENT_ID
	service := instagramFix.TagsService{Client: c}
	params := &instagram.Parameters{Count: 100}
	media, _, err := service.RecentMediaFix(tag, params)
	if err != nil {
		log.Fatalf("Can't load data from instagram: %v", err)
	}

	return randMedia(media)
}

func GetWallpaper(searchType string, value string, httpClient *http.Client) (image.Image, error) {
	value = url.QueryEscape(value)
	switch searchType {
	case "tag":
		return GenerateWallpaper(getImages(searchByTag(value, httpClient), httpClient)), nil
	case "user":
		return GenerateWallpaper(getImages(searchByName(value, httpClient), httpClient)), nil
	default:
		return image.Image(image.NewRGBA(image.Rect(0, 0, 1920, 1280))), errors.New("Wrong searchType %s. Accepted only tag and user\n")
	}
}
