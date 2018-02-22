# About `hc`

`hc` is a command-line tool that runs headless Chrome in isolated
Docker containers (which are automatically created and destroyed)
for safe and reproducible browser automation and data extraction tasks:

1. Generating HTML snapshots for static and dynamically rendered pages;
2. Downloading XHR resources;
3. Evaluating and capturing the output of arbitrary JavaScript code;
4. Generating screenshots.

**Consider this utility EXPERIMENTAL. The list of commands,
their behavior and invocation syntax may change in the future.**

## Quick examples

Output the rendered HTML of the page (as the browser sees it after building the page):

```sh
hc html "http://example.com/"
```

Output the number of paragraphs on a page:

```sh
hc eval "http://example.com/" "return document.getElementsByTagName('p').length"
```

Make a screenshot:

```sh
hc screenshot "http://example.com/" >out.png
```

See more examples below.

## Advantages

1. Ease of deployment: `hc` compiles into a binary with no runtime dependencies
   apart from Docker;

2. Ease of use in shell scripts or other scripting languages;

3. Unix way of working with the data: pipe the fetched HTML, JSON or binary
   resources through other streaming tools. `hc` maintains a clean separation
   between data (STDOUT) and logging / error reporting (STDERR),
   and uses meaningful process exit codes;

4. Security and reproducibility: `hc` uses Docker to temporarily spin up and
   shut down the container for each command invocation. This guarantees that
   Chrome starts in the same clean state when running a command (think of it
   as a per-command incognito mode);

5. Resiliency: if the script doesn't finish execution within a given deadline,
   it is shut down automatically, and the container is killed, so your scripts
   never get stuck;

6. Headless Chrome container runs only for the duration of the command execution,
   so when you don't need it, it doesn't waste your system resources.

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

## Evaluating JavaScript on a page

Save the list of href attributes of all the links on the page to a file:

```sh
$ hc eval \
    --output-file "links-{TIMESTAMP}.txt" \
    "https://httpbin.org/" \
    "return Array(...document.getElementsByTagName('a')).map(el => el.getAttribute('href')).join('\n')"
```

Here the output is redirected to a file with `links-{TIMESTAMP}.txt` name template;
`{TIMESTAMP}` will be replaced automatically with the current date and time
in `YYYY-MM-DD-hh-mm-ss` format, so the final file name will look like this:
`links-2018-02-12-15-34-59.txt`

## Get the contents of a web page

Output the rendered HTML document:

```sh
$ hc html "https://httpbin.org/status/418"
```

The command above is equivalient to:

```sh
$ hc eval "https://httpbin.org/status/418" "return document.documentElement.outerHTML"
```

If you need just the contents of the <body> tag, use:

```sh
$ hc eval "https://httpbin.org/status/418" "return document.body.innerHTML"
```

## Load a resource in the context of a web page

Note how this method is different from loading the resource URL directly:
the resource is loaded by the host page itself, with proper headers and
cookies, and `hc` just captures its content. This allows for easy capturing
of XHR resources.

Output the value of the resource with the exact URL match:

```sh
$ hc resource "http://example.com/" "http://example.com/xhr/someData.js"
```

Output the value of the first resource with the URL starting with a given prefix:

```sh
$ hc resource --match contains "https://httpbin.org/" "tracker.js"
```

Output the value of the first resource with the URL matching a given
regular expression (here the resource is a binary file, so the best option is
to redirect the output to a file, or use the `--output-file` flag as described
in one of the previous examples):

```sh
$ hc resource --match regexp https://httpbin.org/ "forkme.*?\.png" > ~out.png
```

## Save a screenshot of a web page

Make screenshot of a web page and save it to `out.png`:

```sh
hc screenshot "http://example.com/" >out.png
```

When rendering the page, viewport size is set to 1024x768 by default. The final
dimensions of the screenshot are determined by the page content, but you can
control the initial size to imitate different devices:

```sh
hc screenshot --initial-width 800 --initial-height 600 "http://example.com/" >out.png
```

In the command above the initial viewport size is set to 800x600 prior to
rendering the page.

In addition to limiting the initial viewport size, there's an option to limit
the maximum viewport size:

```sh
hc screenshot --max-width 1000 --max-height 1000 "http://example.com/" >out.png
```

Here the maximum viewport size is limited to 1000x1000px. If the content doesn't fit
in this viewport, scrollbars will appear on the screenshot.

When no maximum height or width are defined, the viewport size will be adjusted
to accommodate the content so that an entire page is captured without scrollbars.

# Feedback

Feel free to provide your feedback, suggestions or bug reports here in the <a href="https://github.com/iafan/hc/issues">bug tracker</a>, or message [@afan](https://gophers.slack.com/messages/@afan/) in the [Gophers Slack channel](https://gophersinvite.herokuapp.com/).

# Credits

1. `godet` library (Remote client for Chrome DevTools): Copyright (c) 2017 Raffaele Sena [[link](https://github.com/raff/godet)]
2. `justinribeiro/chrome-headless` (Headless Chrome Docker image): Copyright (c) 2015 Justin Ribeiro [[link](https://hub.docker.com/r/justinribeiro/chrome-headless/)]
3. `chrome.json` seccomp descriptor file: Copyright (c) 2015 Jessie Frazelle
   [[link](https://github.com/jessfraz/dotfiles/blob/master/etc/docker/seccomp/chrome.json)]
