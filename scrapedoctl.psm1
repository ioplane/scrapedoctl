<#
.SYNOPSIS
    Modernized Scrapedoctl PowerShell Module.
.DESCRIPTION
    Provides a native PowerShell experience for the Scrape.do CLI tool,
    including cmdlet wrapping, error handling, and context-aware completion.
#>

function Invoke-Scrapedoctl {
    <#
    .SYNOPSIS
        Wraps the native scrapedoctl binary.
    .DESCRIPTION
        This function enables native command error action preference and provides
        rich argument completion.
    #>
    [CmdletBinding()]
    param(
        [Parameter(ValueFromRemainingArguments = $true)]
        [ArgumentCompleter({
            param($commandName, $parameterName, $wordToComplete, $commandAst, $fakeBoundParameters)

            # Reference completer parameters to satisfy PSScriptAnalyzer (PSReviewUnusedParameter)
            $null = $commandName, $parameterName, $wordToComplete, $fakeBoundParameters

            $exe = Get-ScrapedoctlBinary
            if (-not $exe) { return }

            # Prepare the command to request completions
            $elements = $commandAst.CommandElements | ForEach-Object { "$_" }
            # Remove the first element (the command itself)
            $cmdArgs = $elements[1..($elements.Count - 1)] -join " "

            # Call native binary for completion results
            # Disable ActiveHelp which is not supported for Powershell
            $env:SCRAPEDOCTL_ACTIVE_HELP = 0
            $results = & $exe __complete $cmdArgs 2>$null

            if ($results.Count -eq 0) { return }

            # Parse results (skip the last line which is the directive)
            $items = $results[0..($results.Count - 2)]

            foreach ($line in $items) {
                $parts = $line.Split("`t", 2)
                $name = $parts[0]
                $desc = if ($parts.Count -gt 1) { $parts[1] } else { " " }

                [System.Management.Automation.CompletionResult]::new(
                    $name,
                    $name,
                    'ParameterValue',
                    $desc
                )
            }
        })]
        [string[]]$Arguments
    )

    process {
        # Enable native command error action preference (PS 7.3+)
        $local:PSNativeCommandUseErrorActionPreference = $true

        # Locate binary
        $exe = Get-ScrapedoctlBinary
        if (-not $exe) {
            Write-Error "scrapedoctl binary not found in path or module directory."
            return
        }

        # Execute with call operator
        & $exe @Arguments
    }
}

function Get-ScrapedoctlBinary {
    <#
    .SYNOPSIS
        Locates the scrapedoctl binary.
    #>
    # 1. Check module directory
    $modDir = $PSScriptRoot
    $ext = if ($IsWindows) { ".exe" } else { "" }
    $localExe = Join-Path $modDir "scrapedoctl$ext"
    if (Test-Path -LiteralPath $localExe -PathType Leaf -ErrorAction SilentlyContinue) { return $localExe }

    # 2. Check PATH for the native binary (exclude aliases/functions to avoid circular resolution)
    $pathExe = Get-Command scrapedoctl -CommandType Application -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($pathExe) { return $pathExe.Source }

    return $null
}

# Aliases and Exports
Set-Alias -Name scrapedoctl -Value Invoke-Scrapedoctl -Description "Scrape.do CLI"
Export-ModuleMember -Function Invoke-Scrapedoctl, Get-ScrapedoctlBinary -Alias scrapedoctl
