# Installing Plannet on Linux

## Quick Install (Recommended)

### For x86_64 (AMD64)

```bash
# Download the latest release
curl -L https://github.com/plannet-dev/plannet/releases/latest/download/plannet-linux-amd64 -o plannet

# Move to /usr/local/bin (requires sudo)
sudo mv plannet /usr/local/bin/

# Verify installation
plannet --version
```

### For ARM64

```bash
# Download the latest release
curl -L https://github.com/plannet-dev/plannet/releases/latest/download/plannet-linux-arm64 -o plannet

# Move to /usr/local/bin (requires sudo)
sudo mv plannet /usr/local/bin/

# Verify installation
plannet --version
```

## Alternative Installation Methods

### Manual PATH Setup

If you prefer not to use `/usr/local/bin`, you can add Plannet to your PATH:

1. Create a directory for Plannet:
```bash
mkdir -p ~/bin
```

2. Download and move Plannet:
```bash
# For x86_64
curl -L https://github.com/plannet-dev/plannet/releases/latest/download/plannet-linux-amd64 -o ~/bin/plannet

# For ARM64
curl -L https://github.com/plannet-dev/plannet/releases/latest/download/plannet-linux-arm64 -o ~/bin/plannet
```

3. Add to your shell configuration:

For bash (add to `~/.bashrc`):
```bash
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

For zsh (add to `~/.zshrc`):
```bash
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

For fish (add to `~/.config/fish/config.fish`):
```fish
echo 'set -gx PATH $PATH $HOME/bin' >> ~/.config/fish/config.fish
source ~/.config/fish/config.fish
```

4. Verify installation:
```bash
plannet --version
``` 