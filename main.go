package main

import (
	"context"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
)

type stats struct {
	pulled int
	cloned int
	empty  int
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	organizationURL := os.Getenv("ORGANIZATION_URL")
	personalAccessToken := os.Getenv("PAT")
	downloadPath := os.Getenv("DOWNLOAD_PATH")
	if downloadPath == "" {
		downloadPath = "."
	}

	existingRepos := findGitFolders(downloadPath)
	log.Println("Repos already on disk", len(existingRepos))

	connection := azuredevops.NewPatConnection(organizationURL, personalAccessToken)
	ctx := context.Background()

	gitClient, err := git.NewClient(ctx, connection)
	response, err := gitClient.GetRepositories(ctx, git.GetRepositoriesArgs{})
	if err != nil {
		log.Fatal(err)
	}

	s := stats{}
	for _, repo := range *response {
		log.Println("Cloning", *repo.Name, "...")
		outPath := filepath.Join(downloadPath, *repo.Project.Name, *repo.Name)
		existingRepos = remove(existingRepos, *repo.Project.Name+"/"+*repo.Name)

		// Replace user with PAT
		u, err := url.Parse(*repo.RemoteUrl)
		if err != nil {
			log.Fatal(err)
		}

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
		oldLocation := filepath.Join(downloadPath, f)
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

func findGitFolders(downloadPath string) []string {
	gitRepos := make([]string, 0, 45)
	err := filepath.Walk(downloadPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasPrefix(path, downloadPath+"/archive") {
				return filepath.SkipDir
			}

			if strings.HasSuffix(path, "/.git") {
				parent := filepath.Dir(path)
				folder := strings.Replace(parent, downloadPath+"/", "", 1)
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
