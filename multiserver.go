package servermanager

import "errors"

type MultiServerManager struct {
	store Store
	carManager *CarManager
	raceControlHub *RaceControlHub

	servers []*Server
}

type Server struct {
	RaceControl *RaceControl
	ServerProcess ServerProcess
	ServerConfig GlobalServerConfig
}

func (msm *MultiServerManager) GetServer(i int) (*Server, error) {
	if i < 0 || i > len(msm.servers) {
		return nil, errors.New("servermanager: server out out bounds")
	}

	return msm.servers[i], nil
}

func (msm *MultiServerManager) ListServers() []*Server {
	return msm.servers
}

var ErrNoFreeServerFound = errors.New("servermanager: no free server found")

func (msm *MultiServerManager) FreeServer() (*Server, error) {
	for _, server := range msm.servers {
		if !server.ServerProcess.IsRunning() {
			return server, nil
		}
	}

	return nil, ErrNoFreeServerFound
}

func (msm *MultiServerManager) AddServer(serverConfig GlobalServerConfig) error {
	process := NewAssettoServerProcess(nil, NewContentManagerWrapper(msm.store, msm.carManager))

	msm.servers = append(msm.servers, &Server{
		RaceControl:   NewRaceControl(msm.raceControlHub, filesystemTrackData{}, process),
		ServerProcess: process,
		ServerConfig:  serverConfig,
	})

	return nil
}