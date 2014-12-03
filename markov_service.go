package main

import (
	"github.com/gofun/markov"
	"log"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
	"net/http"
	"strconv"
)

type MarkovService struct {
	chain *markov.Chain
}

type AddPhrasesRequest struct {
	Phrases []string  `json:"phrases"`
}

type GetPhrasesResponse struct {
	Phrases []string   `json:"phrases"`
}

func NewMarkovService() *MarkovService {
	return &MarkovService{markov.NewChain(100)}
}

func (ms MarkovService) Register() {
	ws := new(restful.WebService)
	ws.Path("/markov")

	ws.Route(ws.POST("/phrases").To(ms.addPhrase).
		// docs
		Doc("Add a phrase to the markov chain.").
		Reads(AddPhrasesRequest{}).
		Operation("addPhrase").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Returns(200, "OK", ms))

	ws.Route(ws.GET("/phrases").To(ms.getPhrase).
		// docs
		Doc("Get a randomly generated phrase.").
		Operation("getPhrase").
		Param(ws.QueryParameter("num-phrases", "Number of phrases to get.").DataType("int")).
		Produces(restful.MIME_JSON).
		Writes(GetPhrasesResponse{})) // on the response

	restful.Add(ws)
}

func (ms MarkovService) addPhrase(request *restful.Request, response *restful.Response) {
	phrases := &AddPhrasesRequest{}
	err := request.ReadEntity(&phrases)
	if err == nil {
		ms.chain.Build(phrases.Phrases)
	} else {
		log.Printf("error: %+v\n", err)
		response.WriteError(http.StatusInternalServerError, err)
	}
}

func (ms MarkovService) getPhrase(request *restful.Request, response *restful.Response) {
	res := &GetPhrasesResponse{Phrases: []string{""}}
	num, _ := strconv.Atoi(request.QueryParameter("num-phrases"))
	res.Phrases[0] = ms.chain.Generate(num)
	response.WriteEntity(&res)
}

func main() {
	ms := NewMarkovService()
	ms.Register()
	log.Printf("start listening on localhost:8080")
	config := swagger.Config{
		WebServices:    restful.RegisteredWebServices(), // you control what services are visible
		WebServicesUrl: "http://localhost:8080",
		ApiPath:        "/apidocs.json",

		// Optionally, specifiy where the UI is located
		SwaggerPath:     "/apidocs/",
		SwaggerFilePath: "/Users/vshef00/development/tools/swagger-ui/dist"}
	swagger.InstallSwaggerService(config)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
