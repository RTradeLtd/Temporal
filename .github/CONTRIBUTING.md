# Contributing Guidelines

The Contribution guidelines will continually evolve as we gain a better understanding of our development userbase, and development workflow that needs to be reworked.
In the meantime, all code contributions need to be submitted through a pull request. Please include a short, detailed description of what your pull request changes, any breaking features,and if introducing a new feature, why it's of use.
All code submitted must include tests that handle proper requests, and improper requests. Code must also be adequately commented.

## BASH Script Guidelines

All new scripts must pass validation by [shellcheck](https://www.shellcheck.net/)

## Golang Code

For functions where no return values are kept:

```Golang
if _, err = um.ChangeEthereumAddress( ... ); if err != nil {
    ...
}
```
