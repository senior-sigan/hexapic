package core

import (
	"fmt"
	"github.com/blan4/hexapic/instagramFix"
	"github.com/carbocation/go-instagram/instagram"
	"github.com/oleiade/lane"
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

func (self *SearchApi) randMedia(media []instagram.Media) []instagram.Media {
	res := make([]instagram.Media, len(media))
	rand.Seed(time.Now().UTC().UnixNano())
	list := rand.Perm(len(media))
	for i, n := range list {
		res[i] = media[n]
	}

	return res
}

func (self *SearchApi) getImages(orderedMedia []instagram.Media) []image.Image {
	if len(orderedMedia) < self.Count {
		log.Fatalf("Not enough media %v, expected %v", len(orderedMedia), self.Count)
	}
	var wg sync.WaitGroup
	var mediaQueue = lane.NewQueue()
	media := self.randMedia(orderedMedia)
	images := make([]image.Image, self.Count)
	for _, m := range media[0:] {
		mediaQueue.Enqueue(m)
	}

	for index := 0; index < self.Count; index++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			for {
				value := mediaQueue.Dequeue()
				if value == nil {
					log.Fatal("Not enough media")
				}
				m := value.(instagram.Media)
				log.Printf("Url: %v\tWidth: %v, Height: %v\n",
					m.Images.StandardResolution.URL,
					m.Images.StandardResolution.Width,
					m.Images.StandardResolution.Height)
				resp, err := self.httpClient.Get(m.Images.StandardResolution.URL)
				if err != nil {
					log.Printf("Can't get image %s: %v", m.Images.StandardResolution.URL, err)
					continue
				}
				defer resp.Body.Close()

				img, format, err := image.Decode(resp.Body)
				if err != nil {
					log.Printf("Can't decode image %s of format %s: %v", m.Images.StandardResolution.URL, format, err)
					continue
				}

				if IsSquare(img) {
					images[index] = img
					return
				}
				log.Print("skipped")
			}
		}(index)
	}

	wg.Wait()
	for _, i := range images {
		if i == nil {
			log.Fatalf("Not enough media expected %v", self.Count)
		}
	}
	return images
}

func (self *SearchApi) SearchByName(userName string) []image.Image {
	fmt.Printf("Searching by username %s\n", userName)
	users, _, err := self.client.Users.Search(userName, nil)
	if err != nil {
		log.Fatalf("Can't find user with name %s\n", userName)
	}
	var user *instagram.User
	for _, u := range users {
		if u.Username == userName {
			user = &u
			break
		}
	}
	if user == nil {
		log.Fatalf("Can't find user with username %v", userName)
	}
	fmt.Printf("Found user %s", user.Username)
	params := &instagram.Parameters{Count: 100}
	var media []instagram.Media
	for {
		nextMedia, page, err := self.client.Users.RecentMedia(user.ID, params)
		if err != nil {
			log.Fatalf("Can't load data from instagram: %v\n", err)
		}
		media = append(media, filter(nextMedia)...)
		if len(media) >= self.Count*2 {
			break
		}
		log.Printf("We get %v images. We need %v. So load more", len(media), self.Count*2)
		params.MaxID = page.NextMaxID
	}
	return self.getImages(media)
}

func filter(media []instagram.Media) []instagram.Media {
	var filtered []instagram.Media
	log.Print("Start media filter")
	for _, m := range media {
		if len(m.Tags) > 14 {
			log.Printf("Skipped but for tags %v", len(m.Tags))
			continue
		}
		if m.Type != "image" {
			log.Printf("Skipped: it's not image")
			continue
		}
		if len(m.UsersInPhoto) > 10 {
			log.Printf("Skipped but for users in photo %v", len(m.UsersInPhoto))
			continue
		}
		if m.Images.StandardResolution.Height < 300 || m.Images.StandardResolution.Width < 300 {
			log.Printf("Skipped but for low resolution %vx%v", m.Images.StandardResolution.Height, m.Images.StandardResolution.Width)
			continue
		}
		filtered = append(filtered, m)
	}
	return filtered
}

func (self *SearchApi) SearchByTag(tag string) []image.Image {
	// Instagram return 33 images, we actually need self.Count images. But some of them could be bad. 3 times more images could be okay
	count := self.Count * 3
	fmt.Printf("Searching by tag %s\n", tag)
	service := instagramFix.TagsService{Client: self.client}
	params := &instagram.Parameters{Count: 100}
	var media []instagram.Media
	for {
		nextMedia, page, err := service.RecentMediaFix(tag, params)
		if err != nil {
			log.Fatalf("Can't load data from instagram: %v", err)
		}
		media = append(media, filter(nextMedia)...)
		if len(media) >= count {
			break
		}
		log.Printf("We get %v images. We need %v. So load more", len(media), count)
		params.MaxID = page.NextMaxID
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
