# URI schemes

Sourcegraph uses the following URI schemes. They are required because Sourcegraph exposes APIs and communicates with other systems that assume files and directories can be identified by a string.

## `repo:` and `repo+rev:`

URIs of the form `repo://REPO/PATH` and `repo+rev://REPO/REV/PATH` refer to a file or directory at `PATH` in a repository named `REPO` at (optional) revision `REV`. URIs with `repo:` scheme refer to a repository or file therein at the repository's default branch.

The goals of this URI scheme are to:

- identify resources inside repositories
- be human-readable
- encourage good design decisions in clients

These URIs are intentionally ambiguous and can't be parsed into their components (`REPO`, `REV`, and `PATH`) without additional information. This constraint discourages clients from making poor design decisions (as detailed for the deprecated `git:` URI scheme below).

### Parsing `repo:` URIs

To parse the components of the `repo:` URI given a URI of the form `repo://COMPONENTS`:

- The set of repository names must be known.
- No repository name may be a prefix of another repository's name. (For example, the repositories `a/b` and `a/b/c` can't coexist.)

The parser scans all repository names to find one that is a prefix of `COMPONENTS`. That repository name becomes `REPO`. The rest of the `COMPONENTS` are the `PATH`.

For example, suppose we have a URI `repo://github.com/alice/myrepo/mydir/myfile.txt` and the known set of repository names `{github.com/alice/myrepo, github.com/bob/myrepo}`. The URI is parsed into `REPO=github.com/alice/myrepo` and `PATH=mydir/myfile.txt` because no other repository name matches.

### Parsing `repo+rev:` URIs

To parse the components of the `repo+rev:` URI given a URI of the form `repo+rev://COMPONENTS`:

- All requirements and assumptions for `repo:` URIs (above) apply here, plus:
- The set of Git ref names for the repository must be known.
- Git already enforces the constraint that no ref name may be a prefix of another ref name. (This may or may not hold for other VCS systems, but currently only Git is supported and in scope.)

Git revspecs (such as branch names) may contain slashes, so we use the same kind of assumptions and process as for `REPO` to determine where to split the URI to parse the various components.

For example, suppose we have a URI `repo+rev://github.com/alice/myrepo/wip/mybranch/mydir/myfile.txt` and the known set of Git ref names `{master, wip/mybranch}`. The URI is parsed into `REPO=github.com/alice/myrepo`, `REV=wip/mybranch`, and `PATH=mydir/myfile.txt` because no other Git ref name matches `wip` (so the path component after `wip` must also be part of the Git revspec).

(Note: If you don't believe that these URIs are unambiguous, consider a GitHub URL like https://github.com/facebook/nuclide/blob/master/scripts/create-package.py. How does GitHub know that the branch name is `master`, not `master/scripts`? It can eliminate the latter possibility by knowing that `master` is a branch and therefore `master/scripts` can't be a branch. This also makes sense given the structure of `.git/refs`: you couldn't simultaneously have `.git/refs/heads/master` and `.git/refs/heads/master/scripts` *both* be files.)

## `git:` (deprecated)

> NOTE: This URI scheme is deprecated (see below).

URIs of the form `git://REPO?REV#PATH` refer to a file or directory (or other Git object) at `PATH` in a Git repository named `REPO` at revision `REV`. For example:

- `git://github.com/gorilla/mux?master#route.go` refers to https://github.com/gorilla/mux/blob/master/route.go
- `git://github.com/gorilla/mux?3d80bc801bb034e17cae38591335b3b1110f1c47#route.go` refers to https://github.com/gorilla/mux/blob/3d80bc801bb034e17cae38591335b3b1110f1c47/route.go

This URI scheme is used by `lsp-proxy`, search- and file location-related GraphQL APIs, and Sourcegraph extensions.

### Deprecated

The `git:` URI scheme is **deprecated**.

When communicating with external tools that need a single URI to represent files/directories, use the opaque `repo:` and `repo+rev:` URI schemes defined above because:

- Many tools (such as language servers and editors) do not support it because they ignore the URI query and fragment when asking whether two URIs refer to the same file. Therefore they behave as though all `git:` URIs for the same repository are actually for the same file. This leads to very confusing behavior. It is possible to work around this problem by adding a translation layer, but that adds a lot of complexity.
- Being able to parse the repository name, revision, and file path from the URI (and construct the URIs manually) leads to clients making poor design decisions that frequently lead to bugs.

  Example: Sourcegraph compares repository names case insensitively. If a client constructs a URI manually from an uppercase repository name (or otherwise obtains such a URL) and compares it against an equivalent lowercase URI, the client must know that they can be compared case-insensitively, or else it will incorrectly treat them as distinct. The best solution is for Sourcegraph to canonicalize all URIs it generates, but this breaks if many clients are manually constructing URLs (and not canonicalizing them).

When possible (such as when communicating only among internal Sourcegraph services), pass along each individual field (`REPO`, `REV`, and `PATH`) separately in a Go struct or JSON object to avoid needing to depend on parsing and serializing those values. This also gives you more control over the behavior (such as wanting to preserve the user's input revision in the URL or UI but resolve it to a full SHA for the underlying operations).
