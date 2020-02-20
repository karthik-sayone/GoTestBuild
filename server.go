package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sync"
)

const downloadPath = "./data"

var templates = template.Must(template.ParseFiles(
	"tmpl/home.html"))

func copyAndCapture(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			_, err := w.Write(d)
			if err != nil {
				return out, err
			}
		}
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
}

func buildHandler(w http.ResponseWriter, r *http.Request) {
	projectName := r.FormValue("projectName")
	bundleid := r.FormValue("bundleId")
	url := r.FormValue("url")
	commandString := "web2app_build_bash -n " + projectName + " -b " + bundleid + " -u " + url
	cmd := exec.Command("/bin/bash", "-c", commandString)

	var stdout, stderr []byte
	var errStdout, errStderr error
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	err := cmd.Start()
	if err != nil {
		fmt.Printf(err.Error())
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		stdout, errStderr = copyAndCapture(os.Stdout, stdoutIn)
		wg.Done()
	}()

	stderr, errStderr = copyAndCapture(os.Stderr, stderrIn)

	wg.Wait()

	err = cmd.Wait()
	if err != nil {
		fmt.Printf("cmd.Run() failed with %s", err)
	}
	if errStdout != nil || errStderr != nil {
		fmt.Printf("failed to capture stdout or stderr")
	}

	outStr, errStr := string(stdout), string(stderr)
	fmt.Printf("\nout:\n%s\nerr:\n%s\n", outStr, errStr)

	http.Redirect(w, r, "/download/"+projectName+"_debug.apk", http.StatusFound)
}

func homeRouteHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "home.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {

	fs := http.FileServer(http.Dir(downloadPath))

	http.Handle("/download/", http.StripPrefix("/download", fs))
	http.HandleFunc("/", homeRouteHandler)
	http.HandleFunc("/build/", buildHandler)
	http.ListenAndServe(":8080", nil)
}
