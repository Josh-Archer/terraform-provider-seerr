---
description: How to run and validate OpenTofu integration tests for the Seerr provider
---

This workflow describes how to run the OpenTofu integration tests located in the `tests/` directory.

### Prerequisites

1.  **Overseerr/Jellyseerr Instance**: You must have a running instance accessible via HTTP.
2.  **API Key**: You must have a valid API key for the admin user.
3.  **Local Provider Build**: The provider binary should be built and discoverable by OpenTofu.

### Steps

1.  Build the provider binary:
    // turbo

    ```powershell
    go build -o terraform-provider-seerr.exe
    ```

2.  Navigate to the tests directory:

    ```powershell
    cd tests
    ```

3.  Initialize OpenTofu (ensure dev_overrides are configured in your `.tofurc` if running locally):

    ```powershell
    tofu init
    ```

4.  Run the tests by providing the required variables:
    ```powershell
    tofu test -var="url=${SEERR_URL}" -var="api_key=${SEERR_API_KEY}"
    ```

### Validation

- Ensure all `run` blocks in the `.tftest.hcl` files pass.
- Check assertion messages for any failures related to IP, port, or permissions.
