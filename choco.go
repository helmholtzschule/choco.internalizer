package main

import (
    "errors"
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
 Checks if chocolatey is installed on the computer
 */
func IsChocolateyInstalled() bool {

    // Check if chocolatey is installed by trying to run the executable
    _, err := exec.LookPath("choco.exe")

    // If there was no error, chocolatey is installed
    return err == nil
}


func DownloadChocolateyPackage(name string, version string, source string) (string, error) {

    // Get a temporary file path
    file, err := ioutil.TempFile("", "cinternalize")
    defer file.Close()
    if err != nil {
        log.Println(err)
        return "", err
    }

    // Is the source a folder or a webserver?
    if strings.HasPrefix(source, "http") {

        // It is a webserver
        resp, err := http.Get(source + "/package/" + name + "/" + version)
        defer resp.Body.Close()
        if err != nil {
            return "", errors.New("The specified package does not exist.")
        }

        // Check if the file was found, and if not check the next source
        if resp.StatusCode == 404 {
            return "", errors.New("The specified package does not exist.")
        }

        // Save the package to the disk
        _, err = io.Copy(file, resp.Body)
        if err != nil {
            return "", errors.New("The specified package does not exist.")
        }

        return file.Name(), nil

    } else {

        // If the source isn't an url, it is a folder, so we can just copy the file
        filename := name + "." + version + ".nupkg"
        path := filepath.Join(source, filename)

        // Does the file exist?
        if _, err := os.Stat(path); os.IsNotExist(err) {
            return "", errors.New("The specified package does not exist.")
        }

        // Copy it
        from, err := os.Open(path)
        if err != nil {
            return "", errors.New("The specified package does not exist.")
        }
        defer from.Close()
        _, err = io.Copy(file, from)
        if err != nil {
            return "", errors.New("The specified package does not exist.")
        }

        return file.Name(), nil
    }
}

/*
 Packages the modififed package as a new chocolatey package
 */
func ChocolateyPack(path string) (string, error) {
    cmd := exec.Command("cpack")
    cmd.Dir = path
    err := cmd.Run()
    if err != nil {
        return "", err
    }

    // Fetch the generated file and return the path
    files, err := ioutil.ReadDir(path)
    if err != nil {
        return "", err
    }

    for _, f := range files {
        if strings.HasSuffix(f.Name(),".nupkg") {
            return f.Name(), nil
        }
    }
    return "", errors.New("Something went wrong")
}