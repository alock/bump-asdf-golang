package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/semver"
)

const (
	asdfGoRootComponent        = `<component name="GOROOT" url="file://$USER_HOME$/.asdf/installs/golang/`
	toolversionsFilename       = ".tool-versions"
	jetbrainsWorkspaceFilename = "workspace.xml"
)

//var wsFilesNames = []string{"ws.xml"}

var (
	version   = flag.String("v", "", "golang version to update the files with")
	updateAll = flag.Bool("a", false, "do not maintain minor versions, force update all")
	debugMode = flag.Bool("debug", false, "debug logs to help")
)

type fileToUpdate struct {
	filePathAndName      string // maybe make a file object
	currentGolangVersion string
}

func main() {
	flag.Parse()
	if !semver.IsValid(fmt.Sprintf("v%v", *version)) {
		log.Fatal("version flag is required and must be a valid semver")
	}
	log.Printf("version: %v, updateAll: %v\n", *version, *updateAll)
	workingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatalf("could not get working directory\n%v", err)
	}
	gopath := os.Getenv("GOPATH")
	var workspaceFiles []string
	var toolversionsFiles []string
	err = filepath.Walk(workingDirectory, func(path string, f os.FileInfo, err error) error {
		if gopath != "" && path == gopath && f.IsDir() {
			if *debugMode {
				log.Printf("skipping $GOPATH=%v and contents\n", path)
			}
			return filepath.SkipDir
		}
		if f.IsDir() && (f.Name() == ".git" || f.Name() == "vendor" || f.Name() == "node_modules") {
			if *debugMode {
				log.Printf("skipping %v and contents\n", path)
			}
			return filepath.SkipDir
		}
		if err != nil {
			log.Println(err)
			return nil
		}
		if f.Name() == toolversionsFilename {
			toolversionsFiles = append(toolversionsFiles, path)
		}
		if f.Name() == jetbrainsWorkspaceFilename && strings.Contains(path, ".idea") {
			workspaceFiles = append(workspaceFiles, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("failed to walk the path finding the files\n%v", err)
	}
	passedMajorMinor := semver.MajorMinor(fmt.Sprintf("v%v", *version))
	if len(workspaceFiles) > 0 {
		var wsFilesToUpdate []fileToUpdate
		for _, wsFileName := range workspaceFiles {
			workspaceFileContent, err := os.Open(wsFileName)
			if err != nil {
				log.Fatal(err)
			}
			// When rewrite is ready
			//fixedWsFile := strings.Replace(string(workspaceFileContent),
			//	".asdf/installs/golang/1.19.3/go",
			//	".asdf/installs/golang/1.19.4/go", 1)
			//os.WriteFile(wsFileName, []byte(fixedWsFile), 0644)
			scanner := bufio.NewScanner(workspaceFileContent)
			for scanner.Scan() {
				if strings.Contains(scanner.Text(), asdfGoRootComponent) {
					currentSemVer := strings.Split(scanner.Text(), "/")[6]
					majorMinor := semver.MajorMinor(fmt.Sprintf("v%v", currentSemVer))
					if *updateAll || passedMajorMinor == majorMinor {
						if *version != currentSemVer {
							wsFilesToUpdate = append(wsFilesToUpdate, fileToUpdate{
								filePathAndName:      wsFileName,
								currentGolangVersion: currentSemVer,
							})
						}
					}
				}
			}
		}
		fmt.Printf("%v files to update\n", len(wsFilesToUpdate))
		for _, fileToBump := range wsFilesToUpdate {
			fmt.Printf("%v: %v\n", fileToBump.filePathAndName, fileToBump.currentGolangVersion)
		}
	}
	//make smarter since it's almost the same process
	if len(toolversionsFiles) > 0 {
		var asdfFilesToUpdate []fileToUpdate
		for _, asdfFileName := range toolversionsFiles {
			asdfFileContent, err := os.Open(asdfFileName)
			if err != nil {
				log.Fatal(err)
			}
			scanner := bufio.NewScanner(asdfFileContent)
			for scanner.Scan() {
				if strings.Contains(scanner.Text(), "golang ") {
					currentSemVer := strings.Split(scanner.Text(), " ")[1]
					majorMinor := semver.MajorMinor(fmt.Sprintf("v%v", currentSemVer))
					if *updateAll || passedMajorMinor == majorMinor {
						if *version != currentSemVer {
							asdfFilesToUpdate = append(asdfFilesToUpdate, fileToUpdate{
								filePathAndName:      asdfFileName,
								currentGolangVersion: currentSemVer,
							})
						}
					}
				}
			}
		}
		fmt.Printf("%v files to update\n", len(asdfFilesToUpdate))
		for _, fileToBump := range asdfFilesToUpdate {
			fmt.Printf("%v: %v\n", fileToBump.filePathAndName, fileToBump.currentGolangVersion)
		}
	}
}
