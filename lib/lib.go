package lib

import (
	"fmt"

	chkr "github.com/tweithoener/checker"
)

func init() {
	if err := chkr.AddCheckMaker(
		pingMkr, httpMkr, proxyMkr, dnsMkr, failMkr, peerMkr, cmdMkr, sshMkr,
	); err != nil {
		panic(fmt.Sprintf("configuration error: %v", err))
	}
	if err := chkr.AddNotifierMaker(
		pushoverMkr, logginMkr, lessMkr,
	); err != nil {
		panic(fmt.Sprintf("configuration error: %v", err))
	}

}
