package core

import (
	"fmt"
	"github.com/blan4/hexapic/instagramFix"
	"github.com/carbocation/go-instagram/instagram"
	"image"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type SearchApi struct {
	client     *instagram.Client
	httpClient *http.Client
	Count      int
}

const Count int = 6

func randMedia(media []instagram.Media, count int) []instagram.Media {
	if len(media) < count {
		log.Fatalf("Not enough media")
	}

	res := make([]instagram.Media, count)
	rand.Seed(time.Now().UTC().UnixNano())
	list := rand.Perm(len(media))[0:count]
	for i, n := range list {
		res[i] = media[n]
	}

	return res
}

func getImages(media []instagram.Media, httpClient *http.Client) []image.Image {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	images := make([]image.Image, len(media))
	var wg sync.WaitGroup
	for i, m := range media[0:] {
		wg.Add(1)
		go func(i int, m instagram.Media) {
			defer wg.Done()
			log.Printf("Url: %v\n", m.Link)
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

func (self *SearchApi) SearchByName(userName string) []image.Image {
	fmt.Printf("Searching by username %s\n", userName)
	users, _, err := self.client.Users.Search(userName, nil)
	if err != nil {
		log.Fatalf("Can't find user with name %s\n", userName)
	}
	fmt.Printf("Found user %s", users[0].Username)
	params := &instagram.Parameters{Count: 100}
	media, _, err := self.client.Users.RecentMedia(users[0].ID, params)
	if err != nil {
		log.Fatalf("Can't load data from instagram: %v\n", err)
	}

	m := randMedia(media, self.Count)
	return getImages(m, self.httpClient)
}

func (self *SearchApi) SearchByTag(tag string) []image.Image {
	fmt.Printf("Searching by tag %s\n", tag)
	service := instagramFix.TagsService{Client: self.client}
	params := &instagram.Parameters{Count: 100}
	media, _, err := service.RecentMediaFix(tag, params)
	if err != nil {
		log.Fatalf("Can't load data from instagram: %v", err)
	}

	m := randMedia(media, self.Count)
	return getImages(m, self.httpClient)
}

func (self *SearchApi) SearchByLocation(lat float64, lng float64) []image.Image {
	fmt.Printf("Searching by location area [%s, %s]", lat, lng)
	opt := &instagram.Parameters{Count: 100, Lat: lat, Lng: lng}
	media, _, err := self.client.Media.Search(opt)
	if err != nil {
		log.Fatalf("Can't load data from instagram: %v\n", err)
	}

	m := randMedia(media, self.Count)
	return getImages(m, self.httpClient)
}

func NewSearchApi(clientId string, httpClient *http.Client) (s *SearchApi) {
	inst_client := instagram.NewClient(httpClient)
	inst_client.ClientID = clientId

	s = &SearchApi{httpClient: httpClient, client: inst_client, Count: Count}

	return
}
