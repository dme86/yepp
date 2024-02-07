package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Release struct {
	TagName string `json:"tag_name"`
}

func main() {
	url := "https://gist.githubusercontent.com/dme86/20de09977037ab339ef613e5a928de14/raw/85309c431600a3ef110b66772b582debbe672f9d/gistfile1.txt"

	// Fetch the content of the URL
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching the URL:", err)
		return
	}
	defer response.Body.Close()

	// Read and print every line and check for releases
	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		line := scanner.Text()
		concatenatedURL := concatenateURL(line)

		// Check for releases and retrieve version information
		version, err := getGitHubReleaseVersion(concatenatedURL)
		if err == nil {
			fmt.Printf("Release version for %s: %s\n", concatenatedURL, version)
		} else {
			fmt.Printf("No release found for %s\n", concatenatedURL)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from the response body:", err)
	}
}

// concatenateURL concatenates "https://github.com/" with the given line
func concatenateURL(line string) string {
	return "https://github.com/" + line
}

// getGitHubReleaseVersion checks for releases on the GitHub repository and returns the version if available
func getGitHubReleaseVersion(repoURL string) (string, error) {
	// Convert the GitHub repository URL to the API URL
	apiURL := strings.Replace(repoURL, "https://github.com/", "https://api.github.com/repos/", 1)

	// Fetch releases from the GitHub API
	response, err := http.Get(apiURL + "/releases/latest")
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	// Parse the JSON response
	var release Release
	if err := json.NewDecoder(response.Body).Decode(&release); err != nil {
		return "", err
	}

	return release.TagName, nil
}

