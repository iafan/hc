# About `hc`

### _Headless Chrome automation for your command line._

**This utility should be treated as EXPERIMENTAL.
The list of commands, their behavior and invocation syntax may change in the future.**

```sh
hc eval "http://example.com/" "return document.getElementsByTagName('p').length"

# Output: `2`
```

### Advantages

1. Ease of deployment: `hc` compiles into a binary with no system dependencies;

2. Ease of use in shell scripts or other scripting languages;

3. Unix way of working with the data: pipe the fetched HTML or JSON
   through other streaming tools. `hc` maintains a clean separation
   between data (STDOUT) and logging / error reporting (STDERR),
   and uses meaningful process exit codes;

4. Security and reproducibility: `hc` uses Docker to temporarily spin up and
   shut down the container for each command invocation. This guarantees that
   Chrome starts in the same clean state when running a command (think of it
   as a per-command incognito mode);

5. Resiliency: if the script doesn't finish execution within a given deadline,
   it is shut down automatically, and the container is killed, so your scripts
   never get stuck;

6. Resource-friendliness. Container runs only when needed, so when you
   don't use headless Chrome, it doesn't waste your system resources.

# Installation

```sh
$ go get github.com/iafan/hc
```

# Prerequisites

[Docker](https://www.docker.com/community-edition) and [Go](https://golang.org/dl/).
As for the actual Docker image, `hc` uses [justinribeiro/chrome-headless](https://hub.docker.com/r/justinribeiro/chrome-headless/) by default
(which will be installed automatically). If you prefer some other image,
use the `--docker-image` command-line flag.

**Note:** The first time you run some `hc` command that requires headless Chrome,
Docker will download and install the missing image. Please be patient.

# Examples

## Load a page, then evaluate JavaScript code and save the result to a file

```sh
$ hc eval \
    --output-file "links-{TIMESTAMP}.txt" \
    "https://httpbin.org/" \
    "return Array(...document.getElementsByTagName('a')).map(el => el.getAttribute('href')).join('\n')"

# Output: the list of href attributes of all the links on the page will be saved to a file.
```

Here the output is redirected to a file with `links-{TIMESTAMP}.txt` name template;
`{TIMESTAMP}` will be replaced automatically with the current date and time
in `YYYY-MM-DD-hh-mm-ss` format, so the final file name will look like this:
`links-2018-02-12-15-34-59.txt`

## Get the contents of a web page

```sh
$ hc html "https://httpbin.org/status/418"

# Output: the rendered HTML document.


# The command above is equivalient to:

$ hc eval "https://httpbin.org/status/418" "return document.documentElement.outerHTML"


# If you need just the contents of the <body> tag, use:

$ hc eval "https://httpbin.org/status/418" "return document.body.innerHTML"
```

## Load a resource in the context of a web page

Note how this method is different from loading the resource URL directly:
the resource is loaded by the host page itself, with proper headers and
cookies, and `hc` just captures its content. This allows for easy capturing
of XHR resources.

```sh
$ hc resource "http://example.com/" "http://example.com/xhr/someData.js"

# Output: the value of the resource with the exact URL match.


$ hc resource --match contains "https://httpbin.org/" "tracker.js"

# Output: the value of the first resource with the URL starting with a given prefix.


$ hc resource --match regexp https://httpbin.org/ "forkme.*?\.png" > ~out.png

# Output: the value of the first resource with the URL matching
# a given regular expression. Here the resource is a binary file,
# so the best option is to redirect the output to a file, or use the
# `--output-file` flag as described in one of the previous examples.
```

## Feedback

Feel free to provide your feedback, suggestions or bug reports here in the <a href="https://github.com/iafan/hc/issues">bug tracker</a>, or message [@afan](https://gophers.slack.com/messages/@afan/) in the [Gophers Slack channel](https://gophersinvite.herokuapp.com/).

# Credits

1. `godet` library (Remote client for Chrome DevTools): Copyright (c) 2017 Raffaele Sena [[link](https://github.com/raff/godet)]
2. `justinribeiro/chrome-headless` (Headless Chrome Docker image): Copyright (c) 2015 Justin Ribeiro [[link](https://hub.docker.com/r/justinribeiro/chrome-headless/)]
3. `chrome.json` seccomp descriptor file: Copyright (c) 2015 Jessie Frazelle
   [[link](https://github.com/jessfraz/dotfiles/blob/master/etc/docker/seccomp/chrome.json)]
