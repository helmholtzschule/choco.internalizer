package main

import (
    "io/ioutil"
    "os"
    "os/exec"
    "strings"
)

func ImportChocolateyAPI(src string) string {
    return `if (Test-Path env:CINTERNALIZE_LOG) {
    Import-Module C:/ProgramData/chocolatey/helpers/chocolateyInstaller.psm1
}

` + src
}

func InstallChocolateyPackageWrapped(src string) string {
    return `function Install-ChocolateyPackage-Wrapped {
    param(
        [parameter(Mandatory=$true, Position=0)][string] $packageName,
        [parameter(Mandatory=$false, Position=1)]
        [alias("installerType","installType")][string] $fileType = 'exe',
        [parameter(Mandatory=$false, Position=2)][string[]] $silentArgs = '',
        [parameter(Mandatory=$false, Position=3)][string] $url = '',
        [parameter(Mandatory=$false, Position=4)]
        [alias("url64")][string] $url64bit = '',
        [parameter(Mandatory=$false)] $validExitCodes = @(0),
        [parameter(Mandatory=$false)][string] $checksum = '',
        [parameter(Mandatory=$false)][string] $checksumType = '',
        [parameter(Mandatory=$false)][string] $checksum64 = '',
        [parameter(Mandatory=$false)][string] $checksumType64 = '',
        [parameter(Mandatory=$false)][hashtable] $options = @{Headers=@{}},
        [alias("fileFullPath")][parameter(Mandatory=$false)][string] $file = '',
        [alias("fileFullPath64")][parameter(Mandatory=$false)][string] $file64 = '',
        [parameter(Mandatory=$false)]
        [alias("useOnlyPackageSilentArgs")][switch] $useOnlyPackageSilentArguments = $false,
        [parameter(Mandatory=$false)][switch]$useOriginalLocation,
        [parameter(ValueFromRemainingArguments = $true)][Object[]] $ignoredArguments
    )
    [string]$silentArgs = $silentArgs -join ' '

    if ($PSBoundParameters.ContainsKey('url')) {
        if (Test-Path env:CINTERNALIZE_LOG) {
            Write-Host "cinternalize: $url - $fileType"
            if ($PSBoundParameters.ContainsKey('url64bit')) {
                Write-Host "cinternalize: $url64bit - $fileType"
            }
        } else {
            $toolsDir = Split-Path $MyInvocation.ScriptName
            $hasher = new-object System.Security.Cryptography.MD5CryptoServiceProvider
            $hashByteArray = $hasher.ComputeHash([System.Text.Encoding]::UTF8.GetBytes($PSBoundParameters['url']))
            foreach($byte in $hashByteArray) {
                $result += "{0:X2}" -f $byte
            }
            $PSBoundParameters['url'] = ''
            $PSBoundParameters['file'] = Join-Path $toolsDir (Join-Path 'cinternalize' ($result + '.' + $fileType))

            if ($PSBoundParameters.ContainsKey('url64bit')) {
                $hashByteArray = $hasher.ComputeHash([System.Text.Encoding]::UTF8.GetBytes($PSBoundParameters['url64bit']))
                $result = ''
                foreach($byte in $hashByteArray) {
                    $result += "{0:X2}" -f $byte
                }
                $PSBoundParameters['url64bit'] = ''
                $PSBoundParameters['file64'] = Join-Path $toolsDir (Join-Path 'cinternalize' ($result + '.' + $fileType))
            }
        }
    }

    if (!(Test-Path env:CINTERNALIZE_LOG)) {
        Install-ChocolateyPackage @PSBoundParameters
    }
}

` + strings.Replace(src, "Install-ChocolateyPackage", "Install-ChocolateyPackage-Wrapped", -1)
}

func InstallChocolateyZipPackageWrapped(src string) string {
    return `function Install-ChocolateyZipPackage-Wrapped {
    param(
        [parameter(Mandatory=$true, Position=0)][string] $packageName,
        [parameter(Mandatory=$false, Position=1)][string] $url = '',
        [parameter(Mandatory=$true, Position=2)]
        [alias("destination")][string] $unzipLocation,
        [parameter(Mandatory=$false, Position=3)]
        [alias("url64")][string] $url64bit = '',
        [parameter(Mandatory=$false)][string] $specificFolder ='',
        [parameter(Mandatory=$false)][string] $checksum = '',
        [parameter(Mandatory=$false)][string] $checksumType = '',
        [parameter(Mandatory=$false)][string] $checksum64 = '',
        [parameter(Mandatory=$false)][string] $checksumType64 = '',
        [parameter(Mandatory=$false)][hashtable] $options = @{Headers=@{}},
        [alias("fileFullPath")][parameter(Mandatory=$false)][string] $file = '',
        [alias("fileFullPath64")][parameter(Mandatory=$false)][string] $file64 = '',
        [parameter(ValueFromRemainingArguments = $true)][Object[]] $ignoredArguments
    )
    [string]$silentArgs = $silentArgs -join ' '

    if ($PSBoundParameters.ContainsKey('url')) {
        if (Test-Path env:CINTERNALIZE_LOG) {
            Write-Host "cinternalize: $url - zip"
            if ($PSBoundParameters.ContainsKey('url64bit')) {
                Write-Host "cinternalize: $url64bit - zip"
            }
        } else {
            $toolsDir = Split-Path $MyInvocation.ScriptName
            $hasher = new-object System.Security.Cryptography.MD5CryptoServiceProvider
            $hashByteArray = $hasher.ComputeHash([System.Text.Encoding]::UTF8.GetBytes($PSBoundParameters['url']))
            foreach($byte in $hashByteArray) {
                $result += "{0:X2}" -f $byte
            }
            $PSBoundParameters['url'] = ''
            $PSBoundParameters['file'] = Join-Path $toolsDir (Join-Path 'cinternalize' ($result + '.zip'))

            if ($PSBoundParameters.ContainsKey('url64bit')) {
                $hashByteArray = $hasher.ComputeHash([System.Text.Encoding]::UTF8.GetBytes($PSBoundParameters['url64bit']))
                $result = ''
                foreach($byte in $hashByteArray) {
                    $result += "{0:X2}" -f $byte
                }
                $PSBoundParameters['url64bit'] = ''
                $PSBoundParameters['file64'] = Join-Path $toolsDir (Join-Path 'cinternalize' ($result + '.zip'))
            }
        }
    }

    if (!(Test-Path env:CINTERNALIZE_LOG)) {
        Install-ChocolateyZipPackage @PSBoundParameters
    }
}

` + strings.Replace(src, "Install-ChocolateyZipPackage", "Install-ChocolateyZipPackage-Wrapped", -1)
}

func GetChocolateyWebFileWrapped(src string) string {
    return `function Get-ChocolateyWebFile-Wrapped {
    param(
        [parameter(Mandatory=$true, Position=0)][string] $packageName,
        [parameter(Mandatory=$true, Position=1)][string] $fileFullPath,
        [parameter(Mandatory=$false, Position=2)][string] $url = '',
        [parameter(Mandatory=$false, Position=3)]
        [alias("url64")][string] $url64bit = '',
        [parameter(Mandatory=$false)][string] $checksum = '',
        [parameter(Mandatory=$false)][string] $checksumType = '',
        [parameter(Mandatory=$false)][string] $checksum64 = '',
        [parameter(Mandatory=$false)][string] $checksumType64 = $checksumType,
        [parameter(Mandatory=$false)][hashtable] $options = @{Headers=@{}},
        [parameter(Mandatory=$false)][switch] $getOriginalFileName,
        [parameter(Mandatory=$false)][switch] $forceDownload,
        [parameter(ValueFromRemainingArguments = $true)][Object[]] $ignoredArguments
    )
    [string]$silentArgs = $silentArgs -join ' '

    if ($PSBoundParameters.ContainsKey('url')) {
        if (Test-Path env:CINTERNALIZE_LOG) {
            Write-Host "cinternalize: $url - dat"
            if ($PSBoundParameters.ContainsKey('url64bit')) {
                Write-Host "cinternalize: $url64bit - dat"
            }
        } else {
            $toolsDir = Split-Path $MyInvocation.ScriptName
            $hasher = new-object System.Security.Cryptography.MD5CryptoServiceProvider
            $hashByteArray = $hasher.ComputeHash([System.Text.Encoding]::UTF8.GetBytes($PSBoundParameters['url']))
            foreach($byte in $hashByteArray) {
                $result += "{0:X2}" -f $byte
            }
            $PSBoundParameters['url'] = 'file:///' + (Join-Path $toolsDir (Join-Path 'cinternalize' ($result + '.dat')))

            if ($PSBoundParameters.ContainsKey('url64bit')) {
                $hashByteArray = $hasher.ComputeHash([System.Text.Encoding]::UTF8.GetBytes($PSBoundParameters['url64bit']))
                $result = ''
                foreach($byte in $hashByteArray) {
                    $result += "{0:X2}" -f $byte
                }
                $PSBoundParameters['url64bit'] = 'file:///' + (Join-Path $toolsDir (Join-Path 'cinternalize' ($result + '.dat')))
            }
        }
    }

    if (!(Test-Path env:CINTERNALIZE_LOG)) {
        Get-ChocolateyWebFile @PSBoundParameters
    }
}

` + strings.Replace(src, "Get-ChocolateyWebFile", "Get-ChocolateyWebFile-Wrapped", -1)
}

/*
 Modifies a script and wraps the chocolatey downloading functions
 */
func ModifyScript(file string) error {
    bytes, err := ioutil.ReadFile(file)
    if err != nil {
        return err
    }

    content := string(bytes)
    content = GetChocolateyWebFileWrapped(content)
    content = InstallChocolateyZipPackageWrapped(content)
    content = InstallChocolateyPackageWrapped(content)
    content = ImportChocolateyAPI(content)

    return ioutil.WriteFile(file, []byte(content), 0755)
}

/*
 Executes a script and returns its output
 */
func RunScript(file string, log bool, params string) (string, error) {
    cmd := exec.Command("powershell", file)
    cmd.Env = os.Environ()
    if log {
        cmd.Env = append(cmd.Env, "CINTERNALIZE_LOG=True", "chocolateyPackageParameters=" + params, "ChocolateyForce=true")
    }
    buffer, err := cmd.Output()
    return string(buffer), err
}