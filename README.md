[![Build Status](https://gitlab.com/Northern.tech/Mender/mender-setup/badges/master/pipeline.svg)](https://gitlab.com/Northern.tech/Mender/mender-setup/pipelines)
[![Coverage Status](https://coveralls.io/repos/github/mendersoftware/mender-setup/badge.svg?branch=master)](https://coveralls.io/github/mendersoftware/mender-setup?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/mendersoftware/mender-setup)](https://goreportcard.com/report/github.com/mendersoftware/mender-setup)

`mender-setup` tool
===================

Mender is an open source over-the-air (OTA) software updater for embedded Linux
devices. Mender comprises a client running at the embedded device, as well as
a server that manages deployments across many devices.

This repository contains the `mender-setup` tool to configure the Mender client. It writes the
configuration in the Mender client configuration file `/etc/mender/mender.conf`.

It guides the user through an interactive wizard to define the required fields for the configuration
file. All fields can also be specified with command lines options. See `mender-setup --help` for
more details.

![Mender logo](https://mender.io/user/pages/04.resources/logos/logoS.png)

## Getting started

To start using Mender, we recommend that you begin with the Getting started
section in [the Mender documentation](https://docs.mender.io/).

## Contributing

We welcome and ask for your contribution. If you would like to contribute to Mender, please read our
guide on how to best get started [contributing code or
documentation](https://github.com/mendersoftware/mender/blob/master/CONTRIBUTING.md).

## License

Mender is licensed under the Apache License, Version 2.0. See
[LICENSE](https://github.com/mendersoftware/artifacts/blob/master/LICENSE) for the
full license text.

## Security disclosure

We take security very seriously. If you come across any issue regarding
security, please disclose the information by sending an email to
[security@mender.io](security@mender.io). Please do not create a new public
issue. We thank you in advance for your cooperation.

## Connect with us

* Join the [Mender Hub discussion forum](https://hub.mender.io)
* Follow us on [Twitter](https://twitter.com/mender_io). Please
  feel free to tweet us questions.
* Fork us on [Github](https://github.com/mendersoftware)
* Create an issue in the [bugtracker](https://northerntech.atlassian.net/projects/MEN)
* Email us at [contact@mender.io](mailto:contact@mender.io)
* Connect to the [#mender IRC channel on Libera](https://web.libera.chat/?#mender)
