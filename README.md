# Github-backup application 

This application clone your github repository with all commits, branch, tags etc. to your local disk

## Dependencies

This App use 'git' and 'gh' (github-cli) applications which shoud be preinstalled on the host. The 'git' should be configured to has access to your repositories by ssh. The 'gh' should be logged in to your github account before call this app.

Application parameters:

    -users  <[user-or-organisation-comma-separated-list]>
    -limit  [user-repo-comma-separated-list]
    -output [local-folder-name], default: ./repos

Usage examples:

    go run . -users=kirill-scherba -limit=kirill-scherba/teonet-go -output=./tmp

