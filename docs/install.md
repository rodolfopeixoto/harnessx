# Installing HarnessX

Pick the path that matches your OS. All paths land you on the same
binary version (currently v0.15.0) and `harness update` keeps you there
afterwards.

---

## macOS + Linux — install.sh (recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/rodolfopeixoto/harnessx/main/scripts/install.sh | bash
harness version
```

Defaults to `/usr/local/bin`. Override with `HARNESS_PREFIX`:

```bash
HARNESS_PREFIX="$HOME/.local/bin" \
  curl -fsSL https://raw.githubusercontent.com/rodolfopeixoto/harnessx/main/scripts/install.sh | bash
```

---

## macOS + Linux — Homebrew tap

```bash
brew tap rodolfopeixoto/tap
brew install harnessx
harness version
```

The tap repository ships `Formula/harnessx.rb` regenerated on every
release by `scripts/gen-brew-formula.sh`. Each release of this repo
contains the same formula file under `Formula/harnessx.rb` for
reference.

To maintain the tap manually:

```bash
git clone https://github.com/rodolfopeixoto/homebrew-tap
cp /path/to/harnessx/Formula/harnessx.rb homebrew-tap/Formula/harnessx.rb
cd homebrew-tap && git commit -am "harnessx vX.Y.Z" && git push
```

---

## Windows — manual unzip

```powershell
$tag = "v0.15.0"
$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "arm64" }
$url = "https://github.com/rodolfopeixoto/harnessx/releases/download/$tag/harness-windows-$arch.zip"
$tmp = "$env:TEMP\harnessx-$tag.zip"
Invoke-WebRequest -Uri $url -OutFile $tmp
Expand-Archive -Force $tmp -DestinationPath "$env:USERPROFILE\.harness\bin"
Move-Item -Force "$env:USERPROFILE\.harness\bin\harness-windows-$arch.exe" "$env:USERPROFILE\.harness\bin\harness.exe"
Remove-Item $tmp
```

Then add `%USERPROFILE%\.harness\bin` to `PATH` via System Properties →
Environment Variables.

Verify:

```powershell
harness version
```

### Windows — Scoop bucket (community-maintained)

If you maintain a Scoop bucket, drop this `harnessx.json`:

```json
{
  "version": "0.15.0",
  "url": "https://github.com/rodolfopeixoto/harnessx/releases/download/v0.15.0/harness-windows-amd64.zip",
  "hash": "<sha256 from the .sha256 file>",
  "extract_dir": ".",
  "bin": [["harness-windows-amd64.exe", "harness"]],
  "checkver": "github",
  "autoupdate": {
    "url": "https://github.com/rodolfopeixoto/harnessx/releases/download/v$version/harness-windows-amd64.zip"
  }
}
```

Then `scoop install harnessx`.

---

## Build from source (any OS)

```bash
git clone https://github.com/rodolfopeixoto/harnessx
cd harnessx
make build
./bin/harness version
```

Requires Go 1.23+ and Node 20+ (only for the dashboard).

---

## Verify any install path

```bash
harness version
harness doctor
harness update status
```

---

## Updating later

```bash
harness update                  # latest stable
harness update --channel beta   # opt into pre-releases
harness update --tag vX.Y.Z     # pin a tag
```

`harness update` works on every platform once you are past v0.4.0.
First install via this guide once; from there `harness update` keeps
you current.
