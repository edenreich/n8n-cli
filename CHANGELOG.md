# Changelog

All notable changes to this project will be documented in this file.

## [0.6.0](https://github.com/edenreich/n8n-cli/compare/v0.5.2...v0.6.0) (2025-09-21)

### ✨ Features

* **executions:** Implement workflow execution history retrieval ([#13](https://github.com/edenreich/n8n-cli/issues/13)) ([dfb1ad8](https://github.com/edenreich/n8n-cli/commit/dfb1ad8e43459d8187e60d7e11991f3b186cec34))

### ♻️ Improvements

* Update generated types and OpenAPI spec for execution status and retry functionality ([#29](https://github.com/edenreich/n8n-cli/issues/29)) ([c78a1f1](https://github.com/edenreich/n8n-cli/commit/c78a1f111be2e6ae9724b98ae52862f4a6aecb9a))
* **workflow:** Rename AI Summary Agent to Summary Agent and update connections ([034395d](https://github.com/edenreich/n8n-cli/commit/034395dbaeae39e174663cc85f658e1b30386f5e))

### 👷 CI

* Add claude GitHub actions ([#27](https://github.com/edenreich/n8n-cli/issues/27)) ([90692cc](https://github.com/edenreich/n8n-cli/commit/90692cc0a84b49c86684cebd61dd6514ee2bac0e))

### 📚 Documentation

* **examples:** Add GitHub Actions workflows for syncing and restoring n8n workflows ([dfe2b08](https://github.com/edenreich/n8n-cli/commit/dfe2b087745ba6b1f3eb43f91c7efe547f61623d))
* Revise CLAUDE.md for improved structure and clarity ([a7aee87](https://github.com/edenreich/n8n-cli/commit/a7aee87de67e73bde539e53f1dbcdc97280fb2d6))

### 🔧 Miscellaneous

* Add issue templates for bug reports, feature requests, and refactor requests ([60b29eb](https://github.com/edenreich/n8n-cli/commit/60b29ebb85b890dfa04b317d15973e0587e147e3))
* **copilot:** Enable instruction files for GitHub Copilot chat code generation ([7dec111](https://github.com/edenreich/n8n-cli/commit/7dec11146f13ce32286190e74cef961f1f5f1159))
* **copilot:** Update Copilot instructions for commit message and PR description formats ([0fd45c4](https://github.com/edenreich/n8n-cli/commit/0fd45c4bd066d0087f3bccc70eea3ab5d556775e))
* Rename it to workflows-creator ([f2364f8](https://github.com/edenreich/n8n-cli/commit/f2364f81a2e9a9de09a57ad845977f2873fd7986))
* **todos:** Update TODOs - remove some auto generated TODOs that doesn't make sense ([d7dc125](https://github.com/edenreich/n8n-cli/commit/d7dc12556d8c17255a61ddd6aaf76d8ee77accb0))
* Update README.md ([861b2d4](https://github.com/edenreich/n8n-cli/commit/861b2d4f899698b760fa7e81c8c161fc66de389b))
* Update tags management section in TODOs to reflect completed tasks ([a411678](https://github.com/edenreich/n8n-cli/commit/a411678176276daec3687930444c1e4736a72485))

## [0.5.2](https://github.com/edenreich/n8n-cli/compare/v0.5.1...v0.5.2) (2025-05-20)

### ♻️ Improvements

* **sync:** Use early return instead of nested if checks ([#23](https://github.com/edenreich/n8n-cli/issues/23)) ([a0ebc43](https://github.com/edenreich/n8n-cli/commit/a0ebc436a0871f79ae3954e2a92c0123d501dc7f))

### 🐛 Bug Fixes

* **docs:** Correct typo in response suggestions in README ([85d13ab](https://github.com/edenreich/n8n-cli/commit/85d13ab7b6f87ca47070d6803317a0dc38c99bfc))
* **docs:** Update README to reflect Groq API usage instead of OpenAI API ([766a0af](https://github.com/edenreich/n8n-cli/commit/766a0af6786f09d20f1046b01427df40c472f39b))
* Enhance CleanWorkflow to preserve and clean tags ([#24](https://github.com/edenreich/n8n-cli/issues/24)) ([e125820](https://github.com/edenreich/n8n-cli/commit/e125820b7fd0b6d1a70514591d58c856ff02ef9a))
* **tags:** Implement tag creation and retrieval in n8n client ([#25](https://github.com/edenreich/n8n-cli/issues/25)) ([6cc9058](https://github.com/edenreich/n8n-cli/commit/6cc90583b02ecda8062347462d3a70650a751918))

### 📚 Documentation

* Update Copilot instructions for clarity and structure ([495d134](https://github.com/edenreich/n8n-cli/commit/495d134251149ed02fefdba3e2ac6c6b1f34184b))
* Update README to include examples for contact form and AI-enhanced contact form ([f6ca96f](https://github.com/edenreich/n8n-cli/commit/f6ca96f505ceda1fee87d9d6e7b003b9a5e732fa))

### 🔧 Miscellaneous

* Add instruction for running tests before submitting code in Copilot guidelines ([f930b53](https://github.com/edenreich/n8n-cli/commit/f930b53d2346e7b50ebc7f59a7673b01ba589125))
* Add instruction for verifying linter status in Copilot guidelines ([dd3b0cb](https://github.com/edenreich/n8n-cli/commit/dd3b0cb41c52719d3690976ad9bac9a98ca98ae1))
* Cleanup - some leftover comments from the LLM ([24a1416](https://github.com/edenreich/n8n-cli/commit/24a141661002548fc3222cd92f86f07ed72e630c))
* Correct Claude's instructions for early returns and test writing ([f66ce4e](https://github.com/edenreich/n8n-cli/commit/f66ce4e86dde8dca69752de806f8ecbf4574e524))
* Update Copilot instructions for commit message format and correct grammar ([c61f080](https://github.com/edenreich/n8n-cli/commit/c61f0807411cb339e425b5a6538bf092059e0532))

## [0.5.1](https://github.com/edenreich/n8n-cli/compare/v0.5.0...v0.5.1) (2025-05-19)

### ♻️ Improvements

* **sync:** Enhance sync command with output format and all workflows refresh options ([#22](https://github.com/edenreich/n8n-cli/issues/22)) ([5ba967c](https://github.com/edenreich/n8n-cli/commit/5ba967c0ccd7d5ad659a64c59b873dad0d75dd2c))

## [0.5.0](https://github.com/edenreich/n8n-cli/compare/v0.4.0...v0.5.0) (2025-05-19)

### ✨ Features

* **workflows-sync:** Add --refresh flag to sync command and improve refresh functionality ([#21](https://github.com/edenreich/n8n-cli/issues/21)) ([e36a284](https://github.com/edenreich/n8n-cli/commit/e36a284b512f39b664341cd65db9802187b1b9aa))

### 📚 Documentation

* **examples:** Add AI-Enhanced Contact Form Example with n8n Integration ([#16](https://github.com/edenreich/n8n-cli/issues/16)) ([24f18d6](https://github.com/edenreich/n8n-cli/commit/24f18d61095f0e85d21be180cdffa5c7916414c9))

### 🔧 Miscellaneous

* **docs:** Correct filename for AI-enhanced n8n workflow definition in README ([166c951](https://github.com/edenreich/n8n-cli/commit/166c9517d935e9cc642e564c2353b6f71467c5b1))

## [0.4.0](https://github.com/edenreich/n8n-cli/compare/v0.3.1...v0.4.0) (2025-05-19)

### ✨ Features

* **workflows:** add GetWorkflow method and corresponding tests on the client ([#17](https://github.com/edenreich/n8n-cli/issues/17)) ([7577f0b](https://github.com/edenreich/n8n-cli/commit/7577f0babbe8f35c76a5865a0aff1d662727ea07))

### ♻️ Improvements

* **workflows-sync:** Enhance sync command documentation and improve test setup for workflows ([#18](https://github.com/edenreich/n8n-cli/issues/18)) ([210b6b8](https://github.com/edenreich/n8n-cli/commit/210b6b869a2ff78b48c86d5187cf46f58baf9b4a)), closes [#19](https://github.com/edenreich/n8n-cli/issues/19)

### 🔧 Miscellaneous

* RC testing refactor improve the maintainability of sync command ([#20](https://github.com/edenreich/n8n-cli/issues/20)) ([3cbe4d2](https://github.com/edenreich/n8n-cli/commit/3cbe4d24140012e5756d0fc9c58ccaaf38e3f9d8)), closes [#19](https://github.com/edenreich/n8n-cli/issues/19)

## [0.4.0-rc.2](https://github.com/edenreich/n8n-cli/compare/v0.4.0-rc.1...v0.4.0-rc.2) (2025-05-19)

### 🐛 Bug Fixes

* **encoder:** Change YAML indentation back to 2 spaces ([8115d5b](https://github.com/edenreich/n8n-cli/commit/8115d5bf236ec488cf027c1526781dc67cb77b3c))

## [0.4.0-rc.1](https://github.com/edenreich/n8n-cli/compare/v0.3.1...v0.4.0-rc.1) (2025-05-19)

### ✨ Features

* **client:** Add GetWorkflowTags and UpdateWorkflowTags methods to manage workflow tags ([4ecf4de](https://github.com/edenreich/n8n-cli/commit/4ecf4de793aa23db1d4da003145893f77c3d676e))
* **workflows:** add GetWorkflow method and corresponding tests on the client ([#17](https://github.com/edenreich/n8n-cli/issues/17)) ([7577f0b](https://github.com/edenreich/n8n-cli/commit/7577f0babbe8f35c76a5865a0aff1d662727ea07))
* **workflows:** Implement tag management in workflow synchronization and add logging capabilities ([a62e2d6](https://github.com/edenreich/n8n-cli/commit/a62e2d68e7cf3bfea0d10818b727a29efbd30aa6)), closes [#19](https://github.com/edenreich/n8n-cli/issues/19)

### ♻️ Improvements

* **workflows-sync:** Enhance sync command documentation and improve test setup for workflows ([021072b](https://github.com/edenreich/n8n-cli/commit/021072bd49f6fea469581202c4bec2973248a611))
* **workflows:** Implement workflow encoding and decoding with clean functionality and a drift detection using reflect deep equal for comparing two objects (actual with desired state) ([604ab8f](https://github.com/edenreich/n8n-cli/commit/604ab8fd766b89e1a699fa9388942c73b7511ac0))

### 🐛 Bug Fixes

* **sync:** Update command examples to use 'n8n' instead of 'n8n-cli' ([9fec9c5](https://github.com/edenreich/n8n-cli/commit/9fec9c581872f9931b9960a006e8cff6fb853489))

## [0.3.1](https://github.com/edenreich/n8n-cli/compare/v0.3.0...v0.3.1) (2025-05-17)

### 🐛 Bug Fixes

* **workflows-sync:** exclude ID and active fields when creating workflows ([#15](https://github.com/edenreich/n8n-cli/issues/15)) ([6dea048](https://github.com/edenreich/n8n-cli/commit/6dea04824fbad4e83587593c29b9ea1e2ca088ef))

## [0.3.0](https://github.com/edenreich/n8n-cli/compare/v0.2.6...v0.3.0) (2025-05-15)

### ✨ Features

* Improve Refresh functionality - implement --minimal flag ([#12](https://github.com/edenreich/n8n-cli/issues/12)) ([3f74baf](https://github.com/edenreich/n8n-cli/commit/3f74baf77e254ec29bbba008d917868c9dc2cb43))

### 📚 Documentation

* **README:** Add contributing section with link to CONTRIBUTING.md ([09bdf08](https://github.com/edenreich/n8n-cli/commit/09bdf08976ad018a7d8e8cdc12413a12081e6244))
* **README:** Clarify instructions for creating and using the .env file ([12e7c1c](https://github.com/edenreich/n8n-cli/commit/12e7c1cbea7edceb26752483d9ba29426ab9b366))

### 🔧 Miscellaneous

* **todos:** Add validate command to apply static analysis on workflow files ([1a6a70f](https://github.com/edenreich/n8n-cli/commit/1a6a70ff82c3782a7472df5ddee4812c21a33d4e))
* **todos:** Change priorities - will implement soon credentials syncer ([02ebe3d](https://github.com/edenreich/n8n-cli/commit/02ebe3d9beaacccb4e11a8501f3a9649afdbf3bf))
* **todos:** Deprioritize list workflows by name tags and active status - it's a nice-to-have ([a95b102](https://github.com/edenreich/n8n-cli/commit/a95b1028094b333dcc4708311a683710e5b709c3))
* **todos:** Fix formatting for workflow listing filter capabilities ([f3bac06](https://github.com/edenreich/n8n-cli/commit/f3bac06619f9a279e38b49e15b9f929aebe9870c))
* **todos:** Simplify description of the task ([b9e0227](https://github.com/edenreich/n8n-cli/commit/b9e02277245a26c305aa0d9234f0576e139aec7a))

## [0.2.6](https://github.com/edenreich/n8n-cli/compare/v0.2.5...v0.2.6) (2025-05-15)

### 🐛 Bug Fixes

* **dotenv:** Fix dotenv is not loading from the current working directory ([#11](https://github.com/edenreich/n8n-cli/issues/11)) ([ad40abd](https://github.com/edenreich/n8n-cli/commit/ad40abd467059fe97174296e4f72d34e6505ba65))

### 📚 Documentation

* **README:** Fix punctuation in installation instructions for clarity ([a2268f2](https://github.com/edenreich/n8n-cli/commit/a2268f2d093b6b997d1f4461dc182c6440084285))
* **README:** Improve formatting for installation instructions ([b6b9594](https://github.com/edenreich/n8n-cli/commit/b6b95944f5eb6ca9bbb05f24c894db7b1c6b507f))

### 🔧 Miscellaneous

* **install:** Add version specification option to installation script ([8fc4150](https://github.com/edenreich/n8n-cli/commit/8fc4150a4c59facb3da27502ee9bd9405ab44b7e))
* **todos:** Update workflow management checklist for accuracy ([34bb51f](https://github.com/edenreich/n8n-cli/commit/34bb51f4569c993733564796f018a6874579cb4d))

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
