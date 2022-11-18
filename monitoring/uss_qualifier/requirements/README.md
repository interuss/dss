# Requirements

## Overview

The primary purpose of uss_qualifier is to identify the requirements with which a test participant complies and does not comply.  The existence of these requirements should be documented within this folder.

## Documentation format

A requirement's identifier is `<PACKAGE>.<NAME>` where `<PACKAGE>` is a Python-style package reference to a .md file (without extension) relative to this folder (`uss_qualifier/requirements`).  For instance, the `<PACKAGE>` for the file located at `./astm/f3548/v21.md` would be `astm.f3548.v21`.  `<NAME>` is an identifier defined in the file described by `<PACKAGE>` by enclosing it in a `<tt>` tag; for instance: `<tt>USS0105</tt>`.

## Usage

Requirements tested by a scenario should be defined in that [scenario's documentation](../scenarios/README.md).
