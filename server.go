package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os/exec"
)

var templates = template.Must(template.ParseFiles(
	"tmpl/home.html"))

func buildHandler(w http.ResponseWriter, r *http.Request) {
	projectName := r.FormValue("projectName")
	bundleid := r.FormValue("bundleId")
	command := "bash_build -n " + projectName + " -b " + bundleid
	err := exec.Command("/bin/bash", "-c", command).Run()
	if err != nil {
		fmt.Printf(err.Error())
	}
}

func homeRouteHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "home.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/", homeRouteHandler)
	http.HandleFunc("/build/", buildHandler)
	http.ListenAndServe(":8080", nil)
}
