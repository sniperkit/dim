package server

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/distribution/notifications"
	"github.com/mailgun/manners"
	"github.com/nhurel/dim/lib/index"
	"github.com/nhurel/dim/types"
	"net/http"
	"strings"
	"time"
)

// Server type handle  indexation of a docker registry and serves the search endpoint
type Server struct {
	*manners.GracefulServer
	index *index.Index
}

// NewServer creates a new Server instance to listen on given port and use given index
func NewServer(port string, index *index.Index) *Server {
	http.HandleFunc("/v1/search", handler(index, Search))
	http.HandleFunc("/dim/notify", handler(index, NotifyImageChange))
	http.HandleFunc("/v2/_catalog", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "{}")
	})
	return &Server{manners.NewWithServer(&http.Server{Addr: port, Handler: http.DefaultServeMux}), index}
}

// Run starts the server instance
func (s *Server) Run() error {
	return s.ListenAndServe()
}

// Handler injects an index into an HandlerFunc
func handler(i *index.Index, dhf DimHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dhf(i, w, r)
	}
}

// NotifyImageChange handles docker registry events
func NotifyImageChange(i *index.Index, w http.ResponseWriter, r *http.Request) {

	logrus.Infoln("Receiving event from registry")
	defer r.Body.Close()

	enveloppe := &notifications.Envelope{}

	if err := json.NewDecoder(r.Body).Decode(enveloppe); err != nil {
		logrus.WithError(err).Errorln("Failed to parse event")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Notiifcation handled !")

	logrus.WithField("enveloppe", enveloppe).Debugln("Processing event")

	for _, event := range enveloppe.Events {
		switch event.Action {
		case notifications.EventActionDelete:
			i.DeleteImage(string(event.Target.Digest))
		case notifications.EventActionPush:
			if event.Target.MediaType == schema2.MediaTypeManifest {
				if err := i.GetImageAndIndex(event.Target.Repository, event.Target.Tag, event.Target.Digest); err != nil {
					logrus.WithField("EventTarget", event.Target).WithError(err).Errorln("Failed to reindex image")
				}
			} else {
				logrus.WithField("mediatype", event.Target.MediaType).WithField("Event", event).Debugln("Event safely ignored because mediatype is unknown")
			}
		default:
			logrus.WithField("Action", event.Action).WithField("Event", event).Debugln("Event safely ignored")
		}
	}
}

// Search handles docker search request
func Search(i *index.Index, w http.ResponseWriter, r *http.Request) {

	var err error
	var b []byte

	q := r.FormValue("q")
	a := r.FormValue("a")
	f := r.FormValue("f")

	if q == "" && a == "" {
		http.Error(w, "No search criteria provided", http.StatusBadRequest)
		return
	}

	var sr *bleve.SearchResult

	request := bleve.NewSearchRequest(i.BuildQuery(q, a))
	request.Fields = []string{"Name", "Tag", "FullName", "Labels", "Envs"}
	l := logrus.WithField("request", request).WithField("query", request.Query)
	l.Debugln("Running search")
	if sr, err = i.Search(request); err != nil {
		http.Error(w, "An error occured while procesing your request", http.StatusInternalServerError)
		l.WithError(err).Errorln("Error occured when processing search")
		return
	}

	results := types.SearchResults{NumResults: int(sr.Total), Query: q}
	l.WithField("#results", results.NumResults).WithField("full", f).Debugln("Found results")

	if results.Results, err = buildResults(i, sr, (f == "full")); err != nil {
		http.Error(w, "Error happened while building response", http.StatusInternalServerError)
		return
	}

	if b, err = json.Marshal(results); err != nil {
		http.Error(w, "Failed to serialize the response", http.StatusInternalServerError)
		l.WithError(err).Errorln("Error occured while serializing search results")
	} else {
		fmt.Fprint(w, string(b))
	}
}

func buildResults(i *index.Index, sr *bleve.SearchResult, fillDetails bool) ([]types.SearchResult, error) {
	logrus.WithField("fillDetails", fillDetails).Debugln("Entering buildResult")
	images := make([]types.SearchResult, 0, sr.Total)
	var err error
	for _, h := range sr.Hits {
		doc := h
		if fillDetails {
			if doc, err = searchDetails(i, doc); err != nil {
				return images, err
			}
		}
		images = append(images, documentToSearchResult(doc))
	}
	return images, nil
}

func searchDetails(i *index.Index, doc *search.DocumentMatch) (*search.DocumentMatch, error) {
	logrus.WithField("doc", doc).Debugln("Entering searchDetails")
	request := bleve.NewSearchRequest(bleve.NewDocIDQuery([]string{doc.ID}))
	request.Fields = []string{"Name", "Tag", "FullName", "Volumes", "ExposedPorts", "Env", "Size"}
	if doc.Fields["Labels"] != nil {
		switch f := doc.Fields["Labels"].(type) {
		case string:
			request.Fields = append(request.Fields, fmt.Sprintf("Label.%s", f))
		case []interface{}:
			for _, f := range doc.Fields["Labels"].([]interface{}) {
				request.Fields = append(request.Fields, fmt.Sprintf("Label.%s", f))
			}
		}
	}
	if doc.Fields["Envs"] != nil {
		switch f := doc.Fields["Labels"].(type) {
		case string:
			request.Fields = append(request.Fields, fmt.Sprintf("Env.%s", f))
		case []interface{}:
			for _, f := range doc.Fields["Envs"].([]interface{}) {
				request.Fields = append(request.Fields, fmt.Sprintf("Env.%s", f))
			}
		}
	}

	var sr *bleve.SearchResult
	var err error
	if sr, err = i.Search(request); err != nil {
		return nil, fmt.Errorf("Failed to fetch all image info : %v", err)
	}

	return sr.Hits[0], err
}

func documentToSearchResult(h *search.DocumentMatch) types.SearchResult {
	logrus.WithField("hit", h).Debugln("Entering documentToSearchResult")
	result := types.SearchResult{
		Name:        h.Fields["Name"].(string),
		Description: h.Fields["Tag"].(string),
		Tag:         h.Fields["Tag"].(string),
		FullName:    h.Fields["FullName"].(string),
	}

	if h.Fields["Created"] != nil {
		if t, err := time.Parse(time.RFC3339, h.Fields["Created"].(string)); err == nil {
			result.Created = t
		} else {
			logrus.WithError(err).WithField("time", h.Fields["Created"].(string)).Errorln("Failed to parse time")
		}
	}

	labels := make(map[string]string, 10)
	envs := make(map[string]string, 10)
	for k, v := range h.Fields {
		if strings.HasPrefix(k, "Label.") {
			labels[strings.TrimPrefix(k, "Label.")] = v.(string)
		} else if strings.HasPrefix(k, "Env.") {
			envs[strings.TrimPrefix(k, "Env.")] = v.(string)
		}
	}

	if len(labels) > 0 {
		result.Label = labels
	}
	if h.Fields["Volumes"] != nil {
		result.Volumes = []string{h.Fields["Volumes"].(string)}
	}
	if h.Fields["ExposedPorts"] != nil {
		result.ExposedPorts = h.Fields["ExposedPorts"].([]int)
	}
	if len(envs) > 0 {
		result.Env = envs
	}
	if h.Fields["Size"] != nil {
		result.Size = h.Fields["Size"].(int64)
	}

	return result
}

// DimHandlerFunc injects index into a HandlerFunc function
type DimHandlerFunc func(i *index.Index, w http.ResponseWriter, r *http.Request)
