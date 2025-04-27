# 🧪 Enraijin

*A little bit different brute-force tool.*
  
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
<h1>⚗ About this tool</h1>
<p>I'm Naru Koshin the creator of this wonderful tool. If you are still wondering what this tool is for, why I spent so much time on creating it, and what you can do with it, then I will try to explain you as simply as possible.</p>
<p>I'm studying and working as a penetration tester, and IT Security analyst, call it as you want, I'm hacking servers, but most websites, okay?</p>
<p>Most of the tools are pretty hard to use especially if you are hacking for many days. I don't like to write automated code for every project that will run a Hydra or any other tool that will brute-force passwords for me. And no, I'm not a script kiddie. I just don't like to type very long commands and then figure out why the heck the script is not working as I want. My tool is very simple to use and the config is easy to read. You can share a config file, you can store it for how long you need, etc.</p>
<p>As I mentioned before, This tool is for brute-forcing aka cracking the website passwords. There's nothing difficult to understand. Just type the data about the website, and set your options, for example, you can send a password when it will be found to the email so you can leave this tool to work on your server or somewhere else.</p>
<p>Why I'm spending this tool so much of my time? The reason is simple. I just want to crack passwords gently. Write the config, check the config, everything looks fine, start it, and wait for the password. I know how my tool works better than anyone else. If there is any bug, I'm fixing it.. or at least trying to fix it. In the previous release, I added a new awesome feature - Getting passwords in an email. This will be very useful when I'm working with the team.</p>
<p>Yes, my tool works only on websites, but it's still better than Hydra. 😂 For other protocols I'm using Ncrack.</p>
<h1>📚 Getting started</h1>
<p>To download this tool, type the command below:</p>

```sh
  git clone https://github.com/narukoshin/custom-bruteforce
```
... or download binaries from the releases page.

<h1>📅 TODO</h1>

- [x] Proxy Feature <br>
      + added in commit: <a href="https://github.com/narukoshin/custom-bruteforce/commit/ba5ab6fefc17f29476e31eae98774edc23e94815">`ba5ab6f`</a><br>
      + changelog: <a href="https://github.com/narukoshin/custom-bruteforce/releases/tag/v2.3-beta">`v2.3-beta`</a>
- [x] Idea about `import` option where you can import config file with the custom name like `import: my_website.yml` <br>
      + added in commit: <a href="https://github.com/narukoshin/custom-bruteforce/commit/823b14f907ce92a44d69174510f681ba0da31c6e">`823b14f`</a><br>
      + changelog: <a href="https://github.com/narukoshin/custom-bruteforce/releases/tag/v2.3-beta">`v2.3-beta`</a>
- [x] Email notifications <br>
      + added in commit: <a href="https://github.com/narukoshin/custom-bruteforce/commit/a98c4631dd29cfcf6d50ef45bb5b1a98b67e3aa3">`a98c463`</a><br>
      + changelog: <a href="https://github.com/narukoshin/custom-bruteforce/releases/tag/v2.4.3">`v2.4.3`</a>

💭 If you have any suggestion about new features, please open a new issue with the enhancement label.

<h1>⚙ Creating configuration</h1>
<p>Before you start using the tool, you need to create a config file called <code>config.yml</code></p>

```sh
touch config.yml
... or
vim config.yml
... or you can use pre-made config
mv config.sample.yml config.yml
```

<p>Next, you need to fill the config file with the information about your target to brute-force.</p>

```yaml
#config.yml

# You can import another config file with a custom name, for example, my-project.yml
import: my-project.yml
# after import, the following lines will be ignored.

# You can also include config by separate files
include:
      - file1.yml
      - website.com/file2.yml

site:
    host: https://website.com/login # the login page that you want to crack.
    method: POST # request method for making a request
bruteforce:
    field: password # the field that you want to brute-force (important)
    
    # There are 3 ways from where you can load a wordlist
    # method 1 - from the file
    from: file
    file: /usr/share/wordlists/rockyou.txt # the path, where is your wordlist located at
    # method 2 - from the list
    from: list
    list:
        - password1
        - password2
        - password3
    # method 3 - from the stdin
    # info: be careful with this method because of the RAM leak.
    # more info about the bug: https://github.com/narukoshin/custom-bruteforce/issues/2
    from: stdin
```
<p>When you are using <b>stdin</b> method, type the command as shown below:</p>

```sh
    # example 1
    cat /usr/share/wordlists/rockyou.txt | ./linux
    # example 2
    crunch 8 8 0123456789 | ./linux
    # ...etc
```

```yaml
#config.yml

    # Next, you need to specify how many threads you want to use. The default value is 5
    threads: 30
    
    # If you don't want to see messages like "trying password...", you can turn it off with the option below:
    # It's optional, so if you don't need to turn it off, you can skip this option
    no_verbose: true
    
    # By default, when the tool finds out the password, the password will be printed on the screen, 
    # ...but if you want you can set it to save in the file
    output: /home/naru/my_target/password.txt
    
# Setting the headers
# There's nothing difficult
headers:
    - name: Content-Type
      value: application/x-www-form-urlencoded; charset=utf-8
    - name: User-Agent
      value: Mozilla/5.0 (X11; U; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4747.211 Safari/537.36
      
# Setting the static fields
fields:
    # Setting the username that we want to brute-force
    - name: username  # the input name
      value: admin
      
# Adding error message if the password is incorrect
# p.s. This will be ignored if you add on_pass option
on_fail:
    message: incorrect password
    status_code: 401 # default value is 0
    
# Adding the successful message, if, for example, we are in the admin panel
on_pass:
    message: Welcome, 
    status_code: 200 # default value is 200
    
# And the last cherry of this tool is the crawl option
# This option can help you find the token if there is any and will add it to your request
crawl:
    url: <token-url> # If the token is not located in the original request, then we will set a new one to get the token
    name: token # the name of the field where the token will be passed to the request
    search: "token = '([a-z0-9]{32})" # to find the token, use regex

# To apply proxy setting use this option.
proxy:
    # v1 way (deprecated, soon will be removed)
    socks: socks5://127.0.0.1:9050?timeout=5s # for Tor proxy

    # v2 way
    addr: socks5://127.0.0.1:9050
    timeout: 5s # optional
    verify_url: http://httpbin.org/ip # custom URL for checking if the proxy is working properly
    
# email settings
email:
  # Email settings that will send the email
  server:
    host: your.server.name
    port: 587
    timeout: 3 # default 30
    email: your.email@address.com
    password: your.password123
  # mail settings
  mail:
    # method 1
    recipients: your.email@address.com
    
    # method 2 for multiple recipients
    recipients:
      - your.email@address.com
      - another.email@address.com
      - random.email@address.com
    subject: Your subject text is here
    name: Who Am I?
    message: "Password: <password>" # a real password will appear in <password> place.
```
<p>And that's it, now you are a professional cracker.</p>
