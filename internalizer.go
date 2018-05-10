package main

import (
    "crypto/md5"
    "flag"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strings"
)

const VERSION = "0.0.1"

func main() {

    // Say hello
    log.SetOutput(os.Stdout)
    log.Println("choco.internalizer -- Version: " + VERSION)

    // Parse commandline parameters
    pkg := flag.String("package", "", "The package that should get internalized.")
    version := flag.String("version", "", "The version of the package that should be used.")
    source := flag.String("source", "https://chocolatey.org/api/v2", "The chocolatey package source that should be used.")
    output := flag.String("output", "", "The directory where the internalized file gets saved.")
    params := flag.String("parameters", "", "Installation parameters for the script.")
    flag.Parse()

    // Validate them
    if len(*pkg) == 0 {
        log.Println("[ERR]: Invalid package name!")
        os.Exit(1)
    }
    if len(*version) == 0 {
        log.Println("[ERR]: Invalid version!")
        os.Exit(2)
    }
    if len(*output) == 0 {
        log.Println("[ERR]: Invalid output path!")
        os.Exit(3)
    }

    // Check if chocolatey is installed
    if !IsChocolateyInstalled() {
        log.Fatal("[ERR]: Chocolatey is not installed!")
        os.Exit(4)
    }

    // Download the package
    log.Println("Downloading package " + *pkg + " (Version: " + *version + ") from " + *source)
    file, err := DownloadChocolateyPackage(*pkg, *version, *source)
    if err != nil {
        log.Fatal(err)
        os.Exit(5)
    }

    // Unzip the package
    log.Println("Unpacking " + *pkg)
    dir, err := ioutil.TempDir("", "cinternalize")
    if err != nil {
        log.Fatal(err)
        os.Exit(6)
    }
    _, err = Unzip(file, dir)
    if err != nil {
        log.Fatal(err)
        os.Exit(7)
    }

    // Clean up
    os.RemoveAll(filepath.Join(dir, "package"))
    os.RemoveAll(filepath.Join(dir, "_rels"))
    os.Remove(filepath.Join(dir, "[Content_Types].xml"))

    // Modify the installation script
    log.Println("Injecting internalizer into " + *pkg)
    err = ModifyScript(filepath.Join(dir, "tools", "chocolateyInstall.ps1"))
    if err != nil {
        log.Fatal(err)
        os.Exit(8)
    }

    // Run the script and fetch all urls
    out, err := RunScript(filepath.Join(dir, "tools", "chocolateyInstall.ps1"), true, *params)
    if err != nil {
        log.Fatal(err)
        os.Exit(9)
    }
    lines := strings.Split(out, "\n")
    err = os.MkdirAll(filepath.Join(dir, "tools", "cinternalize"), 0755)
    if err != nil {
        log.Fatal(err)
        os.Exit(10)
    }
    for _,line := range lines {
        if !strings.HasPrefix(line, "cinternalize: ") {
            continue;
        }
        split := strings.Split(strings.Replace(line, "cinternalize: ", "", 1), " - ")
        url := split[0]
        filetype := split[1]
        filename := fmt.Sprintf("%X", md5.Sum([]byte(url))) + "." + filetype

        // Download the file
        log.Println("Downloading " + url)
        installer, err := os.Create(filepath.Join(dir, "tools", "cinternalize", filename))
        if err != nil {
            log.Fatal(err)
            os.Exit(11)
        }
        resp, err := http.Get(url)
        if err != nil {
            log.Fatal(err)
            os.Exit(12)
        }

        // Check if the file was found
        if resp.StatusCode == 404 {
            log.Fatal(err)
            os.Exit(13)
        }

        // Save the file to the disk
        log.Println("Saving to " + installer.Name())
        _, err = io.Copy(installer, resp.Body)
        if err != nil {
            log.Fatal(err)
            os.Exit(14)
        }

        resp.Body.Close()
        installer.Close()
    }

    // Repackage the package
    log.Println("Packing " + *pkg)
    newPkg, err := ChocolateyPack(dir)
    if err != nil {
        log.Fatal(err)
        os.Exit(15)
    }

    // Create the output dir if it doesn't exist
    if _, err := os.Stat(*output); os.IsNotExist(err) {
        err = os.MkdirAll(*output, 0755)
        if err != nil {
            log.Fatal(err)
            os.Exit(16)
        }
    }

    // Copy the package to the destination
    buffer, err := ioutil.ReadFile(filepath.Join(dir, newPkg))
    if err != nil {
        log.Fatal(err)
        os.Exit(17)
    }
    ioutil.WriteFile(filepath.Join(*output, newPkg), buffer, 0755)

    // Cleanup
    os.RemoveAll(dir)
    os.Remove(file)
}
