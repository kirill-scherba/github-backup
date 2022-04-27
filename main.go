// Copyright 2022 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Github-backup application save your github repository to local disk
//
// App use 'git' and 'gh' (github-cli) applications which shoud be preinstalled
// on the host. The 'git' should be configured to has access to your
// repositories by ssh. The 'gh' should be logged in to your github account
// before call this app.
//
// Application parameters:
//
//   -users  <[user-or-organisation-comma-separated-list]>
//   -limit  [user-repo-comma-separated-list]
//   -output [local-folder-name], default: ./repos
//   -printonly
//   -starsonly
//   -stars
//
// Usage examples:
//
//   go run . -users=kirill-scherba -limit=kirill-scherba/teonet-go -output=./tmp
//   go run . -users=kirill-scherba -stars -output=./tmp
//
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func main() {

	// Parse parameters
	var userslist, limitslist, output, maxrepo string
	var stars, starsonly, printonly bool
	//
	flag.StringVar(&userslist, "users", "", "user or organisation comma separated list")
	flag.StringVar(&limitslist, "limit", "", "user/repository comma separated list to backup, all if empty")
	flag.StringVar(&output, "output", "repos", "local folder name to save repositories")
	flag.BoolVar(&stars, "stars", false, "backup starred repositories also")
	flag.BoolVar(&starsonly, "starsonly", false, "backup starred repositories only")
	flag.StringVar(&maxrepo, "maxrepo", "1000", "maximum number of users repositories to be cloned")
	flag.BoolVar(&printonly, "printonly", false, "print repositories but does not clone it")
	flag.Parse()

	// Parse users and limit
	var limit []string
	users := strings.Split(userslist, ",")
	if len(strings.TrimSpace(limitslist)) != 0 {
		limit = strings.Split(limitslist, ",")
	}

	// Get list of repos with gh cli application
	var repos []string
	for _, user := range users {
		if !starsonly {
			r := getRepos(output, strings.TrimSpace(user), maxrepo, limit,
				printonly)
			repos = append(repos, r...)
		}
		if stars || starsonly {
			r := getStars(output, strings.TrimSpace(user), maxrepo, limit,
				printonly)
			repos = append(repos, r...)
		}
	}
}

// Number of repositories to show in print
var reponum int

// getRepos get list of reopsitories and clone it
func getRepos(dir, user, maxrepo string, limit []string,
	printonly bool) (repos []string) {

	// Get list of reopsitories with gh
	out, err := exec.Command("gh", "repo", "list", user, "-L", maxrepo).Output()
	if err != nil {
		log.Fatal(err)
	}

	// Parse gh ouput
	strs := strings.Split(string(out), "\n")
	for i := range strs {
		// Skip empty string
		if len(strs[i]) == 0 {
			continue // or break because the last line of 'out' is empty
		}

		// Get first column from 'gh repo list' output, it's repo name
		words := strings.Split(strs[i], "\t")
		repos = append(repos, words[0])
	}

	// Clone repos
	return cloneRepos(repos, limit, dir, printonly)
}

// getStars get list of starred reopsitories and clone it
func getStars(dir, user, maxrepo string, limit []string,
	printonly bool) (repos []string) {

	// Get list of starred reopsitories with gh by api
	// Loop through pages with 100 entries per page
	for p := 1; ; p++ {
		endpoint := fmt.Sprintf("/users/%s/starred?per_page=100&page=%d",
			user, p)
		out, err := exec.Command("gh", "api", endpoint).Output()
		if err != nil {
			log.Fatal(err)
		}

		// Umarshal github api output
		type starsData struct {
			FullName string `json:"full_name,omitempty"`
		}
		var jsonData []starsData
		if err := json.Unmarshal(out, &jsonData); err != nil {
			log.Printf("Can't parse response body to json: %s\n%s", err,
				string(out))
			return nil
		}

		// Exit form loop
		if len(jsonData) == 0 {
			break
		}

		// Parse github api output
		for i := range jsonData {
			repos = append(repos, jsonData[i].FullName)
		}
	}

	// Clone repos
	return cloneRepos(repos, limit, dir, printonly)
}

// cloneRepos from list of full repo name
func cloneRepos(repos []string, limit []string, dir string,
	printonly bool) (cloned []string) {

	for _, repo := range repos {
		// Get all repos if 'limit' slice is empty or get 'repo' exists in
		// 'limit' slice
		if !(len(limit) == 0 || inSlise(repo, limit)) {
			continue
		}

		// Print repo name
		reponum++
		fmt.Printf("repo %3d: %s\n", reponum, repo)

		// Skip clone if printonly flag set
		if printonly {
			continue
		}

		// Clone repo
		err := exec.Command("git", "clone", "--mirror", "git@github.com:"+repo+
			".git", dir+"/"+repo+".git").Run()
		if err != nil {
			log.Fatal(err)
		}
		cloned = append(cloned, repo)

		// Clone wiki repo
		err = exec.Command("git", "clone", "--mirror", "git@github.com:"+repo+
			".wiki.git", dir+"/"+repo+".wiki.git").Run()
		if err != nil {
			continue
		}
		cloned = append(cloned, repo+".wiki")
	}
	return
}

// inSlise return true if string 'el' exists in 'ar' string slice
func inSlise(el string, ar []string) bool {
	for i := range ar {
		if strings.TrimSpace(ar[i]) == el {
			return true
		}
	}
	return false
}
