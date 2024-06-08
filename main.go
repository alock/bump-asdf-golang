package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/mod/semver"
)

const (
	//asdfGoInstallPath          = `.asdf/installs/golang/`
	//miseAsdfGoSymlink          = `.asdf/installs/go/`
	miseGoInstallPath          = `.local/share/mise/installs/go/`
	jetbrainsIdeaMarker        = `<component name="GOROOT" url="file://$USER_HOME$/`
	golangSpace                = "golang "
	toolversionsFilename       = ".tool-versions"
	jetbrainsWorkspaceFilename = "workspace.xml"
)

var (
	version                    = flag.String("v", "", "golang version to update the files with")
	updateAll                  = flag.Bool("all", false, "do not maintain minor versions, force update all")
	minorBump                  = flag.Bool("minor", false, "flag to use when moving minor versions like 1.19.X to 1.20")
	debugMode                  = flag.Bool("debug", false, "debug logs to help")
	ideaWorkspacePlusGoInstall = fmt.Sprintf("%s%s", jetbrainsIdeaMarker, miseGoInstallPath)
)

type fileInfo struct {
	filePathAndName      string // maybe make a file object
	currentGolangVersion string
}

func main() {
	flag.Parse()
	if !semver.IsValid(fmt.Sprintf("v%v", *version)) {
		log.Fatalln("version flag is required and must be a valid semver")
	}
	if *updateAll && *minorBump {
		log.Fatalln("both flags \"minor\" and \"all\" cannot be set, please use one only")
	}
	log.Printf("version: %v, updateAll: %v, minorBump: %v\n", *version, *updateAll, *minorBump)
	workspaceFiles, toolversionsFiles := findPotentialFilesToUpdate()
	handlePotentialFilesToUpdate(workspaceFiles)
	handlePotentialFilesToUpdate(toolversionsFiles)
}

func findPotentialFilesToUpdate() (workspaceFiles, toolversionsFiles []fileInfo) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatalf("could not get working directory\n%v", err)
	}
	usersHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("could not get users home directory\n%v", err)
	}
	gopath := os.Getenv("GOPATH")
	err = filepath.Walk(workingDirectory, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			log.Println(err)
			return nil
		}
		// do not process gopath files
		if gopath != "" && path == gopath && f.IsDir() {
			if *debugMode {
				log.Printf("skipping $GOPATH=%v and contents\n", path)
			}
			return filepath.SkipDir
		}
		// skip ~/Library
		if path == fmt.Sprintf("%s/Library", usersHomeDir) && f.IsDir() {
			if *debugMode {
				log.Printf("skipping %v and contents\n", path)
			}
			return filepath.SkipDir
		}
		// skip dev folders
		if f.IsDir() && (f.Name() == ".git" || f.Name() == "vendor" || f.Name() == "node_modules") {
			if *debugMode {
				log.Printf("skipping %v and contents\n", path)
			}
			return filepath.SkipDir
		}
		if f.Name() == toolversionsFilename {
			if path == fmt.Sprintf("%s/%s", usersHomeDir, toolversionsFilename) {
				if *debugMode {
					log.Printf("skipping global tools-version file: %s\n", path)
				}
				return nil
			}
			currentVersion := getCurrentVersion(path)
			toolversionsFiles = append(toolversionsFiles, fileInfo{
				filePathAndName:      path,
				currentGolangVersion: currentVersion,
			})
		}
		if f.Name() == jetbrainsWorkspaceFilename && strings.Contains(path, ".idea") {
			currentVersion := getCurrentVersion(path)
			workspaceFiles = append(workspaceFiles, fileInfo{
				filePathAndName:      path,
				currentGolangVersion: currentVersion,
			})
		}
		return nil
	})
	if err != nil {
		log.Fatalf("failed to walk the path finding the files\n%v", err)
	}
	return
}

func getCurrentVersion(filePathAndName string) string {
	fileContent, err := os.Open(filePathAndName)
	if err != nil {
		log.Fatalln(err)
	}
	scanner := bufio.NewScanner(fileContent)
	for scanner.Scan() {
		if strings.Contains(filePathAndName, jetbrainsWorkspaceFilename) {
			if strings.Contains(scanner.Text(), ideaWorkspacePlusGoInstall) {
				// calculate the 8 from split miseGoInstallPath by / + 3
				// file://$USER_HOME$/.local/share/mise/installs/go/1.22.4/go
				return strings.Split(scanner.Text(), "/")[8]
			}
		}
		if strings.Contains(filePathAndName, toolversionsFilename) {
			if strings.Contains(scanner.Text(), golangSpace) {
				return strings.Split(scanner.Text(), " ")[1]
			}
		}
	}
	return ""
}

func getFilesToUpdate(version string, allFilesFound []fileInfo) (filesToUpdate []fileInfo) {
	passedMajorMinor := semver.MajorMinor(fmt.Sprintf("v%v", version))
	versionMinorMinusOne := ""
	if *minorBump {
		splitMajorMinor := strings.Split(passedMajorMinor, ".")
		minor, err := strconv.Atoi(splitMajorMinor[1])
		if err != nil {
			log.Fatalf("cannot calcuate minor version")
		}
		versionMinorMinusOne = semver.MajorMinor(fmt.Sprintf("%s.%v", splitMajorMinor[0], minor-1))
	}
	for _, f := range allFilesFound {
		majorMinor := semver.MajorMinor(fmt.Sprintf("v%v", f.currentGolangVersion))
		if (*updateAll && f.currentGolangVersion != "") || passedMajorMinor == majorMinor || (*minorBump && majorMinor == versionMinorMinusOne) {
			if version != f.currentGolangVersion {
				filesToUpdate = append(filesToUpdate, f)
			}
		}
	}
	return
}

func handlePotentialFilesToUpdate(potentialFilesToUpdate []fileInfo) {
	filesToUpdate := getFilesToUpdate(*version, potentialFilesToUpdate)
	printFilesAndCurrentVersion(filesToUpdate)
	fmt.Printf("%v file(s) to update\n", len(filesToUpdate))
	if len(filesToUpdate) > 0 {
		if yesNo(fmt.Sprintf("Do you want to update the files above to use golang %s?", *version)) {
			for _, fileToBump := range filesToUpdate {
				rewriteFile(fileToBump)
			}
		}
	}
}

func printFilesAndCurrentVersion(files []fileInfo) {
	for _, f := range files {
		fmt.Printf("%v: %v\n", f.filePathAndName, f.currentGolangVersion)
	}
}

func rewriteFile(file fileInfo) {
	fileContents, err := os.ReadFile(file.filePathAndName)
	if err != nil {
		log.Fatalln(err)
	}
	var updatedFileContents string
	if strings.Contains(file.filePathAndName, jetbrainsWorkspaceFilename) {
		updatedFileContents = strings.Replace(string(fileContents),
			fmt.Sprintf("%s%s/go", ideaWorkspacePlusGoInstall, file.currentGolangVersion),
			fmt.Sprintf("%s%s/go", ideaWorkspacePlusGoInstall, *version), 1)
	}
	if strings.Contains(file.filePathAndName, toolversionsFilename) {
		updatedFileContents = strings.Replace(string(fileContents),
			fmt.Sprintf("%s%s", golangSpace, file.currentGolangVersion),
			fmt.Sprintf("%s%s", golangSpace, *version), 1)
	}
	err = os.WriteFile(file.filePathAndName, []byte(updatedFileContents), 0644)
	if err != nil {
		log.Fatalln(err)
	}
}

func yesNo(s string) bool {
	for {
		fmt.Printf("%s [y/n] ", s)
		reader := bufio.NewReader(os.Stdin)
		input, _, err := reader.ReadRune()
		if err != nil {
			log.Fatalln(err)
		}
		switch unicode.ToLower(input) {
		case 'y':
			return true
		case 'n':
			return false
		default:
			log.Fatalf("invalid keypress\n")
		}
	}
}
