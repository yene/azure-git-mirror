package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v6"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v6/git"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v6/wiki"
)

type stats struct {
	pulled int
	cloned int
	empty  int
}

var gitCommit string    // Git sha, set by pipeline
var buildVersion string // human readable version number, set by pipeline

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	downloadWiki := flag.Bool("wiki", false, "download project wikis")
	flag.Parse()

	if len(flag.Args()) >= 1 && flag.Args()[0] == "version" {
		fmt.Println(buildVersion, gitCommit)
		os.Exit(0)
	}

	organizationURL := os.Getenv("ORGANIZATION_URL")
	personalAccessToken := os.Getenv("PAT")
	downloadPath := os.Getenv("DOWNLOAD_PATH")
	if downloadPath == "" {
		downloadPath = "."
	}

	if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
		log.Fatal("Given DOWNLOAD_PATH does not exist")
	}

	repoPath := filepath.Join(downloadPath, "repos")
	existingRepos := findGitFolders(repoPath)
	log.Println("Repos already on disk", len(existingRepos))

	connection := azuredevops.NewPatConnection(organizationURL, personalAccessToken)
	ctx := context.Background()

	gitClient, err := git.NewClient(ctx, connection)
	if err != nil {
		log.Fatal(err)
	}
	repoResponse, err := gitClient.GetRepositories(ctx, git.GetRepositoriesArgs{})
	if err != nil {
		log.Fatal(err)
	}

	if *downloadWiki {
		wikiClient, err := wiki.NewClient(ctx, connection)
		if err != nil {
			log.Fatal(err)
		}
		wikiResponse, err := wikiClient.GetAllWikis(ctx, wiki.GetAllWikisArgs{})
		if err != nil {
			log.Fatal(err)
		}

		for _, w := range *wikiResponse {
			if *w.Type == "projectWiki" {
				log.Println("Found Wiki", *w.Name)
				s := (*w.RepositoryId).String()
				resp, err := gitClient.GetRepository(ctx, git.GetRepositoryArgs{RepositoryId: &s})
				if err != nil {
					log.Fatal(err)
				}
				*repoResponse = append(*repoResponse, *resp)
			}
		}
	}

	s := stats{}
	for _, repo := range *repoResponse {
		log.Println("Cloning", *repo.Name, "...")
		outPath := filepath.Join(repoPath, *repo.Project.Name, *repo.Name)
		existingRepos = remove(existingRepos, *repo.Project.Name+"/"+*repo.Name)

		u, err := url.Parse(*repo.RemoteUrl)
		if err != nil {
			log.Fatal(err)
		}
		// Replace the username in the HTTPS git URL with the PAT token.
		url := strings.Replace(*repo.RemoteUrl, u.User.Username(), personalAccessToken, 1)
		cmd := exec.Command("git", "clone", url, outPath)
		err = cmd.Run()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				if exitError.ExitCode() == 128 {
					log.Println("	->", outPath, "exists")
					log.Println("Pulling", *repo.Name, "...")
					cmd := exec.Command("git", "-C", outPath, "pull", "--all", "--prune")
					err := cmd.Run()
					if err != nil {
						log.Println("Checking if", *repo.Name, "is empty...")
						cmd := exec.Command("git", "-C", outPath, "rev-parse", "HEAD")
						err := cmd.Run()
						if exitError, ok := err.(*exec.ExitError); ok {
							if exitError.ExitCode() == 128 {
								log.Println("	-> Repo is empty")
								s.empty = s.empty + 1
							} else {
								log.Fatal("git rev-parse failed", err)
							}
						}
					} else {
						s.pulled = s.pulled + 1
					}
				}
			}
		} else {
			s.cloned = s.cloned + 1
		}
	}

	foldersToArchive := existingRepos
	for _, f := range foldersToArchive {
		oldLocation := filepath.Join(repoPath, f)
		newLocation := filepath.Join(downloadPath, "archive", f)
		parentFolder := filepath.Dir(newLocation)
		os.MkdirAll(parentFolder, os.ModePerm)
		err := os.Rename(oldLocation, newLocation)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Println("Archived folders", foldersToArchive)
	log.Println(s.cloned, "cloned", s.pulled, "pulled", s.empty, "empty", len(foldersToArchive), "archived")

}

func findGitFolders(ppp string) []string {
	gitRepos := make([]string, 0, 45)
	err := filepath.Walk(ppp,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasPrefix(path, ppp+"/archive") {
				return filepath.SkipDir
			}

			if strings.HasSuffix(path, "/.git") {
				parent := filepath.Dir(path)
				folder := strings.Replace(parent, ppp+"/", "", 1)
				gitRepos = append(gitRepos, folder)
				return filepath.SkipDir
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}
	return gitRepos
}

func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
