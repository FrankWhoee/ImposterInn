package main

import "github.com/gofrs/uuid/v5"


// Registered: Client has a pid <-> exists in pidSet
// InLobby: Client has a pid AND a wid
type IdBroker struct {
	pidSet map[string]bool
	pidToWid map[string]int
	widToPid []string
}

func (idb *IdBroker) issuePid() string {
	pidbytes, _ := uuid.Must(uuid.NewV4()).MarshalText()
	pid := string(pidbytes)

	idBroker.pidSet[pid] = true

	return pid
}

func (idb *IdBroker) assignWid(pid string) int {
	wid := len(idb.widToPid)
	idb.widToPid = append(idb.widToPid, pid)
	idb.pidToWid[pid] = wid

	return wid
}

func (idb *IdBroker) isRegistered(pid string) bool {
	return idBroker.pidSet[pid]
}

func (idb *IdBroker) getWid(pid string) (int,bool) {
	wid,isInLobby := idb.pidToWid[pid]
	return wid, isInLobby
}

func (idb *IdBroker) getPid(wid int) string {
	return idb.widToPid[wid]
}