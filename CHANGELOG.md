# Changelog

## 0.4.0 (November 6, 2017)

Added delegation to existing validators

BREAKING CHANGES:

* Bonding replaced with DeclareCandidacy
* Unbonding used for both delegation and candidates self-bonding
* Unbonding must use the `--pubkey` flag

## 0.3.0 (October 28, 2017)

BREAKING CHANGES:

* don't change AppHash every block

## 0.2.0 (October 12, 2017)

BREAKING CHANGES:

* persist global params in store
* change default bonding token to "fermion"


## 0.1.0 (October 10, 2017)

Working bond and unbond through latest sdk develop
