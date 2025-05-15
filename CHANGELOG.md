# Changelog

All notable changes to this project will be documented in this file.

## [0.2.6-rc.1](https://github.com/edenreich/n8n-cli/compare/v0.2.5...v0.2.6-rc.1) (2025-05-15)

### 🐛 Bug Fixes

* **dotenv:** Fix dotenv is not loading from the current working directory ([4ddb777](https://github.com/edenreich/n8n-cli/commit/4ddb7774d364f6aebe2d30f38f9e5d2c488f07e5))

### 🔧 Miscellaneous

* **todos:** Update workflow management checklist for accuracy ([34bb51f](https://github.com/edenreich/n8n-cli/commit/34bb51f4569c993733564796f018a6874579cb4d))

## [0.2.5](https://github.com/edenreich/n8n-cli/compare/v0.2.4...v0.2.5) (2025-05-15)

### 🐛 Bug Fixes

* **tests:** Correct version command output format for consistency ([39baa38](https://github.com/edenreich/n8n-cli/commit/39baa38e42393c29fa8365f49a10594b21128b27))

## [0.2.4](https://github.com/edenreich/n8n-cli/compare/v0.2.3...v0.2.4) (2025-05-15)

### 🐛 Bug Fixes

* Update version command output for consistency in naming and messaging ([1a6f824](https://github.com/edenreich/n8n-cli/commit/1a6f8240480094d8c3d09b6f24b879e49bc2b47a))

## [0.2.3](https://github.com/edenreich/n8n-cli/compare/v0.2.2...v0.2.3) (2025-05-15)

### 🐛 Bug Fixes

* Update CLI references from n8n-cli to n8n in scripts and documentation ([#10](https://github.com/edenreich/n8n-cli/issues/10)) ([f63143d](https://github.com/edenreich/n8n-cli/commit/f63143d16d64ca5af5888a4553294e4bc608d4ca))

## [0.2.2](https://github.com/edenreich/n8n-cli/compare/v0.2.1...v0.2.2) (2025-05-15)

### ♻️ Improvements

* Rename CLI from n8n-cli to n8n and update installation script ([#8](https://github.com/edenreich/n8n-cli/issues/8)) ([7b51844](https://github.com/edenreich/n8n-cli/commit/7b51844c76fc737eb6810f7f199bd5764d3d6445))

### 📚 Documentation

* Enhance workflows management in README with new commands for list, refresh, activate, and deactivate ([#9](https://github.com/edenreich/n8n-cli/issues/9)) ([9d1838f](https://github.com/edenreich/n8n-cli/commit/9d1838f1722cd69c2e1aa040e897253c6cd5b0bc))

## [0.2.1](https://github.com/edenreich/n8n-cli/compare/v0.2.0...v0.2.1) (2025-05-15)

### 🐛 Bug Fixes

* Add Version Constants to Config and Fix LDFLAGS metadata ([#7](https://github.com/edenreich/n8n-cli/issues/7)) ([cc7f65d](https://github.com/edenreich/n8n-cli/commit/cc7f65ddf02eb05c96ce06410897fca772dd2b37))

## [0.2.0](https://github.com/edenreich/n8n-cli/compare/v0.1.4...v0.2.0) (2025-05-15)

### ✨ Features

* Add workflows management commands for listing, activating, and deactivating workflows ([#6](https://github.com/edenreich/n8n-cli/issues/6)) ([55b6014](https://github.com/edenreich/n8n-cli/commit/55b6014974348a671ad8ea4778aae76b178a20c4))

## [0.1.4](https://github.com/edenreich/n8n-cli/compare/v0.1.3...v0.1.4) (2025-05-12)

### ♻️ Improvements

* Refactor configuration make it shared between commands ([d529c9f](https://github.com/edenreich/n8n-cli/commit/d529c9f01af0cfc247f5a05a915a30dd7a790c72))
* Remove redundant comment in LoadConfig function ([ff30fd0](https://github.com/edenreich/n8n-cli/commit/ff30fd042d19c3461030889de5396a20be704cd2))

### 🐛 Bug Fixes

* Improve error handling in command help execution ([83f5efd](https://github.com/edenreich/n8n-cli/commit/83f5efd7a90eb11f7dff04fb2e303d749b5a9677))

### 🔧 Miscellaneous

* Add initial documentation for project instructions, structure, tools, and development workflow for more efficient claude development ([22666e5](https://github.com/edenreich/n8n-cli/commit/22666e50424db104cff0732b57a92752c2120ef7))
* Add version command to display n8n-cli version information ([a3b6b5a](https://github.com/edenreich/n8n-cli/commit/a3b6b5af6147297e9536dff63ddd48d4386a6729))

### ✅ Miscellaneous

* Add tests ([#4](https://github.com/edenreich/n8n-cli/issues/4)) ([e22b523](https://github.com/edenreich/n8n-cli/commit/e22b5230f6b79c5da0b147a812a119bb9d6dc652))

## [0.1.3](https://github.com/edenreich/n8n-cli/compare/v0.1.2...v0.1.3) (2025-05-11)

### 🐛 Bug Fixes

* Enhance sync command to manage workflow IDs and add server workflow fetching ([09a7f3b](https://github.com/edenreich/n8n-cli/commit/09a7f3bc95c15fd19924450637994d9aa38570f4))

## [0.1.2](https://github.com/edenreich/n8n-cli/compare/v0.1.1...v0.1.2) (2025-05-11)

### 📚 Documentation

* Add completion instructions for bash, zsh, and fish to README ([9003997](https://github.com/edenreich/n8n-cli/commit/90039970f6ed4f2cb26fea3fdffb707f2699239c))
* Remove redundant installer details from README ([2b7238d](https://github.com/edenreich/n8n-cli/commit/2b7238d724949847da29f3885f296cb3b357ffb7))
* Update README Table of Contents for better navigation ([b929d4e](https://github.com/edenreich/n8n-cli/commit/b929d4e6a9b6185c5c988733d8fdfbcf575c1a96))
* Update README to enhance visual appeal with badges and improved header ([40c6175](https://github.com/edenreich/n8n-cli/commit/40c617501c60aed4b4d6a4df219a82385e2ce69e))

### 🔧 Miscellaneous

* Mark install.sh as vendored for linguist ([9752331](https://github.com/edenreich/n8n-cli/commit/9752331b0644db01ed1e94f9bf39b19c8d9aafea))

### 📦 Miscellaneous

* Optimize Go build flags for smaller binary size and ship it statically ([dae0c30](https://github.com/edenreich/n8n-cli/commit/dae0c30bb9dde6034f4573e24bfd726f89cc918c))

## [0.1.1](https://github.com/edenreich/n8n-cli/compare/v0.1.0...v0.1.1) (2025-05-11)

### 🐛 Bug Fixes

* Update root command description for clarity and relevance ([9480f27](https://github.com/edenreich/n8n-cli/commit/9480f27bea322858b5507060ccfeff3003f12c3a))

### 👷 CI

* Enhance release workflow to support version input and improve artifact upload logic ([ecdb8e3](https://github.com/edenreich/n8n-cli/commit/ecdb8e387bc9408790847aa395606835f53e40b5))
* Refactor artifact upload process in release workflow ([f12eebb](https://github.com/edenreich/n8n-cli/commit/f12eebb76a2b5ac55050eddc6ac26321f5883e67))
* Refactor build matrix for cross-platform support in release workflow ([50d960f](https://github.com/edenreich/n8n-cli/commit/50d960fbac0be3049450c0879c5b1c89417353b1))

### 📚 Documentation

* Enhance README with installation instructions and add install script ([6de1b43](https://github.com/edenreich/n8n-cli/commit/6de1b43a9d40af84092fd12ec529d74337ab636d))

### 🔧 Miscellaneous

* Update semantic-release and plugins to specific versions in Dockerfile and release workflow ([4b2e25b](https://github.com/edenreich/n8n-cli/commit/4b2e25b925f57bc4eb0c14b414a15e3ffdcac89b))
* Update semantic-release and plugins to specific versions in release workflow ([d8ea527](https://github.com/edenreich/n8n-cli/commit/d8ea527999bccdfe07d50214cf0fddf072770ea5))
