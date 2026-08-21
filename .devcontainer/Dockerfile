FROM mcr.microsoft.com/devcontainers/go:1-1.24-bookworm

# Install OpenTofu and Terraform CLI
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    gnupg \
    unzip \
    jq \
    && rm -rf /var/lib/apt/lists/*

# Install OpenTofu
RUN curl -fsSL https://get.opentofu.org/opentofu.gpg | tee /etc/apt/trusted.gpg.d/opentofu.gpg >/dev/null \
    && echo "deb [signed-by=/etc/apt/trusted.gpg.d/opentofu.gpg] https://packages.opentofu.org/opentofu/tofu/any/ any main" | tee /etc/apt/sources.list.d/opentofu.list \
    && apt-get update && apt-get install -y tofu \
    && rm -rf /var/lib/apt/lists/*

# Install Terraform CLI
RUN curl -fsSL https://apt.releases.hashicorp.com/gpg | gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg \
    && echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | tee /etc/apt/sources.list.d/hashicorp.list \
    && apt-get update && apt-get install -y terraform \
    && rm -rf /var/lib/apt/lists/*
