_dl_cmd kill "terminate a runaway container"
_dl_kill () {
    CID=$1
    [ "$CID" ] || _dl_error "must specify container to kill"
    CID=$(_dl_resolve containers $CID)
    lxc-stop -n $CID
}