package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "strings"
    "io/ioutil"
    "path/filepath"
)

type Release struct {
    TagName string `json:"tag_name"`
    Assets  []struct {
        Name string `json:"name"`
    } `json:"assets"`
}

// InstalledPackage represents an installed package with its version
type InstalledPackage struct {
    Name    string `json:"name"`
    Version string `json:"version"`
}

// InstalledPackages holds all installed packages
type InstalledPackages struct {
    Packages []InstalledPackage `json:"packages"`
}

func main() {
    url := "https://gist.githubusercontent.com/dme86/20de09977037ab339ef613e5a928de14/raw/85309c431600a3ef110b66772b582debbe672f9d/gistfile1.txt"

    // Fetch the GitHub API token from the environment variable
    githubToken := os.Getenv("GITHUB_TOKEN")

    apiURL := "https://api.github.com/"

    // Call the function and pass the API URL
    err := checkGitHubAPIAccess(apiURL, githubToken)
    if err != nil {
        fmt.Println("Error:", err)
        os.Exit(1)
    }

    // Load installed packages from the local configuration file
    installedPackages, err := loadInstalledPackages("installed_packages.json")
    if err != nil {
        fmt.Println("Error loading installed packages:", err)
        return
    }

    // Fetch the content of the URL
    response, err := http.Get(url)
    if err != nil {
        fmt.Println("Error fetching the URL:", err)
        return
    }
    defer response.Body.Close()

    // Read and process each line
    scanner := bufio.NewScanner(response.Body)
    for scanner.Scan() {
        line := scanner.Text()
        repoURL := concatenateURL(line)

        // Get the package name from the repo URL
        packageName := strings.Split(line, "/")[1]

        // Check for releases and retrieve version information
        release, err := getGitHubReleaseInfo(repoURL)
        if err != nil {
            fmt.Printf("Error checking release for %s: %v\n", repoURL, err)
            continue
        }

        // Get the installed version
        installedVersion := getInstalledVersion(installedPackages, packageName)

        // Compare versions
        if installedVersion == "" {
            fmt.Printf("Package %s is not installed. Latest version is %s.\n", packageName, release.TagName)
            // Prompt user to install
            if promptUser(fmt.Sprintf("Do you want to install %s?", packageName)) {
                // Install the package (you need to implement installPackage)
                installPackage(packageName, release.TagName)
                // Update the installed packages list
                installedPackages = addOrUpdateInstalledPackage(installedPackages, packageName, release.TagName)
            }
        } else if installedVersion != release.TagName {
            fmt.Printf("Package %s has an update available: %s -> %s\n", packageName, installedVersion, release.TagName)
            // Prompt user to update
            if promptUser(fmt.Sprintf("Do you want to update %s?", packageName)) {
                // Update the package (you need to implement updatePackage)
                updatePackage(packageName, release.TagName)
                // Update the installed packages list
                installedPackages = addOrUpdateInstalledPackage(installedPackages, packageName, release.TagName)
            }
        } else {
            fmt.Printf("Package %s is up to date (%s).\n", packageName, installedVersion)
        }
    }

    if err := scanner.Err(); err != nil {
        fmt.Println("Error reading from the response body:", err)
    }

    // Save the updated installed packages list
    err = saveInstalledPackages("installed_packages.json", installedPackages)
    if err != nil {
        fmt.Println("Error saving installed packages:", err)
    }
}

// checkGitHubAPIAccess checks if you can access the GitHub API
func checkGitHubAPIAccess(apiURL string, token string) error {
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

// loadInstalledPackages loads the installed packages from a JSON file
func loadInstalledPackages(filename string) (*InstalledPackages, error) {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        if os.IsNotExist(err) {
            // Return an empty list if the file does not exist
            return &InstalledPackages{}, nil
        }
        return nil, err
    }

    var installedPackages InstalledPackages
    err = json.Unmarshal(data, &installedPackages)
    if err != nil {
        return nil, err
    }

    return &installedPackages, nil
}

// saveInstalledPackages saves the installed packages to a JSON file
func saveInstalledPackages(filename string, installedPackages *InstalledPackages) error {
    data, err := json.MarshalIndent(installedPackages, "", "  ")
    if err != nil {
        return err
    }

    err = ioutil.WriteFile(filename, data, 0644)
    if err != nil {
        return err
    }

    return nil
}

// getInstalledVersion returns the installed version of a package
func getInstalledVersion(installedPackages *InstalledPackages, packageName string) string {
    for _, pkg := range installedPackages.Packages {
        if pkg.Name == packageName {
            return pkg.Version
        }
    }
    return ""
}

// addOrUpdateInstalledPackage adds or updates a package in the installed packages list
func addOrUpdateInstalledPackage(installedPackages *InstalledPackages, packageName, version string) *InstalledPackages {
    for i, pkg := range installedPackages.Packages {
        if pkg.Name == packageName {
            installedPackages.Packages[i].Version = version
            return installedPackages
        }
    }
    installedPackages.Packages = append(installedPackages.Packages, InstalledPackage{Name: packageName, Version: version})
    return installedPackages
}

// promptUser prompts the user with a yes/no question
func promptUser(question string) bool {
    fmt.Printf("%s [y/N]: ", question)
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    response := strings.ToLower(strings.TrimSpace(scanner.Text()))
    return response == "y" || response == "yes"
}

// installPackage installs the package (you need to implement this function)
func installPackage(packageName, version string) {
    fmt.Printf("Installing %s version %s...\n", packageName, version)
    // Implement the installation logic here
}

// updatePackage updates the package (you need to implement this function)
func updatePackage(packageName, version string) {
    fmt.Printf("Updating %s to version %s...\n", packageName, version)
    // Implement the update logic here
}

