---
## 1.0.3 - 2026-01-08


### Bug fixes


- Allow certain values to start with `-`
([MEN-9203](https://northerntech.atlassian.net/browse/MEN-9203)) ([b1cb5d6](https://github.com/mendersoftware/mender-setup/commit/b1cb5d6bd12b1662358a15497ead59ac3ef7b7d7))  by @danielskinstad





  Fixed an issue where a tenant token or a password starting with
  `-` would fail with a `--<flag> requires a non-empty value` error.






## 1.0.2 - 2025-12-19


### Bug fixes


- Allow empty string for `--server-cert`
([MEN-9178](https://northerntech.atlassian.net/browse/MEN-9178)) ([d98e10d](https://github.com/mendersoftware/mender-setup/commit/d98e10d9b1f37172854265c2f647ced52458c6bf))  by @lluiscampos






  Amends 24a5f0deee8c1f6e2794b6c1436e3a71aab8e74e
  
  This flag is optional. An empty value indicates no custom certificate,
  which is the typical use case for servers that use well known CA for
  certificate signing (like hosted Mender ;)






## 1.0.1 - 2025-12-18


### Bug fixes


- Do not crash when configuration flags miss a value
 ([24a5f0d](https://github.com/mendersoftware/mender-setup/commit/24a5f0deee8c1f6e2794b6c1436e3a71aab8e74e))  by @pasinskim




  Validate configuration flags and return user friendly error message that
  the configuration parameter is missing. Do not print the whole trace but
  simple message instead.





## mender-setup 1.0.0

_Released 01.15.2024_

### Changelogs

#### mender-setup (1.0.0)

* First release of mender-setup
