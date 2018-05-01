package main

import (
    "encoding/json"
    "errors"
    "github.com/blang/semver"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
)

/*
 Returns all configured package feeds, ordered by their priority
 */
func GetChocolateySources() []string {
    return []string{"https://chocolatey.org/api/v2"}
}

/*
 Checks if chocolatey is installed on the computer
 */
func IsChocolateyInstalled() bool {

    // Check if chocolatey is installed by trying to run the executable
    _, err := exec.LookPath("choco.exe")

    // If there was no error, chocolatey is installed
    return err == nil
}

/*
 Gets the latest package version of the package, based on semantic versioning
 */
func GetPackageVersion(name string, selector string) (string, error) {

    // Check all configured sources
    for _, source := range GetChocolateySources() {

        // Is the source a http API?
        if (strings.HasPrefix(source, "http")) {

            // Get all versions of the package
            resp, err := http.Get(source + "/package-versions/" + name)
            if err != nil {
                log.Println(err)
                continue
            }
            versions := []string{}
            err = json.NewDecoder(resp.Body).Decode(&versions)
            if err != nil {
                log.Println(err)
                continue
            }

            if len(versions) == 0 {
                continue
            }

            i := len(versions) - 1
            v, err := semver.Parse(ToSemantic(versions[i]))
            matches := func(x semver.Version) bool { return true }
            r, rErr := semver.ParseRange(selector)
            if rErr == nil {
                matches = func(x semver.Version) bool { return r(x) }
            }
            for !matches(v) || err != nil {
                i--
                v, err = semver.Parse(ToSemantic(versions[i]))
            }
            return versions[i], nil

        } else {

            // The source must be a directory
            // Get all files inside
            files, err := ioutil.ReadDir(source)
            if err != nil {
                log.Println(err)
                continue
            }

            // All versions we can find
            versions := []string{}

            for _,file := range files {

                // Ignore subdirectories
                if file.IsDir() {
                    continue
                }

                // Is the file a version for the requested package?
                if !strings.HasPrefix(file.Name(), name) {
                    continue
                }

                // Extract the version from the filename
                version := strings.Replace(file.Name(), name + ".", "", 1)
                version = strings.Replace(version, ".nupkg", "", 1)
                versions = append(versions, version)
            }

            if len(versions) == 0 {
                continue
            }

            i := len(versions) - 1
            v, err := semver.Parse(ToSemantic(versions[i]))
            matches := func(x semver.Version) bool { return true }
            r, rErr := semver.ParseRange(selector)
            if rErr == nil {
                matches = func(x semver.Version) bool { return r(x) }
            }
            for !matches(v) || err != nil {
                i--
                v, err = semver.Parse(ToSemantic(versions[i]))
            }
            return versions[i], nil
        }
    }
    return "", errors.New("The specified package does not exist.")
}

/*
 Converts a chocolatey version into a semantic version
 */
func ToSemantic(version string) string {
    split := strings.Split(version, ".")
    major := "0"
    minor := "0"
    patch := "0"
    if len(split) > 0 {
        major = split[0]
    }
    if len(split) > 1 {
        minor = split[1]
    }
    if len(split) > 2 {
        patch = split[2]
    }
    return major + "." + minor + "." + patch
}


func DownloadChocolateyPackage(name string, version string) (string, error) {

    // Get a temporary file path
    file, err := ioutil.TempFile("", "")
    defer file.Close()
    if err != nil {
        log.Println(err)
        return "", err
    }

    // Check every package source
    for _,source := range GetChocolateySources() {

        // Is the source a folder or a webserver?
        if strings.HasPrefix(source, "http") {

            // It is a webserver
            resp, err := http.Get(source + "/package/" + name + "/" + version)
            defer resp.Body.Close()
            if err != nil {
                log.Println(err)
                continue
            }

            // Check if the file was found, and if not check the next source
            if resp.StatusCode == 404 {
                log.Println(err)
                continue
            }

            // Save the package to the disk
            _, err = io.Copy(file, resp.Body)
            if err != nil {
                log.Println(err)
                continue
            }

            return file.Name(), nil

        } else {

            // If the source isn't an url, it is a folder, so we can just copy the file
            filename := name + "." + version + ".nupkg"
            path := filepath.Join(source, filename)

            // Does the file exist?
            if _, err := os.Stat(path); os.IsNotExist(err) {
                log.Println(err)
                continue
            }

            // Copy it
            from, err := os.Open(path)
            if err != nil {
                log.Println(err)
                continue
            }
            defer from.Close()
            _, err = io.Copy(file, from)
            if err != nil {
                log.Println(err)
                continue
            }

            return file.Name(), nil
        }
    }
    return "", errors.New("The specified package does not exist.")

}