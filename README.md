# snoo

<p align="center">
  <img src="https://raw.githubusercontent.com/snoofox/snoo/main/assets/space-gopher.png " alt="snoo gopher" style="max-width: 100%; height: auto;"/>
</p>

A terminal feed reader that doesn't suck (yet).

## Install

```bash
# needs cgo for sqlite
CGO_ENABLED=1 go install github.com/snoofox/snoo@latest
```

Or build yourself:

```bash
git clone https://github.com/snoofox/snoo
cd snoo
go build
```

## Preview
![snoo preview](https://raw.githubusercontent.com/snoofox/snoo/main/assets/demo.gif )

## Quick start

Add feeds:

```bash
# reddit
snoo sub add golang
snoo sub add rust:new
snoo sub add programming:top

# rss
snoo sub rss https://lwn.net/headlines/rss
snoo sub rss https://hnrss.org/frontpage

# lobsters
snoo sub lobsters active
```

Read:

```bash
snoo
```

Pick a theme (optional):

```bash
snoo theme dracula
```

## Commands

### Manage subscriptions

```
snoo sub add <subreddit>[:sort]     # reddit
snoo sub rss <url>                  # any rss/atom
snoo sub lobsters active|recent     # lobsters
snoo sub list                       # show all
snoo sub rm <id>                    # remove one
```

### View feed

```
snoo              # same as 'snoo feed'
snoo feed
```

### Themes

```
snoo theme        # list
snoo theme <name> # set
```

Available: default, catppuccin, dracula, github, peppermint

## Navigation

### Feed list:

```
j/k         move
Enter       open post
f           filter sources
s           sort posts
q           quit
```

### Inside post:

```
j/k         scroll
g/G         jump to top/bottom
r           read full article
s           sort comments
Esc         back
q           quit
```

### Filter menu:

```
Space       toggle source
a           enable all
d           disable all
Esc         back
```

## Details

- Stores everything in `data.sqlite3`
- Caches posts for 1 hour
- No login required

## Contributing
If you are interested in contributing to snoo, please feel free to submit a pull request or open an issue on our GitHub repository.

## License
```
MIT License

Copyright (c) 2022

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## Thanks For Visiting
Hope you liked it. Wanna support?

- **[Star This Repository](https://github.com/snoofox/snoo)**
- **[Buy Me A Coffee](https://www.buymeacoffee.com/shoto)**
