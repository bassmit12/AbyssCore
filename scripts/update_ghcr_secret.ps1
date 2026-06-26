# Update-GhcrSecret.ps1
# Helper script to update the GitHub Container Registry image pull secret (ghcr-secret)
# across all required namespaces in the Kubernetes cluster.

param (
    [Parameter(Mandatory=$true, HelpMessage="Your new GitHub Personal Access Token (PAT) with read:packages scope")]
    [string]$Pat
)

$namespaces = @("abysscore", "argocd", "monitoring")
$kubeconfig = Join-Path $PSScriptRoot "..\abysscore_cluster.conf"

if (-not (Test-Path $kubeconfig)) {
    Write-Error "Could not find abysscore_cluster.conf at: $kubeconfig. Please make sure the cluster config file is in the project root."
    exit 1
}

Write-Host "Updating ghcr-secret across namespaces: $namespaces..." -ForegroundColor Cyan

foreach ($ns in $namespaces) {
    Write-Host "Updating namespace: $ns" -ForegroundColor Yellow
    
    # Delete the old secret if it exists
    kubectl --kubeconfig=$kubeconfig delete secret ghcr-secret --namespace $ns --ignore-not-found
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to delete old secret in namespace $ns"
        continue
    }

    # Create the new secret
    kubectl --kubeconfig=$kubeconfig create secret docker-registry ghcr-secret `
      --namespace $ns `
      --docker-server=ghcr.io `
      --docker-username=bassmit12 `
      --docker-password=$Pat
      
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Successfully updated ghcr-secret in namespace: $ns" -ForegroundColor Green
    } else {
        Write-Error "Failed to create secret in namespace $ns"
    }
}

Write-Host "`nSecret update completed. Pods should automatically retry pulling the images within a few minutes." -ForegroundColor Green
