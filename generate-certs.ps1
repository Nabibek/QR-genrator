# PowerShell script to generate self-signed SSL certificates using built-in Windows tools

Write-Host "Generating self-signed SSL certificates..."
Write-Host ""

# Go to nginx/certs directory
$certsDir = ".\nginx\certs"
if (-not (Test-Path $certsDir)) {
    New-Item -ItemType Directory -Path $certsDir -Force | Out-Null
}

# Certificate file paths
$certFile = Join-Path $certsDir "server.crt"
$keyFile = Join-Path $certsDir "server.key"
$pfxFile = Join-Path $certsDir "server.pfx"

# Remove old certificates if they exist
if (Test-Path $certFile) { Remove-Item $certFile -Force }
if (Test-Path $keyFile) { Remove-Item $keyFile -Force }
if (Test-Path $pfxFile) { Remove-Item $pfxFile -Force }

# Create self-signed certificate using PowerShell
Write-Host "Creating certificate for localhost..."

$cert = New-SelfSignedCertificate -CertStoreLocation cert:\currentuser\my `
    -DnsName localhost `
    -FriendlyName "Warehouse Local HTTPS" `
    -NotAfter (Get-Date).AddYears(1) `
    -KeyUsage DigitalSignature, KeyEncipherment `
    -TextExtension @("2.5.29.37={text}1.3.6.1.5.5.7.3.1")

Write-Host "Certificate created: $($cert.Thumbprint)"

# Export certificate to PFX (with private key)
$password = ConvertTo-SecureString -String "temp123" -AsPlainText -Force
Export-PfxCertificate -Cert $cert -FilePath $pfxFile -Password $password -Force | Out-Null

write-Host "PFX exported"

# Extract cert and key from PFX using openssl (requires Git or standalone OpenSSL)
# Try to find openssl in common locations
$openssl_paths = @(
    "openssl",
    "C:\Program Files\Git\usr\bin\openssl.exe",
    "C:\Program Files (x86)\Git\usr\bin\openssl.exe"
)

$found = $false
foreach ($path in $openssl_paths) {
    try {
        & $path version 2>&1 | Out-Null
        $openssl = $path
        $found = $true
        break
    } catch {}
}

if ($found) {
    Write-Host "Found OpenSSL at: $openssl"
    Write-Host "Extracting certificate and key..."
    
    # Extract certificate
    & $openssl pkcs12 -in $pfxFile -clcerts -nokeys -out $certFile -password pass:temp123
    
    # Extract private key
    & $openssl pkcs12 -in $pfxFile -nocerts -out $keyFile.tmp -password pass:temp123 -passin pass:temp123
    & $openssl rsa -in $keyFile.tmp -out $keyFile
    Remove-Item $keyFile.tmp -Force
    
    Write-Host ""
    Write-Host "Certificates successfully created!"
    Write-Host ""
    Write-Host "Locations:"
    Write-Host "   Certificate: $certFile"
    Write-Host "   Key:         $keyFile"
} else {
    Write-Host ""
    Write-Host "OpenSSL not found. Please install Git:"
    Write-Host "  https://git-scm.com/download/win"
    Write-Host ""
    Write-Host "Git includes openssl needed to extract certificates from PFX"
    Write-Host ""
    Write-Host "Temporary PFX file created: $pfxFile"
    Write-Host "You can manually extract it after installing OpenSSL"
    exit 1
}

# Verify certificates
if ((Test-Path $certFile) -and (Test-Path $keyFile)) {
    Write-Host ""
    Write-Host "Verification:"
    Write-Host "   Certificate size: $(Get-ChildItem $certFile | % Length) bytes"
    Write-Host "   Key size: $(Get-ChildItem $keyFile | % Length) bytes"
    Write-Host ""
    Write-Host "Self-signed certificate valid for 365 days"
    Write-Host "Browsers will show a security warning (normal for dev)"
}
