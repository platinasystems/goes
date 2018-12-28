# Platina Documentation Style Guide

## Overview

This document contains a quick reference for formatting and style in the user-facing documentation. This is to try to maintain consistency.

For better or for worse, documentation is currently published in GitHub Format Markdown format. A handy reference for GFM is [available here](https://github.com/adam-p/markdown-here/wiki/Markdown-Cheatsheet). This document covers the Platina-specific use of the syntax described there.

## Nomenclature

1. *Platina Appliance Head* or *Platina Appliance Head Device* refers to the whole packaged Platina ToR switch device.
1. *GOES* refers to the GOES software package/service. The use of *GoES* is deprecated. Use `goes` when referring specifically to the command used on the command line.

## Markdown

### Headings

Use headings and subheadings rather than bold/italics or large font size. These are denoted using hash '#' characters.

Also remember to update the table of contents when modifying headings.

### Code/Command examples and output

Code or command examples should be formatted as preformatted fixed-width text. This is generally done by indenting with four spaces in the case of a block of output. For example:

    mgeddes@sjc01-pd1-lf05:~$ sudo goes status
    GOES status
    ======================
      Mode            - XETH
      PCI             - OK
      Check daemons   - OK
      Check Redis     - OK
      Check vnet      - OK

In cases where a command name is given inside a paragraph of text, it can be wrapped in `backticks`.

### Notes

When making a note to accompany paragraphs of text, use *italics*. Not bold. This is done by wrapping the text in between asterisk characters.

