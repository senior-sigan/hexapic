package core

import (
	"fmt"
	"github.com/blan4/hexapic/instagramFix"
	"github.com/carbocation/go-instagram/instagram"
	"image"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type SearchApi struct {
	client     *instagram.Client
	httpClient *http.Client
	Count      int
}

const Count int = 6

func (self *SearchApi) randMedia(media []instagram.Media) []instagram.Media {
	if len(media) < self.Count {
		log.Fatalf("Not enough media")
	}

	res := make([]instagram.Media, len(media))
	rand.Seed(time.Now().UTC().UnixNano())
	list := rand.Perm(len(media))
	for i, n := range list {
		res[i] = media[n]
	}

	return res
}

func (self *SearchApi) getImages(orderedMedia []instagram.Media) []image.Image {
	media := self.randMedia(orderedMedia)
	images := make([]image.Image, self.Count)
	needToLoad := self.Count

	i := 0
	for needToLoad > 0 && i < len(media) {
		m := media[i]
		log.Printf("Url: %v\n", m.Images.StandardResolution.URL)
		resp, err := self.httpClient.Get(m.Images.StandardResolution.URL)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		img, format, err := image.Decode(resp.Body)
		if err != nil {
			log.Fatalf("Can't decode image %s of format %s: %v", m.Images.StandardResolution.URL, format, err)
		}

		if IsSquare(img) {
			images[needToLoad-1] = img
			needToLoad--
		} else {
			log.Print("skipped")
		}
		i++
		log.Println()
	}
	fmt.Println()

	if needToLoad > 0 {
		log.Fatalf("Not enough media: can't find good pictures")
	}

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

	return self.getImages(media)
}

func (self *SearchApi) SearchByTag(tag string) []image.Image {
	fmt.Printf("Searching by tag %s\n", tag)
	service := instagramFix.TagsService{Client: self.client}
	params := &instagram.Parameters{Count: 100}
	media, _, err := service.RecentMediaFix(tag, params)
	if err != nil {
		log.Fatalf("Can't load data from instagram: %v", err)
	}

	return self.getImages(media)
}

func (self *SearchApi) SearchByLocation(lat float64, lng float64) []image.Image {
	fmt.Printf("Searching by location area [%s, %s]", lat, lng)
	opt := &instagram.Parameters{Count: 100, Lat: lat, Lng: lng}
	media, _, err := self.client.Media.Search(opt)
	if err != nil {
		log.Fatalf("Can't load data from instagram: %v\n", err)
	}

	return self.getImages(media)
}

func NewSearchApi(clientId string, httpClient *http.Client) (s *SearchApi) {
	inst_client := instagram.NewClient(httpClient)
	inst_client.ClientID = clientId

	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	s = &SearchApi{httpClient: httpClient, client: inst_client, Count: Count}

	return
}
