package main

import "fmt"

func main() {
    version,err := GetPackageVersion("firefox", "<57.0.0")
    if err != nil {
        panic(err)
    }
    fmt.Println(DownloadChocolateyPackage("firefox", version))
}
