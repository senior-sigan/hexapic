package hexapic

import (
	"appengine"
	"fmt"
	"github.com/go-martini/martini"
	"net/http"
)

func init() {
	m := martini.Classic()
	m.Use(AppEngine)
	imgContent := imgMimeType()

	m.Group("/api", func(r martini.Router) {
		r.Get("/tag/:name", imgContent, GetImageByTag)
		r.Get("/user/:name", imgContent, GetImageByName)
		r.Get("/question/:id", imgContent, GetImageByQuestionId)
		r.Get("/question", NewQuestion)
		r.Post("/question/:id", AnswerQuestion)
	})

	http.Handle("/", m)
}

func AppEngine(c martini.Context, r *http.Request) {
	c.Map(appengine.NewContext(r))
}

func imgMimeType() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "image/jpeg")
	}
}

func GetImageByTag(c appengine.Context, params martini.Params) {

}

func GetImageByName(c appengine.Context, params martini.Params) {

}

func GetImageByQuestionId(c appengine.Context, params martini.Params) string {
	return fmt.Sprintf("Image for %s", params["id"])
}

func NewQuestion(c appengine.Context, params martini.Params) string {
	return "New question"
}

func AnswerQuestion(c appengine.Context, params martini.Params) {

}
