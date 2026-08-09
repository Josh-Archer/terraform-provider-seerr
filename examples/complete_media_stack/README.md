# Complete Media Stack OpenTofu / Terraform Example

This example demonstrates how to configure an entire **Seerr / Jellyseerr / Overseerr** instance using OpenTofu or Terraform Infrastructure as Code (IaC).

## Features Configured
- **Application General Settings**: Title, base URL, trust proxy.
- **Sonarr Integration**: Connects TV series downloader service with default quality profile and root path.
- **Radarr Integration**: Connects movie downloader service with minimum availability rules.
- **User Permission Tiers**: Configures full Admin users and Power Users with custom request quotas.
- **Notification Agents**: Sets up Discord Webhook and Pushover integration for real-time request alerts.

## Quickstart

1. **Copy example variables file**:
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   ```

2. **Fill in your server URLs and API Keys** in `terraform.tfvars`.

3. **Initialize OpenTofu / Terraform**:
   ```bash
   tofu init
   # or
   terraform init
   ```

4. **Apply Configuration**:
   ```bash
   tofu apply
   ```
