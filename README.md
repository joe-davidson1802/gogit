# gogit

Scrape a git org for go package versions


```
gogit -pat-path ~/.gogitpat myorg main github.com/gorilla/mux
```

This will produce a list of repos:

```
my-api                                         v1.8.0
my-front-end                                   none
other-api                                      v.1.7.0
```

## Installation 

Install the package:

```
go install github.com/joe-davidson1802/gogit/cmd/gogit@latest
```

(Optional) Configure your GitHub Personal Access Token.

- Generate a new PAT <https://github.com/settings/tokens>
- Save the new pat in `~/.gogit
