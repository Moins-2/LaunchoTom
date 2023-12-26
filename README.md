The goal of this tool is to run commands easily in background. They are still accessible, but still run even if the terminal close. All of this using [tmux]().
Thought to be customizable, all commands and used tools are modifiable.

# Installation

Follow go instructions at first launch

# Usage

Launch command :

```bash
go run .
```

# Configuration

The file containing commands is scan.conf.json. It is a json file constructed like that:

```json
{
  "title": "L37's H4CK!",
  "items": [
    {
      "title": "FFUF",
      "content": "Finding Web Hidden Treasures",
      "items": [
        { "title": "Common (quick)", "content": "ffuf -config ./ffufrc-common.toml" },
        { "title": "Files", "content": "ffuf -config ./ffufrc-files.toml" },
        { "title": "Directories", "content": "ffuf -config ./ffufrc-dir.toml" }
      ]
    },
    {
      "title": "nmap",
      "content": "Sherlock of Ports",
      "items": [
        { "title": "Vanilla", "content": "echo 'vanilla'" },
        { "title": "TCP", "content": "sudo nmap -sS -sV -PN -T4 -p- -iL ./short_url.txt -oA nmap_TCP_scan" },
        { "title": "UDP", "content": "sudo nmap -sU -sV -Pn -T4 -iL short_url.txt -oA nmap_UDP_scan" }
      ]
    },
  ]
}
```

You can add as many sections as you want to.
