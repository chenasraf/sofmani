# Changelog

## [1.11.1](https://github.com/chenasraf/sofmani/compare/v1.11.0...v1.11.1) (2025-07-04)


### Bug Fixes

* **docker:** fix pull command ([d605e1a](https://github.com/chenasraf/sofmani/commit/d605e1a95dedcd1ed253c317f440f88200f7160e))

## [1.11.0](https://github.com/chenasraf/sofmani/compare/v1.10.1...v1.11.0) (2025-07-03)


### Features

* **brew:** add cask option ([2d4df23](https://github.com/chenasraf/sofmani/commit/2d4df23be6537c799d85395c7b253e2ddd0d0a04))
* **docker:** add flag to skip if docker is not running ([b56e934](https://github.com/chenasraf/sofmani/commit/b56e934ffcbbd59ac46db20ae6c8aac68dfdf68e))


### Bug Fixes

* **docker:** pull during update and not during check ([861e257](https://github.com/chenasraf/sofmani/commit/861e257c04a653349fc0e11e9ec72ee92053194a))

## [1.10.1](https://github.com/chenasraf/sofmani/compare/v1.10.0...v1.10.1) (2025-06-27)


### Bug Fixes

* docker pull to get correct digest ([5f7449c](https://github.com/chenasraf/sofmani/commit/5f7449c8e35a9880de07780444b992340bd0f594))

## [1.10.0](https://github.com/chenasraf/sofmani/compare/v1.9.6...v1.10.0) (2025-06-27)


### Features

* docker installer ([f316fa9](https://github.com/chenasraf/sofmani/commit/f316fa94ecc441113a83c72d9d530a8fd0f9f160))


### Bug Fixes

* show output for shell install/update checks ([fc7e2b9](https://github.com/chenasraf/sofmani/commit/fc7e2b9bf42f1855e9f1c662fadc913e281f83ea))

## [1.9.6](https://github.com/chenasraf/sofmani/compare/v1.9.5...v1.9.6) (2025-06-26)


### Bug Fixes

* brew outdated status code parsing ([f5aa76b](https://github.com/chenasraf/sofmani/commit/f5aa76b94ec2c92888de6e39ac9fa5ec28552bda))

## [1.9.5](https://github.com/chenasraf/sofmani/compare/v1.9.4...v1.9.5) (2025-06-26)


### Bug Fixes

* brew update exit-code/logic ([ad7464c](https://github.com/chenasraf/sofmani/commit/ad7464c7bd7362d3cf21e9c8dbd59ed46194c2e6))

## [1.9.4](https://github.com/chenasraf/sofmani/compare/v1.9.3...v1.9.4) (2025-06-25)


### Bug Fixes

* improve brew update output ([803823b](https://github.com/chenasraf/sofmani/commit/803823bba232785daf8d38b48e7a7a9e2e372e23))

## [1.9.3](https://github.com/chenasraf/sofmani/compare/v1.9.2...v1.9.3) (2025-06-22)


### Bug Fixes

* use tap name on brew update ([e2d37d1](https://github.com/chenasraf/sofmani/commit/e2d37d1045d8df5d76154dd54e5d29067771ec73))

## [1.9.2](https://github.com/chenasraf/sofmani/compare/v1.9.1...v1.9.2) (2025-06-19)


### Bug Fixes

* append installers after validation ([3352ae6](https://github.com/chenasraf/sofmani/commit/3352ae6cfb52fedd61534dba76757409e65d7109))

## [1.9.1](https://github.com/chenasraf/sofmani/compare/v1.9.0...v1.9.1) (2025-06-19)


### Bug Fixes

* panic on missing debug/check_updates ([dc5c3ca](https://github.com/chenasraf/sofmani/commit/dc5c3caa815d5a21e8c9b85af30e02b58cb55f4c))
* wrong config validations ([28f3c1c](https://github.com/chenasraf/sofmani/commit/28f3c1c6dbc6e50cb78674099bd9f3e5e4cf29cb))

## [1.9.0](https://github.com/chenasraf/sofmani/compare/v1.8.0...v1.9.0) (2025-06-19)


### Features

* validations ([70357d1](https://github.com/chenasraf/sofmani/commit/70357d1436e41cf5dde9e5796d09c6d9688cd66a))


### Bug Fixes

* config file/cli overrides ([f92093f](https://github.com/chenasraf/sofmani/commit/f92093f6dc9f43ebd890bec029671fb835022e90))

## [1.8.0](https://github.com/chenasraf/sofmani/compare/v1.7.0...v1.8.0) (2025-01-26)


### Features

* github release installer ([00934e9](https://github.com/chenasraf/sofmani/commit/00934e98f9b675eaea3d0b17ea85e2dc4bd6a756))


### Bug Fixes

* always log everything to file ([fa59a40](https://github.com/chenasraf/sofmani/commit/fa59a4006bcd938cfe8873c4a5b2b9001ae330c2))

## [1.7.0](https://github.com/chenasraf/sofmani/compare/v1.6.0...v1.7.0) (2025-01-20)


### Features

* apk installer ([9d0a6fc](https://github.com/chenasraf/sofmani/commit/9d0a6fc173f74c9d2b302a568d32f834f01cf373))
* pipx installer ([fcebf7c](https://github.com/chenasraf/sofmani/commit/fcebf7c176d1faa6e75ad080f1ab3933f8b8747a))

## [1.6.0](https://github.com/chenasraf/sofmani/compare/v1.5.1...v1.6.0) (2025-01-18)


### Features

* **installer:** add `enabled` option ([e5460d2](https://github.com/chenasraf/sofmani/commit/e5460d255ea61f86de92e76faee3702394306877))

## [1.5.1](https://github.com/chenasraf/sofmani/compare/v1.5.0...v1.5.1) (2025-01-18)


### Bug Fixes

* temp dir location ([d76a3f9](https://github.com/chenasraf/sofmani/commit/d76a3f9f757f8e555b1cf9393fa54c452ddc2709))

## [1.5.0](https://github.com/chenasraf/sofmani/compare/v1.4.3...v1.5.0) (2025-01-16)


### Features

* add `--filter` flag ([3f1cfb6](https://github.com/chenasraf/sofmani/commit/3f1cfb6aed088ed8d9bfa2f6a37199b139657a8f))
* platform-specific env ([3d25d68](https://github.com/chenasraf/sofmani/commit/3d25d68ce2c501cb78afa819160c8e68c7ec7ef7))
* tag filters ([28c9264](https://github.com/chenasraf/sofmani/commit/28c9264bfa4d8ffc994ce6ae9da5cc9d3619df6c))


### Bug Fixes

* filter behavior ([f373216](https://github.com/chenasraf/sofmani/commit/f373216de2993aef2f70da635d016714c8b395ef))
* platform env nil pointer ([94491be](https://github.com/chenasraf/sofmani/commit/94491be4dc574337abafd9cde937a4568c6e101e))
* remove tmpdir from commands ([008165d](https://github.com/chenasraf/sofmani/commit/008165d6765ee4dce70133f64f5ca24e21ab3863))

## [1.4.3](https://github.com/chenasraf/sofmani/compare/v1.4.2...v1.4.3) (2025-01-14)


### Bug Fixes

* remove excess log ([a81bcc6](https://github.com/chenasraf/sofmani/commit/a81bcc678c6f3ea7f689e367ab048d9938a5e2bc))

## [1.4.2](https://github.com/chenasraf/sofmani/compare/v1.4.1...v1.4.2) (2025-01-13)


### Bug Fixes

* show version number ([d62c643](https://github.com/chenasraf/sofmani/commit/d62c643be97158c47b56c3dc67fc03545810d4e9))
* version output trim ([187df75](https://github.com/chenasraf/sofmani/commit/187df7523209c8e58371bbd5bb582d9e71681099))

## [1.4.1](https://github.com/chenasraf/sofmani/compare/v1.4.0...v1.4.1) (2025-01-11)


### Bug Fixes

* bind stdin to process + expand env home dir ([53d3ec6](https://github.com/chenasraf/sofmani/commit/53d3ec645d8994b8a095b0d0db3d556e369055db))
* run install on group updates ([6292fec](https://github.com/chenasraf/sofmani/commit/6292fec0bfdc80d4fc24bb47a8906ce81b8e22ff))

## [1.4.0](https://github.com/chenasraf/sofmani/compare/v1.3.0...v1.4.0) (2025-01-09)


### Features

* add git installer ([b807696](https://github.com/chenasraf/sofmani/commit/b807696014bdfae348779d312a206710566ea7f0))
* add remote manifest installer ([362ee12](https://github.com/chenasraf/sofmani/commit/362ee121682eacaf9cb793d4be848b7f0c8f0793))


### Bug Fixes

* check exit code shell ([4e54373](https://github.com/chenasraf/sofmani/commit/4e54373828bc49545ea67bcea7e18cc71c2b0cdd))
* file argument ([362ee12](https://github.com/chenasraf/sofmani/commit/362ee121682eacaf9cb793d4be848b7f0c8f0793))

## [1.3.0](https://github.com/chenasraf/sofmani/compare/v1.2.0...v1.3.0) (2025-01-06)


### Features

* add apt installer ([5fe683a](https://github.com/chenasraf/sofmani/commit/5fe683a6530043d94ea3feb2bd3a9c722ad43f39))


### Bug Fixes

* env shell platform map ([1367921](https://github.com/chenasraf/sofmani/commit/13679214acf3b2d5b2750efa5d54a322dc987b37))

## [1.2.0](https://github.com/chenasraf/sofmani/compare/v1.1.0...v1.2.0) (2025-01-06)


### Features

* shell update command + env shell ([0d51d26](https://github.com/chenasraf/sofmani/commit/0d51d260f339120a4c47140eaff2d9962bdf1945))

## [1.1.0](https://github.com/chenasraf/sofmani/compare/v1.0.1...v1.1.0) (2024-12-24)


### Features

* add env support ([957968f](https://github.com/chenasraf/sofmani/commit/957968f2d00beab4b78467ae70dfb18da4d18b54))
* improve windows support ([55347b2](https://github.com/chenasraf/sofmani/commit/55347b2ece9993df15db3ee50f4902224de0cc6d))
* npm installer ([1047177](https://github.com/chenasraf/sofmani/commit/104717717acda2937fa813a6025f3bb75fc54edf))


### Bug Fixes

* correctly install brew by name ([be3cd37](https://github.com/chenasraf/sofmani/commit/be3cd37bd6c5549a6cda1e2bd7516406b04ce99b))
* env path resolution ([7b69333](https://github.com/chenasraf/sofmani/commit/7b693334e5b9a98fa26e45ccd33481f6536e2b2c))
* group should always "have update" ([b5af709](https://github.com/chenasraf/sofmani/commit/b5af70985d7f0b7d7f50a303357a5ee49e14070d))
* log directory ([64de503](https://github.com/chenasraf/sofmani/commit/64de5037a2155bd31cb74e15b9e103ef75d16c51))

## [1.0.1](https://github.com/chenasraf/sofmani/compare/v1.0.0...v1.0.1) (2024-12-24)


### Bug Fixes

* find config in .config dir ([18c7933](https://github.com/chenasraf/sofmani/commit/18c7933c0b354a958ab4cae4d407f33674f889ff))

## 1.0.0 (2024-12-24)


### Features

* add bin check to group ([ae2c2df](https://github.com/chenasraf/sofmani/commit/ae2c2dfbe2b101a9ba1d8c328c7238875004b719))
* add custom installed check, add pre/post command ([3b5f720](https://github.com/chenasraf/sofmani/commit/3b5f720441a2f534411c9002c6f627d178dd9e54))
* basic brew installer ([285f278](https://github.com/chenasraf/sofmani/commit/285f278e0952557c39007a2cdc9de79b0c22763e))
* cli args ([6995d46](https://github.com/chenasraf/sofmani/commit/6995d4671b63729f94080727fc4e5f05a7d8b648))
* common post/pre commands, default overrides ([e2f0a35](https://github.com/chenasraf/sofmani/commit/e2f0a352003abb2bac63fbba6d3f5d2252ec8ed3))
* dynamic config file ([ceced91](https://github.com/chenasraf/sofmani/commit/ceced91b5deafb7808e1828b8131b938d6670e2c))
* group installer ([d39c36e](https://github.com/chenasraf/sofmani/commit/d39c36ec55ce41cdc2602ed6a76a3d53e2e38bc8))
* improve logging/flow ([ab85fe7](https://github.com/chenasraf/sofmani/commit/ab85fe77beabfa78048407b0ee300c69ecc308b1))
* initial commit ([ad022ef](https://github.com/chenasraf/sofmani/commit/ad022ef14466cdf06825dd897ef81ed643c35b22))
* logger ([c773a6c](https://github.com/chenasraf/sofmani/commit/c773a6c1400d4150c3dceb3dba78a26de55ff535))
* rsync installer ([d1f3de3](https://github.com/chenasraf/sofmani/commit/d1f3de3d8c74da373ef7123f39761f64b7ecdc66))
* shell installer ([8aecf9a](https://github.com/chenasraf/sofmani/commit/8aecf9af3646261b529455e9f8d6b56662ac48e0))
* use yaml/json instead of pkl ([ad45c24](https://github.com/chenasraf/sofmani/commit/ad45c24e56980e974a8797411b98ecaedf22c6c8))


### Bug Fixes

* cli overrides ([ee29d14](https://github.com/chenasraf/sofmani/commit/ee29d149059e3588729378b911ac0e9469cadae0))
* pkl load ([66b2fb0](https://github.com/chenasraf/sofmani/commit/66b2fb04674e4a73747d31b2e7b614748bac2f32))
* rsync install check ([72002eb](https://github.com/chenasraf/sofmani/commit/72002ebae8e972e263f21bbab51d90d053c68f63))
