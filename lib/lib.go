package lib

import (
	"fmt"

	chkr "github.com/tweithoener/checker"
)

func init() {
	if err := chkr.AddCheckMaker(
		pingMkr, httpMkr, proxyMkr, dnsMkr, failMkr, cmdMkr, sshMkr,
		memMkr, cpuMkr, diskMkr, uptimeMkr, loadMkr, swapMkr, procExistsMkr, sysProcsMkr,
	); err != nil {
		panic(fmt.Sprintf("configuration error: %v", err))
	}
	if err := chkr.AddNotifierMaker(
		pushoverMkr, loggingMkr, lessMkr, debugMkr, emailMkr,
	); err != nil {
		panic(fmt.Sprintf("configuration error: %v", err))
	}

}
