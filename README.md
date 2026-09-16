# g: GCP shortcuts for Alfred

## Download

An exported workflow is available under [releases](https://github.com/jarlefosen/alfred-gcloud-shortcuts/releases) (or your fork's releases). Make sure you meet the [requirements](#requirements).

### macOS Gatekeeper Authorization (First Run)

Because the workflow binaries (`bin/products` and `bin/projects`) are natively compiled in Go from GitHub Actions without an Apple Developer subscription, macOS Gatekeeper attaches a quarantine attribute to downloaded `.alfredworkflow` bundles.

If macOS displays a warning stating the binary *cannot be opened because the developer cannot be verified*:

**Option A: Authorize via System Settings (Recommended)**
1. Open **System Settings $\rightarrow$ Privacy & Security**.
2. Scroll down to the **Security** section.
3. You will see a notice that `products` (or `projects`) was blocked.
4. Click **"Open Anyway"** (or **"Allow Anyway"**). You only need to do this once.

**Option B: Strip Quarantine via Terminal**
Before importing, remove the quarantine attribute from the downloaded file:
```bash
xattr -d com.apple.quarantine ~/Downloads/alfred-gcloud-shortcuts.alfredworkflow
```
Or, if already imported into Alfred:
```bash
xattr -cr "$HOME/Library/Application Support/Alfred/Alfred.alfredpreferences/workflows/"*alfred-gcloud*
```

## Usage

Commands:
- `g <query>` for search
- `g-refresh` for updating list of projects


### Refresh projects list

Initially run `g-refresh` in Alfred to update the list of authenticated projects.

### Open product page

`g <project filter>` ↩️️ `BigQuery` ➡️ Opens BigQuery for the selected project.

`g My Project` ↩️ `Kube` ➡️ Opens Kubernetes Engine in GCP for project My Project.

### Configuration

**hotkey**

Changing the variable `hotkey` from `g` to `gcp` results in commands like `gcp <query>` and `gcp-refresh`.

**authuser**

Specify the `authuser` query parameter in case you are logged into multiple Google accounts and you want to open links logged into a specific user.

## Requirements

```
If you initialized gcloud recently, make sure to save the authentication locally, see below.
```

- installed and authenticated `gcloud` https://cloud.google.com/sdk/
- coreutils `brew install coreutils`
- save [auth locally](https://github.com/jarlefosen/alfred-gcloud-shortcuts/issues/5#issuecomment-537852834): `gcloud auth application-default login`

## Development

Build native universal binaries (Apple Silicon arm64 + Intel amd64) and package the workflow:

```bash
# Build universal binaries (bin/products, bin/projects)
make build

# Package the .alfredworkflow file in target/
make workflow

# Build and trigger installation in Alfred
make install
```

Alternatively, link the repository directly into your Alfred workflows directory for local development:

```bash
./scripts/link-alfred.sh
```

