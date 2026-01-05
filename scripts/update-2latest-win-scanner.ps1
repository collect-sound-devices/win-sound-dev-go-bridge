# pick latest stable if it's newer than any rc; otherwise latest rc
$mod = 'github.com/collect-sound-devices/sound-win-scanner/v4'

# need Go on PATH
if (-not (Get-Command go -ErrorAction SilentlyContinue)) { throw 'go not found in PATH' }

# list published module versions
$all = (go list -m -versions $mod) -split '\s+'

# no versions -> nothing to pick
if (-not $all -or $all.Count -lt 2) { throw "no versions returned for $mod" }

# v4.<minor>.<patch>[-rc.<n>]
$items = foreach ($v in $all) {
    if ($v -match '^v4\.(\d+)\.(\d+)(?:-rc\.(\d+))?$') {
        $minor = [int]$Matches[1]
        $patch = [int]$Matches[2]
        $rc    = if ($Matches[3]) { [int]$Matches[3] } else { $null }

        # simple order key: higher minor/patch wins; stable > rc for same minor/patch
        $rcKey = if ($rc -eq $null) { 999 } else { $rc }
        $key = ($minor * 1000000) + ($patch * 1000) + $rcKey

        [pscustomobject]@{ Tag = $v; Rc = $rc; Key = $key }
    }
}

$stable = $items | Where-Object { $_.Rc -eq $null } | Sort-Object Key | Select-Object -Last 1
$rcTop  = $items | Where-Object { $_.Rc -ne $null } | Sort-Object Key | Select-Object -Last 1

$chosen = if ($stable -and (!$rcTop -or $stable.Key -gt $rcTop.Key)) { $stable.Tag } else { $rcTop.Tag }

# print just the version tag
Write-Output $chosen

# pick the version and install it
go get $mod@$chosen
go mod tidy
