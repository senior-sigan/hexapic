package hexapic

import (
	"appengine"
	"appengine/urlfetch"
	"fmt"
	hexapic "github.com/blan4/hexapic/core"
	"github.com/go-martini/martini"
	_ "image"
	"image/jpeg"
	"log"
	"net/http"
)

const (
	CLIENT_ID string = "417c3ee8c9544530b83aa1c24de2abb3"
	//UUID_LENGTH int    = 15
)

var (
	imageQuality = jpeg.Options{Quality: jpeg.DefaultQuality}
	//imageBadQuality = jpeg.Options{Quality: 50}
	//index           []byte
	//tagList         = [...]string{"pomodoro", "birthday", "twins", "hospital", "sport", "pet", "duckface", "rainbow", "tatoo", "car", "champion", "makeup", "best_friends", "snowman", "pigeon", "beard", "sunglasses", "pool", "piano", "butterfly", "internet_explorer", "pigeoff", "steampunk", "bike", "military", "graffiti", "starwars", "delorean", "dwarffortress", "punkisnotdead", "cookies", "gdg", "devfest", "retrofuturism", "thecakeisalie", "selfie", "minimalism_world", "explorerussia", "latte", "marshmellow", "cupcake", "macarons"}
	//letters         = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func init() {
	m := martini.Classic()
	m.Use(AppEngine)

	m.Group("/api", func(r martini.Router) {
		r.Get("/tag/:id", GetImageByTag)
		r.Get("/user/:id", GetImageByName)
		r.Get("/question/:id", GetImageByQuestionId)
		r.Get("/question", GetNewQuestion)
		r.Post("/question/:id", AnswerQuestion)
	})

	http.Handle("/", m)
}

func AppEngine(c martini.Context, r *http.Request) {
	engine := appengine.NewContext(r)
	httpClient := urlfetch.Client(engine)
	api := hexapic.NewSearchApi(CLIENT_ID, httpClient)
	c.Map(engine)
	c.Map(api)
}

func GetImageByTag(w http.ResponseWriter, api *hexapic.SearchApi, params martini.Params) {
	log.Println(params["tag"])
	height := 2
	width := 3
	api.Count = height * width
	imgs := api.SearchByTag(params["id"])
	img := hexapic.GenerateCollage(imgs, height, width)
	w.Header().Set("Content-Type", "image/jpeg")
	if err := jpeg.Encode(w, img, &imageQuality); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func GetImageByName(w http.ResponseWriter, api *hexapic.SearchApi, params martini.Params) {
	height := 2
	width := 3
	api.Count = height * width
	imgs := api.SearchByName(params["id"])
	img := hexapic.GenerateCollage(imgs, height, width)
	w.Header().Set("Content-Type", "image/jpeg")
	if err := jpeg.Encode(w, img, &imageQuality); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func GetImageByQuestionId(c appengine.Context, params martini.Params) string {
	return fmt.Sprintf("Image for %s", params["id"])
}

func GetNewQuestion(c appengine.Context, params martini.Params) string {
	return "New question"
}

func AnswerQuestion(c appengine.Context, params martini.Params) {

}
