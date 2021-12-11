# gmi-utils

A collection of small command-line utilities for working with Gemini capsules (see https://gemini.circumlunar.space/ or gemini://gemini.circumlunar.space/).

Name|Description
---|---
`gmiget`|Retrieves a given Gemini page
`gmifmt`|Formats a gemini page supplied on `stdin` or a file, allowing you to set display margins and colours

They are designed to be chained together in classic UNIX-style, for example:

```
$ gmiget gemini://gemini.circumlunar.space/ | gmifmt -m 5 | less -r
```

These are primarily written to serve my day-to-day needs when Gemini browsing, whilst scratching an itch to write code to handle the protocol. Your mileage may vary, as they say.

## gmiget
`gmiget` retrieves a single specified Gemini page, and outputs it to `stdout`.

It currently sits somewhere between a basic and advanced client as defined in the [Gemini protocol specification](https://gemini.circumlunar.space/docs/specification.gmi), but over time it will move further toward a more fully-featured client (well, as far as a non-interactive utility allows). 

## gmifmt
`gmifmt` formats gemtext supplied via `stdin` or a given file, applying margins and colourising output via a simple configuration file.

### Configuring gmifmt
`gmifmt` looks for a configuration file in the following locations in the listed order:
* `${XDG_CONFIG_HOME}/gemini/.gmifmtconf`
* `${HOME}/.config/gemini/.gmifmtconf`
* `${HOME}/.gmifmtconf`

You can specify a specific file with the `-f`/`--file` flags if preferred, which will take precedence over the above locations.

The configuration file is very simple - a small list of key-value pairs where the keys indicate which Gemtext line types can be configured, and the values are hex colour values of the form `#rrggbb`. The following key values may be used:

Key|Description
---|---
`header`|Colour for Header 1 lines (`#`)
`header2`|Colour for Header 2 lines (`##`)
`header3`|Colour for Header 3 lines (`###`)
`preformatted`|Colour for preformatted lines
`quoted`|Colour for Quoted text
`link`|Colour for URLs (applies to link text and URL)

For example, here is a sample `gmifmt` configuration colouring some Gemtext output in Nord colours:

```
header=#81a1c1
header2=#88c0d0
header3=#88c0d0
preformatted=#ebcb8b
quoted=#a3be8c
link=#5e81a
```
