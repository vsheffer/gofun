package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/libgit2/git2go"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var repo *git.Repository
var sig *git.Signature
var repoDir string
var staticDir string

const (
	Success string = "success"
	Error   string = "error"
)

type CommitRequest struct {
	FileName string `json:"fileName"`
}

type FileListResponse struct {
	FileList []string `json:"fileList"`
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func saveSpecFileHandler(w http.ResponseWriter, r *http.Request) {
	fileBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Can't read bytes: %+v", err)
	}

	log.Println("fileBytes = %s", string(fileBytes))

	fileName := mux.Vars(r)["filename"]
	ioutil.WriteFile(repoDir+fileName, fileBytes, 0644)
	json.NewEncoder(w).Encode(Response{Status: Success, Message: "File " + fileName + " saved."})
}

func getRepoDirListingHandler(w http.ResponseWriter, r *http.Request) {
	// Return list of files.

	fileList, _ := ioutil.ReadDir(repoDir)

	// Get the number of files that will be returned...

	numFilesToReturn := 0
	for _, fileInfo := range fileList {
		if strings.Index(fileInfo.Name(), ".") == 0 || fileInfo.IsDir() {
			continue
		}

		numFilesToReturn += 1
	}

	fileListResponse := FileListResponse{FileList: make([]string, numFilesToReturn)}
	listIndex := 0
	for _, fileInfo := range fileList {
		if strings.Index(fileInfo.Name(), ".") == 0 || fileInfo.IsDir() {
			continue
		}
		fileListResponse.FileList[listIndex] = fileInfo.Name()
		listIndex += 1
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fileListResponse)
}

func getSpecFileHandler(w http.ResponseWriter, r *http.Request) {
	fileName := mux.Vars(r)["filename"]
	bytes, err := ioutil.ReadFile(repoDir + fileName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.Header().Set("Content-Type", "text/yaml")
		w.Write(bytes)
	}
}

func commitFileHandler(w http.ResponseWriter, r *http.Request) {
	index, err := repo.Index()
	if err != nil {
		log.Printf("Can't open index %+v", err)
		return
	}

	log.Printf("index = %+v", index)

	fileName := mux.Vars(r)["filename"]
	err = index.AddByPath(fileName)
	if err != nil {
		log.Printf("Can't AddByPath %+v", err)
	}

	treeId, err := index.WriteTree()
	if err != nil {
		log.Printf("WriteTree error: %+v", err)
	}
	tree, err := repo.LookupTree(treeId)
	if err != nil {
		log.Printf("LookupTree error: %+v", err)
	}

	msg := "Updated " + fileName + "."
	var commitErr error
	currentBranch, err := repo.Head()
	log.Printf("currentBranch = %+v", currentBranch)
	if currentBranch != nil {
		currentTip, _ := repo.LookupCommit(currentBranch.Target())
		_, commitErr = repo.CreateCommit("HEAD", sig, sig, msg, tree, currentTip)
	} else {
		_, commitErr = repo.CreateCommit("HEAD", sig, sig, msg, tree)
	}

	if commitErr != nil {
		log.Printf("Commit error: %+v", commitErr)
	}
	index.Write()
	fmt.Fprintf(w, "Filename = %s", fileName)
}

func main() {
	flag.StringVar(&repoDir, "repo-dir", "", "The directory where the Git repository will be saved.")
	flag.StringVar(&staticDir, "static-dir", "", "The directory containing static content to be served.")
	flag.Parse()
	if len(repoDir) == 0 {
		log.Fatalf("repo-dir is required.")
	}

	if len(staticDir) == 0 {
		staticDir = repoDir + "/static"
	}

	if !strings.HasSuffix(repoDir, "/") {
		repoDir = fmt.Sprintf("%s%s", repoDir, "/")
	}
	log.Printf("repoDir = %s, staticDir = %s", repoDir, staticDir)
	var err error
	repo, err = git.InitRepository(repoDir, false)
	if err != nil {
		log.Fatalf("Can't initialize repository %+v", err)
		return
	}
	repo.Head()
	repo.DefaultSignature()
	sig, _ = repo.DefaultSignature()
	log.Printf("sig = %+v", sig)
	sig.Name = "vince.sheffer"
	sig.Email = "vince.sheffer@bhnetwork.com"

	log.Printf("repos = %+v", repo)

	r := mux.NewRouter().StrictSlash(false)
	r.HandleFunc("/specfiles", getRepoDirListingHandler).Methods("GET")
	r.HandleFunc("/specfiles/{filename}", getSpecFileHandler).Methods("GET")
	r.HandleFunc("/specfiles/{filename}", saveSpecFileHandler).Methods("PUT")
	r.HandleFunc("/commitfile/{filename}", commitFileHandler).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(staticDir)))
	http.Handle("/", r)
	http.ListenAndServe(":8080", r)
}
