package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/gorilla/mux"
	"github.com/libgit2/git2go"
	auth "github.com/vsheffer/go-http-auth"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

var repo *git.Repository
var sig *git.Signature
var repoDir string
var staticDir string
var corsAllowedHost string

const (
	Success string = "success"
	Error   string = "error"
)

type CommitRequest struct {
	FileName string `json:"fileName"`
}

type LogEntry struct {
	CommitterUsername string    `json:"committer"`
	CommittedTime     time.Time `json:"commitedTime"`
	Message           string    `json:"message"`
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
	log.Printf("fileName = %s", fileName)
	bytes, err := ioutil.ReadFile(repoDir + fileName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		acceptHeader := r.Header.Get("Accept")
		log.Printf("acceptsHeader = %s, %d", acceptHeader, strings.Index(acceptHeader, "application/json"))
		w.Header().Set("Content-Type", "application/yaml")
		if strings.Index(acceptHeader, "yaml") < 0 {
			bytes, _ = yaml.YAMLToJSON(bytes)
			w.Header().Set("Content-Type", "application/json")
		}
		w.Header().Set("Access-Control-Allow-Origin", corsAllowedHost)
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Write(bytes)
	}
}

func commitFileHandler(w http.ResponseWriter, r *http.Request) {
	commitMessage := r.Header.Get("Commit-Message")
	committer := r.Header.Get("Committer")

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

	var commitErr error
	currentBranch, err := repo.Head()
	log.Printf("currentBranch = %+v", currentBranch)
	sig.Name = committer
	sig.When = time.Now()
	sig.Email = committer
	if currentBranch != nil {
		currentTip, _ := repo.LookupCommit(currentBranch.Target())
		_, commitErr = repo.CreateCommit("HEAD", sig, sig, commitMessage, tree, currentTip)
	} else {
		_, commitErr = repo.CreateCommit("HEAD", sig, sig, commitMessage, tree)
	}

	if commitErr != nil {
		log.Printf("Commit error: %+v", commitErr)
	}
	index.Write()
	fmt.Fprintf(w, "Filename = %s", fileName)
}

func historyHandler(w http.ResponseWriter, r *http.Request) {
	fileName := mux.Vars(r)["filename"]
	log.Printf("filename = %s", fileName)
	revspec, err := repo.Revparse("HEAD^{tree}")
	if err != nil {
		log.Printf("revparse err = %+v", err)
	}
	head := revspec.From().Id()
	log.Printf("head = %+v", head)
	tree, err := repo.LookupTree(head)
	if err != nil {
		log.Printf("err = %+v", err)
	}
	entry := tree.EntryByName(fileName)
	log.Printf("entry = %+v", entry)
	walk, _ := repo.Walk()
	walk.Push(entry.Id)
	historyResponse := make([]LogEntry, 5)
	numWalked := 0
	walk.Iterate(func(commit *git.Commit) bool {
		if numWalked > len(historyResponse) {
			return false
		}
		log.Printf("oid = %+v", commit.Id())
		historyResponse[numWalked] = LogEntry{
			CommitterUsername: commit.Committer().Name,
			CommittedTime:     commit.Committer().When,
			Message:           commit.Message(),
		}

		numWalked += 1
		return false
	})

	json.NewEncoder(w).Encode(historyResponse)
}

func main() {
	var passwordFile string

	flag.StringVar(&repoDir, "repo-dir", "", "The directory where the Git repository will be saved.")
	flag.StringVar(&staticDir, "static-dir", "", "The directory containing static content to be served.")
	flag.StringVar(&corsAllowedHost, "cors-allowed-origin", "*", "The hostname of the allowed origin for cors support.  All hosts are allowed by default.")
	flag.StringVar(&passwordFile, "password-file", "htpasswd", "The path to the password file that should match the format of htpasswd.")
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

	secrets := auth.HtpasswdFileProvider(passwordFile)
	authenticator := auth.NewBasicAuthenticator("gitrest", secrets)

	r.HandleFunc("/specfiles", getRepoDirListingHandler).Methods("GET")
	r.HandleFunc("/specfiles/{filename}", getSpecFileHandler).Methods("GET")
	r.HandleFunc("/specfiles/{filename}", saveSpecFileHandler).Methods("PUT")
	r.HandleFunc("/commitfile/{filename}", commitFileHandler).Methods("POST")
	r.HandleFunc("/history/{filename}", historyHandler).Methods("GET")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(staticDir)))
	http.Handle("/", authenticator.Wrap(func(w http.ResponseWriter, ar *auth.AuthenticatedRequest) {
		w.Header().Add("X-Basic-Auth-Username", ar.Username)
		r.ServeHTTP(w, &ar.Request)
	}))
	s := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
