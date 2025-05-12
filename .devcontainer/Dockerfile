FROM mcr.microsoft.com/devcontainers/go:1-1.24-bookworm

ENV GOPATH=/home/vscode/go \
    PATH="/home/vscode/go/bin:${PATH}"

RUN mkdir -p /home/vscode/go && chown -R vscode:golang /home/vscode/go

# Tool Versions
ENV TASK_VERSION=v3.43.3 \
    GOLANG_CI_LINT_VERSION=v2.1.5 \
    GOLANG_COBRA_CLI_VERSION=v1.3.0 \
    GITHUB_CLI_VERSION=2.72.0 \
    SEMANTIC_RELEASE_VERSION=v24.2.3

# Install common
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash -
RUN apt-get update && apt-get install -y apt-transport-https ca-certificates curl gnupg nodejs

# Install task
RUN curl -fsSL https://taskfile.dev/install.sh | sh -s -- -b /usr/local/bin ${TASK_VERSION}

# Install gh
RUN curl -fsSL https://github.com/cli/cli/releases/download/v${GITHUB_CLI_VERSION}/gh_${GITHUB_CLI_VERSION}_linux_$(dpkg --print-architecture).tar.gz | tar -xzv -C /usr/local/bin --strip-components=2 gh_${GITHUB_CLI_VERSION}_linux_$(dpkg --print-architecture)/bin/gh

# Install semantic-release
RUN npm install -g \
    semantic-release@${SEMANTIC_RELEASE_VERSION} \
    @semantic-release/changelog@6.0.3 \
    @semantic-release/git@10.0.1 \
    @semantic-release/github@11.0.2 \
    @semantic-release/commit-analyzer@13.0.1 \
    @semantic-release/release-notes-generator@14.0.3 \
    conventional-changelog-conventionalcommits@8.0.0

# Install spectral
RUN npm install -g @stoplight/spectral-cli

# Install powerlevel10k theme for zsh
RUN git clone --depth=1 https://github.com/romkatv/powerlevel10k.git /home/vscode/.oh-my-zsh/custom/themes/powerlevel10k
RUN echo 'source /home/vscode/.oh-my-zsh/custom/themes/powerlevel10k/powerlevel10k.zsh-theme' >> /home/vscode/.zshrc
RUN echo 'POWERLEVEL9K_DISABLE_CONFIGURATION_WIZARD=true' >> /home/vscode/.zshrc

WORKDIR /workspaces/n8n-cli
USER vscode:golang

# Install golangci-lint
RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANG_CI_LINT_VERSION}

# Install cobra-cli
RUN go install github.com/spf13/cobra-cli@${GOLANG_COBRA_CLI_VERSION}

# Auto-completions
RUN echo '# Enable autocompletion' >> /home/vscode/.zshrc
RUN echo 'autoload -Uz compinit && compinit' >> /home/vscode/.zshrc
RUN echo 'source <(task --completion zsh)' >> /home/vscode/.zshrc
RUN echo 'source <(gh completion -s zsh)' >> /home/vscode/.zshrc
RUN echo 'source <(golangci-lint completion zsh)' >> /home/vscode/.zshrc
