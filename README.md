> A fork of bradfitz's [git-brute][upstream] with some refactoring and
performance improvements for multiple CPUs.

## Original README

gitbrute brute-forces a pair of author+committer timestamps such that
the resulting git commit has your desired prefix.

It will find the most recent time that satisfies your prefix.

Shorter prefixes match more quickly, of course. The author &
committer timestamp are not kept in sync.

Example: https://github.com/bradfitz/deadbeef

Usage:

    gitbrute -prefix=000000

This amends the last commit of the current repository.


## Changes

- `-dryrun` and `-v` verbose flags.
- set a default prefix via `$GITBRUTE_PREFIX`.
- refactored code for profiling and benchmarking.
- modernized for ease of maintenance to use some standard library additions.
- improved mechanism for parallel trial exploration, increases performance
  significantly on multiple CPU cores.<sup>*</sup>

<small>\* On a 2020 MacBook Air (4p/4e core), approximately ~70% performance
improvement in time to find a solution. [[ref]]</small>

:headstone: The original gitbrute ["is kinda a joke and I don't want to maintain
it"][upstream-status], so I am maintaining this as my own fork for personal use
rather than working on pull requests and giving someone else more maintenance
chores.


[upstream]: https://github.com/bradfitz/gitbrute/
[ref]: https://github.com/mroth/gitbrute/pull/1
[upstream-status]: https://github.com/bradfitz/gitbrute/issues/8#issuecomment-168887530