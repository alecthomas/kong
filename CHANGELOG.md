# Changelog

All notable changes to this project will be documented in this file. See [conventional commits](https://www.conventionalcommits.org/) for commit guidelines.

## [unreleased]

- **feat:** Placeholder string interpolation. (#510)&nbsp;&nbsp;·&nbsp;&nbsp;[44be791](https://github.com/squareup/sq/commit/44be791798047748290119ce70a4e390a986324a)
- **feat:** Allow configuring global hooks via Kong's functional options (#511)&nbsp;&nbsp;·&nbsp;&nbsp;[78d4066](https://github.com/squareup/sq/commit/78d4066dabe5432d604eb28ab5376a3ad1708e29)
- **chore(deps):** update all non-major dependencies (#506)&nbsp;&nbsp;·&nbsp;&nbsp;[1edf069](https://github.com/squareup/sq/commit/1edf069f4a1ad7499f6d4cc34b1e7af0d06d89b1)

## [1.9.0](https://github.com/squareup/sq/compare/v1.8.1..v1.9.0)&nbsp;&nbsp;·&nbsp;&nbsp;2025-03-11

- **feat:** Support singleton providers (#501)&nbsp;&nbsp;·&nbsp;&nbsp;[3b9af5b](https://github.com/squareup/sq/commit/3b9af5bdced7d12010db4133bf3d0ef0d1f6c3ee)
- **feat:** Allow ignoring fields from embedded structs (#499)&nbsp;&nbsp;·&nbsp;&nbsp;[300cba8](https://github.com/squareup/sq/commit/300cba8c27f2a60a3a65ca7537530bc7ec86a8dd)
- **feat:** add IgnoreDefault optional interface&nbsp;&nbsp;·&nbsp;&nbsp;[0c495e4](https://github.com/squareup/sq/commit/0c495e4936b62e3599434bb9938cb5eb4e20ab53)
- **feat:** Allow kong.Path to describe remaining unparsed args (#472)&nbsp;&nbsp;·&nbsp;&nbsp;[5b36573](https://github.com/squareup/sq/commit/5b36573738d800fbd511e9ed3c5711db6c7b6661)
- **feat:** Allow Kong to exit with semantic exit codes (#507)&nbsp;&nbsp;·&nbsp;&nbsp;[a86adbb](https://github.com/squareup/sq/commit/a86adbbb25c02f414c222d42bf4cec9bcb824f96)
- **fix:** ignore --help flag for determining optional flag usage&nbsp;&nbsp;·&nbsp;&nbsp;[73db2e8](https://github.com/squareup/sq/commit/73db2e86a5dee444a2089c62beff8703c82467b4)
- **refactor:** ParseError can carry an exit code&nbsp;&nbsp;·&nbsp;&nbsp;[3d03233](https://github.com/squareup/sq/commit/3d03233b16d42450d1deb4b9f2aa18a837cc0f6c)
- **chore(deps):** update module github.com/alecthomas/kong to v1.8.1 (#503)&nbsp;&nbsp;·&nbsp;&nbsp;[7f94c90](https://github.com/squareup/sq/commit/7f94c902b9e2a7e5b84194d1a4c61ce62ec0361d)
- **ci:** Test with Go 1.23 and 1.24 (#502)&nbsp;&nbsp;·&nbsp;&nbsp;[9f71a49](https://github.com/squareup/sq/commit/9f71a497678f07c938979a3a7971ad4ba1b06cd6)

## [1.8.1](https://github.com/squareup/sq/compare/v1.8.0..v1.8.1)&nbsp;&nbsp;·&nbsp;&nbsp;2025-02-13

- **chore:** add test for decoding&nbsp;&nbsp;·&nbsp;&nbsp;[cab639a](https://github.com/squareup/sq/commit/cab639ab83b14162986cc0a34700047274514d06)

## [1.8.0](https://github.com/squareup/sq/compare/v1.7.0..v1.8.0)&nbsp;&nbsp;·&nbsp;&nbsp;2025-02-09

- **refactor:** minor simplification of getMethods&nbsp;&nbsp;·&nbsp;&nbsp;[5765c11](https://github.com/squareup/sq/commit/5765c1152209cc4f20c3afcffb18ca0d4367e5f2)
- **chore(deps):** update all non-major dependencies (#486)&nbsp;&nbsp;·&nbsp;&nbsp;[6590294](https://github.com/squareup/sq/commit/6590294c3d15753cb8ded03ed7bbc24cba5a41f3)
- **chore:** add lefthook&nbsp;&nbsp;·&nbsp;&nbsp;[a6efd40](https://github.com/squareup/sq/commit/a6efd403cc72d824c9c948d7c7e29d2c56ed481a)
- **chore:** reduce lefthook output&nbsp;&nbsp;·&nbsp;&nbsp;[705e259](https://github.com/squareup/sq/commit/705e259cd6cb87a0195eceac6f31d9bf7cf01e58)

## [1.7.0](https://github.com/squareup/sq/compare/v1.6.1..v1.7.0)&nbsp;&nbsp;·&nbsp;&nbsp;2025-01-30

- **feat:** allow use of providers that don't return errors&nbsp;&nbsp;·&nbsp;&nbsp;[4e1757c](https://github.com/squareup/sq/commit/4e1757c0e858abc58762bb54014c9bae5d1bbc42)
- **chore(deps):** update all non-major dependencies (#485)&nbsp;&nbsp;·&nbsp;&nbsp;[8d238c8](https://github.com/squareup/sq/commit/8d238c88faef4e3a6ef331352d81f40397ee46fb)

## [1.6.1](https://github.com/squareup/sq/compare/v1.6.0..v1.6.1)&nbsp;&nbsp;·&nbsp;&nbsp;2025-01-04

- **feat:** allow hooks to be declared on embedded fields&nbsp;&nbsp;·&nbsp;&nbsp;[840220c](https://github.com/squareup/sq/commit/840220c2ed477d11bf362a32428bc26da7a3051a)
- **fix(Context.Run):** Don't panic on unselected root node (#484)&nbsp;&nbsp;·&nbsp;&nbsp;[7ca8467](https://github.com/squareup/sq/commit/7ca846736ce870e28be6cbfbfa4e27e2b0609770)
- **fix:** don't call Apply() twice&nbsp;&nbsp;·&nbsp;&nbsp;[a14bb20](https://github.com/squareup/sq/commit/a14bb2072c00a4aa1fccf1dcad2a16ad230099e8)
- **fix:** don't append ... for fields with an explicit type&nbsp;&nbsp;·&nbsp;&nbsp;[cacaace](https://github.com/squareup/sq/commit/cacaace969ab71cfa8815a06590c3e0b03501ee6)
- **fix:** add an `xorprefix:"..."` option for prefixing xor/and groups&nbsp;&nbsp;·&nbsp;&nbsp;[47b090f](https://github.com/squareup/sq/commit/47b090f2f44b4509d00edd0dcf20c0f52f1ff28e)
- **fix:** whoops, forgot to prefix "and"&nbsp;&nbsp;·&nbsp;&nbsp;[b811e32](https://github.com/squareup/sq/commit/b811e322436d3dfb266ea76b2e2a987bbf204a2f)
- **refactor(test):** replace os.Setenv() with t.Setenv()&nbsp;&nbsp;·&nbsp;&nbsp;[3625768](https://github.com/squareup/sq/commit/36257680f1d0b7bf1e63c751b2b5e7e8aaa78a2c)
- **chore(deps):** update all non-major dependencies (#476)&nbsp;&nbsp;·&nbsp;&nbsp;[565ae9b](https://github.com/squareup/sq/commit/565ae9b740578e038691242cb239aaa1206e698c)
- **chore:** `interface{}` -> `any`&nbsp;&nbsp;·&nbsp;&nbsp;[a32b94b](https://github.com/squareup/sq/commit/a32b94b70530e64716c954ff2be62d14f26cc7f6)

## [1.6.0](https://github.com/squareup/sq/compare/v1.5.1..v1.6.0)&nbsp;&nbsp;·&nbsp;&nbsp;2024-12-07

- **feat:** add support for `Provide*() (<type>, error)` methods on commands&nbsp;&nbsp;·&nbsp;&nbsp;[388ba35](https://github.com/squareup/sq/commit/388ba35f1a32c1cf97108a7777b8a36a84dbc1a4)
- **chore(deps):** update all non-major dependencies (#474)&nbsp;&nbsp;·&nbsp;&nbsp;[87ee7dc](https://github.com/squareup/sq/commit/87ee7dc3dbfa9f97fa17f5f32e37343dc6f53608)

## [1.5.1](https://github.com/squareup/sq/compare/v1.5.0..v1.5.1)&nbsp;&nbsp;·&nbsp;&nbsp;2024-12-02

- **fix:** regression where all args became passthrough&nbsp;&nbsp;·&nbsp;&nbsp;[81d0c29](https://github.com/squareup/sq/commit/81d0c29303442ba216395c690b07526ede3ca34a)

## [1.5.0](https://github.com/squareup/sq/compare/v1.4.0..v1.5.0)&nbsp;&nbsp;·&nbsp;&nbsp;2024-12-01

- **feat:** add old "passthrough" behaviour back in as an option&nbsp;&nbsp;·&nbsp;&nbsp;[96647c3](https://github.com/squareup/sq/commit/96647c30af56cdf1a273f3f95de0b73c6d7a0b38)
- **chore(deps):** update all non-major dependencies (#469)&nbsp;&nbsp;·&nbsp;&nbsp;[5221ebb](https://github.com/squareup/sq/commit/5221ebb8fb83249597a43fa977170e1803890fc8)
- **chore:** fix all golangci-lint warnings&nbsp;&nbsp;·&nbsp;&nbsp;[88e13d7](https://github.com/squareup/sq/commit/88e13d750a44855f5517ab596191697e01dc94e2)

## [1.4.0](https://github.com/squareup/sq/compare/v1.3.0..v1.4.0)&nbsp;&nbsp;·&nbsp;&nbsp;2024-11-05

- **feat:** support optionally passing kong.Context to Validate()&nbsp;&nbsp;·&nbsp;&nbsp;[1b9d57e](https://github.com/squareup/sq/commit/1b9d57eec1a2667072589e00d6ad4b63beba9c46)
- **feat:** add AfterRun() hook&nbsp;&nbsp;·&nbsp;&nbsp;[2544d3f](https://github.com/squareup/sq/commit/2544d3f00855f6857cf121ce04608296b5858e4a)
- **fix:** format enum value&nbsp;&nbsp;·&nbsp;&nbsp;[64229c9](https://github.com/squareup/sq/commit/64229c9fe78f0cfabce5c3713508ca9bb23b1a09)
- **fix:** update enum+default vars after interpolation&nbsp;&nbsp;·&nbsp;&nbsp;[d0beaf7](https://github.com/squareup/sq/commit/d0beaf7df321c4630d1e6b0460882038c498e0ca)
- **fix:** NPE due to checking if error is nil when it can be a value&nbsp;&nbsp;·&nbsp;&nbsp;[f388f6c](https://github.com/squareup/sq/commit/f388f6cd39ae7fdbc4ffc75205d16a798c5a56bf)
- **chore(deps):** update module github.com/alecthomas/kong to v1.3.0 (#467)&nbsp;&nbsp;·&nbsp;&nbsp;[c90c673](https://github.com/squareup/sq/commit/c90c6732cc4d66d4ee45212f37ca2b19d306941d)

## [1.3.0](https://github.com/squareup/sq/compare/v1.2.1..v1.3.0)&nbsp;&nbsp;·&nbsp;&nbsp;2024-11-01

- **feat:** support recursive injection of provider parameters&nbsp;&nbsp;·&nbsp;&nbsp;[7bbb0b7](https://github.com/squareup/sq/commit/7bbb0b76ada1610f18cf71c54cca74209da88bd8)
- **refactor:** reuse callAnyFunction for callFunction&nbsp;&nbsp;·&nbsp;&nbsp;[373692a](https://github.com/squareup/sq/commit/373692af87b177d48898c89ad53b6054f5b339bf)
- **chore(deps):** update module github.com/alecthomas/kong to v1 (#457)&nbsp;&nbsp;·&nbsp;&nbsp;[3b28c2c](https://github.com/squareup/sq/commit/3b28c2c83a3ee74b1f7d8087f50f88766755544f)
- **chore(deps):** update module github.com/alecthomas/assert/v2 to v2.11.0 (#459)&nbsp;&nbsp;·&nbsp;&nbsp;[c9fbc30](https://github.com/squareup/sq/commit/c9fbc305f4d0629395f729e83f633bf70e77a8fe)
- **chore(deps):** update all non-major dependencies (#462)&nbsp;&nbsp;·&nbsp;&nbsp;[56e990f](https://github.com/squareup/sq/commit/56e990fc15ee8b1426a07e14d22013a38d272c67)

## [1.2.1](https://github.com/squareup/sq/compare/v1.2.0..v1.2.1)&nbsp;&nbsp;·&nbsp;&nbsp;2024-09-12

- **fix:** allow duplicate custom negated flags&nbsp;&nbsp;·&nbsp;&nbsp;[26c1c9a](https://github.com/squareup/sq/commit/26c1c9ad052ae46eb77fda47fea6b2b7cfb6a605)

## [1.2.0](https://github.com/squareup/sq/compare/v1.1.0..v1.2.0)&nbsp;&nbsp;·&nbsp;&nbsp;2024-09-11

- **fix** [**breaking**]**:** Include `--` in passthrough args (#436)&nbsp;&nbsp;·&nbsp;&nbsp;[9924ec4](https://github.com/squareup/sq/commit/9924ec4461bc4fec32a85c16970c9a2a691d41c9)

## [1.1.0](https://github.com/squareup/sq/compare/v1.0.0..v1.1.0)&nbsp;&nbsp;·&nbsp;&nbsp;2024-09-10

- **chore(deps):** update all non-major dependencies (#455)&nbsp;&nbsp;·&nbsp;&nbsp;[b297ae9](https://github.com/squareup/sq/commit/b297ae97f2a54d75c9a68f3015e615e0be9fa310)

## [1.0.0](https://github.com/squareup/sq/compare/v0.9.0..v1.0.0)&nbsp;&nbsp;·&nbsp;&nbsp;2024-08-21

- **feat:** allow non-structs to be used as commands (#428)&nbsp;&nbsp;·&nbsp;&nbsp;[e864bb0](https://github.com/squareup/sq/commit/e864bb02209592cb6e8259b74d325edf4236f47c)
- **fix:** When a Grammar combines flags with passthrough args, see if an unrecognized flag may be treated as a positional argument (#435)&nbsp;&nbsp;·&nbsp;&nbsp;[fcb5e05](https://github.com/squareup/sq/commit/fcb5e05c0706be6be5027c32691c29d61dc4e6d4)
- **fix:** typo in README (#451)&nbsp;&nbsp;·&nbsp;&nbsp;[6c216a3](https://github.com/squareup/sq/commit/6c216a37cee562d79039309b7f7ba489e9fe2f94)
- **chore(deps):** update all non-major dependencies (#418)&nbsp;&nbsp;·&nbsp;&nbsp;[3f756c1](https://github.com/squareup/sq/commit/3f756c1dbd40606b03a131f516e20328e44601e8)
- **chore(deps):** update module github.com/gliderlabs/ssh to v0.3.7 (#421)&nbsp;&nbsp;·&nbsp;&nbsp;[0376ce4](https://github.com/squareup/sq/commit/0376ce4e069a1b7d2c3637dc29b8e38f7bdf3bd1)
- **chore(deps):** update all non-major dependencies (#424)&nbsp;&nbsp;·&nbsp;&nbsp;[2ab5733](https://github.com/squareup/sq/commit/2ab5733f117949ba9d1e24c1123d9c977d1808e6)
- **chore(deps):** update all non-major dependencies (#425)&nbsp;&nbsp;·&nbsp;&nbsp;[605cdd6](https://github.com/squareup/sq/commit/605cdd64a921c795cab5abc722861d3fd5e2645d)
- **chore(deps):** update all non-major dependencies (#445)&nbsp;&nbsp;·&nbsp;&nbsp;[d113f61](https://github.com/squareup/sq/commit/d113f61e2acf60e555e5227d3305156af9ed7fe9)
- **chore:** bump go + golangci-lint&nbsp;&nbsp;·&nbsp;&nbsp;[2ad9498](https://github.com/squareup/sq/commit/2ad9498bdfbb7dcae73ae57508692826353a918d)

## [0.9.0](https://github.com/squareup/sq/compare/v0.8.1..v0.9.0)&nbsp;&nbsp;·&nbsp;&nbsp;2024-03-08

- **feat:** Add info about optional flags to usage (#416)&nbsp;&nbsp;·&nbsp;&nbsp;[dc13080](https://github.com/squareup/sq/commit/dc13080e39e9e4843493f3bf1ae76fe2b57f4ac1)
- **fix:** use LookupEnv to check if envar is actually set&nbsp;&nbsp;·&nbsp;&nbsp;[0ecd272](https://github.com/squareup/sq/commit/0ecd272e23814bddc938ba6776c47e5f33df723e)
- **fix:** set trailing positional arguments to active&nbsp;&nbsp;·&nbsp;&nbsp;[575d5b1](https://github.com/squareup/sq/commit/575d5b1d349d24c96a5dbfe4569dbc8a2cf1d0d7)
- **fix:** broken test&nbsp;&nbsp;·&nbsp;&nbsp;[8e675d6](https://github.com/squareup/sq/commit/8e675d6130741b200afeb065c90e229e235ea18e)
- **fix:** slices/maps of existingfile would fail to work&nbsp;&nbsp;·&nbsp;&nbsp;[a41b2e8](https://github.com/squareup/sq/commit/a41b2e8f4ed34d80305fff68cf819f819c06ae7e)
- **chore(deps):** update all non-major dependencies (#398)&nbsp;&nbsp;·&nbsp;&nbsp;[79809cb](https://github.com/squareup/sq/commit/79809cbb473aee5e404b955c8ba10cca62e0f667)
- **chore(deps):** update module github.com/chzyer/readline to v1 (#401)&nbsp;&nbsp;·&nbsp;&nbsp;[d63c4d7](https://github.com/squareup/sq/commit/d63c4d71e9ef386ff723e269c2b25b4fb234dba8)
- **chore(deps):** update actions/checkout action to v4 (#399)&nbsp;&nbsp;·&nbsp;&nbsp;[27806be](https://github.com/squareup/sq/commit/27806be88ff59de3357f292a0281ab132074701b)
- **chore(deps):** update actions/setup-go action to v5 (#400)&nbsp;&nbsp;·&nbsp;&nbsp;[6997afd](https://github.com/squareup/sq/commit/6997afd4f945f313494f6e8915f03e03d56ce365)
- **chore(deps):** update all non-major dependencies (#405)&nbsp;&nbsp;·&nbsp;&nbsp;[4907efc](https://github.com/squareup/sq/commit/4907efc9c8534c570080311781870d8a3de76155)
- **chore(deps):** update all non-major dependencies (#406)&nbsp;&nbsp;·&nbsp;&nbsp;[5e5a0cf](https://github.com/squareup/sq/commit/5e5a0cf53d4d70881a92b856936c1a83a001c4a0)
- **chore(deps):** update module github.com/alecthomas/assert/v2 to v2.5.0 (#407)&nbsp;&nbsp;·&nbsp;&nbsp;[30e8461](https://github.com/squareup/sq/commit/30e84613fe34bbc57e9421303b14af459d845f74)
- **chore(deps):** update all non-major dependencies (#410)&nbsp;&nbsp;·&nbsp;&nbsp;[088cd78](https://github.com/squareup/sq/commit/088cd7874e49bfbe3e390851bd89a31db93f2205)
- **chore(renovate):** disable golangci-lint&nbsp;&nbsp;·&nbsp;&nbsp;[51e3de8](https://github.com/squareup/sq/commit/51e3de86cfe4a6adcab9de266205e711e88f1c58)
- **ci:** Test against Go 1.20 and 1.21 (#404)&nbsp;&nbsp;·&nbsp;&nbsp;[de00f3b](https://github.com/squareup/sq/commit/de00f3b71ce15c105b967e52f9858ebaab90f930)

## [0.8.1](https://github.com/squareup/sq/compare/v0.8.0..v0.8.1)&nbsp;&nbsp;·&nbsp;&nbsp;2023-10-07

- **fix:** remove showing app name twice in errors (#368)&nbsp;&nbsp;·&nbsp;&nbsp;[074ccd0](https://github.com/squareup/sq/commit/074ccd090604a69363b9e6f56b0205bafb79884d)
- **fix:** support IsBool() on MapperValue implementations.&nbsp;&nbsp;·&nbsp;&nbsp;[46c0384](https://github.com/squareup/sq/commit/46c03841edae90a3daa987b58623589d47ddf1ba)

## [0.8.0](https://github.com/squareup/sq/compare/v0.7.1..v0.8.0)&nbsp;&nbsp;·&nbsp;&nbsp;2023-02-25

- **feat:** Embed() option and Context.Call()&nbsp;&nbsp;·&nbsp;&nbsp;[bf0cbf5](https://github.com/squareup/sq/commit/bf0cbf5d7ccdf573efa31446a09f91fa939b17eb)
- **feat:** add filecontent named mapper&nbsp;&nbsp;·&nbsp;&nbsp;[919e701](https://github.com/squareup/sq/commit/919e70103f7a7068ebccb991e24836f2933529fd)
- **feat:** add FlagNamer option&nbsp;&nbsp;·&nbsp;&nbsp;[95a465b](https://github.com/squareup/sq/commit/95a465b4b5324acaa5d9cf6733fd019dde718de0)
- **feat:** support multiple env variables (#349)&nbsp;&nbsp;·&nbsp;&nbsp;[9610ed6](https://github.com/squareup/sq/commit/9610ed62d93992d4f13e210ada9fd0917530f079)
- **fix:** ensure pointers can be detected as bools&nbsp;&nbsp;·&nbsp;&nbsp;[9b08b89](https://github.com/squareup/sq/commit/9b08b8939602e1a55972e5543d64a017feb13daf)
- **fix:** handle contents of tags properly by unquoting them when necessary&nbsp;&nbsp;·&nbsp;&nbsp;[37e8014](https://github.com/squareup/sq/commit/37e801405f5ee2b18697a7b57130afee254fe0b9)

## [0.7.1](https://github.com/squareup/sq/compare/v0.7.0..v0.7.1)&nbsp;&nbsp;·&nbsp;&nbsp;2022-11-12

- **feat:** add support for camelCase JSON keys&nbsp;&nbsp;·&nbsp;&nbsp;[d974d72](https://github.com/squareup/sq/commit/d974d7270ae0502ee5c9336ecbe4c39dcd4661e5)

## [0.7.0](https://github.com/squareup/sq/compare/v0.6.1..v0.7.0)&nbsp;&nbsp;·&nbsp;&nbsp;2022-10-17

- **fix:** os-specific builds&nbsp;&nbsp;·&nbsp;&nbsp;[81070a0](https://github.com/squareup/sq/commit/81070a043b434467613a0d472cb8bf099f23854c)
- **fix:** go1.19 specific tests for text wrapping&nbsp;&nbsp;·&nbsp;&nbsp;[3371bd6](https://github.com/squareup/sq/commit/3371bd616773115a609ccfc4b84d6cfde66e5b64)
- **fix:** tests&nbsp;&nbsp;·&nbsp;&nbsp;[8b82618](https://github.com/squareup/sq/commit/8b826182d82d66ee22cfdce0342026df3223e96e)
- **refactor:** switch to alecthomas/assert&nbsp;&nbsp;·&nbsp;&nbsp;[e75e1ca](https://github.com/squareup/sq/commit/e75e1ca88a47dfec13d46902aff7e93c2bf745d3)

## [0.6.1](https://github.com/squareup/sq/compare/v0.6.0..v0.6.1)&nbsp;&nbsp;·&nbsp;&nbsp;2022-06-15

- **fix:** Reset() should always be called before anything else&nbsp;&nbsp;·&nbsp;&nbsp;[0066abb](https://github.com/squareup/sq/commit/0066abb97338a1de4b69d15ceeb54029970d6305)

## [0.6.0](https://github.com/squareup/sq/compare/v0.5.0..v0.6.0)&nbsp;&nbsp;·&nbsp;&nbsp;2022-06-11

- **fix:** values that look like flags would not be parsed correctly&nbsp;&nbsp;·&nbsp;&nbsp;[7c6ff10](https://github.com/squareup/sq/commit/7c6ff10d3388dbc9ec5977785f12e5404b339f09)
- **fix:** existing file/dir always checks default&nbsp;&nbsp;·&nbsp;&nbsp;[0aaa4c1](https://github.com/squareup/sq/commit/0aaa4c11997b409c9258b597b86f6388cbe6cc49)
- **chore:** pin Go-1.17.8&nbsp;&nbsp;·&nbsp;&nbsp;[556f8b7](https://github.com/squareup/sq/commit/556f8b773b24e22c5dcb05a1a740567c19d1ef99)

## [0.5.0](https://github.com/squareup/sq/compare/v0.4.1..v0.5.0)&nbsp;&nbsp;·&nbsp;&nbsp;2022-03-11

- **feat:** AutoGroup option&nbsp;&nbsp;·&nbsp;&nbsp;[7a63525](https://github.com/squareup/sq/commit/7a635254200602a6de251035c6de8a9259abd8f1)

## [0.4.1](https://github.com/squareup/sq/compare/v0.4.0..v0.4.1)&nbsp;&nbsp;·&nbsp;&nbsp;2022-02-16

- **fix:** hydrate pointer fields&nbsp;&nbsp;·&nbsp;&nbsp;[72024dd](https://github.com/squareup/sq/commit/72024dd44aafe32cde2c2c06e3e1eb4e7a047d6f)

## [0.4.0](https://github.com/squareup/sq/compare/v0.3.0..v0.4.0)&nbsp;&nbsp;·&nbsp;&nbsp;2022-01-27

- **fix:** correct line number for json example (#265)&nbsp;&nbsp;·&nbsp;&nbsp;[38c5622](https://github.com/squareup/sq/commit/38c56226ba8a6963c7202f05584a7b662dbaf53a)

## [0.3.0](https://github.com/squareup/sq/compare/v0.2.22..v0.3.0)&nbsp;&nbsp;·&nbsp;&nbsp;2022-01-05

- **fix:** issue with symlinked temp dirs&nbsp;&nbsp;·&nbsp;&nbsp;[3987efb](https://github.com/squareup/sq/commit/3987efba8647711f11290d443e07106f22f7393f)

## [0.2.21](https://github.com/squareup/sq/compare/v0.2.20..v0.2.21)&nbsp;&nbsp;·&nbsp;&nbsp;2021-12-13

- **fix:** overly enthusiastic validation of tags&nbsp;&nbsp;·&nbsp;&nbsp;[fa08e70](https://github.com/squareup/sq/commit/fa08e7027c35633d457b29984486873558f46bdd)

## [0.2.19](https://github.com/squareup/sq/compare/v0.2.18..v0.2.19)&nbsp;&nbsp;·&nbsp;&nbsp;2021-12-02

- **feat:** add support for variable interpolation into env tag key (#234)&nbsp;&nbsp;·&nbsp;&nbsp;[79f896f](https://github.com/squareup/sq/commit/79f896f0c0cbaef813d02006acd321d64e9cf55d)

## [0.2.17](https://github.com/squareup/sq/compare/v0.2.16..v0.2.17)&nbsp;&nbsp;·&nbsp;&nbsp;2021-06-06


## [0.2.16](https://github.com/squareup/sq/compare/v0.2.15..v0.2.16)&nbsp;&nbsp;·&nbsp;&nbsp;2021-03-02


## [0.1.16](https://github.com/squareup/sq/compare/v0.1.15..v0.1.16)&nbsp;&nbsp;·&nbsp;&nbsp;2019-03-29

- **feat:** add a command tree help view (#32)&nbsp;&nbsp;·&nbsp;&nbsp;[2a90ca2](https://github.com/squareup/sq/commit/2a90ca2ad300aa9820c9cf299ae21fb0385f964c)
- **fix:** `hidden` tag is not passed to the flag object&nbsp;&nbsp;·&nbsp;&nbsp;[fcf5f9b](https://github.com/squareup/sq/commit/fcf5f9bc1a4fc3388f490d77c7477a5681b1f97a)

<!-- generated by git-cliff -->
