// replication-manager - Replication Manager Monitoring and CLI for MariaDB and MySQL
// Copyright 2017 Signal 18 SARL
// Authors: Guillaume Lefranc <guillaume@signal18.io>
//          Stephane Varoqui  <svaroqui@gmail.com>
// This source code is licensed under the GNU General Public License, version 3.
// Redistribution/Reuse of this code is permitted under the GNU v3 license, as
// an additional term, ALL code must carry the original Author(s) credit in comment form.
// See LICENSE in this directory for the integral text.

package cluster

import (
	"errors"
	"os"
	"sync"
	"time"
)

func (cluster *Cluster) WaitFailoverEndState() {
	for cluster.sme.IsInFailover() {
		time.Sleep(time.Second)
		cluster.LogPrintf(LvlInfo, "Waiting for failover stopped.")
	}
	time.Sleep(recoverTime * time.Second)
}

func (cluster *Cluster) WaitFailoverEnd() error {
	cluster.WaitFailoverEndState()
	return nil

}

func (cluster *Cluster) WaitFailover(wg *sync.WaitGroup) {
	cluster.LogPrintf(LvlInfo, "Waiting failover end")
	defer wg.Done()
	exitloop := 0
	ticker := time.NewTicker(time.Millisecond * 2000)
	for exitloop < 15 {
		select {
		case <-ticker.C:
			cluster.LogPrintf(LvlInfo, "Waiting failover end")
			exitloop++
		case <-cluster.failoverCond.Recv:
			cluster.LogPrintf(LvlInfo, "Failover end receive from channel failoverCond")
			return
		}
	}
	if exitloop == 100 {
		cluster.LogPrintf(LvlInfo, "Failover end")
	} else {
		cluster.LogPrintf(LvlErr, "Failover end timeout")
		return
	}
	return
}

func (cluster *Cluster) WaitSwitchover(wg *sync.WaitGroup) {

	defer wg.Done()
	exitloop := 0
	ticker := time.NewTicker(time.Millisecond * 2000)
	for exitloop < 15 {
		select {
		case <-ticker.C:
			cluster.LogPrintf(LvlInfo, "Waiting switchover end")
			exitloop++
		case <-cluster.switchoverCond.Recv:
			return
		}
	}
	if exitloop == 100 {
		cluster.LogPrintf(LvlInfo, "Switchover end")
	} else {
		cluster.LogPrintf(LvlErr, "Switchover end timeout")
		return
	}
	return
}

func (cluster *Cluster) WaitRejoin(wg *sync.WaitGroup) {

	defer wg.Done()

	exitloop := 0

	ticker := time.NewTicker(time.Millisecond * 2000)
	for exitloop < 15 {

		select {
		case <-ticker.C:
			cluster.LogPrintf(LvlInfo, "Waiting Rejoin")
			exitloop++
		case <-cluster.rejoinCond.Recv:
			return

		}

	}
	if exitloop < 15 {
		cluster.LogPrintf(LvlInfo, "Rejoin Finished")

	} else {
		cluster.LogPrintf(LvlErr, "Rejoin timeout")
		return
	}
	return
}

func (cluster *Cluster) WaitClusterStop() error {
	exitloop := 0
	ticker := time.NewTicker(time.Millisecond * 2000)
	cluster.LogPrintf(LvlInfo, "Waiting for cluster shutdown")
	for exitloop < 10 {
		select {
		case <-ticker.C:
			cluster.LogPrintf(LvlInfo, "Waiting for cluster shutdown")
			exitloop++
			// All cluster down
			if cluster.sme.IsInState("ERR00021") == true {
				exitloop = 100
			}

		}
	}
	if exitloop == 100 {
		cluster.LogPrintf(LvlInfo, "Cluster is shutdown")
	} else {
		cluster.LogPrintf(LvlErr, "Cluster shutdown timeout")
		return errors.New("Failed to stop the cluster")
	}
	return nil
}

func (cluster *Cluster) WaitProxyEqualMaster() error {
	exitloop := 0
	ticker := time.NewTicker(time.Millisecond * 2000)
	cluster.LogPrintf(LvlInfo, "Waiting for proxy to join master")
	for exitloop < 60 {
		select {
		case <-ticker.C:
			cluster.LogPrintf(LvlInfo, "Waiting for proxy to join master")
			exitloop++
			// All cluster down
			if cluster.IsProxyEqualMaster() == true {
				exitloop = 100
			}
		}
	}
	if exitloop == 100 {
		cluster.LogPrintf(LvlInfo, "Proxy can join master")
	} else {
		cluster.LogPrintf(LvlErr, "Proxy to join master timeout")
		return errors.New("Failed to join master via proxy")
	}
	return nil
}

func (cluster *Cluster) WaitMariaDBStop(server *ServerMonitor) error {
	exitloop := 0
	ticker := time.NewTicker(time.Millisecond * 2000)
	for exitloop < 30 {
		select {
		case <-ticker.C:
			cluster.LogPrintf(LvlInfo, "Waiting MariaDB shutdown")
			exitloop++
			_, err := os.FindProcess(server.Process.Pid)
			if err != nil {
				exitloop = 100
			}

		}
	}
	if exitloop == 100 {
		cluster.LogPrintf(LvlInfo, "MariaDB shutdown")
	} else {
		cluster.LogPrintf(LvlInfo, "MariaDB shutdown timeout")
		return errors.New("Failed to Stop MariaDB")
	}
	return nil
}

func (cluster *Cluster) WaitDatabaseStart(server *ServerMonitor) error {
	exitloop := 0
	ticker := time.NewTicker(time.Millisecond * 2000)
	for exitloop < 30 {
		select {
		case <-ticker.C:

			exitloop++

			err := server.Refresh()
			if err == nil {

				exitloop = 100
			} else {
				cluster.LogPrintf(LvlInfo, "Waiting for database start on %s failed with error %s ", server.URL, err)
			}
		}
	}
	if exitloop == 100 {
		cluster.LogPrintf(LvlInfo, "Database started")
	} else {
		cluster.LogPrintf(LvlInfo, "Database start timeout")
		return errors.New("Failed to Start MariaDB")
	}
	return nil
}

func (cluster *Cluster) WaitBootstrapDiscovery() error {
	cluster.LogPrintf(LvlInfo, "Waiting Bootstrap and discovery")
	exitloop := 0
	ticker := time.NewTicker(time.Millisecond * 2000)
	for exitloop < 30 {
		select {
		case <-ticker.C:
			cluster.LogPrintf(LvlInfo, "Waiting Bootstrap and discovery")
			exitloop++
			if cluster.sme.IsDiscovered() {
				exitloop = 100
			}

		}
	}
	if exitloop == 100 {
		cluster.LogPrintf(LvlInfo, "Cluster is Bootstraped and discovery")
	} else {
		cluster.LogPrintf(LvlErr, "Bootstrap timeout")
		return errors.New("Failed Bootstrap timeout")
	}
	return nil
}

func (cluster *Cluster) waitMasterDiscovery() error {
	cluster.LogPrintf(LvlInfo, "Waiting Master Found")
	exitloop := 0
	ticker := time.NewTicker(time.Millisecond * 2000)
	for exitloop < 30 {
		select {
		case <-ticker.C:
			cluster.LogPrintf(LvlInfo, "Waiting Master Found")
			exitloop++
			if cluster.GetMaster() != nil {
				exitloop = 100
			}

		}
	}
	if exitloop == 100 {
		cluster.LogPrintf(LvlInfo, "Master founded")
	} else {
		cluster.LogPrintf(LvlErr, "Master found timeout")
		return errors.New("Failed Master search timeout")
	}
	return nil
}

func (cluster *Cluster) AllDatabaseCanConn() bool {
	for _, s := range cluster.Servers {
		if s.IsDown() {
			return false
		}
	}
	return true
}

func (cluster *Cluster) WaitDatabaseCanConn() error {
	exitloop := 0
	ticker := time.NewTicker(time.Millisecond * 2000)

	cluster.LogPrintf(LvlInfo, "Waiting for cluster to start")
	for exitloop < 30 {
		select {
		case <-ticker.C:
			cluster.LogPrintf(LvlInfo, "Waiting for cluster to start")
			exitloop++
			if cluster.AllDatabaseCanConn() && cluster.HasAllDbUp() {
				exitloop = 100
			}

		}
	}
	if exitloop == 100 {
		cluster.LogPrintf(LvlInfo, "All databases can connect")
	} else {
		cluster.LogPrintf(LvlErr, "Timeout waiting for database to be connected")
		return errors.New("Connections to databases failure")
	}
	return nil
}
