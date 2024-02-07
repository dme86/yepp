package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type Release struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
	} `json:"assets"`
}

func main() {
	url := "https://gist.githubusercontent.com/dme86/20de09977037ab339ef613e5a928de14/raw/85309c431600a3ef110b66772b582debbe672f9d/gistfile1.txt"

	apiURL := "https://api.github.com/"

	err := checkGitHubAPIAccess(apiURL)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// Fetch the content of the URL
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching the URL:", err)
		return
	}
	defer response.Body.Close()

	// Read and print every line and check for releases and binaries
	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		line := scanner.Text()
		concatenatedURL := concatenateURL(line)

		// Check for releases and retrieve version information
		release, err := getGitHubReleaseInfo(concatenatedURL)
		if err != nil {
			fmt.Printf("Error checking release for %s: %v\n", concatenatedURL, err)
			continue
		}

		// Print version and binary availability
		fmt.Printf("Release version for %s: %s |bin=%s\n", concatenatedURL, release.TagName, checkBinaryAvailability(release, "MyBinary"))
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from the response body:", err)
	}
}

// checkGitHubAPIAccess checks if you can access the GitHub API
func checkGitHubAPIAccess(apiURL string) error {
	// Make a request to the GitHub API
	response, err := http.Get(apiURL)
	if err != nil {
		return fmt.Errorf("Error making request: %v", err)
	}
	defer response.Body.Close()

	// Check the HTTP status code
	if response.StatusCode == http.StatusOK {
		fmt.Println("You can access the GitHub API.")
		return nil
	} else if response.StatusCode == http.StatusForbidden {
		return fmt.Errorf("Unable to access the GitHub API. Status code: %d", response.StatusCode)
	}

	return fmt.Errorf("Unexpected status code: %d", response.StatusCode)
}

// concatenateURL concatenates "https://github.com/" with the given line
func concatenateURL(line string) string {
	return "https://github.com/" + line
}

// getGitHubReleaseInfo checks for releases on the GitHub repository and returns the Release struct if available
func getGitHubReleaseInfo(repoURL string) (*Release, error) {
	// Convert the GitHub repository URL to the API URL
	apiURL := strings.Replace(repoURL, "https://github.com/", "https://api.github.com/repos/", 1)

	// Fetch releases from the GitHub API
	response, err := http.Get(apiURL + "/releases/latest")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Parse the JSON response
	var release Release
	if err := json.NewDecoder(response.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// checkBinaryAvailability checks if there are binary assets in the latest release
func checkBinaryAvailability(release *Release, binaryName string) string {
	// Print out information about assets for debugging
	// fmt.Println("Assets for release:", release.TagName)
	// for _, asset := range release.Assets {
	// 	fmt.Println(asset.Name)
	// }

	// Define valid binary suffixes
	validSuffixes := []string{".zip", ".tar.gz"}

	// Iterate through assets of the release and check for binary presence
	for _, asset := range release.Assets {
		// Print intermediate result
		fmt.Printf("Checking asset: %s\n", asset.Name)

		// Check if the asset name contains ".zip" or ".tar.gz"
		for _, suffix := range validSuffixes {
			if strings.HasSuffix(asset.Name, suffix) {
				return fmt.Sprintf("yepp (%s)", asset.Name)
			}
		}
	}

	// If no binary is found, return the asset names for further inspection
	var assetNames []string
	for _, asset := range release.Assets {
		assetNames = append(assetNames, asset.Name)
	}

	return fmt.Sprintf("nope for %s. Asset names: %v", release.TagName, assetNames)
}
