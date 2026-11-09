package domain

import (
	"path/filepath"
	"strconv"
)

// PythonDownload is a ready-made Python the launcher can fetch for itself when the computer has
// none: a python-build-standalone archive, unpacked in place with no installer or password.
type PythonDownload struct {
	Version string // shown to the user, "3.12.15"
	URL     string
	SHA256  string // the archive's checksum, so a broken or altered download is refused
	SizeMB  int    // roughly, to say how long it may take
}

// pythonRelease is the python-build-standalone release the downloads come from. Python 3.12 is the
// newest with pyrealsense2 wheels for every platform below.
const pythonRelease = "https://github.com/astral-sh/python-build-standalone/releases/download/20261013/"

// pythonDownloads maps GOOS/GOARCH to the archive for that computer.
var pythonDownloads = map[string]PythonDownload{
	"linux/amd64":  {Version: "3.12.15", URL: pythonRelease + "cpython-3.12.15%2B20261013-x86_64-unknown-linux-gnu-install_only.tar.gz", SHA256: "5d546fc004fc2435fa142eb8f42ed8de5eb4725963dc22f7e6671014a80eb560", SizeMB: 67},
	"linux/arm64":  {Version: "3.12.15", URL: pythonRelease + "cpython-3.12.15%2B20261013-aarch64-unknown-linux-gnu-install_only.tar.gz", SHA256: "34a43010cc05982d8b5debbad7d09f7b1200117ebf018f7acfd656895e408205", SizeMB: 52},
	"darwin/arm64": {Version: "3.12.15", URL: pythonRelease + "cpython-3.12.15%2B20261013-aarch64-apple-darwin-install_only.tar.gz", SHA256: "ba32036ad8f4c434d97a1ee7feef8b7edc0fac16777639b97e39c74a108e4440", SizeMB: 25},
	"darwin/amd64": {Version: "3.12.15", URL: pythonRelease + "cpython-3.12.15%2B20261013-x86_64-apple-darwin-install_only.tar.gz", SHA256: "9686ddd8639b855905447ea4a16eb54a804a9ef12afbe7c9ecb1412d9c1fbc98", SizeMB: 25},
}

// PythonDownloadFor returns the Python to download on goos/goarch, if there is one.
func PythonDownloadFor(goos, goarch string) (PythonDownload, bool) {
	download, ok := pythonDownloads[goos+"/"+goarch]
	return download, ok
}

// DownloadedPython is the interpreter inside dir once download has been unpacked there.
func DownloadedPython(dir string) string {
	return filepath.Join(dir, "python", "bin", "python3")
}

// pythonInstallScript downloads the archive ($1) to a temporary file, checks it against the
// checksum ($2) and unpacks it into the folder ($3) through a sibling folder, so a cut connection
// never leaves a half-installed Python that looks finished.
const pythonInstallScript = `set -e
url=$1 sum=$2 dir=$3
mkdir -p "$(dirname "$dir")"
archive="$dir.tar.gz" partial="$dir.partial"
rm -rf "$archive" "$partial"
curl -fsSL --retry 3 -o "$archive" "$url"
if command -v sha256sum >/dev/null; then got=$(sha256sum "$archive"); else got=$(shasum -a 256 "$archive"); fi
if [ "${got%% *}" != "$sum" ]; then rm -f "$archive"; echo "La descarga de Python está dañada. Vuelve a intentarlo." >&2; exit 1; fi
mkdir "$partial"
tar -xzf "$archive" -C "$partial"
rm -rf "$dir"
mv "$partial" "$dir"
rm -f "$archive"`

// PythonInstallSteps are the steps that put download into dir, to run before SetupPlan with
// DownloadedPython(dir).
func PythonInstallSteps(download PythonDownload, dir string) []SetupStep {
	label := "Descargando Python " + download.Version + " (unos " + strconv.Itoa(download.SizeMB) + " MB)"
	return []SetupStep{{Label: label, Program: "bash", Args: []string{"-c", pythonInstallScript, "bash", download.URL, download.SHA256, dir}}}
}
