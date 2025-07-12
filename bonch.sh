# Build the YAML in the CC variable
CC=$(cat <<'EOF'
#cloud-config
runcmd:
  - curl -fsSL https://bun.sh/install | bash
  - ln -s /root/.bun/bin/bun /usr/local/bin/bun
  - bun add chalk
  - echo 'import chalk from "chalk"; console.log(chalk.red("Hello from cloud-init!"))' > /root/index.ts
EOF
)

# Base64-encode in one line, strip newlines (-w0; if your base64
# lacks -w use `tr -d '\n'` afterwards)
UD=$(printf '%s' "$CC" | base64 -w0)

echo $UD