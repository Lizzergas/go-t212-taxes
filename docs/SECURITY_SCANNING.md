# Security Scanning

This project uses automated security scanning to identify potential security vulnerabilities in the codebase.

## Tools Used

### Gosec
- **Purpose**: Static analysis security scanner for Go code
- **Format**: Generates SARIF (Static Analysis Results Interchange Format) files
- **Integration**: Results are uploaded to GitHub Security tab

## GitHub Actions Integration

### Permissions Required
The security scanning workflow requires specific GitHub token permissions:

```yaml
permissions:
  contents: read
  security-events: write  # Required for SARIF upload
  actions: read
```

### Workflow Steps

1. **Install Gosec**: Downloads and installs the latest version of gosec
2. **Run Security Scan**: Executes gosec with SARIF output format
3. **Check SARIF File**: Verifies the scan output was generated
4. **Display Summary**: Shows human-readable security issues in workflow logs
5. **Upload SARIF**: Uploads results to GitHub Security tab
6. **Fallback Artifact**: Saves SARIF file as workflow artifact if upload fails

### Expected Behavior

- **Exit Code 1**: Normal when security issues are found
- **Continue on Error**: Workflow continues even if SARIF upload fails
- **Fallback Storage**: SARIF files are always saved as artifacts

## Current Security Issues

As of the latest scan, the following types of security issues are detected:

- **G304**: File inclusion via variable (CLI file operations)
- **G204**: Subprocess launched with variable (browser opening functionality)
- **G601**: Implicit memory aliasing in loops (false positives)

These issues are expected for a CLI application and have been reviewed for security implications.

## Troubleshooting

### "Resource not accessible by integration"

This error occurs when:
1. Repository doesn't have GitHub Advanced Security enabled
2. Token lacks `security-events: write` permission
3. Repository is private without proper license

**Solutions**:
1. Enable GitHub Advanced Security in repository settings
2. Ensure workflow has proper permissions (already configured)
3. Check repository security settings

### SARIF Upload Failed

If SARIF upload fails:
1. Check workflow artifacts for the `gosec-sarif` file
2. Download and review SARIF content manually
3. Verify repository has security scanning enabled

### No Security Issues Found

If gosec reports no issues:
1. Verify gosec is running correctly
2. Check if code patterns changed
3. Ensure gosec rules are up to date

## Manual Security Scan

To run security scanning locally:

```bash
# Install gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run scan with SARIF output
gosec -fmt sarif -out gosec.sarif ./...

# Run scan with human-readable output
gosec ./...

# Run scan with specific confidence level
gosec -confidence medium ./...
```

## Security Policy

- All security findings are reviewed during development
- Critical and high-severity issues must be addressed before release
- Medium and low-severity issues are evaluated case-by-case
- False positives are documented and excluded appropriately

## References

- [Gosec Documentation](https://github.com/securego/gosec)
- [GitHub Security Scanning](https://docs.github.com/en/code-security/code-scanning)
- [SARIF Format](https://docs.github.com/en/code-security/code-scanning/integrating-with-code-scanning/sarif-support-for-code-scanning) 