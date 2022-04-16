package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	if _, err := os.Stat(primaryWorkspace); err != nil {
		workspace = secondaryWorkspace
	}
	workspace = primaryWorkspace
}

const (
	primaryWorkspace   = "/dev/shm/"
	secondaryWorkspace = "/tmp/"
	sessionIDFstring   = "%016x"
)

var (
	workspace  string
	serializer sync.Mutex
)

type Session struct {
	id string
	ip string
}

func trimPortFromIP(ip string) string {
	if !strings.Contains(ip, ":") {
		return ip
	}
	splitIP := strings.Split(ip, ":")
	return splitIP[0]
}

func NewSession(ip string) *Session {
	serializer.Lock()
	defer serializer.Unlock()
	session := &Session{
		id: fmt.Sprintf(sessionIDFstring, rand.Uint64()),
		ip: trimPortFromIP(ip),
	}
	if err := os.Mkdir(session.Workspace(), 0777); err != nil {
		log.Fatalln("Fatal Error: could not create workspace", session.Workspace(), err)
	}
	return session
}

func (s *Session) Workspace() string {
	return workspace + s.id + "/"
}

func (s *Session) ExtractedPDFDirectory() string {
	return s.Workspace() + "extracted_pdfs" + "/"
}

func (s *Session) ZipTarget() string {
	return s.Workspace() + "extracted_pdfs.zip"
}

func (s *Session) Cleanup() {
	os.RemoveAll(s.Workspace())
}
