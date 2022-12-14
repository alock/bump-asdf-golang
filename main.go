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

const asdfGoRootComponent = `<component name="GOROOT" url="file://$USER_HOME$/.asdf/installs/golang/`

//var wsFilesNames = []string{"ws.xml"}

var (
	version   = flag.String("version", "", "new golang version to update")
	updateAll = flag.Bool("update-all", false, "do not maintain minor versions")
)

type fileToUpdate struct {
	filePathAndName      string // maybe mage a file object
	currentGolangVersion string
}

func main() {
	flag.Parse()
	if !semver.IsValid(fmt.Sprintf("v%v", *version)) {
		log.Fatal("version flag is required and must be a valid semver")
	}
	fmt.Printf("version: %v, updateAll: %v\n", *version, *updateAll)
	workingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatalf("could not get working directory\n%v", err)
	}
	var workspaceFiles []string
	var toolversionFiles []string
	err = filepath.Walk(workingDirectory, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() && (f.Name() == ".git" || f.Name() == "vendor" || f.Name() == "node_modules") {
			log.Printf("skipping %v and contents\n", path)
			return filepath.SkipDir
		}
		if f.IsDir() && f.Name() == "mod" {
			if strings.Contains(path, "pkg") {
				log.Printf("skipping %v and contents\n", path)
				return filepath.SkipDir
			}
		}
		if err != nil {
			fmt.Println(err)
			return nil
		}
		if f.Name() == ".tool-versions" {
			toolversionFiles = append(toolversionFiles, path)
		}
		if f.Name() == "workspace.xml" && strings.Contains(path, ".idea") {
			workspaceFiles = append(workspaceFiles, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("failed to walk the path finding the files\n%v", err)
	}
	if len(workspaceFiles) > 0 {
		passedMajorMinor := semver.MajorMinor(fmt.Sprintf("v%v", *version))
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
						wsFilesToUpdate = append(wsFilesToUpdate, fileToUpdate{
							filePathAndName:      wsFileName,
							currentGolangVersion: currentSemVer,
						})
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
	//if len(toolversionFiles) > 0 {
	//	fmt.Println("tool-versions files:")
	//	fmt.Println(strings.Join(toolversionFiles, "\n"))
	//}
}
