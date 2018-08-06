# Change Log

This project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased]

## [0.2.0]
### Added
- Add plugin registry; in-process and separate processes
- Add plugin unmarshaller; add new cobra commands dynamically
- Crane.yml file for easy orchestration and volume sync on MacOSX
- Multi-stage dockerfile for cross-compilation
- Multi-stage dockerfile for alpine/golang developpment
- Makefile targets for docker based builds
- Makefile targets for local based builds
- Makefile helpers (info, environement)

### Updated
- project source tree refactoring; create cmd, pkg and plugin dir
- docker-compose.yml format version from 2.2 to 3.6

### Fixed
- Fix: ./pkg/wrapper/docker/client.go
- Fix: switch docker reference declaration from 'github.com/docker/docker/reference' to 'github.com/docker/distribution/reference'
- Fix: bleve package imports and function/methods calls for latest package release changes

## [0.1.0] - 2016-06-27
### Initial Release

Please see the [releases](https://github.com/sniperkit/dim/releases).
[Unreleased]: https://github.com/sniperkit/dim/compare/v0.5.0...HEAD
[0.2.0]: https://github.com/sniperkit/dim/compare/v0.1.0...v0.2.0