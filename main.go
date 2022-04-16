package main

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Response struct {
	Message string
}

func NewResponse(message string) Response {
	return Response{Message: message}
}

const (
	Port      = ":8080"
	IndexFile = "./index.html"
)

var (
	Server = &http.Server{
		IdleTimeout: time.Second * 15,
		Addr:        Port,
	}
)

func compressResults(session *Session) {
	cmd := exec.Command("zip", "-r", session.ZipTarget(), "extracted_pdfs"+"/")
	cmd.Dir = session.Workspace()
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalln("Fatal Error:", err)
	}
	fmt.Println(string(out))
}

func processPDFS(session *Session) {
	files, err := os.ReadDir(session.Workspace())
	if err != nil {
		log.Fatalln("Fatal Error reading workspace directory", session.Workspace(), err)
	}
	for i := range files {
		filename := files[i].Name()
		filepath := session.Workspace() + filename
		if strings.HasSuffix(strings.ToLower(filename), "pdf") {
			// do processing
			cmd := exec.Command("python3", "converter.py", filepath, session.ExtractedPDFDirectory())
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Fatalln("Fatal Error from converter.py", string(output), err)
			}
		}
	}
}

func doUploadImage(w http.ResponseWriter, r *http.Request) {
	session := NewSession(r.RemoteAddr)
	defer session.Cleanup()
	r.ParseMultipartForm(10 << 20)
	formData := r.MultipartForm
	files := formData.File["file"]
	for i, h := range files {
		file, err := files[i].Open()
		if err != nil {
			log.Fatalln("Fatal Error opening input file", h.Filename, err)
		}
		defer file.Close()
		out, err := os.OpenFile(session.Workspace()+h.Filename, os.O_CREATE|os.O_WRONLY, 0644)
		//out, err := ioutil.TempFile(session.Workspace(), "file*"+path.Ext(h.Filename))
		if err != nil {
			log.Fatalln("Fatal Error creating output file:", err)
		}
		defer out.Close()
		_, err = io.Copy(out, file)
		if err != nil {
			log.Fatalln("Fatal Error copying to output", err)
		}
		fmt.Println("Successfully uploaded", out.Name())
	}
	fmt.Println("Finished doUploadImage")
	processPDFS(session)
	compressResults(session)
	theFile, err := os.Open(session.ZipTarget())
	if err != nil {
		log.Fatalln("Fatal Error reading .zip", err)
	}
	defer theFile.Close()
	//w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	cd := mime.FormatMediaType("attachment", map[string]string{"filename": "extracted_images.zip"})
	//w.Header().Set("Content-Disposition", "attachment; filename=extracted_images.zip")
	w.Header().Set("Content-Disposition", cd)
	http.ServeContent(w, r, session.ZipTarget(), time.Now(), theFile)
}

func doIndex(w http.ResponseWriter, r *http.Request) {
	//session := NewSession(r.RemoteAddr)
	//log.Println(session)
	file, err := os.ReadFile(IndexFile)
	if err != nil {
		log.Fatalln("Fatal Error:", IndexFile, "not found.", err)
	}
	w.Write(file)
}

func setupHandlers() {
	http.HandleFunc("/", doIndex)
	http.HandleFunc("/upload/image", doUploadImage)
}

func main() {
	setupHandlers()
	log.Fatalln(Server.ListenAndServe())
}
