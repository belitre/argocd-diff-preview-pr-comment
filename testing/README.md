# Testing ArgoCD Diff Preview

This directory contains testing resources for generating ArgoCD application diffs.

## Contents

- **argocd-application.yaml**: Sample ArgoCD Application manifest that uses the ArgoCD Helm chart
- **values.yaml**: Helm values file for customizing the ArgoCD installation
- **generate-diff.sh**: Script to download argocd-diff-preview and generate diffs

## Important Note

`argocd-diff-preview` works with **GitHub repositories**, not standalone files. It compares ArgoCD applications between Git branches in a GitHub repository.

## Usage

### Generate a Diff from a GitHub Repository

Set environment variables to specify your repository:

```bash
cd testing

# Set your repository information
export REPO_OWNER="your-github-username"
export REPO_NAME="your-repository-name"
export BASE_BRANCH="main"  # Optional: base branch to compare against (default: main)

./generate-diff.sh
```

The script will:
1. Detect your operating system (Linux or macOS) and architecture
2. Download the appropriate argocd-diff-preview binary from the latest GitHub release
3. Automatically use your **current git branch** as the target branch
4. Run argocd-diff-preview with `--max-diff-length 999999`
5. Generate a diff comparing ArgoCD applications between your current branch and the base branch

### Default Behavior (Without Environment Variables)

If you run the script without setting environment variables, it will use defaults:

```bash
./generate-diff.sh
```

Defaults:
- Repository: `your-org/your-repo`
- Target branch: Current git branch (automatically detected)
- Base branch: `main`

### Re-download Binary

If you need to update to a newer version of argocd-diff-preview:

```bash
cd testing
rm argocd-diff-preview
./generate-diff.sh
```

## Sample ArgoCD Application Files

This directory contains sample ArgoCD application files for reference:

- **argocd-application.yaml**: Example ArgoCD Application using the ArgoCD Helm chart
- **values.yaml**: Example Helm values file

These files can be used as templates for creating ArgoCD applications in your repository.

## Requirements

- bash
- curl
- git
- Internet connection (for downloading argocd-diff-preview and accessing GitHub)
- Linux or macOS operating system
- A GitHub repository with ArgoCD application manifests

## How argocd-diff-preview Works

`argocd-diff-preview` analyzes ArgoCD applications in a GitHub repository by:

1. Cloning the repository
2. Comparing ArgoCD application manifests between branches
3. Rendering the applications with their sources (Helm charts, Kustomize, etc.)
4. Generating a diff of the resulting Kubernetes manifests

**Command-line flags used:**
- `--repo`: GitHub repository in `OWNER/REPO` format (required)
- `--target-branch`: Branch with changes to preview (required)
- `--base-branch`: Base branch to compare against (default: `main`)
- `--max-diff-length`: Maximum diff message character count (default: 65536)

## Notes

- The binary is downloaded to the `testing/` directory
- The script caches the binary and won't re-download unless you delete it
- Maximum diff length is set to 999999 to capture full diffs
- This directory is initialized as a git repository for local testing purposes
