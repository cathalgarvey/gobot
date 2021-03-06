# Contributing to Gobot

This document is based on the [io.js contribution guidelines](https://github.com/nodejs/io.js/blob/master/CONTRIBUTING.md)

## Issue Contributions

When opening new issues or commenting on existing issues on this repository
please make sure discussions are related to concrete technical issues with the
Gobot software.

## Code Contributions

The Gobot project welcomes new contributors.

This document will guide you through the contribution process.

What do you want to contribute?

- I want to otherwise correct or improve the docs or examples
- I want to report a bug
- I want to add some feature or functionality to an existing hardware platform
- I want to add support for a new hardware platform

Descriptions for each of these will eventually be provided below.

## General Guidelines

* All active development is in the `dev` branch. New or updated features must be added to the `dev` branch. Hotfixes will be considered on the `master` branch in situations where it does not alter behaviour or features, only fixes a bug.
* All patches must be provided under the Apache 2.0 License
* Please use the -s option in git to "sign off" that the commit is your work and you are providing it under the Apache 2.0 License
* Submit a Github Pull Request to the appropriate branch and ideally discuss the changes with us in IRC.
* We will look at the patch, test it out, and give you feedback.
* Avoid doing minor whitespace changes, renamings, etc. along with merged content. These will be done by the maintainers from time to time but they can complicate merges and should be done seperately.
* Take care to maintain the existing coding style.
* `golint` and `go fmt` your code.
* Add unit tests for any new or changed functionality.
* All pull requests should be "fast forward"
  * If there are commits after yours use “git rebase -i <new_head_branch>”
  * If you have local changes you may need to use “git stash”
  * For git help see [progit](http://git-scm.com/book) which is an awesome (and free) book on git

## Creating Pull Requests
Because Gobot makes use of self-referencing import paths, you will want
to implement the local copy of your fork as a remote on your copy of the
original Gobot repo. Katrina Owen has [an excellent post on this workflow](https://splice.com/blog/contributing-open-source-git-repositories-go/).

The basics are as follows:

1. Fork the project via the GitHub UI

2. `go get` the upstream repo and set it up as the `upstream` remote and your own repo as the `origin` remote:

`go get github.com/hybridgroup/gobot`
`cd $GOPATH/src/github.com/hybridgroup/gobot`
`git remote rename origin upstream`
`git remote add origin git@github.com/YOUR_GITHUB_NAME/gobot`

All import paths should now work fine assuming that you've got the
proper branch checked out.


## Landing Pull Requests
(This is for committers only. If you are unsure whether you are a committer, you are not.)

1. Set the contributor's fork as an upstream on your checkout

`git remote add contrib1 https://github.com/contrib1/gobot`

2. Fetch the contributor's repo

`git fetch contrib1`

3. Checkout a copy of the PR branch

`git checkout pr-1234 --track contrib1/branch-for-pr-1234`

4. Review the PR as normal

5. Land when you're ready via the GitHub UI

## Developer's Certificate of Origin 1.0

By making a contribution to this project, I certify that:

* (a) The contribution was created in whole or in part by me and I
  have the right to submit it under the open source license indicated
  in the file; or
* (b) The contribution is based upon previous work that, to the best
  of my knowledge, is covered under an appropriate open source license
  and I have the right under that license to submit that work with
  modifications, whether created in whole or in part by me, under the
  same open source license (unless I am permitted to submit under a
  different license), as indicated in the file; or
* (c) The contribution was provided directly to me by some other
  person who certified (a), (b) or (c) and I have not modified it.


## Code of Conduct

This Code of Conduct is adapted from [Rust's wonderful
CoC](http://www.rust-lang.org/conduct.html).

* We are committed to providing a friendly, safe and welcoming
  environment for all, regardless of gender, sexual orientation,
  disability, ethnicity, religion, or similar personal characteristic.
* Please avoid using overtly sexual nicknames or other nicknames that
  might detract from a friendly, safe and welcoming environment for
  all.
* Please be kind and courteous. There's no need to be mean or rude.
* Respect that people have differences of opinion and that every
  design or implementation choice carries a trade-off and numerous
  costs. There is seldom a right answer.
* Please keep unstructured critique to a minimum. If you have solid
  ideas you want to experiment with, make a fork and see how it works.
* We will exclude you from interaction if you insult, demean or harass
  anyone.  That is not welcome behaviour. We interpret the term
  "harassment" as including the definition in the [Citizen Code of
  Conduct](http://citizencodeofconduct.org/); if you have any lack of
  clarity about what might be included in that concept, please read
  their definition. In particular, we don't tolerate behavior that
  excludes people in socially marginalized groups.
* Private harassment is also unacceptable. No matter who you are, if
  you feel you have been or are being harassed or made uncomfortable
  by a community member, please contact one of the channel ops or any
  of the TC members immediately with a capture (log, photo, email) of
  the harassment if possible.  Whether you're a regular contributor or
  a newcomer, we care about making this community a safe place for you
  and we've got your back.
* Likewise any spamming, trolling, flaming, baiting or other
  attention-stealing behaviour is not welcome.
* Avoid the use of personal pronouns in code comments or
  documentation. There is no need to address persons when explaining
  code (e.g. "When the developer")
