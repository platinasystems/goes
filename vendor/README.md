This directory contains [subtrees] of the external packages imported by `goes`.

These were added by either...

```console
$ git subtree add -P vendor/PREFIX --squash REMOTE REF
```

or,

```console
$ git remote add REMOTE REPOS
$ git fetch REMOTE
$ git subtree add -P vendor/PREFIX --squash REMOTE/master
```

These pull from upstream with...

```console
$ git subtree pull -P vendor/PREFIX --squash REPOS REF
```


or,

```console
$ git fetch REMOTE
$ git subtree merge -P vendor/PREFIX REMOTE/master
```

Any commits of subtree changes are segregated from changes to other areas and
have message headlines that reference the subtree.

Subtree commits are pushed upstream with...

```console
$ git remote add REMOTE REPOS
$ git fetch REMOTE
$ git branch my-REMOTE REMOTE/master
$ git subtree push -P vendor/PREFIX . my-REMOTE
$ git push -n REMOTE my-REMOTE:master
# verify commits!
$ git push REMOTE my-REMOTE:master
```

---

*&copy; 2015-2016 Platina Systems, Inc. All rights reserved.
Use of this source code is governed by this BSD-style [LICENSE].*

[LICENSE]: ../LICENSE
[subtrees]: https://github.com/git/git/blob/master/contrib/subtree/git-subtree.txt
