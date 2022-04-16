package main

import (
	"html/template"
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

func trimFilename(filename string) string {
	var results string
	for i := range filename {
		if filename[i] != ' ' && filename[i] != '\t' {
			results += string(filename[i])
		}
	}
	return results
}

func compressResults(session *Session) {
	if _, err := os.Stat(session.ExtractedPDFDirectory()); err != nil {
		return
	}
	cmd := exec.Command("zip", "-r", session.ZipTarget(), "extracted_pdfs"+"/")
	cmd.Dir = session.Workspace()
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalln("Fatal Error:", err)
	}
	log.Println(string(out))
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

type ErrorStruct struct {
	Message string
}

func doUploadImage(w http.ResponseWriter, r *http.Request) {
	session := NewSession(r.RemoteAddr)
	//defer session.Cleanup()
	r.ParseMultipartForm(10 << 20)
	formData := r.MultipartForm
	files := formData.File["file"]
	log.Println("doUploadImage()")
	for i, h := range files {
		log.Println(h.Filename)
		file, err := files[i].Open()
		if err != nil {
			log.Fatalln("Fatal Error opening input file", h.Filename, err)
		}
		defer file.Close()
		out, err := os.OpenFile(session.Workspace()+trimFilename(h.Filename), os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalln("Fatal Error creating output file:", err)
		}
		defer out.Close()
		_, err = io.Copy(out, file)
		if err != nil {
			log.Fatalln("Fatal Error copying to output", err)
		}
		log.Println("Successfully uploaded", out.Name())
	}
	log.Println("Finished doUploadImage")
	processPDFS(session)
	compressResults(session)
	if _, err := os.Stat(session.ZipTarget()); err != nil {
		file, err := template.ParseFiles(IndexFile)
		if err != nil {
			log.Fatalln("Fatal Error parsing", IndexFile, "from doUploadImages:", err)
		}
		err = file.Execute(w, ErrorStruct{"Error: no .pdf files uploaded. Please try again"})
		if err != nil {
			log.Fatalln("Fatal Error executing", IndexFile, "from doUploadImages", err)
		}
	} else {
		theFile, err := os.Open(session.ZipTarget())
		if err != nil {
			log.Fatalln("Fatal Error reading .zip", err)
		}
		defer theFile.Close()
		w.Header().Set("Content-Type", "application/octet-stream")
		cd := mime.FormatMediaType("attachment", map[string]string{"filename": "extracted_images.zip"})
		w.Header().Set("Content-Disposition", cd)
		http.ServeContent(w, r, session.ZipTarget(), time.Now(), theFile)
	}
}

func doIndex(w http.ResponseWriter, r *http.Request) {
	file, err := template.ParseFiles(IndexFile)
	if err != nil {
		log.Fatalln("Fatal Error:", IndexFile, "not found.", err)
	}
	file.Execute(w, nil)
}

func setupHandlers() {
	http.HandleFunc("/", doIndex)
	http.HandleFunc("/upload", doUploadImage)
}

func main() {
	setupHandlers()
	logFile, err := os.OpenFile("api.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln("Fatal Error creating api.log", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.Fatalln(Server.ListenAndServe())
}
