# Contributing to Perun Node

Want to contribute? That's great! Any contribution is welcome, e.g.
documentation, bug reports, feature request, issues, blog posts,
tutorials, feature implementations, etc. You can contribute code or
documentation through the standard GitHub pull request model. For large
contributions we do encourage you to file a ticket in the GitHub issue
tracking system prior to any development to coordinate with the DST
development team early in the process. Coordinating up front helps to
avoid frustration later on.

*Note that this repository only contains the source code of perun-node. For
issues / pull requests related only to documentation please refer to the
[perun-doc](https://github.com/direct-state-transfer/perun-doc) repository.*

## Contribution Guideline

* We use the GitHub [issue
  tracker](https://github.com/direct-state-transfer/perun-node/issues) and
  this [branching
  model](https://nvie.com/posts/a-successful-git-branching-model/).

* When you want to submit a patch to any DST project, you must [sign a
  CLA (once) and sign your commits](#sign-the-cla-and-sign-your-work).

* Your contribution must be licensed under the Apache-2.0 license, the
  license used by this project.

* When submitting a pull-request to the project, please note / ensure
  the following:
    1. If possible, keep the changes small and simple which makes review
       process easier.
    2. [Add / retain copyright notices](#add-retain-copyright-notices).
    3. Document the code by following [Go commentary
       guidelines](https://golang.org/doc/effective_go.html#commentary).
    4. Make sure to run linter with the project's [linter
       config](build/linterConfig.json) and fix any errors or warnings.
    5. Verify that the code builds cleanly and passes all unit tests.
    6. Ensure good test coverage by including unit tests when adding
       significant amount of functionalities.
    7. Please follow these
       [guidelines](https://chris.beams.io/posts/git-commit/) on how to
       write a good **commit** message. If available, please include the
       ticket number. And don't forget the
       [Signed-Off-By](#sign-the-cla-and-sign-your-work) line.
    8. Create a pull request against the develop branch.

## Legal stuff

### Add / retain copyright notices

Include a copyright notice and license in each new file to be
contributed, consistent with the style used by this project. If your
contribution contains code under the copyright of a third party,
document its origin, license, and copyright holders.

### Sign the CLA and sign your work

We believe that if successful the DST project should move to a neutral
home in the long run (e.g. to some not-for-profit foundation). To allow
us to retain the necessary open source licensing flexibility please
ensure you have signed the [Contributor License Agreement
(CLA)](https://cla-assistant.io/direct-state-transfer/perun-node) before
creating any pull request. All contributors will retain ownership in
their contributions while granting us the necessary legal rights. The
CLA only needs to be signed once and it covers all [Direct State
Transfer](https://github.com/direct-state-transfer) repositories.

This project also tracks patch provenance and licensing using
Signed-off-by tags. With the sign-off in a commit message you certify
that you authored the patch or otherwise have the right to submit it
under an open source license and you acknowledge that the CLA signed for
any DST project also applies to this contribution. The procedure is
simple: just append a line

    Signed-off-by: Random J Developer <random@developer.example.org>

to every commit message using your real name or your pseudonym and a valid
email address.

If you have set your `user.name` and `user.email` git configs you can
automatically sign the commit by running the git-commit command with the
`-s` option.  There may be multiple sign-offs if more than one developer
was involved in authoring the contribution.
