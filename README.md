# 🧪 Enraijin

*A small, flexible web brute-force framework — designed for long runs, clarity, and easy configuration.*

<img src="https://c.tenor.com/gOP4dRPvzWcAAAAi/angry-mafumafu.gif" align="right" width="160">
<div>
  <img src="https://img.shields.io/github/go-mod/go-version/narukoshin/custom-bruteforce">
  <img src="https://img.shields.io/github/v/release/narukoshin/custom-bruteforce">
  <img src="https://img.shields.io/github/last-commit/narukoshin/custom-bruteforce">
  <img src="https://img.shields.io/github/contributors/narukoshin/custom-bruteforce">
  <br><br>
  <div>
    <a target="_blank" href="https://twitter.com/enkosan_p"><img src="https://media4.giphy.com/media/iFUiSYMNPvIJZDpMKN/giphy.gif?cid=ecf05e471v5jn6vuhczu1tflu2wm7qt11atwybfwcgaqxz38&rid=giphy.gif&ct=s" align="middle" width="120"></a>
    <a target="_blank" href="https://instagram.com/enko.san"><img src="https://media1.giphy.com/media/Wu9Graz2W46frtHFKc/giphy.gif?cid=ecf05e47h46mbuhq40rgevni5rbxgadpw5icrr71vr9nu8d4&rid=giphy.gif&ct=s" align="middle" width="120"></a>
  </div>
</div>

Enraijin is a focused tool for automating credential brute-force against web forms. It prioritizes readable configuration, reliable proxy support, token crawling, and convenient notifications so you can set it up once and run long tests without constant babysitting.

> Important: Use this tool only against systems you own or have explicit permission to test. Unauthorized access is illegal. Always follow your organization’s rules and applicable laws.

---

📚 Table of contents
- About
- Quick start
- Configuration (simple → advanced)
- Plugins (Agentix, Proxmania)
- Common workflows & examples
- Proxy & crawl details
- Email notifications
- Troubleshooting & tips
- TODO & changelog
- Contributing & license

---

# ⚗ About this tool

Hi — I'm Naru Koshin, author of Enraijin. I built this to make web brute-force runs easier to manage and repeatable across engagements. Many pentest tasks require the same patterns (form fields, tokens, proxies, output), and copying ad-hoc scripts for each target gets tedious and error-prone. Enraijin gives you a single, human-editable YAML config to cover most use cases with sensible defaults.

If you need protocol-level cracking (FTP, SSH, RDP, etc.), consider Ncrack or other specialized tools — Enraijin focuses on HTTP(S) web forms.

---

# 📚 Quick start

Clone the repo:

```sh
git clone https://github.com/narukoshin/custom-bruteforce
cd custom-bruteforce
```

Build or download a release from the Releases page, then run the binary:

- Run a local binary (Linux example):
  ```sh
  ./enraijin
  ```

- Install via `go install` (recommended if you use Go toolchain):
  ```sh
  go install github.com/narukoshin/EnRaiJin/v2@latest
  ```
  After `go install`, the binary will be placed in `$GOBIN` (or `$GOPATH/bin` if `$GOBIN` is not set). Run it as `EnRaiJin` or by full path.

If the repository includes platform-specific binaries, pick the one for your OS.

---

# ⚙ Creating configuration

Enraijin uses a single YAML config file (default name: `config.yml`). The file is intentionally straightforward — fill in your target, choose a wordlist source, and set a few optional behaviors.

Minimum config example (simple):

```yaml
# config.yml
site:
  host: "https://example.com/login"
  method: POST

bruteforce:
  field: password        # the form input name to brute-force
  from: file             # 'file' | 'list' | 'stdin'
  file: /usr/share/wordlists/rockyou.txt
  threads: 5
```

Advanced, annotated example (recommended for real runs):

```yaml
# config.yml

# import another config file (if present, this file is ignored)
# import: my-project.yml

# include additional partial configs (merged)
# import:
#   - common-headers.yml
#   - site-specific.yml

site:
  host: "https://website.com/login"   # login URL (or a page that accepts the auth request)
  method: POST                        # HTTP method used for auth POST/GET

# bruteforce options
bruteforce:
  field: password                     # name of the input to brute-force
  from: file                          # file | list | stdin
  file: /usr/share/wordlists/rockyou.txt
  # OR
  # from: list
  # list:
  #   - P@ssw0rd
  #   - password123
  # OR (stdin)
  # from: stdin

  threads: 30                         # number of concurrent attempts (default: 5)
  no_verbose: false                   # true to silence "trying password..." lines
  output: /home/naru/results/passwords.txt  # save successful credentials

# static fields included with each attempt (e.g., username)
fields:
  - name: username
    value: admin

# add or override headers sent with requests
headers:
  - name: Content-Type
    value: application/x-www-form-urlencoded; charset=utf-8
  - name: User-Agent
    value: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36"

# detect failed attempts by message or by status code
on_fail:
  message: "incorrect password"
  status_code: 401

# detect successful login by message or by status code
on_pass:
  message: "Welcome,"
  status_code: 200

# crawl + token extraction (useful for CSRF tokens)
crawl:
  url: "https://website.com/login"                 # optional: fetch another page first
  name: token                                      # form field to include in the request
  search: "token = '([a-z0-9]{32})"                # regex to extract token

# proxy support
proxy:
  # legacy (v1) example (deprecated)
  socks: "socks5://127.0.0.1:9050?timeout=5s"

  # v2 (recommended)
  addr: "socks5://127.0.0.1:9050"
  timeout: 5s
  verify_url: "http://httpbin.org/ip"  # optional check URL

# email alerts on success
email:
  server:
    host: smtp.example.com
    port: 587
    timeout: 3
    email: your.email@example.com
    password: your.smtp.password
  mail:
    recipients:
      - you@example.com
      - team@example.com
    subject: "Enraijin: password found"
    name: "Enraijin"
    message: "Password: <password>"
```

Notes:
- "from" selects where candidate passwords come from:
  - file: path to a wordlist
  - list: inline small lists (useful for quick checks)
  - stdin: pipe from other programs like crunch (be careful with memory and long runs)
- If both message and status_code are provided in on_pass/on_fail, both are used to determine outcome (configurable per version).

---

# 🔁 Importing, include, and deprecated parameter

- import: completely replaces the current config with another file (useful for managed projects).

Deprecated parameter
- The older parameter named `include` is deprecated and has been fully replaced by `import`.
- If you see `include: <file>` in your configs, change it to `import: <file>`.

Migration example:

Old (deprecated):
```yaml
# deprecated usage — DO NOT use
include: my-project.yml
```

New (use this):
```yaml
# preferred
import: my-project.yml
```

---

# 🔌 Plugins: loading, Agentix, and Proxmania

Enraijin supports loading plugins as shared objects (.so). Plugins are added under the `bruteforce` section using the `plugins` key. Two formats are accepted:

- Single plugin (string):
```yaml
bruteforce:
  plugins: ./plugins/proxmania/proxmania.so
```

- Multiple plugins (list):
```yaml
bruteforce:
  plugins:
    - ./plugins/proxmania/proxmania.so
    - ./plugins/agentix/agentix.so
```

When the binary loads plugins, it will attempt to initialize them. Plugin availability may depend on how the binary was built (plugins may be optional in some releases). Plugins should be placed in a plugins directory inside the project or referenced by absolute path.

Plugin behavior and configuration:
- Plugins may expose their own configuration sections at the top level of the YAML (example: `proxmania:`). The plugin loader will read those sections and pass them to the plugin during initialization.
- If a plugin does not implement configuration handling yet, it will run with built-in defaults.

Proxmania (example plugin configuration)
- Purpose: fetch, validate, and rotate proxies from an external source (example uses a public dataset URL).

Example config:
```yaml
proxmania:
  # URL to fetch the SOCKS5 proxy data set
  proxy_data_set: "https://raw.githubusercontent.com/proxifly/free-proxy-list/refs/heads/main/proxies/protocols/socks5/data.txt"

  # maximum number of proxies to use
  max_proxies: 15

  # timeout for each proxy request
  timeout: 30s
```

Notes on Proxmania:
- The plugin will download the proxy list from `proxy_data_set`, validate proxies (respecting `timeout`), and keep up to `max_proxies` in the local pool.
- The plugin typically hands proxies to Enraijin, which then assigns them to threads. Check plugin logs for detailed behavior and rate-limit handling.
- Keep API/data URLs and provider limits in mind; do not overload public services.

Agentix (current status)
- Purpose: rotate user-agents, randomize headers, and add per-agent session handling (reduces fingerprinting).
- Configuration: at the time of this writing, Agentix has no configuration implemented — if you load the plugin, it will run with built-in defaults. The plugin is being prepared for configurable options (user agent lists, rotation mode, jitter, etc.) in a future release.

Security & best practices
- Keep secrets (API keys or provider credentials) out of committed configs. Use environment variables or an external secrets manager if possible.
- Start with small thread counts and small proxy pools when enabling rotation. Validate plugin behavior in a short smoke test before long runs.

---

# Common workflows & examples

1. Quick run with a local wordlist:
   ```sh
   ./enraijin
   ```

2. Pipe from stdin (e.g., crunch):
   ```sh
   crunch 8 8 0123456789 | ./enraijin
   ```

3. Use Tor via SOCKS5 (ensure Tor is running):
   ```yaml
   proxy:
     addr: "socks5://127.0.0.1:9050"
     timeout: 5s
   ```

4. Save results to a file:
   ```yaml
   bruteforce:
     output: /home/me/targets/siteA.txt
   ```

---

# Proxy, crawling & token extraction

- Proxy support is built-in with a v2 configuration (addr + timeout + verify_url). A legacy v1 `socks` option exists but is deprecated.
- Crawl option fetches a page (may be the same as the host or a separate URL), runs a regex against the response to extract a token, then injects the token into the configured `name` for subsequent requests.
- Regex must be a quoted pattern; capture group 1 will be used.

Example crawl config:
```yaml
crawl:
  url: "https://website.com/session"
  name: csrf_token
  search: "<input name=\"csrf_token\" value=\"([a-z0-9]{32})\""
```

---

# Email notifications

Configure your SMTP server and recipient list to receive an email when a credential is found. Replace placeholders with secure credentials and consider using app-specific passwords or a throwaway relay for testing.

---

# Troubleshooting & tips

- If candidates from stdin stop abruptly, check resource usage — piping massive lists can consume memory depending on version/platform.
- Use conservative `threads` on production web servers to avoid crashing services and to reduce the chance of being blocked.
- Use `no_verbose: true` for long runs to reduce stdout noise and log only successes.
- Add `verify_url` under `proxy` to ensure proxy checks before brute-forcing.
- When using plugins, run a short dry-run to confirm plugin behavior (rotation, pool size, API limits). Check plugin logs/outputs for initialization details.
- If a plugin fails to load, verify the .so path and that the binary supports plugin loading (some releases may be built without plugin support).

---

# 📅 TODO & changelog (high level)

- [x] Proxy Feature
  - commit: ba5ab6f (see Releases)
  - changelog: v2.3-beta
- [x] Import config option
  - commit: 823b14f
  - changelog: v2.3-beta
- [x] Email notifications
  - commit: a98c463
  - changelog: v2.4.3
- [ ] Agentix plugin: add configuration support (planned)
- [ ] Improve plugin docs and examples in releases

If you have feature suggestions, please open an issue with the "enhancement" label.

---

# Contributing

Contributions, bug reports, and ideas are welcome. Open an issue or submit a PR. Please include reproducible steps and config files (redacting any secrets).

---

# License

Check the LICENSE file in the repository for license details.

---

Thanks for trying Enraijin — keep things legal and responsible.
