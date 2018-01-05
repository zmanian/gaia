# Changelog

## 0.5.0 (December 29, 2017)

BREAKING CHANGES:

* Update to Tendermint v0.15.0

## 0.4.0 (December 6, 2017)

Added delegation to existing validators

BREAKING CHANGES:

* Consolidation of binaries:
  * `gaia` -> `gaia node` 
  * `gaiacli` -> `gaia client` 
* `bond` replaced with `declare-candidacy`
* `unbond` used for both delegation and candidates self-bonding
* Unbonding must use the `--pubkey` flag

IMPROVEMENTS: 

* Delegation to existing validators with `gaia client tx delegate`
* Added REST server support

## 0.3.0 (October 28, 2017)

BREAKING CHANGES:

* don't change AppHash every block

## 0.2.0 (October 12, 2017)

BREAKING CHANGES:

* persist global params in store
* change default bonding token to "fermion"


## 0.1.0 (October 10, 2017)

Working bond and unbond through latest sdk develop
